package backup

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/tianrking/ClawRemove/internal/model"
)

// Snapshot represents a backup taken before execution.
type Snapshot struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Product   string    `json:"product"`
	Items     []Item    `json:"items"`
}

// Item represents a single backed-up file or directory.
type Item struct {
	OriginalPath string `json:"originalPath"`
	BackupPath   string `json:"backupPath"`
	IsDir        bool   `json:"isDir"`
}

// Manager handles creating snapshots and rolling them back.
type Manager struct {
	backupDir string
}

// NewManager creates a new backup Manager storing snapshots in the given base directory.
func NewManager(baseDir string) *Manager {
	return &Manager{
		backupDir: filepath.Join(baseDir, "snapshots"),
	}
}

// CreateSnapshot creates a backup of all files queued for removal in the given plan.
func (m *Manager) CreateSnapshot(product string, plan model.Plan) (string, error) {
	if len(plan.Actions) == 0 {
		return "", nil
	}

	id := time.Now().UTC().Format("20060102150405")
	snapshotDir := filepath.Join(m.backupDir, id)
	if err := os.MkdirAll(snapshotDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create snapshot dir: %w", err)
	}

	snapshot := Snapshot{
		ID:        id,
		Timestamp: time.Now().UTC(),
		Product:   product,
	}

	for i, action := range plan.Actions {
		if action.Target == "" || action.Kind != "filesystem" {
			continue // Only backup filesystem actions for now
		}

		info, err := os.Stat(action.Target)
		if err != nil {
			if os.IsNotExist(err) {
				continue // File doesn't exist, nothing to backup
			}
			return "", fmt.Errorf("stat failed for %s: %w", action.Target, err)
		}

		backupPath := filepath.Join(snapshotDir, fmt.Sprintf("item_%d", i))
		item := Item{
			OriginalPath: action.Target,
			BackupPath:   backupPath,
			IsDir:        info.IsDir(),
		}

		if err := copyPath(action.Target, backupPath); err != nil {
			return "", fmt.Errorf("failed to backup %s: %w", action.Target, err)
		}

		snapshot.Items = append(snapshot.Items, item)
	}

	manifestPath := filepath.Join(snapshotDir, "manifest.json")
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal manifest: %w", err)
	}
	if err := os.WriteFile(manifestPath, data, 0600); err != nil {
		return "", fmt.Errorf("failed to write manifest: %w", err)
	}

	return id, nil
}

// Rollback restores a specific snapshot by ID.
func (m *Manager) Rollback(id string) error {
	snapshotDir := filepath.Join(m.backupDir, id)
	manifestPath := filepath.Join(snapshotDir, "manifest.json")
	
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read snapshot manifest: %w", err)
	}

	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	for _, item := range snapshot.Items {
		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(item.OriginalPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent dir for %s: %w", item.OriginalPath, err)
		}

		// Remove current item if it exists so we can safely overwrite
		_ = os.RemoveAll(item.OriginalPath)

		if err := copyPath(item.BackupPath, item.OriginalPath); err != nil {
			return fmt.Errorf("failed to restore %s: %w", item.OriginalPath, err)
		}
	}

	return nil
}

// ListSnapshots returns all available snapshots ordered from newest to oldest.
func (m *Manager) ListSnapshots() ([]Snapshot, error) {
	entries, err := os.ReadDir(m.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read backup dir: %w", err)
	}

	var snapshots []Snapshot
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		manifestPath := filepath.Join(m.backupDir, entry.Name(), "manifest.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue // skip invalid snapshots
		}
		
		var snap Snapshot
		if err := json.Unmarshal(data, &snap); err == nil {
			snapshots = append(snapshots, snap)
		}
	}

	// Sort newest first
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Timestamp.After(snapshots[j].Timestamp)
	})

	return snapshots, nil
}

func copyPath(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	
	info, err := os.Stat(src)
	if err == nil {
		_ = os.Chmod(dst, info.Mode())
	}
	
	return out.Sync()
}

func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

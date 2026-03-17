package backup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tianrking/ClawRemove/internal/model"
)

func TestBackupAndRollback(t *testing.T) {
	// Create temp directories for test
	baseDir := t.TempDir()
	sourceDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(sourceDir, "test.txt")
	testContent := []byte("hello backup")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Initialize manager
	mgr := NewManager(baseDir)

	// Create plan
	plan := model.Plan{
		Actions: []model.Action{
			{Target: testFile, Kind: "filesystem"},
		},
	}

	// Create snapshot
	id, err := mgr.CreateSnapshot("test-product", plan)
	if err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty snapshot ID")
	}

	// Verify snapshot exists in list
	snapshots, err := mgr.ListSnapshots()
	if err != nil {
		t.Fatalf("ListSnapshots failed: %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snapshots))
	}
	if snapshots[0].ID != id {
		t.Errorf("expected snapshot ID %q, got %q", id, snapshots[0].ID)
	}

	// Now "delete" the original file
	if err := os.Remove(testFile); err != nil {
		t.Fatalf("failed to delete test file: %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Fatal("test file should be deleted")
	}

	// Rollback
	if err := mgr.Rollback(id); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// Verify file is restored
	restored, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read restored file: %v", err)
	}
	if string(restored) != string(testContent) {
		t.Errorf("expected %q, got %q", string(testContent), string(restored))
	}
}

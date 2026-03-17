package cleanup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/system"
)

// Scanner discovers cleanup candidates across AI artifact storage.
type Scanner struct {
	runner system.Runner
	home   string
}

// NewScanner creates a new cleanup scanner.
func NewScanner(runner system.Runner) *Scanner {
	home, _ := os.UserHomeDir()
	return &Scanner{runner: runner, home: home}
}

// ScanAll aggregates all cleanup candidates.
func (s *Scanner) ScanAll(ctx context.Context) model.CleanupReport {
	var candidates []model.CleanupCandidate
	candidates = append(candidates, s.ScanModelVersions()...)
	candidates = append(candidates, s.ScanOrphanedCaches()...)
	candidates = append(candidates, s.ScanUnusedVectorDBs(ctx)...)
	candidates = append(candidates, s.ScanLogs()...)

	var totalReclaimable int64
	for _, c := range candidates {
		totalReclaimable += c.Size
	}

	summary := "No cleanup candidates found"
	if len(candidates) > 0 {
		summary = fmt.Sprintf("%d candidates found, %s reclaimable", len(candidates), formatSize(totalReclaimable))
	}

	return model.CleanupReport{
		Candidates:       candidates,
		TotalReclaimable: totalReclaimable,
		Summary:          summary,
	}
}

// ScanModelVersions detects old or duplicate model versions that can be cleaned.
func (s *Scanner) ScanModelVersions() []model.CleanupCandidate {
	var candidates []model.CleanupCandidate

	modelDirs := []struct {
		name string
		path string
	}{
		{"Ollama Models", filepath.Join(s.home, ".ollama", "models")},
		{"LM Studio Models", filepath.Join(s.home, ".lmstudio", "models")},
		{"Hugging Face Models", filepath.Join(s.home, ".cache", "huggingface", "hub")},
		{"GPT4All Models", filepath.Join(s.home, ".cache", "gpt4all")},
	}

	for _, md := range modelDirs {
		if _, err := os.Stat(md.path); err != nil {
			continue
		}

		// Scan for old blobs (Ollama-specific: blobs directory)
		blobsDir := filepath.Join(md.path, "blobs")
		if info, err := os.Stat(blobsDir); err == nil && info.IsDir() {
			candidates = append(candidates, s.scanOllamaBlobs(blobsDir, md.name)...)
			continue
		}

		// Generic model directory: find large files older than 90 days
		candidates = append(candidates, s.scanOldModels(md.path, md.name)...)
	}

	return candidates
}

// scanOllamaBlobs scans Ollama blob storage for unreferenced blobs.
func (s *Scanner) scanOllamaBlobs(blobsDir, source string) []model.CleanupCandidate {
	var candidates []model.CleanupCandidate

	// Collect all manifest references
	manifestDir := filepath.Join(filepath.Dir(blobsDir), "manifests")
	referenced := make(map[string]bool)
	filepath.Walk(manifestDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		content, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		// Ollama manifests reference blobs by sha256 digest
		text := string(content)
		if strings.Contains(text, "sha256:") {
			parts := strings.Split(text, "sha256:")
			for _, part := range parts[1:] {
				// Extract the digest (hex chars)
				digest := extractDigest(part)
				if digest != "" {
					referenced["sha256-"+digest] = true
				}
			}
		}
		return nil
	})

	// Check each blob
	entries, _ := os.ReadDir(blobsDir)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !referenced[entry.Name()] && len(referenced) > 0 {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if info.Size() > 1024*1024 { // Only report >1MB blobs
				candidates = append(candidates, model.CleanupCandidate{
					Path:     filepath.Join(blobsDir, entry.Name()),
					Size:     info.Size(),
					Reason:   "Unreferenced model blob (no manifest points to it)",
					Category: "model_version",
					Risk:     "low",
					Source:   source,
				})
			}
		}
	}

	return candidates
}

// scanOldModels finds model files older than 90 days.
func (s *Scanner) scanOldModels(dir, source string) []model.CleanupCandidate {
	var candidates []model.CleanupCandidate
	cutoff := time.Now().AddDate(0, 0, -90)

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		// Model file extensions
		ext := strings.ToLower(filepath.Ext(info.Name()))
		isModel := ext == ".gguf" || ext == ".bin" || ext == ".safetensors" || ext == ".onnx" || ext == ".pt" || ext == ".pth"
		if !isModel {
			return nil
		}
		if info.ModTime().Before(cutoff) && info.Size() > 100*1024*1024 { // >100MB and >90 days old
			candidates = append(candidates, model.CleanupCandidate{
				Path:     path,
				Size:     info.Size(),
				Reason:   fmt.Sprintf("Old model file (last modified %s)", info.ModTime().Format("2006-01-02")),
				Category: "model_version",
				Risk:     "medium",
				Source:   source,
			})
		}
		return nil
	})

	return candidates
}

// ScanOrphanedCaches detects caches for AI runtimes that are no longer installed.
func (s *Scanner) ScanOrphanedCaches() []model.CleanupCandidate {
	var candidates []model.CleanupCandidate

	cacheChecks := []struct {
		cachePath  string
		binaryName string
		name       string
	}{
		{filepath.Join(s.home, ".cache", "huggingface"), "transformers", "Hugging Face Cache"},
		{filepath.Join(s.home, ".cache", "torch"), "python", "PyTorch Cache"},
		{filepath.Join(s.home, ".cache", "tensorflow"), "python", "TensorFlow Cache"},
		{filepath.Join(s.home, ".ollama"), "ollama", "Ollama Cache"},
		{filepath.Join(s.home, ".cache", "gpt4all"), "gpt4all", "GPT4All Cache"},
		{filepath.Join(s.home, ".cache", "pip"), "pip", "pip Cache"},
		{filepath.Join(s.home, ".cache", "sentence_transformers"), "python", "Sentence Transformers Cache"},
	}

	for _, check := range cacheChecks {
		info, err := os.Stat(check.cachePath)
		if err != nil || !info.IsDir() {
			continue
		}

		// Check if the corresponding runtime/binary still exists
		if !commandExists(check.binaryName) {
			size := dirSize(check.cachePath)
			if size > 1024*1024 { // Only report >1MB
				candidates = append(candidates, model.CleanupCandidate{
					Path:     check.cachePath,
					Size:     size,
					Reason:   fmt.Sprintf("Cache for %s but '%s' is not installed", check.name, check.binaryName),
					Category: "orphaned_cache",
					Risk:     "low",
					Source:   check.name,
				})
			}
		}
	}

	// Also check for stale tmp files in agent state dirs
	agentDirs := []string{
		filepath.Join(s.home, ".openclaw", "tmp"),
		filepath.Join(s.home, ".nanobot", "tmp"),
		filepath.Join(s.home, ".picoclaw", "tmp"),
		filepath.Join(s.home, ".cursor", "Cache"),
		filepath.Join(s.home, ".continue", "cache"),
	}
	cutoff := time.Now().AddDate(0, 0, -30)

	for _, dir := range agentDirs {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}
		var staleSize int64
		var staleCount int
		filepath.Walk(dir, func(path string, fi os.FileInfo, werr error) error {
			if werr != nil || fi.IsDir() {
				return nil
			}
			if fi.ModTime().Before(cutoff) {
				staleSize += fi.Size()
				staleCount++
			}
			return nil
		})
		if staleSize > 1024*1024 { // Only report >1MB
			candidates = append(candidates, model.CleanupCandidate{
				Path:     dir,
				Size:     staleSize,
				Reason:   fmt.Sprintf("%d stale temp files (>30 days old)", staleCount),
				Category: "orphaned_cache",
				Risk:     "low",
				Source:   filepath.Base(filepath.Dir(dir)),
			})
		}
	}

	return candidates
}

// ScanUnusedVectorDBs detects vector database storage without an active service.
func (s *Scanner) ScanUnusedVectorDBs(ctx context.Context) []model.CleanupCandidate {
	var candidates []model.CleanupCandidate

	vectorDBChecks := []struct {
		dataPath    string
		serviceName string
		port        int
		name        string
	}{
		{filepath.Join(s.home, ".chromadb"), "chromadb", 8000, "ChromaDB"},
		{filepath.Join(s.home, ".local", "share", "chromadb"), "chromadb", 8000, "ChromaDB"},
		{filepath.Join(s.home, ".local", "share", "milvus"), "milvus", 19530, "Milvus"},
		{filepath.Join(s.home, ".local", "share", "qdrant"), "qdrant", 6333, "Qdrant"},
		{filepath.Join(s.home, ".local", "share", "weaviate"), "weaviate", 8080, "Weaviate"},
	}

	for _, check := range vectorDBChecks {
		info, err := os.Stat(check.dataPath)
		if err != nil || !info.IsDir() {
			continue
		}

		// Check if the service is running (via docker or binary)
		running := false
		if s.runner.Exists(ctx, check.serviceName) {
			running = true
		}
		if s.runner.Exists(ctx, "docker") {
			result := s.runner.Run(ctx, "docker", "ps", "--format", "{{.Names}}")
			if result.OK && strings.Contains(strings.ToLower(result.Stdout), strings.ToLower(check.name)) {
				running = true
			}
		}

		if !running {
			size := dirSize(check.dataPath)
			if size > 1024*1024 { // Only report >1MB
				candidates = append(candidates, model.CleanupCandidate{
					Path:     check.dataPath,
					Size:     size,
					Reason:   fmt.Sprintf("%s data found but service is not running or installed", check.name),
					Category: "unused_vectordb",
					Risk:     "medium",
					Source:   check.name,
				})
			}
		}
	}

	return candidates
}

// ScanLogs finds large or old log files in AI agent directories.
func (s *Scanner) ScanLogs() []model.CleanupCandidate {
	var candidates []model.CleanupCandidate

	// Scan log directories
	logLocations := []struct {
		dir  string
		name string
	}{
		{filepath.Join(s.home, ".openclaw", "logs"), "OpenClaw"},
		{filepath.Join(s.home, ".nanobot", "logs"), "NanoBot"},
		{filepath.Join(s.home, ".picoclaw", "logs"), "PicoClaw"},
		{filepath.Join(s.home, ".cursor", "logs"), "Cursor"},
		{filepath.Join(s.home, ".continue", "logs"), "Continue"},
		{filepath.Join(s.home, ".ollama", "logs"), "Ollama"},
		{filepath.Join(s.home, ".aider", "logs"), "Aider"},
	}

	cutoff := time.Now().AddDate(0, 0, -14) // Logs older than 14 days

	for _, loc := range logLocations {
		info, err := os.Stat(loc.dir)
		if err != nil || !info.IsDir() {
			continue
		}

		var oldLogSize int64
		var logCount int

		filepath.Walk(loc.dir, func(path string, fi os.FileInfo, werr error) error {
			if werr != nil || fi.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(fi.Name()))
			isLog := ext == ".log" || ext == ".txt" || ext == ".jsonl" || strings.Contains(fi.Name(), ".log.")
			if !isLog {
				return nil
			}
			if fi.ModTime().Before(cutoff) {
				oldLogSize += fi.Size()
				logCount++
			}
			return nil
		})

		if oldLogSize > 512*1024 { // Only report >512KB of logs
			candidates = append(candidates, model.CleanupCandidate{
				Path:     loc.dir,
				Size:     oldLogSize,
				Reason:   fmt.Sprintf("%d old log files (>14 days) from %s", logCount, loc.name),
				Category: "log_rotation",
				Risk:     "low",
				Source:   loc.name,
			})
		}
	}

	// Also find individual large log files in state dirs
	stateDirs := []string{
		filepath.Join(s.home, ".openclaw"),
		filepath.Join(s.home, ".nanobot"),
		filepath.Join(s.home, ".picoclaw"),
		filepath.Join(s.home, ".cursor"),
		filepath.Join(s.home, ".ollama"),
	}

	for _, dir := range stateDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if ext != ".log" && ext != ".jsonl" {
				continue
			}
			fi, err := entry.Info()
			if err != nil {
				continue
			}
			if fi.Size() > 50*1024*1024 { // >50MB single log file
				candidates = append(candidates, model.CleanupCandidate{
					Path:     filepath.Join(dir, entry.Name()),
					Size:     fi.Size(),
					Reason:   fmt.Sprintf("Large log file (%s)", formatSize(fi.Size())),
					Category: "log_rotation",
					Risk:     "low",
					Source:   filepath.Base(dir),
				})
			}
		}
	}

	return candidates
}

// Execute removes the given cleanup candidates.
func (s *Scanner) Execute(candidates []model.CleanupCandidate, dryRun bool) []model.CleanupResult {
	var results []model.CleanupResult
	for _, c := range candidates {
		if dryRun {
			results = append(results, model.CleanupResult{
				Path:    c.Path,
				OK:      true,
				DryRun:  true,
				Reclaimed: c.Size,
			})
			continue
		}
		if err := os.RemoveAll(c.Path); err != nil {
			results = append(results, model.CleanupResult{
				Path:  c.Path,
				OK:    false,
				Error: err.Error(),
			})
		} else {
			results = append(results, model.CleanupResult{
				Path:      c.Path,
				OK:        true,
				Reclaimed: c.Size,
			})
		}
	}
	return results
}

// Helper functions

func extractDigest(s string) string {
	var digest []byte
	for _, ch := range s {
		if (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') {
			digest = append(digest, byte(ch))
		} else {
			break
		}
	}
	if len(digest) >= 12 { // Minimum digest length
		return string(digest)
	}
	return ""
}

func commandExists(name string) bool {
	path, ok := os.LookupEnv("PATH")
	if !ok {
		return false
	}
	sep := string(os.PathListSeparator)
	for _, dir := range strings.Split(path, sep) {
		full := filepath.Join(dir, name)
		if _, serr := os.Stat(full); serr == nil {
			return true
		}
		// Also check with .exe on Windows
		if _, serr := os.Stat(full + ".exe"); serr == nil {
			return true
		}
	}
	return false
}

func dirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)
	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.1fTB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.1fGB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1fMB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1fKB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

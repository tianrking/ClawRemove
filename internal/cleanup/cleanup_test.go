package cleanup

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tianrking/ClawRemove/internal/model"
)

func TestScanModelVersionsFindsOldModels(t *testing.T) {
	tmp := t.TempDir()
	scanner := &Scanner{home: tmp}

	// Create a fake model directory with an old model file
	modelDir := filepath.Join(tmp, ".cache", "huggingface", "hub")
	if err := os.MkdirAll(modelDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a model file >100MB and set its mod time to 100 days ago
	modelFile := filepath.Join(modelDir, "old-model.safetensors")
	f, err := os.Create(modelFile)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(110 * 1024 * 1024); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	oldTime := time.Now().AddDate(0, 0, -100)
	if err := os.Chtimes(modelFile, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	candidates := scanner.ScanModelVersions()
	if len(candidates) == 0 {
		t.Fatal("expected at least one model version candidate")
	}

	found := false
	for _, c := range candidates {
		if c.Category == "model_version" && c.Path == modelFile {
			found = true
			if c.Risk != "medium" {
				t.Fatalf("expected risk=medium, got %s", c.Risk)
			}
		}
	}
	if !found {
		t.Fatal("expected to find the old model file as a cleanup candidate")
	}
}

func TestScanLogsFindsOldLogs(t *testing.T) {
	tmp := t.TempDir()
	scanner := &Scanner{home: tmp}

	logDir := filepath.Join(tmp, ".openclaw", "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create several old log files (>512KB total to meet threshold)
	oldTime := time.Now().AddDate(0, 0, -20)
	for _, name := range []string{"agent.log", "error.log", "debug.log"} {
		logFile := filepath.Join(logDir, name)
		f, err := os.Create(logFile)
		if err != nil {
			t.Fatal(err)
		}
		if err := f.Truncate(256 * 1024); err != nil {
			f.Close()
			t.Fatal(err)
		}
		f.Close()
		if err := os.Chtimes(logFile, oldTime, oldTime); err != nil {
			t.Fatal(err)
		}
	}

	candidates := scanner.ScanLogs()
	if len(candidates) == 0 {
		t.Fatal("expected at least one log cleanup candidate")
	}

	found := false
	for _, c := range candidates {
		if c.Category == "log_rotation" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected log_rotation category candidate")
	}
}

func TestScanAllAggregatesEmpty(t *testing.T) {
	tmp := t.TempDir()
	scanner := &Scanner{home: tmp}

	report := scanner.ScanAll(context.Background())
	if len(report.Candidates) != 0 {
		t.Fatalf("expected 0 candidates in empty dir, got %d", len(report.Candidates))
	}
	if report.Summary != "No cleanup candidates found" {
		t.Fatalf("unexpected summary: %s", report.Summary)
	}
}

func TestExecuteDryRunDoesNotDelete(t *testing.T) {
	tmp := t.TempDir()
	scanner := &Scanner{home: tmp}

	testFile := filepath.Join(tmp, "test.log")
	if err := os.WriteFile(testFile, []byte("test data"), 0o644); err != nil {
		t.Fatal(err)
	}

	candidates := []model.CleanupCandidate{{
		Path:     testFile,
		Size:     9,
		Category: "log_rotation",
		Risk:     "low",
	}}

	results := scanner.Execute(candidates, true)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].OK {
		t.Fatal("expected OK=true for dry-run")
	}
	if !results[0].DryRun {
		t.Fatal("expected DryRun=true")
	}

	// File should still exist
	if _, err := os.Stat(testFile); err != nil {
		t.Fatal("file should still exist after dry-run")
	}
}

func TestExecuteRealDeletesFile(t *testing.T) {
	tmp := t.TempDir()
	scanner := &Scanner{home: tmp}

	testFile := filepath.Join(tmp, "delete-me.log")
	if err := os.WriteFile(testFile, []byte("delete this"), 0o644); err != nil {
		t.Fatal(err)
	}

	candidates := []model.CleanupCandidate{{
		Path:     testFile,
		Size:     11,
		Category: "log_rotation",
		Risk:     "low",
	}}

	results := scanner.Execute(candidates, false)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].OK {
		t.Fatalf("expected OK=true, got error: %s", results[0].Error)
	}

	// File should be deleted
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Fatal("file should have been deleted")
	}
}

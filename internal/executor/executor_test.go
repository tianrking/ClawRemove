package executor

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/system"
)

// mockRunner implements system.Runner for testing.
type mockRunner struct {
	runResult    system.CommandResult
	existsResult bool
	runCalled    bool
	existsCalled bool
	lastCommand  []string
}

func (m *mockRunner) Run(ctx context.Context, name string, args ...string) system.CommandResult {
	m.runCalled = true
	m.lastCommand = append([]string{name}, args...)
	return m.runResult
}

func (m *mockRunner) Exists(ctx context.Context, name string) bool {
	m.existsCalled = true
	return m.existsResult
}

func TestExecuteReportOnlyAction(t *testing.T) {
	runner := &mockRunner{}
	exec := New(runner)

	plan := model.Plan{
		Actions: []model.Action{
			{Kind: model.ActionReportOnly, Target: "/test/path", Reason: "test reason"},
		},
	}

	results := exec.Execute(context.Background(), plan, model.Options{})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].OK {
		t.Fatal("expected OK result")
	}
	if !results[0].Skipped {
		t.Fatal("expected skipped to be true for report-only action")
	}
	if results[0].Target != "/test/path" {
		t.Fatalf("expected target /test/path, got %s", results[0].Target)
	}
}

func TestExecuteRunCommand(t *testing.T) {
	runner := &mockRunner{
		runResult: system.CommandResult{OK: true, Stdout: "success"},
	}
	exec := New(runner)

	plan := model.Plan{
		Actions: []model.Action{
			{Kind: model.ActionRunCommand, Command: []string{"echo", "test"}, Target: "echo test", Reason: "test"},
		},
	}

	results := exec.Execute(context.Background(), plan, model.Options{})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].OK {
		t.Fatalf("expected OK result, got error: %s", results[0].Error)
	}
	if !runner.runCalled {
		t.Fatal("expected Run to be called")
	}
}

func TestExecuteRunCommandDryRun(t *testing.T) {
	runner := &mockRunner{}
	exec := New(runner)

	plan := model.Plan{
		Actions: []model.Action{
			{Kind: model.ActionRunCommand, Command: []string{"echo", "test"}, Target: "echo test", Reason: "test"},
		},
	}

	results := exec.Execute(context.Background(), plan, model.Options{DryRun: true})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].OK {
		t.Fatal("expected OK result")
	}
	if !results[0].DryRun {
		t.Fatal("expected DryRun to be true")
	}
	if runner.runCalled {
		t.Fatal("expected Run not to be called in dry-run mode")
	}
}

func TestExecuteRunCommandFailure(t *testing.T) {
	runner := &mockRunner{
		runResult: system.CommandResult{OK: false, Stderr: "command failed"},
	}
	exec := New(runner)

	plan := model.Plan{
		Actions: []model.Action{
			{Kind: model.ActionRunCommand, Command: []string{"false"}, Target: "false", Reason: "test"},
		},
	}

	results := exec.Execute(context.Background(), plan, model.Options{})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].OK {
		t.Fatal("expected failure result")
	}
	if results[0].Error == "" {
		t.Fatal("expected error message")
	}
}

func TestExecuteRunCommandMissingCommand(t *testing.T) {
	runner := &mockRunner{}
	exec := New(runner)

	plan := model.Plan{
		Actions: []model.Action{
			{Kind: model.ActionRunCommand, Command: []string{}, Target: "empty", Reason: "test"},
		},
	}

	results := exec.Execute(context.Background(), plan, model.Options{})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].OK {
		t.Fatal("expected failure for missing command")
	}
	if results[0].Error != "missing command" {
		t.Fatalf("expected 'missing command' error, got: %s", results[0].Error)
	}
}

func TestExecuteRemovePath(t *testing.T) {
	// Create a temporary file to remove
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "testfile")
	if err := os.WriteFile(tmpFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	runner := &mockRunner{}
	exec := New(runner)

	plan := model.Plan{
		Actions: []model.Action{
			{Kind: model.ActionRemovePath, Target: tmpFile, Reason: "test"},
		},
	}

	results := exec.Execute(context.Background(), plan, model.Options{})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].OK {
		t.Fatalf("expected OK result, got error: %s", results[0].Error)
	}

	// Verify file was removed
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Fatal("expected file to be removed")
	}
}

func TestExecuteRemovePathDryRun(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "testfile")
	if err := os.WriteFile(tmpFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	runner := &mockRunner{}
	exec := New(runner)

	plan := model.Plan{
		Actions: []model.Action{
			{Kind: model.ActionRemovePath, Target: tmpFile, Reason: "test"},
		},
	}

	results := exec.Execute(context.Background(), plan, model.Options{DryRun: true})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].OK {
		t.Fatal("expected OK result")
	}
	if !results[0].DryRun {
		t.Fatal("expected DryRun to be true")
	}

	// Verify file still exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Fatal("expected file to still exist in dry-run mode")
	}
}

func TestExecuteRemovePathRootRefusal(t *testing.T) {
	runner := &mockRunner{}
	exec := New(runner)

	plan := model.Plan{
		Actions: []model.Action{
			{Kind: model.ActionRemovePath, Target: "/", Reason: "test"},
		},
	}

	results := exec.Execute(context.Background(), plan, model.Options{})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].OK {
		t.Fatal("expected failure for root path removal")
	}
	if results[0].Error != "refusing to remove root path" {
		t.Fatalf("expected root path refusal error, got: %s", results[0].Error)
	}
}

func TestExecuteCleanShellProfile(t *testing.T) {
	// Create a temporary shell profile with openclaw content
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, ".zshrc")
	content := `# My shell config
export PATH=$PATH:/usr/local/bin
# OpenClaw Completion
eval "$(openclaw completion zsh)"
# End of config
`
	if err := os.WriteFile(profilePath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create temp profile: %v", err)
	}

	runner := &mockRunner{}
	exec := New(runner)

	plan := model.Plan{
		Actions: []model.Action{
			{
				Kind:    model.ActionEditFile,
				Target:  profilePath,
				Reason:  "clean shell profile",
				Markers: []string{"openclaw"},
			},
		},
	}

	results := exec.Execute(context.Background(), plan, model.Options{})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].OK {
		t.Fatalf("expected OK result, got error: %s", results[0].Error)
	}

	// Verify backup was created
	backupPath := profilePath + ".clawremove.bak"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatal("expected backup file to be created")
	}

	// Verify content was cleaned
	cleanedContent, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatalf("failed to read cleaned profile: %v", err)
	}
	cleaned := string(cleanedContent)
	if containsString(cleaned, "openclaw") {
		t.Fatal("expected openclaw content to be removed")
	}
}

func TestExecuteCleanShellProfileDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, ".zshrc")
	content := `# OpenClaw Completion
eval "$(openclaw completion zsh)"
`
	if err := os.WriteFile(profilePath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create temp profile: %v", err)
	}
	originalContent, _ := os.ReadFile(profilePath)

	runner := &mockRunner{}
	exec := New(runner)

	plan := model.Plan{
		Actions: []model.Action{
			{
				Kind:    model.ActionEditFile,
				Target:  profilePath,
				Reason:  "clean shell profile",
				Markers: []string{"openclaw"},
			},
		},
	}

	results := exec.Execute(context.Background(), plan, model.Options{DryRun: true})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].OK {
		t.Fatal("expected OK result")
	}
	if !results[0].DryRun {
		t.Fatal("expected DryRun to be true")
	}

	// Verify content was NOT modified
	currentContent, _ := os.ReadFile(profilePath)
	if string(currentContent) != string(originalContent) {
		t.Fatal("expected profile to remain unchanged in dry-run mode")
	}
}

func TestExecuteCleanShellProfileNotExist(t *testing.T) {
	runner := &mockRunner{}
	exec := New(runner)

	plan := model.Plan{
		Actions: []model.Action{
			{
				Kind:    model.ActionEditFile,
				Target:  "/nonexistent/.zshrc",
				Reason:  "clean shell profile",
				Markers: []string{"openclaw"},
			},
		},
	}

	results := exec.Execute(context.Background(), plan, model.Options{})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].OK {
		t.Fatal("expected OK result for non-existent file")
	}
	if !results[0].Skipped {
		t.Fatal("expected Skipped to be true for non-existent file")
	}
}

func TestRemoveMarkerLines(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		markers  []string
		expected string
	}{
		{
			name: "removes comment with marker",
			content: `line1
# OpenClaw something
line2`,
			markers: []string{"openclaw"},
			expected: `line1
line2`,
		},
		{
			name: "removes eval line following marker comment",
			content: `line1
# OpenClaw Completion
eval "$(openclaw completion zsh)"
line2`,
			markers: []string{"openclaw"},
			expected: `line1
line2`,
		},
		{
			name: "preserves unrelated content",
			content: `line1
# Some other comment
line2`,
			markers: []string{"openclaw"},
			expected: `line1
# Some other comment
line2`,
		},
		{
			name: "handles multiple markers",
			content: `line1
# openclaw config
source ~/.openclaw/init.sh
# moltbot config
eval "$(moltbot init)"
line2`,
			markers: []string{"openclaw", "moltbot"},
			expected: `line1
line2`,
		},
		{
			name: "case insensitive matching",
			content: `line1
# OPENCLAW CONFIG
eval "$(openclaw init)"
line2`,
			markers: []string{"openclaw"},
			expected: `line1
line2`,
		},
		{
			name: "legacy hard-coded comment format",
			content: `line1
# OpenClaw Completion
eval "$(openclaw completion zsh)"
line2`,
			markers: []string{"openclaw"}, // Legacy format still needs markers for eval line
			expected: `line1
line2`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeMarkerLines(tt.content, tt.markers)
			if result != tt.expected {
				t.Errorf("expected:\n%q\ngot:\n%q", tt.expected, result)
			}
		})
	}
}

func TestSlicesClone(t *testing.T) {
	original := []string{"a", "b", "c"}
	cloned := slicesClone(original)

	if len(cloned) != len(original) {
		t.Fatalf("expected same length")
	}
	for i := range original {
		if cloned[i] != original[i] {
			t.Fatalf("expected same elements")
		}
	}

	// Modify clone should not affect original
	cloned[0] = "modified"
	if original[0] == "modified" {
		t.Fatal("expected original to be unchanged")
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{123, "123"},
		{1000, "1000"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := itoa(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

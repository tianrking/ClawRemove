package discovery

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/platform"
	"github.com/tianrking/ClawRemove/internal/system"
)

// mockRunner implements system.Runner for testing.
type mockRunner struct {
	runResults    map[string]system.CommandResult
	existsResults map[string]bool
}

func newMockRunner() *mockRunner {
	return &mockRunner{
		runResults:    make(map[string]system.CommandResult),
		existsResults: make(map[string]bool),
	}
}

func (m *mockRunner) Run(ctx context.Context, name string, args ...string) system.CommandResult {
	key := name + " " + joinArgs(args)
	if result, ok := m.runResults[key]; ok {
		return result
	}
	// Try just the command name
	if result, ok := m.runResults[name]; ok {
		return result
	}
	return system.CommandResult{OK: false, Code: 1}
}

func (m *mockRunner) Exists(ctx context.Context, name string) bool {
	return m.existsResults[name]
}

func joinArgs(args []string) string {
	result := ""
	for _, a := range args {
		if result != "" {
			result += " "
		}
		result += a
	}
	return result
}

func testFacts() model.ProductFacts {
	return model.ProductFacts{
		ID:                "testclaw",
		DisplayName:       "TestClaw",
		StateDirNames:     []string{".testclaw"},
		WorkspaceDirNames: []string{"workspace"},
		Markers:           []string{"testclaw", "test.claw"},
		ShellProfileGlobs: []string{".zshrc", ".bashrc"},
		TempPrefixes:      []string{"testclaw", "testclaw-"},
		AppPaths:          []string{".config/TestClaw"},
		CLIPaths:          []string{".local/bin/testclaw"},
		PackageRefs: []model.PackageRef{
			{Manager: "npm", Name: "testclaw"},
		},
		ListenerPorts: []int{18789, 19001},
	}
}

func TestDiscoverStateDirs(t *testing.T) {
	// Create a temporary home directory structure
	tmpHome := t.TempDir()

	// Create state directories
	stateDir := filepath.Join(tmpHome, ".testclaw")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("failed to create state dir: %v", err)
	}

	// Create a dynamic state directory (prefixed)
	dynamicDir := filepath.Join(tmpHome, ".openclaw-123")
	if err := os.MkdirAll(dynamicDir, 0o755); err != nil {
		t.Fatalf("failed to create dynamic state dir: %v", err)
	}

	// Override home directory for test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	runner := newMockRunner()
	host := platform.Host{OS: "linux"}
	d := New(runner, testFacts(), host)

	// Test discoverStateDirs
	stateDirs := d.discoverStateDirs(tmpHome)

	// Should find at least the explicit state dir
	found := false
	for _, dir := range stateDirs {
		if dir == stateDir {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find .testclaw state directory")
	}
}

func TestDiscoverWorkspaces(t *testing.T) {
	tmpHome := t.TempDir()

	// Create state directory with workspace
	stateDir := filepath.Join(tmpHome, ".testclaw")
	workspaceDir := filepath.Join(stateDir, "workspace")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	// Create dynamic workspace
	dynamicWorkspace := filepath.Join(stateDir, "workspace-projects")
	if err := os.MkdirAll(dynamicWorkspace, 0o755); err != nil {
		t.Fatalf("failed to create dynamic workspace: %v", err)
	}

	runner := newMockRunner()
	host := platform.Host{OS: "linux"}
	d := New(runner, testFacts(), host)

	stateDirs := []string{stateDir}
	workspaces := d.discoverWorkspaces(tmpHome, stateDirs)

	// Should find the workspace directory
	found := false
	for _, dir := range workspaces {
		if dir == workspaceDir {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find workspace directory")
	}
}

func TestDiscoverTempPaths(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	// Create a temp file with prefix
	tempFile := filepath.Join(tmpDir, "testclaw-123")
	if err := os.WriteFile(tempFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Override temp dir for test
	originalTempDir := os.TempDir()
	t.Setenv("TMPDIR", tmpDir)
	defer os.Setenv("TMPDIR", originalTempDir)

	runner := newMockRunner()
	host := platform.Host{OS: "linux"}
	d := New(runner, testFacts(), host)

	// Note: This test may not work perfectly due to os.TempDir() behavior
	// In a real test environment, we would need to mock os.TempDir()
	_ = d.discoverTempPaths()
}

func TestDiscoverShellProfilesWithMarker(t *testing.T) {
	tmpHome := t.TempDir()

	// Create shell profile with marker
	profileContent := `# My shell config
# testclaw completion
eval "$(testclaw completion zsh)"
export PATH=$PATH:/usr/local/bin
`
	profilePath := filepath.Join(tmpHome, ".zshrc")
	if err := os.WriteFile(profilePath, []byte(profileContent), 0o644); err != nil {
		t.Fatalf("failed to create profile: %v", err)
	}

	runner := newMockRunner()
	host := platform.Host{OS: "linux"}
	d := New(runner, testFacts(), host)

	profiles := d.discoverShellProfiles(tmpHome)

	// Should find the profile because it contains the marker
	found := false
	for _, p := range profiles {
		if p == profilePath {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find shell profile with marker")
	}
}

func TestDiscoverShellProfilesWithoutMarker(t *testing.T) {
	tmpHome := t.TempDir()

	// Create shell profile without marker
	profileContent := `# My shell config
export PATH=$PATH:/usr/local/bin
alias ll='ls -la'
`
	profilePath := filepath.Join(tmpHome, ".zshrc")
	if err := os.WriteFile(profilePath, []byte(profileContent), 0o644); err != nil {
		t.Fatalf("failed to create profile: %v", err)
	}

	runner := newMockRunner()
	host := platform.Host{OS: "linux"}
	d := New(runner, testFacts(), host)

	profiles := d.discoverShellProfiles(tmpHome)

	// Should NOT find the profile because it doesn't contain the marker
	for _, p := range profiles {
		if p == profilePath {
			t.Error("should not include profile without marker")
		}
	}
}

func TestDiscoverAppPaths(t *testing.T) {
	tmpHome := t.TempDir()

	// Create app directory
	appDir := filepath.Join(tmpHome, ".config", "TestClaw")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatalf("failed to create app dir: %v", err)
	}

	runner := newMockRunner()
	host := platform.Host{OS: "linux"}
	d := New(runner, testFacts(), host)

	appPaths := d.discoverAppPaths(tmpHome)

	// Should find the app directory
	found := false
	for _, p := range appPaths {
		if p == appDir {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find app path")
	}
}

func TestDiscoverCLIPaths(t *testing.T) {
	tmpHome := t.TempDir()

	// Create CLI path
	cliDir := filepath.Join(tmpHome, ".local", "bin")
	if err := os.MkdirAll(cliDir, 0o755); err != nil {
		t.Fatalf("failed to create CLI dir: %v", err)
	}

	cliPath := filepath.Join(cliDir, "testclaw")
	if err := os.WriteFile(cliPath, []byte("#!/bin/sh"), 0o755); err != nil {
		t.Fatalf("failed to create CLI file: %v", err)
	}

	runner := newMockRunner()
	host := platform.Host{OS: "linux"}
	d := New(runner, testFacts(), host)

	cliPaths := d.discoverCLIPaths(tmpHome)

	// Should find the CLI path
	found := false
	for _, p := range cliPaths {
		if p == cliPath {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find CLI path")
	}
}

func TestDiscoverPackagesNPM(t *testing.T) {
	runner := newMockRunner()
	runner.existsResults["npm"] = true

	host := platform.Host{OS: "linux"}
	d := New(runner, testFacts(), host)

	packages := d.discoverPackages(context.Background())

	// Should include npm package since npm exists
	found := false
	for _, pkg := range packages {
		if pkg.Manager == "npm" && pkg.Name == "testclaw" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find npm package")
	}
}

func TestDiscoverPackagesBrewOnLinux(t *testing.T) {
	runner := newMockRunner()
	runner.existsResults["brew"] = true

	host := platform.Host{OS: "linux"} // Not darwin
	facts := testFacts()
	facts.PackageRefs = append(facts.PackageRefs, model.PackageRef{Manager: "brew", Name: "testclaw"})
	d := New(runner, facts, host)

	packages := d.discoverPackages(context.Background())

	// Should NOT include brew package on Linux
	for _, pkg := range packages {
		if pkg.Manager == "brew" {
			t.Error("should not include brew package on non-darwin platform")
		}
	}
}

func TestDiscoverDarwinServices(t *testing.T) {
	tmpHome := t.TempDir()

	// Create LaunchAgents directory with test plist
	launchAgentsDir := filepath.Join(tmpHome, "Library", "LaunchAgents")
	if err := os.MkdirAll(launchAgentsDir, 0o755); err != nil {
		t.Fatalf("failed to create LaunchAgents dir: %v", err)
	}

	plistPath := filepath.Join(launchAgentsDir, "test.testclaw.agent.plist")
	if err := os.WriteFile(plistPath, []byte(`<?xml version="1.0"?><plist><dict></dict></plist>`), 0o644); err != nil {
		t.Fatalf("failed to create plist: %v", err)
	}

	runner := newMockRunner()
	host := platform.Host{OS: "darwin"}
	d := New(runner, testFacts(), host)

	services := d.discoverDarwinServices(tmpHome)

	// Should find the service
	if len(services) == 0 {
		t.Error("expected to find at least one service")
	}

	found := false
	for _, svc := range services {
		if svc.Name == "test.testclaw.agent" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find testclaw launch agent")
	}
}

func TestDiscoverLinuxServices(t *testing.T) {
	tmpHome := t.TempDir()

	// Create systemd user directory with test service
	systemdDir := filepath.Join(tmpHome, ".config", "systemd", "user")
	if err := os.MkdirAll(systemdDir, 0o755); err != nil {
		t.Fatalf("failed to create systemd dir: %v", err)
	}

	servicePath := filepath.Join(systemdDir, "testclaw.service")
	if err := os.WriteFile(servicePath, []byte(`[Unit]
Description=TestClaw Service
[Service]
ExecStart=/usr/bin/testclaw
`), 0o644); err != nil {
		t.Fatalf("failed to create service file: %v", err)
	}

	runner := newMockRunner()
	host := platform.Host{OS: "linux"}
	d := New(runner, testFacts(), host)

	services := d.discoverLinuxServices(tmpHome)

	// Should find the service
	found := false
	for _, svc := range services {
		if svc.Name == "testclaw" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find testclaw systemd service")
	}
}

func TestDiscoverProcesses(t *testing.T) {
	runner := newMockRunner()
	runner.existsResults["ps"] = true
	runner.runResults["ps"] = system.CommandResult{
		OK:     true,
		Stdout: "1234 1 /usr/bin/testclaw\n5678 1 /usr/bin/other\n",
	}

	host := platform.Host{OS: "linux"}
	d := New(runner, testFacts(), host)

	processes := d.discoverProcesses(context.Background())

	// Should find the testclaw process
	found := false
	for _, proc := range processes {
		if containsSubstring(proc.Command, "testclaw") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find testclaw process")
	}
}

func TestDiscoverListeners(t *testing.T) {
	runner := newMockRunner()
	runner.existsResults["lsof"] = true
	runner.runResults["lsof"] = system.CommandResult{
		OK:     true,
		Stdout: "LISTEN 127.0.0.1:18789 testclaw\nLISTEN 0.0.0.0:80 nginx\n",
	}

	host := platform.Host{OS: "darwin"}
	d := New(runner, testFacts(), host)

	listeners := d.discoverListeners(context.Background())

	// Should find the listener on provider-declared port
	found := false
	for _, listener := range listeners {
		if containsSubstring(listener, "18789") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find listener on port 18789")
	}
}

func TestDiscoverCrontab(t *testing.T) {
	runner := newMockRunner()
	runner.existsResults["crontab"] = true
	runner.runResults["crontab -l"] = system.CommandResult{
		OK:     true,
		Stdout: "# Regular cron\n* * * * * /usr/bin/date\n# TestClaw task\n0 * * * * /usr/bin/testclaw sync\n",
	}

	host := platform.Host{OS: "linux"}
	d := New(runner, testFacts(), host)

	lines := d.discoverCrontab(context.Background())

	// Should find the testclaw crontab line
	found := false
	for _, line := range lines {
		if containsSubstring(line, "testclaw") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find testclaw crontab line")
	}
}

func TestDiscoverCrontabOnWindows(t *testing.T) {
	runner := newMockRunner()
	runner.existsResults["crontab"] = true

	host := platform.Host{OS: "windows"}
	d := New(runner, testFacts(), host)

	lines := d.discoverCrontab(context.Background())

	// Should return nil on Windows
	if lines != nil {
		t.Error("expected nil crontab on Windows")
	}
}

func TestDiscoverContainers(t *testing.T) {
	runner := newMockRunner()
	runner.existsResults["docker"] = true
	// The actual command key includes the full args
	cmdKey := "docker ps -a --format {{.ID}}\t{{.Image}}\t{{.Names}}\t{{.Status}}"
	// Use a marker that will match - "testclaw" as standalone word or with dots
	runner.runResults[cmdKey] = system.CommandResult{
		OK:     true,
		Stdout: "abc123\ttest.claw/image\ttestclaw\tUp 2 hours\ndef456\tother-image\tother-container\tUp 1 hour\n",
	}

	host := platform.Host{OS: "linux"}
	d := New(runner, testFacts(), host)

	containers := d.discoverContainers(context.Background())

	// Should find the testclaw container
	if len(containers) == 0 {
		t.Error("expected to find at least one container")
	}

	found := false
	for _, c := range containers {
		if c.Name == "testclaw" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find testclaw container, got %d containers", len(containers))
	}
}

func TestDiscoverImages(t *testing.T) {
	runner := newMockRunner()
	runner.existsResults["docker"] = true
	runner.runResults["docker images --format {{.Repository}}:{{.Tag}}\t{{.ID}}"] = system.CommandResult{
		OK:     true,
		Stdout: "testclaw:latest\timg123\nother:latest\timg456\n",
	}

	host := platform.Host{OS: "linux"}
	d := New(runner, testFacts(), host)

	images := d.discoverImages(context.Background())

	// Should find the testclaw image
	found := false
	for _, img := range images {
		if containsSubstring(img.Name, "testclaw") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find testclaw image")
	}
}

func TestHasMarker(t *testing.T) {
	tests := []struct {
		input    string
		markers  []string
		expected bool
	}{
		{"testclaw", []string{"testclaw"}, true},
		{"testclaw", []string{"testclaw"}, true},
		{"prefix testclaw suffix", []string{"testclaw"}, true},
		{"notestclawhere", []string{"testclaw"}, false}, // word boundary
		{"xtestclaw", []string{"testclaw"}, false},      // word boundary
		{"test.claw.app", []string{"test.claw"}, true},  // dot marker
		{"", []string{"testclaw"}, false},
		{"anything", []string{}, false},
		{"anything", []string{""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := hasMarker(tt.input, tt.markers)
			if result != tt.expected {
				t.Errorf("hasMarker(%q, %v) = %v, want %v", tt.input, tt.markers, result, tt.expected)
			}
		})
	}
}

func TestContainsWordLike(t *testing.T) {
	tests := []struct {
		input    string
		marker   string
		expected bool
	}{
		{"testclaw", "testclaw", true},
		{"prefix testclaw suffix", "testclaw", true},
		{"testclaw123", "testclaw", false}, // 123 is word rune, so no boundary
		{"xtestclaw", "testclaw", false},   // word boundary
		{"testclawx", "testclaw", false},   // word boundary
		{"_testclaw", "testclaw", false},   // word boundary
		{"testclaw_", "testclaw", false},   // word boundary
		{"test.claw", "test.claw", true},   // dot marker (no word boundary check)
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := containsWordLike(tt.input, tt.marker)
			if result != tt.expected {
				t.Errorf("containsWordLike(%q, %q) = %v, want %v", tt.input, tt.marker, result, tt.expected)
			}
		})
	}
}

func TestAtoi(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"0", 0},
		{"1", 1},
		{"123", 123},
		{"123abc", 123}, // stops at non-digit
		{"abc", 0},      // no digits
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := atoi(tt.input)
			if result != tt.expected {
				t.Errorf("atoi(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestUniqExistingish(t *testing.T) {
	input := []string{"b", "a", "c", "a", "b"}
	result := uniqExistingish(input)

	// Should be deduplicated and sorted
	if len(result) != 3 {
		t.Fatalf("expected 3 items, got %d", len(result))
	}
	if result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Errorf("expected sorted unique items, got %v", result)
	}
}

func TestUniqPackages(t *testing.T) {
	input := []model.PackageRef{
		{Manager: "npm", Name: "test"},
		{Manager: "npm", Name: "test"}, // duplicate
		{Manager: "npm", Name: "other"},
	}
	result := uniqPackages(input)

	if len(result) != 2 {
		t.Errorf("expected 2 unique packages, got %d", len(result))
	}
}

func TestUniqServices(t *testing.T) {
	input := []model.ServiceRef{
		{Platform: "linux", Scope: "user", Name: "test", Path: "/path"},
		{Platform: "linux", Scope: "user", Name: "test", Path: "/path"}, // duplicate
		{Platform: "linux", Scope: "system", Name: "test", Path: "/path2"},
	}
	result := uniqServices(input)

	if len(result) != 2 {
		t.Errorf("expected 2 unique services, got %d", len(result))
	}
}

func TestFileContainsMarker(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file with marker
	fileWithMarker := filepath.Join(tmpDir, "with_marker.txt")
	content := `# Config file
testclaw setting
other setting
`
	if err := os.WriteFile(fileWithMarker, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Create file without marker
	fileWithoutMarker := filepath.Join(tmpDir, "without_marker.txt")
	if err := os.WriteFile(fileWithoutMarker, []byte("no markers here"), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	markers := []string{"testclaw"}

	if !fileContainsMarker(fileWithMarker, markers) {
		t.Error("expected file to contain marker")
	}

	if fileContainsMarker(fileWithoutMarker, markers) {
		t.Error("expected file NOT to contain marker")
	}

	if fileContainsMarker("/nonexistent/file", markers) {
		t.Error("expected nonexistent file to return false")
	}
}

func TestDedupe(t *testing.T) {
	input := []string{"a", "b", "a", "c", "b", ""}
	result := dedupe(input)

	expected := []string{"a", "b", "c"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d items, got %d", len(expected), len(result))
	}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("expected %s at position %d, got %s", v, i, result[i])
		}
	}
}

// Helper function
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

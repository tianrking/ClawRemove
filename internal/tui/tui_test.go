package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tianrking/ClawRemove/internal/cleanup"
	"github.com/tianrking/ClawRemove/internal/core"
	"github.com/tianrking/ClawRemove/internal/executor"
	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/system"
	"github.com/tianrking/ClawRemove/internal/platform"
	"github.com/tianrking/ClawRemove/internal/llm"
)

func TestTUIModelInitialization(t *testing.T) {
	runner := system.NewRunner()
	host := platform.Detect()
	engine := core.NewEngine(runner, llm.NewAdvisorFromEnv(runner, host, nil), host)
	scanner := cleanup.NewScanner(runner)
	exec := executor.New(runner, nil)

	opts := model.Options{}
	m := InitialModel(engine, scanner, exec, opts).(modelTUI)

	if m.state != StateScanning {
		t.Errorf("Expected initial state to be StateScanning, got %v", m.state)
	}

	// Test Init command
	cmd := m.Init()
	if cmd == nil {
		t.Error("Expected Init() to return a command")
	}
}

func TestTUIUpdateAndNavigation(t *testing.T) {
	// Setup model with some mock candidates
	m := modelTUI{
		state: StateSelection,
		candidates: []model.CleanupCandidate{
			{Path: "/test1", Size: 100},
			{Path: "/test2", Size: 200},
			{Path: "/test3", Size: 300},
		},
		selected: make(map[int]struct{}),
		cursor:   0,
	}

	// 1. Test moving down
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(modelTUI)
	if m.cursor != 1 {
		t.Errorf("Expected cursor to move to 1, got %d", m.cursor)
	}

	// 2. Test moving up
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(modelTUI)
	if m.cursor != 0 {
		t.Errorf("Expected cursor to move to 0, got %d", m.cursor)
	}

	// 3. Test selection toggle
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m = updated.(modelTUI)
	if _, ok := m.selected[0]; !ok {
		t.Error("Expected item 0 to be selected")
	}

	// 4. Test toggle all
	m.selected = make(map[int]struct{}) // reset
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = updated.(modelTUI)
	if len(m.selected) != 3 {
		t.Errorf("Expected all 3 items to be selected, got %d", len(m.selected))
	}

	// 5. Test execution trigger
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(modelTUI)
	if m.state != StateExecuting {
		t.Errorf("Expected state to change to StateExecuting, got %v", m.state)
	}
	if cmd == nil {
		t.Error("Expected execution command to be returned")
	}
}

func TestTUIViewRendering(t *testing.T) {
	m := modelTUI{
		state: StateSelection,
		candidates: []model.CleanupCandidate{
			{Path: "/test1", Size: 1024, Category: "test_cache", Reason: "testing"},
		},
		selected: map[int]struct{}{0: {}},
		cursor:   0,
	}

	view := m.View()
	
	if !strings.Contains(view, "ClawRemove") {
		t.Error("View does not contain title")
	}
	if !strings.Contains(view, "[x]") {
		t.Error("View does not show selected item correctly")
	}
	if !strings.Contains(view, "1.0KB") {
		t.Error("View does not format size correctly")
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{500, "500B"},
		{1024, "1.0KB"},
		{1500 * 1024, "1.5MB"},
		{2 * 1024 * 1024 * 1024, "2.0GB"},
	}

	for _, tt := range tests {
		result := formatSize(tt.bytes)
		if result != tt.expected {
			t.Errorf("formatSize(%d): expected %s, got %s", tt.bytes, tt.expected, result)
		}
	}
}

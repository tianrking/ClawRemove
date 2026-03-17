package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tianrking/ClawRemove/internal/cleanup"
	"github.com/tianrking/ClawRemove/internal/core"
	"github.com/tianrking/ClawRemove/internal/executor"
	"github.com/tianrking/ClawRemove/internal/model"
)

type ViewState int

const (
	StateScanning ViewState = iota
	StateSelection
	StateExecuting
	StateDone
)

type modelTUI struct {
	engine     core.Engine
	scanner    *cleanup.Scanner
	exec       executor.Executor
	options    model.Options
	
	state      ViewState
	report     model.CleanupReport
	candidates []model.CleanupCandidate
	selected   map[int]struct{}
	cursor     int
	
	progressMsg string
	execResults []model.Result
	err         error
	
	width  int
	height int
}

type scanResultMsg struct {
	report model.CleanupReport
	err    error
}

type execResultMsg struct {
	results []model.Result
	err     error
}

func InitialModel(engine core.Engine, scanner *cleanup.Scanner, exec executor.Executor, opts model.Options) tea.Model {
	return modelTUI{
		engine:   engine,
		scanner:  scanner,
		exec:     exec,
		options:  opts,
		state:    StateScanning,
		selected: make(map[int]struct{}),
	}
}

func (m modelTUI) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.startScan(),
	)
}

func (m modelTUI) startScan() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		
		report := m.scanner.ScanAll(ctx)
		return scanResultMsg{report: report}
	}
}

func (m modelTUI) executeCleanup() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		
		// Build plan from selected candidates
		var actions []model.Action
		for i, c := range m.candidates {
			if _, ok := m.selected[i]; ok {
				actions = append(actions, model.Action{
					Target: c.Path,
					Kind:   "filesystem",
					Reason: c.Reason,
				})
			}
		}
		
		plan := model.Plan{Actions: actions}
		results := m.exec.Execute(ctx, plan, m.options, "all")
		
		return execResultMsg{results: results}
	}
}

func (m modelTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		
		case "up", "k":
			if m.state == StateSelection && m.cursor > 0 {
				m.cursor--
			}
		
		case "down", "j":
			if m.state == StateSelection && m.cursor < len(m.candidates)-1 {
				m.cursor++
			}
			
		case " ":
			if m.state == StateSelection {
				if _, ok := m.selected[m.cursor]; ok {
					delete(m.selected, m.cursor)
				} else {
					m.selected[m.cursor] = struct{}{}
				}
			}
			
		case "enter":
			if m.state == StateSelection && len(m.selected) > 0 {
				m.state = StateExecuting
				return m, m.executeCleanup()
			} else if m.state == StateDone {
				return m, tea.Quit
			}
			
		case "a":
			if m.state == StateSelection {
				// Toggle all
				if len(m.selected) == len(m.candidates) {
					m.selected = make(map[int]struct{})
				} else {
					for i := range m.candidates {
						m.selected[i] = struct{}{}
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case scanResultMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = StateDone
			return m, nil
		}
		m.report = msg.report
		m.candidates = msg.report.Candidates
		if len(m.candidates) == 0 {
			m.state = StateDone
		} else {
			m.state = StateSelection
			// Pre-select all by default
			for i := range m.candidates {
				m.selected[i] = struct{}{}
			}
		}

	case execResultMsg:
		m.state = StateDone
		m.err = msg.err
		m.execResults = msg.results
	}

	return m, nil
}

func (m modelTUI) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("\n  Error: %v\n\n  Press q to quit.", m.err))
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("ClawRemove - AI Environment Cleanup"))
	b.WriteString("\n\n")

	switch m.state {
	case StateScanning:
		b.WriteString(infoStyle.Render("  🔍 Scanning environment for cleanup candidates..."))
		
	case StateSelection:
		b.WriteString(fmt.Sprintf("  Found %d items. Total reclaimable: %s\n", len(m.candidates), greenStyle.Render(formatSize(m.report.TotalReclaimable))))
		b.WriteString(subtleStyle.Render("  [↑/↓] Navigate  [Space] Toggle  [a] Toggle All  [Enter] Execute  [q] Quit\n\n"))
		
		var selectedSize int64
		for i, c := range m.candidates {
			if _, ok := m.selected[i]; ok {
				selectedSize += c.Size
			}
			
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			
			checked := " "
			if _, ok := m.selected[i]; ok {
				checked = "x"
			}
			
			line := fmt.Sprintf("%s [%s] %s (%s) - %s", 
				cursor, 
				checked, 
				c.Category, 
				formatSize(c.Size),
				c.Reason,
			)
			
			if m.cursor == i {
				b.WriteString(selectedStyle.Render(line) + "\n")
			} else {
				b.WriteString(line + "\n")
			}
		}
		
		b.WriteString(fmt.Sprintf("\n  Selected to reclaim: %s\n", greenStyle.Render(formatSize(selectedSize))))

	case StateExecuting:
		b.WriteString(warnStyle.Render("  🧹 Executing cleanup... Please wait."))

	case StateDone:
		if len(m.candidates) == 0 {
			b.WriteString(infoStyle.Render("  ✨ Environment is clean! No candidates found."))
		} else {
			success := 0
			for _, r := range m.execResults {
				if r.OK {
					success++
				}
			}
			b.WriteString(fmt.Sprintf("  ✅ Cleanup complete. Successfully processed %d/%d items.\n", success, len(m.selected)))
			
			var failed int
			for _, r := range m.execResults {
				if !r.OK {
					failed++
					b.WriteString(errorStyle.Render(fmt.Sprintf("  ❌ Failed: %s - %s\n", r.Target, r.Error)))
				}
			}
			
			if failed == 0 {
				b.WriteString(greenStyle.Render("\n  All selected items were successfully removed!"))
			}
		}
		b.WriteString(subtleStyle.Render("\n\n  Press Enter or q to quit."))
	}

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Left, lipgloss.Top,
		b.String(),
	)
}

// Start launches the interactive TUI.
func Start(engine core.Engine, scanner *cleanup.Scanner, exec executor.Executor, opts model.Options) error {
	p := tea.NewProgram(InitialModel(engine, scanner, exec, opts))
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
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

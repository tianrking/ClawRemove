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
	StateSearch
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

	// Pagination
	pageOffset int
	pageSize   int

	// Search
	searchQuery  string
	filtered     []int // indices of filtered candidates
	searchActive bool

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
		engine:       engine,
		scanner:      scanner,
		exec:         exec,
		options:      opts,
		state:        StateScanning,
		selected:     make(map[int]struct{}),
		pageSize:     15, // Default page size
		searchActive: false,
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
		// Handle search input mode
		if m.searchActive {
			switch msg.String() {
			case "esc":
				m.searchActive = false
				m.searchQuery = ""
				m.filtered = nil
				m.cursor = 0
				m.pageOffset = 0
			case "enter":
				m.searchActive = false
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.updateFilter()
				}
			default:
				if len(msg.String()) == 1 && msg.String()[0] >= 32 {
					m.searchQuery += msg.String()
					m.updateFilter()
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.state == StateSelection && m.cursor > 0 {
				m.cursor--
				m.adjustPageOffset()
			}

		case "down", "j":
			if m.state == StateSelection && m.cursor < m.getDisplayCount()-1 {
				m.cursor++
				m.adjustPageOffset()
			}

		case "pgup":
			if m.state == StateSelection {
				m.cursor -= m.pageSize
				if m.cursor < 0 {
					m.cursor = 0
				}
				m.pageOffset = m.cursor
			}

		case "pgdown":
			if m.state == StateSelection {
				m.cursor += m.pageSize
				max := m.getDisplayCount() - 1
				if m.cursor > max {
					m.cursor = max
				}
				m.adjustPageOffset()
			}

		case " ":
			if m.state == StateSelection {
				idx := m.getRealIndex(m.cursor)
				if _, ok := m.selected[idx]; ok {
					delete(m.selected, idx)
				} else {
					m.selected[idx] = struct{}{}
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
				// Toggle all visible
				visibleIndices := m.getVisibleIndices()
				allSelected := true
				for _, idx := range visibleIndices {
					if _, ok := m.selected[idx]; !ok {
						allSelected = false
						break
					}
				}
				if allSelected {
					for _, idx := range visibleIndices {
						delete(m.selected, idx)
					}
				} else {
					for _, idx := range visibleIndices {
						m.selected[idx] = struct{}{}
					}
				}
			}

		case "/":
			if m.state == StateSelection {
				m.searchActive = true
				m.searchQuery = ""
				m.filtered = nil
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
			// Initialize filtered to show all
			m.filtered = nil
			m.cursor = 0
			m.pageOffset = 0
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
		totalCount := len(m.candidates)
		displayCount := m.getDisplayCount()
		filteredCount := ""
		if m.searchQuery != "" {
			filteredCount = fmt.Sprintf(" (filtered from %d)", totalCount)
		}
		b.WriteString(fmt.Sprintf("  Found %d items%s. Total reclaimable: %s\n",
			displayCount, filteredCount, greenStyle.Render(formatSize(m.report.TotalReclaimable))))

		// Help line with pagination
		helpLine := "  [↑/↓] Navigate  [Space] Toggle  [a] Toggle Visible  [/] Search"
		if displayCount > m.pageSize {
			helpLine += fmt.Sprintf("  [PgUp/PgDn] Page")
		}
		helpLine += "  [Enter] Execute  [q] Quit\n"
		b.WriteString(subtleStyle.Render(helpLine))

		// Search bar if active
		if m.searchActive {
			b.WriteString(fmt.Sprintf("  Search: %s█\n", m.searchQuery))
		} else if m.searchQuery != "" {
			b.WriteString(fmt.Sprintf("  Filter: %s [Esc to clear]\n", m.searchQuery))
		}
		b.WriteString("\n")

		// Calculate visible range
		start := m.pageOffset
		end := start + m.pageSize
		if end > displayCount {
			end = displayCount
		}

		var selectedSize int64
		for i := start; i < end; i++ {
			realIdx := m.getRealIndex(i)
			c := m.candidates[realIdx]

			if _, ok := m.selected[realIdx]; ok {
				selectedSize += c.Size
			}

			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}

			checked := " "
			if _, ok := m.selected[realIdx]; ok {
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

		// Pagination info
		if m.pageSize > 0 && displayCount > m.pageSize {
			totalPages := (displayCount + m.pageSize - 1) / m.pageSize
			currentPage := (m.cursor / m.pageSize) + 1
			b.WriteString(fmt.Sprintf("\n  Page %d/%d (%d-%d of %d)\n",
				currentPage, totalPages, start+1, end, displayCount))
		}

		// Calculate total selected size
		var totalSelectedSize int64
		for i := range m.candidates {
			if _, ok := m.selected[i]; ok {
				totalSelectedSize += m.candidates[i].Size
			}
		}
		b.WriteString(fmt.Sprintf("\n  Selected to reclaim: %s\n", greenStyle.Render(formatSize(totalSelectedSize))))

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

// getDisplayCount returns the number of items to display (filtered or all)
func (m *modelTUI) getDisplayCount() int {
	if len(m.filtered) > 0 {
		return len(m.filtered)
	}
	return len(m.candidates)
}

// getRealIndex converts a display index to the real candidate index
func (m *modelTUI) getRealIndex(displayIdx int) int {
	if len(m.filtered) > 0 {
		return m.filtered[displayIdx]
	}
	return displayIdx
}

// getVisibleIndices returns the indices of visible items on current page
func (m *modelTUI) getVisibleIndices() []int {
	start := m.pageOffset
	end := start + m.pageSize
	count := m.getDisplayCount()
	if end > count {
		end = count
	}

	var indices []int
	for i := start; i < end; i++ {
		indices = append(indices, m.getRealIndex(i))
	}
	return indices
}

// adjustPageOffset adjusts page offset to keep cursor visible
func (m *modelTUI) adjustPageOffset() {
	if m.cursor < m.pageOffset {
		m.pageOffset = m.cursor
	} else if m.cursor >= m.pageOffset+m.pageSize {
		m.pageOffset = m.cursor - m.pageSize + 1
	}
}

// updateFilter updates the filtered list based on search query
func (m *modelTUI) updateFilter() {
	m.filtered = nil
	if m.searchQuery == "" {
		return
	}

	query := strings.ToLower(m.searchQuery)
	for i, c := range m.candidates {
		if strings.Contains(strings.ToLower(c.Category), query) ||
			strings.Contains(strings.ToLower(c.Reason), query) ||
			strings.Contains(strings.ToLower(c.Path), query) {
			m.filtered = append(m.filtered, i)
		}
	}

	// Reset cursor and offset after filter
	m.cursor = 0
	m.pageOffset = 0
}

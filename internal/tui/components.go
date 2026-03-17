package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	infoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575"))

	warnStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E88388"))

	errorStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF0000"))

	greenStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575"))

	subtleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))

	selectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4"))
)

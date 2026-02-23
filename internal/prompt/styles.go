package prompt

import "github.com/charmbracelet/lipgloss"

var (
	activeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6")) // cyan
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green
	faintStyle    = lipgloss.NewStyle().Faint(true)
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // red
	labelStyle    = lipgloss.NewStyle().Bold(true)
)

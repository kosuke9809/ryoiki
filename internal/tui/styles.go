package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170")).Padding(0, 1)
	headerStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("245"))
	selectedStyle     = lipgloss.NewStyle().Background(lipgloss.Color("236")).Bold(true)
	currentMarkerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	statusStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
)

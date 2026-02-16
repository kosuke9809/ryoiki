package tui

import (
	"fmt"
	"strings"
)

func renderHelpView(app *App) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("ryoiki — keybindings"))
	b.WriteString("\n\n")

	bindings := []struct {
		key  string
		desc string
	}{
		{"j / ↓", "Move cursor down"},
		{"k / ↑", "Move cursor up"},
		{"enter", "View workspace detail (single-pane mode)"},
		{"esc", "Go back / cancel"},
		{"d", "Set workspace purpose"},
		{"f", "Forget workspace (removes directory)"},
		{"a", "Add workspace (path first)"},
		{"A", "Quick add workspace (name first)"},
		{"s", "Switch to workspace (cd)"},
		{"/", "Fuzzy search workspaces"},
		{"L", "Toggle stdout pane"},
		{"r", "Refresh workspace list"},
		{"?", "Show this help"},
		{"q", "Quit"},
	}

	for _, bind := range bindings {
		line := fmt.Sprintf("  %-12s %s", bind.key, bind.desc)
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  Press any key to close"))

	return b.String()
}

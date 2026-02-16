package tui

import (
	"fmt"
	"strings"
)

func renderInputOverlay(app *App) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("ryoiki — workspace manager"))
	b.WriteString("\n\n")

	switch app.inputMode {
	case InputDescribe:
		ws := selectedWorkspace(app)
		name := ""
		if ws != nil {
			name = ws.Name
		}
		b.WriteString(fmt.Sprintf("  Purpose for %q:\n", name))
		b.WriteString("  " + app.textInput.View())
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("  enter:confirm  esc:cancel"))

	case InputConfirmForget:
		ws := selectedWorkspace(app)
		name := ""
		if ws != nil {
			name = ws.Name
		}
		b.WriteString(fmt.Sprintf("  Forget workspace %q and remove its directory? (y/n)\n", name))
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("  y:confirm  n/esc:cancel"))

	case InputAddPath:
		b.WriteString("  New workspace path:\n")
		b.WriteString("  " + app.textInput.View())
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("  enter:next  esc:cancel"))

	case InputAddName:
		b.WriteString(fmt.Sprintf("  Path: %s\n", app.addPath))
		b.WriteString("  Workspace name (optional):\n")
		b.WriteString("  " + app.textInput.View())
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("  enter:next  esc:cancel"))

	case InputAddNameFirst:
		b.WriteString("  Quick add — workspace name:\n")
		b.WriteString("  " + app.textInput.View())
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("  Path will be ~/.ryoiki/<repo>-<repohash>/<name>"))
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("  enter:next  esc:cancel"))

	case InputAddPurpose:
		b.WriteString(fmt.Sprintf("  Path: %s\n", app.addPath))
		if app.addName != "" {
			b.WriteString(fmt.Sprintf("  Name: %s\n", app.addName))
		}
		b.WriteString("  Purpose (optional):\n")
		b.WriteString("  " + app.textInput.View())
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("  enter:confirm  esc:cancel"))
	}

	return b.String()
}

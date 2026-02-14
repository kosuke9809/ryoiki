package tui

import (
	"fmt"
	"strings"
)

func renderListView(app *App) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("ryoiki — workspace manager"))
	b.WriteString("\n\n")

	// Header
	header := fmt.Sprintf("  %-15s %-10s %-10s %-30s %s", "NAME", "CHANGE", "COMMIT", "DESC", "PURPOSE")
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	// Workspace rows
	for i, ws := range app.workspaces {
		marker := "  "
		if ws.IsCurrent {
			marker = currentMarkerStyle.Render("* ")
		}

		changeID := shortID(ws.ChangeID, 8)
		commitID := shortID(ws.CommitID, 8)
		desc := ws.Description
		if desc == "" {
			desc = "(empty)"
		}
		desc = truncate(desc, 28)
		purpose := truncate(ws.Purpose, 20)

		row := fmt.Sprintf("%s%-15s %-10s %-10s %-30s %s",
			marker,
			truncate(ws.Name, 15),
			changeID,
			commitID,
			desc,
			purpose,
		)

		if i == app.cursor {
			b.WriteString(selectedStyle.Render(row))
		} else {
			b.WriteString(row)
		}
		b.WriteString("\n")
	}

	if len(app.workspaces) == 0 {
		b.WriteString("  No workspaces found.\n")
	}

	b.WriteString("\n")

	// Status / error messages
	if app.errMsg != "" {
		b.WriteString(errorStyle.Render(app.errMsg))
		b.WriteString("\n")
	} else if app.statusMsg != "" {
		b.WriteString(statusStyle.Render(app.statusMsg))
		b.WriteString("\n")
	}

	// Help
	help := "j/k:move  enter:detail  d:describe  f:forget  a:add  s:switch  r:refresh  q:quit"
	b.WriteString(helpStyle.Render(help))

	return b.String()
}

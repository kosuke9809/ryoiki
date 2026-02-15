package tui

import (
	"fmt"
	"strings"
)

func renderListView(app *App) string {
	var b strings.Builder

	// Title with optional search indicator
	title := "ryoiki — workspace manager"
	if app.searchActive || app.inputMode == InputSearch {
		filteredCount := len(app.workspaces)
		totalCount := len(app.allWorkspaces)
		title = fmt.Sprintf("ryoiki — workspace manager [%d/%d filtered]", filteredCount, totalCount)
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")

	// Search bar (shown when in search mode)
	if app.inputMode == InputSearch {
		searchPrefix := "🔍 search: "
		searchBar := searchPrefix + app.textInput.View()
		b.WriteString(searchBarStyle.Render(searchBar))
		b.WriteString("\n")
	}
	b.WriteString("\n")

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
	var help string
	if app.inputMode == InputSearch {
		help = "type to search  enter:confirm  esc:clear search  q:quit"
	} else {
		help = "j/k:move  enter:detail  d:describe  f:forget  a:add  /:search  r:refresh  q:quit"
	}
	b.WriteString(helpStyle.Render(help))

	return b.String()
}

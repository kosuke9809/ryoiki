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

	visibleRows := app.listVisibleRows()

	total := len(app.workspaces)
	startIdx := app.scrollOffset
	if startIdx > total {
		startIdx = total
	}
	endIdx := startIdx + visibleRows
	if endIdx > total {
		endIdx = total
	}

	// "more above" indicator
	if startIdx > 0 {
		b.WriteString(helpStyle.Render(fmt.Sprintf("  ... %d more above ...", startIdx)))
		b.WriteString("\n")
	}

	// Workspace rows (scrolled window)
	for i := startIdx; i < endIdx; i++ {
		ws := app.workspaces[i]
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

		nameStr := truncate(ws.Name, 15)
		descStr := truncate(desc, 28)
		purposeStr := truncate(ws.Purpose, 20)

		// Use highlighted versions when search is active
		if app.searchActive && app.searchResults != nil {
			if sr := findSearchResult(app, i); sr != nil {
				nameStr = highlightMatches(ws.Name, filterMatchesByField(sr.Matches, "name"), 15)
				purposeStr = highlightMatches(ws.Purpose, filterMatchesByField(sr.Matches, "purpose"), 20)
			}
		}

		row := fmt.Sprintf("%s%-15s %-10s %-10s %-30s %s",
			marker,
			nameStr,
			changeID,
			commitID,
			descStr,
			purposeStr,
		)

		if i == app.cursor {
			b.WriteString(selectedStyle.Render(row))
		} else {
			b.WriteString(row)
		}
		b.WriteString("\n")
	}

	// "more below" indicator
	if endIdx < total {
		b.WriteString(helpStyle.Render(fmt.Sprintf("  ... %d more below ...", total-endIdx)))
		b.WriteString("\n")
	}

	if total == 0 {
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

	b.WriteString(keysBarStyle.Render("KEYS: " + keyHintText(app)))

	return b.String()
}

// findSearchResult returns the SearchResult for workspace at index i, or nil.
func findSearchResult(app *App, wsIndex int) *SearchResult {
	if wsIndex < 0 || wsIndex >= len(app.workspaces) || app.searchResults == nil {
		return nil
	}
	ws := app.workspaces[wsIndex]
	for idx := range app.searchResults {
		if app.searchResults[idx].Workspace.Name == ws.Name {
			return &app.searchResults[idx]
		}
	}
	return nil
}

// filterMatchesByField returns match positions for the given field.
func filterMatchesByField(matches []MatchPosition, field string) []MatchPosition {
	var filtered []MatchPosition
	for _, m := range matches {
		if m.Field == field {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

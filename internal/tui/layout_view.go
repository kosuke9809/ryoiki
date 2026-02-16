package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kosuke9809/ryoiki/internal/display"
)

func renderTriPaneView(app *App) string {
	title := "ryoiki"
	if app.searchActive {
		title = fmt.Sprintf("ryoiki  | search: %s  | %d/%d ws", app.searchQuery, len(app.workspaces), len(app.allWorkspaces))
	} else {
		title = fmt.Sprintf("ryoiki  | %d ws", len(app.workspaces))
	}

	topBar := titleStyle.Render(title)
	if app.inputMode == InputSearch {
		topBar += "\n" + searchBarStyle.Render("search: "+app.textInput.View())
	}

	bodyHeight := app.height - 5
	if app.inputMode == InputSearch {
		bodyHeight--
	}
	if bodyHeight < 12 {
		bodyHeight = 12
	}

	leftWidth := app.width * 40 / 100
	if leftWidth < 36 {
		leftWidth = 36
	}
	rightWidth := app.width - leftWidth - 1
	if rightWidth < 40 {
		rightWidth = 40
	}

	leftTopHeight := bodyHeight * 60 / 100
	if leftTopHeight < 7 {
		leftTopHeight = 7
	}
	leftBottomHeight := bodyHeight - leftTopHeight
	if leftBottomHeight < 6 {
		leftBottomHeight = 6
	}

	leftTop := paneStyle.Width(leftWidth).Height(leftTopHeight).Render(renderWorkspacePane(app, leftTopHeight))
	leftBottom := paneStyle.Width(leftWidth).Height(leftBottomHeight).Render(renderDetailPane(app))
	leftCol := lipgloss.JoinVertical(lipgloss.Left, leftTop, leftBottom)

	jjLogHeight := bodyHeight
	stdoutHeight := 0
	if app.showLogPane {
		jjLogHeight = bodyHeight * 55 / 100
		if jjLogHeight < 7 {
			jjLogHeight = 7
		}
		stdoutHeight = bodyHeight - jjLogHeight
		if stdoutHeight < 5 {
			stdoutHeight = 5
			jjLogHeight = bodyHeight - stdoutHeight
		}
	}

	jjLogPane := paneStyle.Width(rightWidth).Height(jjLogHeight).Render(renderJJLogPane(app, jjLogHeight))
	rightCol := jjLogPane
	if app.showLogPane {
		stdoutPane := paneStyle.Width(rightWidth).Height(stdoutHeight).Render(renderStdoutPane(app, stdoutHeight))
		rightCol = lipgloss.JoinVertical(lipgloss.Left, jjLogPane, stdoutPane)
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)
	keys := keysBarStyle.Width(app.width - 2).Render("KEYS: " + keyHintText(app))

	return topBar + "\n" + body + "\n" + keys
}

func renderWorkspacePane(app *App, paneHeight int) string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("WORKSPACES"))
	b.WriteString("\n")
	b.WriteString("NAME         CHANGE    COMMIT    PURPOSE")
	b.WriteString("\n")

	visibleRows := paneHeight - 4
	if visibleRows < 1 {
		visibleRows = 1
	}
	total := len(app.workspaces)
	startIdx := app.scrollOffset
	if startIdx > total {
		startIdx = total
	}
	endIdx := startIdx + visibleRows
	if endIdx > total {
		endIdx = total
	}

	for i := startIdx; i < endIdx; i++ {
		ws := app.workspaces[i]
		marker := " "
		if ws.IsCurrent {
			marker = "*"
		}
		cursor := " "
		if i == app.cursor {
			cursor = ">"
		}
		nameStr := truncate(ws.Name, 11)
		changeID := shortID(ws.ChangeID, 8)
		commitID := shortID(ws.CommitID, 8)
		purpose := truncate(ws.Purpose, 10)
		row := fmt.Sprintf("%s%s %-11s %-8s  %-8s  %s", cursor, marker, nameStr, changeID, commitID, purpose)
		if i == app.cursor {
			b.WriteString(selectedStyle.Render(row))
		} else {
			b.WriteString(row)
		}
		b.WriteString("\n")
	}

	if total == 0 {
		b.WriteString("No workspaces found.")
	}

	return b.String()
}

func renderDetailPane(app *App) string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("DETAIL"))
	b.WriteString("\n")
	ws := selectedWorkspace(app)
	if ws == nil {
		b.WriteString("No workspace selected.")
		return b.String()
	}

	flags := buildWorkspaceFlags(ws)
	b.WriteString(fmt.Sprintf("name:    %s\n", ws.Name))
	b.WriteString(fmt.Sprintf("path:    %s\n", truncate(ws.Path, 32)))
	b.WriteString(fmt.Sprintf("purpose: %s\n", fallback(ws.Purpose, "(empty)")))
	b.WriteString(fmt.Sprintf("flags:   %s\n", flags))
	if app.errMsg != "" {
		b.WriteString(errorStyle.Render(app.errMsg))
	} else if app.statusMsg != "" {
		b.WriteString(statusStyle.Render(app.statusMsg))
	}
	return b.String()
}

func renderJJLogPane(app *App, paneHeight int) string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("JJ LOG"))
	b.WriteString("\n")

	ws := selectedWorkspace(app)
	if ws == nil {
		b.WriteString("No workspace selected.")
		return b.String()
	}

	if app.jjLogLoading[ws.Name] {
		b.WriteString("Loading jj log...")
		return b.String()
	}
	if errText, ok := app.jjLogErr[ws.Name]; ok {
		b.WriteString(errorStyle.Render("Failed to load jj log: " + truncate(errText, max(10, app.width-65))))
		return b.String()
	}

	lines := app.jjLogCache[ws.Name]
	if len(lines) == 0 {
		b.WriteString("No history available.")
		return b.String()
	}
	maxLines := paneHeight - 3
	if maxLines < 1 {
		maxLines = 1
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	for _, line := range lines {
		b.WriteString(truncate(line, max(10, app.width-50)))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func renderStdoutPane(app *App, paneHeight int) string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("STDOUT / STDERR"))
	b.WriteString("\n")

	ws := selectedWorkspace(app)
	if ws == nil {
		b.WriteString("No workspace selected.")
		return b.String()
	}

	lines := readWorkspaceLogTail(app, ws, paneHeight-3)
	if len(lines) == 0 {
		b.WriteString("No logs yet.")
		return b.String()
	}
	for _, line := range lines {
		b.WriteString(truncate(line, max(10, app.width-50)))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func readWorkspaceLogTail(app *App, ws *display.WorkspaceInfo, maxLines int) []string {
	if maxLines < 1 {
		maxLines = 1
	}

	candidates := []string{
		filepath.Join(app.root, ".ryoiki", "logs", ws.Name+".log"),
		filepath.Join(app.root, ".ryoiki", ws.Name+".log"),
	}
	if ws.Path != "" {
		candidates = append(candidates, filepath.Join(ws.Path, ".ryoiki.log"))
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content := strings.TrimSpace(string(data))
		if content == "" {
			return nil
		}
		lines := strings.Split(content, "\n")
		if len(lines) > maxLines {
			lines = lines[len(lines)-maxLines:]
		}
		return lines
	}
	return nil
}

func buildWorkspaceFlags(ws *display.WorkspaceInfo) string {
	var flags []string
	if ws.Purpose == "" {
		flags = append(flags, "no-purpose")
	}
	if ws.Description == "" {
		flags = append(flags, "no-description")
	}
	if ws.Path == "" {
		flags = append(flags, "no-path")
	}
	if len(flags) == 0 {
		return "ok"
	}
	return strings.Join(flags, ", ")
}

func keyHintText(app *App) string {
	if app.inputMode == InputSearch {
		return "type to search  enter keep  esc clear  q quit"
	}
	if app.inputMode != InputNone {
		return "enter confirm  esc cancel"
	}
	return "j/k move  enter detail  d describe  f forget  a/A add  s switch  / search  L stdout  q quit"
}

func fallback(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

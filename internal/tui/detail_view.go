package tui

import (
	"fmt"
	"strings"

	"github.com/kosuke9809/ryoiki/internal/display"
)

func renderDetailView(app *App) string {
	if app.cursor < 0 || app.cursor >= len(app.workspaces) {
		return "No workspace selected."
	}

	ws := app.workspaces[app.cursor]

	var b strings.Builder

	b.WriteString(titleStyle.Render("ryoiki — workspace detail"))
	b.WriteString("\n\n")

	b.WriteString(renderField("Name", ws.Name))
	if ws.Path != "" {
		b.WriteString(renderField("Path", ws.Path))
	}
	b.WriteString(renderField("Change ID", ws.ChangeID))
	b.WriteString(renderField("Commit ID", ws.CommitID))
	desc := ws.Description
	if desc == "" {
		desc = "(empty)"
	}
	b.WriteString(renderField("Description", desc))
	b.WriteString(renderField("Author", fmt.Sprintf("%s <%s>", ws.AuthorName, ws.AuthorEmail)))
	if ws.Purpose != "" {
		b.WriteString(renderField("Purpose", ws.Purpose))
	}
	if !ws.CreatedAt.IsZero() {
		b.WriteString(renderField("Created", ws.CreatedAt.Format("2006-01-02 15:04:05")))
	}
	if !ws.UpdatedAt.IsZero() {
		b.WriteString(renderField("Updated", ws.UpdatedAt.Format("2006-01-02 15:04:05")))
	}
	if ws.IsCurrent {
		b.WriteString(renderField("Current", "yes"))
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

	help := "esc/q:back  d:describe  f:forget  s:switch"
	b.WriteString(helpStyle.Render(help))

	return b.String()
}

func renderField(label string, value string) string {
	return fmt.Sprintf("  %-14s %s\n", label+":", value)
}

// selectedWorkspace returns the workspace at the current cursor, or nil.
func selectedWorkspace(app *App) *display.WorkspaceInfo {
	if app.cursor < 0 || app.cursor >= len(app.workspaces) {
		return nil
	}
	return &app.workspaces[app.cursor]
}

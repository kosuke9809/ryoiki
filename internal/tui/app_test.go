package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kosuke9809/ryoiki/internal/config"
	"github.com/kosuke9809/ryoiki/internal/display"
	"github.com/kosuke9809/ryoiki/internal/jj"
)

// mockExecutor implements jj.Executor for testing.
type mockExecutor struct {
	output []byte
	err    error
}

func (m *mockExecutor) Execute(args []string) ([]byte, error) {
	return m.output, m.err
}

func newTestApp() App {
	executor := &mockExecutor{}
	ws := jj.NewWorkspaceService(executor)
	store := config.NewMetadataStore("/tmp/test-repo")
	app := NewApp(ws, store, "/tmp/test-repo")
	app.workspaces = []display.WorkspaceInfo{
		{Name: "default", ChangeID: "abc123", CommitID: "def456", IsCurrent: true},
		{Name: "feature", ChangeID: "ghi789", CommitID: "jkl012", Purpose: "new feature"},
		{Name: "bugfix", ChangeID: "mno345", CommitID: "pqr678", Purpose: "fix bug"},
	}
	return app
}

func TestNewApp(t *testing.T) {
	executor := &mockExecutor{}
	ws := jj.NewWorkspaceService(executor)
	store := config.NewMetadataStore("/tmp/test-repo")
	app := NewApp(ws, store, "/tmp/test-repo")

	if app.viewMode != ViewList {
		t.Errorf("expected ViewList, got %d", app.viewMode)
	}
	if app.inputMode != InputNone {
		t.Errorf("expected InputNone, got %d", app.inputMode)
	}
	if app.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", app.cursor)
	}
	if app.workspaces != nil {
		t.Errorf("expected nil workspaces, got %v", app.workspaces)
	}
}

func TestUpdateKeyNavigation(t *testing.T) {
	app := newTestApp()

	// Move down
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	app = model.(App)
	if app.cursor != 1 {
		t.Errorf("expected cursor 1 after j, got %d", app.cursor)
	}

	// Move down again
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	app = model.(App)
	if app.cursor != 2 {
		t.Errorf("expected cursor 2 after j, got %d", app.cursor)
	}

	// Move down at bottom stays at bottom
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	app = model.(App)
	if app.cursor != 2 {
		t.Errorf("expected cursor 2 at bottom, got %d", app.cursor)
	}

	// Move up
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	app = model.(App)
	if app.cursor != 1 {
		t.Errorf("expected cursor 1 after k, got %d", app.cursor)
	}

	// Move up to top
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	app = model.(App)
	if app.cursor != 0 {
		t.Errorf("expected cursor 0 after k, got %d", app.cursor)
	}

	// Move up at top stays at top
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	app = model.(App)
	if app.cursor != 0 {
		t.Errorf("expected cursor 0 at top, got %d", app.cursor)
	}
}

func TestUpdateEnterDetail(t *testing.T) {
	app := newTestApp()

	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app = model.(App)

	if app.viewMode != ViewDetail {
		t.Errorf("expected ViewDetail after enter, got %d", app.viewMode)
	}
}

func TestUpdateQuitFromList(t *testing.T) {
	app := newTestApp()

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	// tea.Quit returns a special batch message; verify it's not nil
	if cmd == nil {
		t.Error("expected quit command, got nil")
	}
}

func TestUpdateEscFromDetail(t *testing.T) {
	app := newTestApp()
	app.viewMode = ViewDetail

	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyEscape})
	app = model.(App)

	if app.viewMode != ViewList {
		t.Errorf("expected ViewList after esc from detail, got %d", app.viewMode)
	}
}

func TestForgetDefaultRejected(t *testing.T) {
	app := newTestApp()
	// cursor is on "default" (index 0)

	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	app = model.(App)

	if app.inputMode != InputNone {
		t.Errorf("expected InputNone (forget rejected for default), got %d", app.inputMode)
	}
	if app.errMsg != "Cannot forget the default workspace" {
		t.Errorf("expected error message about default workspace, got %q", app.errMsg)
	}
}

func TestForgetNonDefault(t *testing.T) {
	app := newTestApp()
	app.cursor = 1 // "feature"

	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	app = model.(App)

	if app.inputMode != InputConfirmForget {
		t.Errorf("expected InputConfirmForget, got %d", app.inputMode)
	}
}

func TestLoadWorkspacesMsg(t *testing.T) {
	app := newTestApp()
	app.workspaces = nil

	infos := []display.WorkspaceInfo{
		{Name: "default", ChangeID: "aaa"},
		{Name: "work", ChangeID: "bbb"},
	}

	model, _ := app.Update(workspacesLoadedMsg{workspaces: infos})
	app = model.(App)

	if len(app.workspaces) != 2 {
		t.Fatalf("expected 2 workspaces, got %d", len(app.workspaces))
	}
	if app.workspaces[0].Name != "default" {
		t.Errorf("expected first workspace 'default', got %q", app.workspaces[0].Name)
	}
	if app.errMsg != "" {
		t.Errorf("expected empty errMsg, got %q", app.errMsg)
	}
}

func TestLoadWorkspacesMsgAdjustsCursor(t *testing.T) {
	app := newTestApp()
	app.cursor = 5 // beyond bounds

	infos := []display.WorkspaceInfo{
		{Name: "default"},
		{Name: "work"},
	}

	model, _ := app.Update(workspacesLoadedMsg{workspaces: infos})
	app = model.(App)

	if app.cursor != 1 {
		t.Errorf("expected cursor adjusted to 1, got %d", app.cursor)
	}
}

func TestViewRendersWithoutPanic(t *testing.T) {
	app := newTestApp()

	// List view
	output := app.View()
	if output == "" {
		t.Error("expected non-empty list view")
	}

	// Detail view
	app.viewMode = ViewDetail
	output = app.View()
	if output == "" {
		t.Error("expected non-empty detail view")
	}

	// Input overlay
	app.inputMode = InputDescribe
	output = app.View()
	if output == "" {
		t.Error("expected non-empty input overlay")
	}
}

func TestWindowSizeMsg(t *testing.T) {
	app := newTestApp()

	model, _ := app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	app = model.(App)

	if app.width != 120 {
		t.Errorf("expected width 120, got %d", app.width)
	}
	if app.height != 40 {
		t.Errorf("expected height 40, got %d", app.height)
	}
}

func TestDescribeEntersInputMode(t *testing.T) {
	app := newTestApp()
	app.cursor = 1 // "feature"

	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	app = model.(App)

	if app.inputMode != InputDescribe {
		t.Errorf("expected InputDescribe, got %d", app.inputMode)
	}
}

func TestAddEntersInputMode(t *testing.T) {
	app := newTestApp()

	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	app = model.(App)

	if app.inputMode != InputAddPath {
		t.Errorf("expected InputAddPath, got %d", app.inputMode)
	}
}

func TestConfirmForgetRejectsOnNo(t *testing.T) {
	app := newTestApp()
	app.cursor = 1
	app.inputMode = InputConfirmForget

	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	app = model.(App)

	if app.inputMode != InputNone {
		t.Errorf("expected InputNone after 'n', got %d", app.inputMode)
	}
}

func TestEscCancelsDescribe(t *testing.T) {
	app := newTestApp()
	app.inputMode = InputDescribe

	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyEscape})
	app = model.(App)

	if app.inputMode != InputNone {
		t.Errorf("expected InputNone after esc, got %d", app.inputMode)
	}
}

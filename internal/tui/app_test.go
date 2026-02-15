package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kosuke9809/ryoiki/internal/config"
	"github.com/kosuke9809/ryoiki/internal/display"
	"github.com/kosuke9809/ryoiki/internal/jj"
)

// mockExecutor implements jj.Executor for testing.
type mockExecutor struct {
	output    []byte
	err       error
	dirOutput []byte
	dirErr    error
}

func (m *mockExecutor) Execute(args []string) ([]byte, error) {
	return m.output, m.err
}

func (m *mockExecutor) ExecuteInDir(args []string, dir string) ([]byte, error) {
	return m.dirOutput, m.dirErr
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

func TestAddQuickEntersNameFirstMode(t *testing.T) {
	app := newTestApp()

	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	app = model.(App)

	if app.inputMode != InputAddNameFirst {
		t.Errorf("expected InputAddNameFirst, got %d", app.inputMode)
	}
}

func TestAddNameFirstSetsAutoPath(t *testing.T) {
	app := newTestApp()
	app.inputMode = InputAddNameFirst
	app.textInput.SetValue("my-workspace")

	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app = model.(App)

	if app.inputMode != InputAddPurpose {
		t.Errorf("expected InputAddPurpose after name entry, got %d", app.inputMode)
	}
	if app.addName != "my-workspace" {
		t.Errorf("expected addName 'my-workspace', got %q", app.addName)
	}
	expectedPath, err := config.GetDefaultWorkspacePath("/tmp/test-repo", "my-workspace")
	if err != nil {
		t.Fatalf("failed to compute default workspace path: %v", err)
	}
	if app.addPath != expectedPath {
		t.Errorf("expected addPath %q, got %q", expectedPath, app.addPath)
	}
}

func TestAddNameFirstEmptyNameCancels(t *testing.T) {
	app := newTestApp()
	app.inputMode = InputAddNameFirst
	app.textInput.SetValue("")

	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app = model.(App)

	if app.inputMode != InputNone {
		t.Errorf("expected InputNone for empty name, got %d", app.inputMode)
	}
}

func TestSwitchDefaultWorkspace(t *testing.T) {
	app := newTestApp()
	app.cursor = 0 // "default"

	model, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	app = model.(App)

	if app.SwitchPath != "/tmp/test-repo" {
		t.Errorf("expected SwitchPath to be root, got %q", app.SwitchPath)
	}
	if cmd == nil {
		t.Error("expected quit command, got nil")
	}
}

func TestSwitchWorkspaceWithPath(t *testing.T) {
	app := newTestApp()
	app.workspaces[1].Path = "/tmp/test-repo/feature-ws"
	app.cursor = 1

	model, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	app = model.(App)

	if app.SwitchPath != "/tmp/test-repo/feature-ws" {
		t.Errorf("expected SwitchPath '/tmp/test-repo/feature-ws', got %q", app.SwitchPath)
	}
	if cmd == nil {
		t.Error("expected quit command, got nil")
	}
}

func TestHelpViewOpensAndCloses(t *testing.T) {
	app := newTestApp()

	// Open help
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	app = model.(App)

	if app.viewMode != ViewHelp {
		t.Errorf("expected ViewHelp, got %d", app.viewMode)
	}

	// Help view renders without panic
	output := app.View()
	if output == "" {
		t.Error("expected non-empty help view")
	}

	// Any key closes help
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	app = model.(App)

	if app.viewMode != ViewList {
		t.Errorf("expected ViewList after closing help, got %d", app.viewMode)
	}
}

func TestScrollOffset(t *testing.T) {
	app := newTestApp()
	app.height = 10 // Small terminal: overhead=6, so visibleRows=4

	// With 3 workspaces and visibleRows=4, no scroll needed
	app.cursor = 2
	app.ensureCursorVisible()
	if app.scrollOffset != 0 {
		t.Errorf("expected scrollOffset 0, got %d", app.scrollOffset)
	}

	// Add more workspaces to test scrolling
	for i := 0; i < 10; i++ {
		app.workspaces = append(app.workspaces, display.WorkspaceInfo{
			Name: "extra",
		})
	}

	// Move cursor beyond visible area
	app.cursor = 8
	app.ensureCursorVisible()
	if app.scrollOffset <= 0 {
		t.Errorf("expected scrollOffset > 0 for cursor=8, got %d", app.scrollOffset)
	}

	// Move cursor back up
	app.cursor = 0
	app.ensureCursorVisible()
	if app.scrollOffset != 0 {
		t.Errorf("expected scrollOffset 0 after moving to top, got %d", app.scrollOffset)
	}
}

func TestForgetDoneMsgWithDirectory(t *testing.T) {
	app := newTestApp()

	// Test with directory removed
	model, _ := app.Update(forgetDoneMsg{name: "feature", dirRemoved: "/tmp/feature"})
	app = model.(App)
	if app.statusMsg == "" {
		t.Error("expected status message for forget with directory")
	}
	if app.inputMode != InputNone {
		t.Errorf("expected InputNone after forget, got %d", app.inputMode)
	}
}

func TestForgetDoneMsgWithDirError(t *testing.T) {
	app := newTestApp()

	model, _ := app.Update(forgetDoneMsg{name: "feature", dirErr: fmt.Errorf("permission denied")})
	app = model.(App)
	if app.statusMsg == "" {
		t.Error("expected status message for forget with dir error")
	}
}

func TestWindowSizeSwitchesToTriPaneLayout(t *testing.T) {
	app := newTestApp()

	model, _ := app.Update(tea.WindowSizeMsg{Width: 140, Height: 30})
	app = model.(App)

	if app.layoutMode != LayoutTriPane {
		t.Fatalf("expected LayoutTriPane, got %v", app.layoutMode)
	}

	model, _ = app.Update(tea.WindowSizeMsg{Width: 90, Height: 30})
	app = model.(App)

	if app.layoutMode != LayoutSingle {
		t.Fatalf("expected LayoutSingle after resize, got %v", app.layoutMode)
	}
}

func TestToggleLogKey(t *testing.T) {
	app := newTestApp()
	app.showLogPane = true

	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'L'}})
	app = model.(App)

	if app.showLogPane {
		t.Fatal("expected log pane to be hidden after L")
	}
}

func TestTriPaneViewRendersKeysBar(t *testing.T) {
	app := newTestApp()
	app.layoutMode = LayoutTriPane
	app.width = 140
	app.height = 30

	out := app.View()
	if !strings.Contains(out, "WORKSPACES") {
		t.Fatal("expected tri-pane to include WORKSPACES section")
	}
	if !strings.Contains(out, "DETAIL") {
		t.Fatal("expected tri-pane to include DETAIL section")
	}
	if !strings.Contains(out, "LOG") {
		t.Fatal("expected tri-pane to include LOG section")
	}
	if !strings.Contains(out, "KEYS:") {
		t.Fatal("expected tri-pane to include KEYS footer")
	}
}

func TestSearchModeKeyHints(t *testing.T) {
	app := newTestApp()
	app.inputMode = InputSearch

	hints := keyHintText(&app)
	if !strings.Contains(hints, "type to search") {
		t.Fatalf("unexpected hints for search mode: %q", hints)
	}
}

func TestJJLogLoadedMsgStoresCache(t *testing.T) {
	app := newTestApp()

	model, _ := app.Update(jjLogLoadedMsg{name: "default", lines: []string{"@ abc", "o def"}})
	app = model.(App)

	if len(app.jjLogCache["default"]) != 2 {
		t.Fatalf("expected cache entries, got %d", len(app.jjLogCache["default"]))
	}
}

func TestJJLogErrMsgStoresError(t *testing.T) {
	app := newTestApp()

	model, _ := app.Update(jjLogErrMsg{name: "default", err: fmt.Errorf("boom")})
	app = model.(App)

	if app.jjLogErr["default"] == "" {
		t.Fatal("expected jj log error to be stored")
	}
}

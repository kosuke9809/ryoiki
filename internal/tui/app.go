package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/kosuke9809/ryoiki/internal/config"
	"github.com/kosuke9809/ryoiki/internal/display"
	"github.com/kosuke9809/ryoiki/internal/jj"
)

// App is the Bubble Tea model for the ryoiki TUI.
type App struct {
	// Services
	ws    *jj.WorkspaceService
	store *config.MetadataStore
	root  string

	// Data
	workspaces []display.WorkspaceInfo

	// UI state
	viewMode  ViewMode
	inputMode InputMode
	cursor    int
	width     int
	height    int

	// Input
	textInput textinput.Model

	// Add workflow
	addPath string
	addName string

	// Messages
	statusMsg string
	errMsg    string
}

// NewApp creates a new App instance.
func NewApp(ws *jj.WorkspaceService, store *config.MetadataStore, root string) App {
	ti := textinput.New()
	ti.CharLimit = 256
	return App{
		ws:        ws,
		store:     store,
		root:      root,
		viewMode:  ViewList,
		inputMode: InputNone,
		textInput: ti,
	}
}

// Init implements tea.Model.
func (app App) Init() tea.Cmd {
	return app.loadWorkspaces()
}

// Update implements tea.Model.
func (app App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		app.width = msg.Width
		app.height = msg.Height
		return app, nil

	case workspacesLoadedMsg:
		app.workspaces = msg.workspaces
		app.errMsg = ""
		if app.cursor >= len(app.workspaces) {
			app.cursor = max(0, len(app.workspaces)-1)
		}
		return app, nil

	case errMsg:
		app.errMsg = fmt.Sprintf("Error: %v", msg.err)
		return app, nil

	case forgetDoneMsg:
		app.statusMsg = fmt.Sprintf("Workspace %q forgotten", msg.name)
		app.inputMode = InputNone
		app.viewMode = ViewList
		return app, app.loadWorkspaces()

	case describeDoneMsg:
		app.statusMsg = fmt.Sprintf("Purpose set for %q", msg.name)
		app.inputMode = InputNone
		return app, app.loadWorkspaces()

	case addDoneMsg:
		app.statusMsg = fmt.Sprintf("Workspace %q added", msg.name)
		app.inputMode = InputNone
		app.viewMode = ViewList
		return app, app.loadWorkspaces()

	case tea.KeyMsg:
		// Handle input modes first
		if app.inputMode != InputNone {
			return app.updateInput(msg)
		}

		switch app.viewMode {
		case ViewList:
			return app.updateList(msg)
		case ViewDetail:
			return app.updateDetail(msg)
		}
	}

	return app, nil
}

// View implements tea.Model.
func (app App) View() string {
	if app.inputMode != InputNone {
		return renderInputOverlay(&app)
	}
	switch app.viewMode {
	case ViewDetail:
		return renderDetailView(&app)
	default:
		return renderListView(&app)
	}
}

// --- List view key handling ---

func (app App) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		return app, tea.Quit

	case key.Matches(msg, keys.Up):
		if app.cursor > 0 {
			app.cursor--
		}
		return app, nil

	case key.Matches(msg, keys.Down):
		if app.cursor < len(app.workspaces)-1 {
			app.cursor++
		}
		return app, nil

	case key.Matches(msg, keys.Enter):
		if len(app.workspaces) > 0 {
			app.viewMode = ViewDetail
		}
		return app, nil

	case key.Matches(msg, keys.Describe):
		if ws := selectedWorkspace(&app); ws != nil {
			app.inputMode = InputDescribe
			app.textInput.SetValue(ws.Purpose)
			app.textInput.Focus()
			return app, app.textInput.Cursor.BlinkCmd()
		}
		return app, nil

	case key.Matches(msg, keys.Forget):
		if ws := selectedWorkspace(&app); ws != nil {
			if ws.Name == "default" {
				app.errMsg = "Cannot forget the default workspace"
				return app, nil
			}
			app.inputMode = InputConfirmForget
		}
		return app, nil

	case key.Matches(msg, keys.Add):
		app.inputMode = InputAddPath
		app.addPath = ""
		app.addName = ""
		app.textInput.SetValue("")
		app.textInput.Placeholder = "/path/to/workspace"
		app.textInput.Focus()
		return app, app.textInput.Cursor.BlinkCmd()

	case key.Matches(msg, keys.Switch):
		// Switch is a placeholder — jj workspace switch requires shell context
		app.statusMsg = "Use 'cd' to switch workspaces"
		return app, nil

	case key.Matches(msg, keys.Refresh):
		app.statusMsg = "Refreshing..."
		return app, app.loadWorkspaces()
	}

	return app, nil
}

// --- Detail view key handling ---

func (app App) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Esc), key.Matches(msg, keys.Quit):
		app.viewMode = ViewList
		app.errMsg = ""
		app.statusMsg = ""
		return app, nil

	case key.Matches(msg, keys.Describe):
		if ws := selectedWorkspace(&app); ws != nil {
			app.inputMode = InputDescribe
			app.textInput.SetValue(ws.Purpose)
			app.textInput.Focus()
			return app, app.textInput.Cursor.BlinkCmd()
		}
		return app, nil

	case key.Matches(msg, keys.Forget):
		if ws := selectedWorkspace(&app); ws != nil {
			if ws.Name == "default" {
				app.errMsg = "Cannot forget the default workspace"
				return app, nil
			}
			app.inputMode = InputConfirmForget
		}
		return app, nil

	case key.Matches(msg, keys.Switch):
		app.statusMsg = "Use 'cd' to switch workspaces"
		return app, nil
	}

	return app, nil
}

// --- Input mode handling ---

func (app App) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch app.inputMode {
	case InputConfirmForget:
		return app.updateConfirmForget(msg)
	case InputDescribe:
		return app.updateDescribeInput(msg)
	case InputAddPath:
		return app.updateAddPath(msg)
	case InputAddName:
		return app.updateAddName(msg)
	case InputAddPurpose:
		return app.updateAddPurpose(msg)
	}
	return app, nil
}

func (app App) updateConfirmForget(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		ws := selectedWorkspace(&app)
		if ws != nil {
			return app, app.forgetWorkspace(ws.Name)
		}
		app.inputMode = InputNone
		return app, nil
	case "n", "N", "esc":
		app.inputMode = InputNone
		return app, nil
	}
	return app, nil
}

func (app App) updateDescribeInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		ws := selectedWorkspace(&app)
		if ws != nil {
			purpose := app.textInput.Value()
			app.textInput.Blur()
			return app, app.describeWorkspace(ws.Name, purpose)
		}
		app.inputMode = InputNone
		return app, nil
	case "esc":
		app.inputMode = InputNone
		app.textInput.Blur()
		return app, nil
	default:
		var cmd tea.Cmd
		app.textInput, cmd = app.textInput.Update(msg)
		return app, cmd
	}
}

func (app App) updateAddPath(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		app.addPath = app.textInput.Value()
		if app.addPath == "" {
			app.inputMode = InputNone
			app.textInput.Blur()
			return app, nil
		}
		app.inputMode = InputAddName
		app.textInput.SetValue("")
		app.textInput.Placeholder = "workspace-name"
		return app, nil
	case "esc":
		app.inputMode = InputNone
		app.textInput.Blur()
		return app, nil
	default:
		var cmd tea.Cmd
		app.textInput, cmd = app.textInput.Update(msg)
		return app, cmd
	}
}

func (app App) updateAddName(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		app.addName = app.textInput.Value()
		app.inputMode = InputAddPurpose
		app.textInput.SetValue("")
		app.textInput.Placeholder = "purpose description"
		return app, nil
	case "esc":
		app.inputMode = InputNone
		app.textInput.Blur()
		return app, nil
	default:
		var cmd tea.Cmd
		app.textInput, cmd = app.textInput.Update(msg)
		return app, cmd
	}
}

func (app App) updateAddPurpose(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		purpose := app.textInput.Value()
		app.textInput.Blur()
		return app, app.addWorkspace(app.addPath, app.addName, purpose)
	case "esc":
		app.inputMode = InputNone
		app.textInput.Blur()
		return app, nil
	default:
		var cmd tea.Cmd
		app.textInput, cmd = app.textInput.Update(msg)
		return app, cmd
	}
}

// --- Async commands ---

func (app *App) loadWorkspaces() tea.Cmd {
	ws := app.ws
	store := app.store
	root := app.root
	return func() tea.Msg {
		workspaces, err := ws.List()
		if err != nil {
			return errMsg{err: err}
		}
		metadataMap, err := store.ListWorkspaces()
		if err != nil {
			return errMsg{err: err}
		}
		currentRoot, _ := ws.Root()
		infos := display.BuildWorkspaceInfos(workspaces, metadataMap, root, currentRoot)
		return workspacesLoadedMsg{workspaces: infos}
	}
}

func (app *App) forgetWorkspace(name string) tea.Cmd {
	ws := app.ws
	store := app.store
	return func() tea.Msg {
		if err := ws.Forget(name); err != nil {
			return errMsg{err: err}
		}
		if err := store.Remove(name); err != nil {
			return errMsg{err: err}
		}
		return forgetDoneMsg{name: name}
	}
}

func (app *App) describeWorkspace(name, purpose string) tea.Cmd {
	store := app.store
	return func() tea.Msg {
		if err := store.SetPurpose(name, purpose); err != nil {
			return errMsg{err: err}
		}
		return describeDoneMsg{name: name, purpose: purpose}
	}
}

func (app *App) addWorkspace(path, name, purpose string) tea.Cmd {
	ws := app.ws
	store := app.store
	return func() tea.Msg {
		opts := jj.AddOptions{Name: name}
		if err := ws.Add(path, opts); err != nil {
			return errMsg{err: err}
		}
		// Determine the actual name for metadata
		wsName := name
		if wsName == "" {
			// jj uses the directory basename as workspace name when no name is given
			wsName = path
		}
		if err := store.AddWorkspace(wsName, path, purpose); err != nil {
			return errMsg{err: err}
		}
		return addDoneMsg{name: wsName}
	}
}

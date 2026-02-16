package tui

import (
	"fmt"
	"os"

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
	viewMode     ViewMode
	layoutMode   LayoutMode
	inputMode    InputMode
	cursor       int
	width        int
	height       int
	scrollOffset int
	showLogPane  bool

	// Switch output (set before tea.Quit to signal shell cd)
	SwitchPath string

	// Input
	textInput textinput.Model

	// Add workflow
	addPath string
	addName string

	// Search state
	allWorkspaces []display.WorkspaceInfo // Original workspace data
	searchQuery   string                  // Current search query
	searchActive  bool                    // Whether search is active
	searchResults []SearchResult          // Search results with highlighting info
	fuzzyMatcher  *FuzzyMatcher           // Fuzzy search engine
	jjLogCache    map[string][]string
	jjLogErr      map[string]string
	jjLogLoading  map[string]bool
	lastSelected  string

	// Messages
	statusMsg string
	errMsg    string
}

// NewApp creates a new App instance.
func NewApp(ws *jj.WorkspaceService, store *config.MetadataStore, root string) App {
	ti := textinput.New()
	ti.CharLimit = 256
	return App{
		ws:           ws,
		store:        store,
		root:         root,
		viewMode:     ViewList,
		layoutMode:   LayoutSingle,
		inputMode:    InputNone,
		textInput:    ti,
		searchActive: false,
		fuzzyMatcher: NewFuzzyMatcher(),
		showLogPane:  true,
		jjLogCache:   make(map[string][]string),
		jjLogErr:     make(map[string]string),
		jjLogLoading: make(map[string]bool),
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
		app.layoutMode = app.detectLayoutMode()
		return app, nil

	case workspacesLoadedMsg:
		app.allWorkspaces = msg.workspaces
		app.errMsg = ""

		// Apply search if active
		if app.searchActive && app.searchQuery != "" {
			app.performSearch()
		} else {
			app.workspaces = app.allWorkspaces
			app.searchResults = nil
		}

		if app.cursor >= len(app.workspaces) {
			app.cursor = max(0, len(app.workspaces)-1)
		}
		return app, app.loadJJLogForSelected(false)

	case jjLogLoadedMsg:
		app.jjLogCache[msg.name] = msg.lines
		delete(app.jjLogErr, msg.name)
		delete(app.jjLogLoading, msg.name)
		return app, nil

	case jjLogErrMsg:
		app.jjLogErr[msg.name] = msg.err.Error()
		delete(app.jjLogLoading, msg.name)
		return app, nil

	case errMsg:
		app.errMsg = fmt.Sprintf("Error: %v", msg.err)
		return app, nil

	case forgetDoneMsg:
		if msg.dirErr != nil {
			app.statusMsg = fmt.Sprintf("Workspace %q forgotten (directory removal failed: %v)", msg.name, msg.dirErr)
		} else if msg.dirRemoved != "" {
			app.statusMsg = fmt.Sprintf("Workspace %q forgotten, directory removed: %s", msg.name, msg.dirRemoved)
		} else {
			app.statusMsg = fmt.Sprintf("Workspace %q forgotten", msg.name)
		}
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
		case ViewHelp:
			// Any key closes help
			app.viewMode = ViewList
			return app, nil
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
	case ViewHelp:
		return renderHelpView(&app)
	default:
		if app.layoutMode == LayoutTriPane {
			return renderTriPaneView(&app)
		}
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
			app.ensureCursorVisible()
		}
		return app, app.loadJJLogForSelected(false)

	case key.Matches(msg, keys.Down):
		if app.cursor < len(app.workspaces)-1 {
			app.cursor++
			app.ensureCursorVisible()
		}
		return app, app.loadJJLogForSelected(false)

	case key.Matches(msg, keys.Enter):
		if len(app.workspaces) > 0 && app.layoutMode == LayoutSingle {
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

	case key.Matches(msg, keys.AddQuick):
		app.inputMode = InputAddNameFirst
		app.addPath = ""
		app.addName = ""
		app.textInput.SetValue("")
		app.textInput.Placeholder = "workspace-name"
		app.textInput.Focus()
		return app, app.textInput.Cursor.BlinkCmd()

	case key.Matches(msg, keys.Switch):
		return app.switchToSelected()

	case key.Matches(msg, keys.Help):
		app.viewMode = ViewHelp
		return app, nil

	case key.Matches(msg, keys.Refresh):
		app.statusMsg = "Refreshing..."
		app.clearJJLogCache()
		return app, app.loadWorkspaces()

	case key.Matches(msg, keys.Search):
		app.startSearch()
		return app, app.textInput.Cursor.BlinkCmd()

	case key.Matches(msg, keys.ToggleLog):
		app.showLogPane = !app.showLogPane
		return app, nil
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
		return app.switchToSelected()
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
	case InputAddNameFirst:
		return app.updateAddNameFirst(msg)
	case InputAddPurpose:
		return app.updateAddPurpose(msg)
	case InputSearch:
		return app.updateSearchInput(msg)
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

func (app App) updateAddNameFirst(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		app.addName = app.textInput.Value()
		if app.addName == "" {
			app.inputMode = InputNone
			app.textInput.Blur()
			return app, nil
		}
		// Auto-generate path: ~/.ryoiki/<repo>/<name>
		defaultPath, err := config.GetDefaultWorkspacePath(app.root, app.addName)
		if err != nil {
			app.errMsg = fmt.Sprintf("Cannot compute workspace path: %v", err)
			app.inputMode = InputNone
			app.textInput.Blur()
			return app, nil
		}
		app.addPath = defaultPath
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

func (app App) updateSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Enter confirms search, stays in search mode
		return app, nil
	case "esc":
		// Escape clears search and exits search mode
		app.clearSearch()
		return app, app.loadJJLogForSelected(false)
	default:
		// Update search query in real-time
		var cmd tea.Cmd
		app.textInput, cmd = app.textInput.Update(msg)

		// Perform search on every keystroke
		query := app.textInput.Value()
		app.updateSearchQuery(query)
		return app, tea.Batch(cmd, app.loadJJLogForSelected(false))
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
		// Resolve workspace path before forgetting (for directory removal)
		wsPath, _ := store.ResolveWorkspacePath(name)

		if err := ws.Forget(name); err != nil {
			return errMsg{err: err}
		}
		if err := store.Remove(name); err != nil {
			return errMsg{err: err}
		}

		// Remove directory
		var dirRemoved string
		var dirErr error
		if wsPath != "" {
			if err := os.RemoveAll(wsPath); err != nil {
				dirErr = err
			} else {
				dirRemoved = wsPath
			}
		}

		return forgetDoneMsg{name: name, dirRemoved: dirRemoved, dirErr: dirErr}
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

// switchToSelected resolves the workspace path, sets SwitchPath, and quits.
func (app App) switchToSelected() (tea.Model, tea.Cmd) {
	ws := selectedWorkspace(&app)
	if ws == nil {
		return app, nil
	}

	var switchPath string
	if ws.Name == "default" {
		switchPath = app.root
	} else if ws.Path != "" {
		switchPath = ws.Path
	} else {
		resolved, err := app.store.ResolveWorkspacePath(ws.Name)
		if err != nil {
			app.errMsg = fmt.Sprintf("Cannot resolve path: %v", err)
			return app, nil
		}
		switchPath = resolved
	}

	app.SwitchPath = switchPath
	return app, tea.Quit
}

// ensureCursorVisible adjusts scrollOffset so the cursor is within the visible viewport.
func (app *App) ensureCursorVisible() {
	visibleRows := app.listVisibleRows()

	if app.cursor < app.scrollOffset {
		app.scrollOffset = app.cursor
	}
	if app.cursor >= app.scrollOffset+visibleRows {
		app.scrollOffset = app.cursor - visibleRows + 1
	}
}

// Search-related methods

// performSearch executes fuzzy search on all workspaces
func (app *App) performSearch() {
	if app.searchQuery == "" {
		// No query - show all workspaces
		app.workspaces = app.allWorkspaces
		app.searchResults = nil
		app.searchActive = false
		return
	}

	// Perform fuzzy search
	app.searchResults = app.fuzzyMatcher.Search(app.searchQuery, app.allWorkspaces)
	app.workspaces = ExtractWorkspaces(app.searchResults)
	app.searchActive = true

	// Reset cursor to first result
	app.cursor = 0
}

// startSearch initiates search mode
func (app *App) startSearch() {
	app.inputMode = InputSearch
	app.searchQuery = ""
	app.textInput.SetValue("")
	app.textInput.Placeholder = "search workspaces..."
	app.textInput.Focus()
}

// clearSearch clears search and returns to showing all workspaces
func (app *App) clearSearch() {
	app.searchActive = false
	app.searchQuery = ""
	app.searchResults = nil
	app.workspaces = app.allWorkspaces
	app.inputMode = InputNone
	app.textInput.Blur()
	app.cursor = 0
}

// updateSearchQuery updates the search query and performs search
func (app *App) updateSearchQuery(query string) {
	app.searchQuery = query
	if query == "" {
		app.clearSearch()
		return
	}
	app.performSearch()
}

func (app *App) clearJJLogCache() {
	app.jjLogCache = make(map[string][]string)
	app.jjLogErr = make(map[string]string)
	app.jjLogLoading = make(map[string]bool)
}

func (app *App) loadJJLogForSelected(force bool) tea.Cmd {
	if app.ws == nil {
		return nil
	}
	ws := selectedWorkspace(app)
	if ws == nil {
		return nil
	}

	name := ws.Name
	if !force {
		if name == app.lastSelected {
			if _, ok := app.jjLogCache[name]; ok {
				return nil
			}
			if _, ok := app.jjLogErr[name]; ok {
				return nil
			}
		}
	}
	if app.jjLogLoading[name] {
		return nil
	}

	path, err := app.resolveWorkspacePath(ws)
	if err != nil {
		return func() tea.Msg {
			return jjLogErrMsg{name: name, err: err}
		}
	}

	app.lastSelected = name
	app.jjLogLoading[name] = true
	wsService := app.ws
	return func() tea.Msg {
		lines, err := wsService.LogGraph(path, 10)
		if err != nil {
			return jjLogErrMsg{name: name, err: err}
		}
		return jjLogLoadedMsg{name: name, lines: lines}
	}
}

func (app *App) resolveWorkspacePath(ws *display.WorkspaceInfo) (string, error) {
	if ws.Name == "default" {
		return app.root, nil
	}
	if ws.Path != "" {
		return ws.Path, nil
	}
	if app.store == nil {
		return "", fmt.Errorf("workspace path is not available")
	}
	return app.store.ResolveWorkspacePath(ws.Name)
}

const (
	minTriPaneWidth  = 120
	minTriPaneHeight = 20
)

func (app App) detectLayoutMode() LayoutMode {
	if app.width >= minTriPaneWidth && app.height >= minTriPaneHeight {
		return LayoutTriPane
	}
	return LayoutSingle
}

func (app App) listVisibleRows() int {
	if app.layoutMode == LayoutTriPane {
		bodyHeight := app.height - 5
		if app.inputMode == InputSearch {
			bodyHeight--
		}
		if bodyHeight < 12 {
			bodyHeight = 12
		}
		leftTopHeight := bodyHeight * 60 / 100
		if leftTopHeight < 7 {
			leftTopHeight = 7
		}
		rows := leftTopHeight - 4
		if rows < 1 {
			rows = 1
		}
		return rows
	}

	// Title(1) + search bar(0-2) + blank(1) + header(1) + blank(1) + status(1) + help(1)
	overhead := 6
	if app.inputMode == InputSearch {
		overhead += 2
	}
	rows := app.height - overhead
	if rows < 1 {
		rows = 1
	}
	return rows
}

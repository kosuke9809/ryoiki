package tui

import "github.com/kosuke9809/ryoiki/internal/display"

// workspacesLoadedMsg is sent when workspace data has been loaded.
type workspacesLoadedMsg struct {
	workspaces []display.WorkspaceInfo
}

// errMsg is sent when an error occurs.
type errMsg struct {
	err error
}

// forgetDoneMsg is sent when a workspace forget operation completes.
type forgetDoneMsg struct {
	name       string
	dirRemoved string // path that was removed (empty if not removed)
	dirErr     error  // directory removal error (nil if success)
}

// describeDoneMsg is sent when a describe (set purpose) operation completes.
type describeDoneMsg struct {
	name    string
	purpose string
}

// addDoneMsg is sent when a workspace add operation completes.
type addDoneMsg struct {
	name string
}

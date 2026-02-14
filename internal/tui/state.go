package tui

// ViewMode represents the current view in the TUI.
type ViewMode int

const (
	ViewList   ViewMode = iota
	ViewDetail
)

// InputMode represents the current input state in the TUI.
type InputMode int

const (
	InputNone          InputMode = iota
	InputDescribe                // Entering purpose text
	InputConfirmForget           // Confirming workspace forget
	InputAddPath                 // Entering path for new workspace
	InputAddName                 // Entering name for new workspace
	InputAddPurpose              // Entering purpose for new workspace
)

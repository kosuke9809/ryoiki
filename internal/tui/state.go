package tui

// ViewMode represents the current view in the TUI.
type ViewMode int

const (
	ViewList ViewMode = iota
	ViewDetail
	ViewHelp
)

// LayoutMode controls how the main TUI content is arranged.
type LayoutMode int

const (
	LayoutSingle LayoutMode = iota
	LayoutTriPane
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
	InputSearch                  // Searching/filtering workspaces
	InputAddNameFirst            // Entering name for name-first add workflow
)

// SearchType represents what field to search in
type SearchType int

const (
	SearchByName SearchType = iota
	SearchByPurpose
	SearchByDescription
	SearchByAll // Search in all fields
)

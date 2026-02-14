package display

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

// WorkspaceInfo combines jj workspace data with ryoiki metadata for display
type WorkspaceInfo struct {
	Name        string
	Path        string
	ChangeID    string
	CommitID    string
	Description string
	AuthorName  string
	AuthorEmail string
	Purpose     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	IsCurrent   bool
}

// PrintListTable prints workspace list in table format
func PrintListTable(w io.Writer, workspaces []WorkspaceInfo) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "  NAME\tPATH\tPURPOSE\n")
	for _, ws := range workspaces {
		marker := " "
		if ws.IsCurrent {
			marker = "*"
		}
		fmt.Fprintf(tw, "%s %s\t%s\t%s\n", marker, ws.Name, ws.Path, ws.Purpose)
	}
	tw.Flush()
}

// PrintStatusTable prints workspace status in table format with commit info
func PrintStatusTable(w io.Writer, workspaces []WorkspaceInfo) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "  NAME\tCHANGE\tCOMMIT\tDESCRIPTION\tPURPOSE\n")
	for _, ws := range workspaces {
		marker := " "
		if ws.IsCurrent {
			marker = "*"
		}
		changeID := shortID(ws.ChangeID, 8)
		commitID := shortID(ws.CommitID, 8)
		desc := ws.Description
		if desc == "" {
			desc = "(empty)"
		}
		if len(desc) > 30 {
			desc = desc[:27] + "..."
		}
		fmt.Fprintf(tw, "%s %s\t%s\t%s\t%s\t%s\n", marker, ws.Name, changeID, commitID, desc, ws.Purpose)
	}
	tw.Flush()
}

// PrintShowDetail prints detailed info for a single workspace
func PrintShowDetail(w io.Writer, ws WorkspaceInfo) {
	fmt.Fprintf(w, "Name:        %s\n", ws.Name)
	if ws.Path != "" {
		fmt.Fprintf(w, "Path:        %s\n", ws.Path)
	}
	fmt.Fprintf(w, "Change ID:   %s\n", ws.ChangeID)
	fmt.Fprintf(w, "Commit ID:   %s\n", ws.CommitID)
	desc := ws.Description
	if desc == "" {
		desc = "(empty)"
	}
	fmt.Fprintf(w, "Description: %s\n", desc)
	fmt.Fprintf(w, "Author:      %s <%s>\n", ws.AuthorName, ws.AuthorEmail)
	if ws.Purpose != "" {
		fmt.Fprintf(w, "Purpose:     %s\n", ws.Purpose)
	}
	if !ws.CreatedAt.IsZero() {
		fmt.Fprintf(w, "Created:     %s\n", ws.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	if !ws.UpdatedAt.IsZero() {
		fmt.Fprintf(w, "Updated:     %s\n", ws.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
	if ws.IsCurrent {
		fmt.Fprintf(w, "Current:     yes\n")
	}
}

func shortID(id string, length int) string {
	if len(id) > length {
		return id[:length]
	}
	return id
}

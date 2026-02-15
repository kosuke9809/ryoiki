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
	if _, err := fmt.Fprintf(tw, "  NAME\tPATH\tPURPOSE\n"); err != nil {
		return
	}
	for _, ws := range workspaces {
		marker := " "
		if ws.IsCurrent {
			marker = "*"
		}
		if _, err := fmt.Fprintf(tw, "%s %s\t%s\t%s\n", marker, ws.Name, ws.Path, ws.Purpose); err != nil {
			return
		}
	}
	if err := tw.Flush(); err != nil {
		return
	}
}

// PrintStatusTable prints workspace status in table format with commit info
func PrintStatusTable(w io.Writer, workspaces []WorkspaceInfo) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintf(tw, "  NAME\tCHANGE\tCOMMIT\tDESCRIPTION\tPURPOSE\n"); err != nil {
		return
	}
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
		if _, err := fmt.Fprintf(tw, "%s %s\t%s\t%s\t%s\t%s\n", marker, ws.Name, changeID, commitID, desc, ws.Purpose); err != nil {
			return
		}
	}
	if err := tw.Flush(); err != nil {
		return
	}
}

// PrintShowDetail prints detailed info for a single workspace
func PrintShowDetail(w io.Writer, ws WorkspaceInfo) {
	if _, err := fmt.Fprintf(w, "Name:        %s\n", ws.Name); err != nil {
		return
	}
	if ws.Path != "" {
		if _, err := fmt.Fprintf(w, "Path:        %s\n", ws.Path); err != nil {
			return
		}
	}
	if _, err := fmt.Fprintf(w, "Change ID:   %s\n", ws.ChangeID); err != nil {
		return
	}
	if _, err := fmt.Fprintf(w, "Commit ID:   %s\n", ws.CommitID); err != nil {
		return
	}
	desc := ws.Description
	if desc == "" {
		desc = "(empty)"
	}
	if _, err := fmt.Fprintf(w, "Description: %s\n", desc); err != nil {
		return
	}
	if _, err := fmt.Fprintf(w, "Author:      %s <%s>\n", ws.AuthorName, ws.AuthorEmail); err != nil {
		return
	}
	if ws.Purpose != "" {
		if _, err := fmt.Fprintf(w, "Purpose:     %s\n", ws.Purpose); err != nil {
			return
		}
	}
	if !ws.CreatedAt.IsZero() {
		if _, err := fmt.Fprintf(w, "Created:     %s\n", ws.CreatedAt.Format("2006-01-02 15:04:05")); err != nil {
			return
		}
	}
	if !ws.UpdatedAt.IsZero() {
		if _, err := fmt.Fprintf(w, "Updated:     %s\n", ws.UpdatedAt.Format("2006-01-02 15:04:05")); err != nil {
			return
		}
	}
	if ws.IsCurrent {
		if _, err := fmt.Fprintf(w, "Current:     yes\n"); err != nil {
			return
		}
	}
}

func shortID(id string, length int) string {
	if len(id) > length {
		return id[:length]
	}
	return id
}

package display

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestPrintListTable(t *testing.T) {
	t.Run("single workspace", func(t *testing.T) {
		var buf bytes.Buffer
		workspaces := []WorkspaceInfo{
			{Name: "default", Path: "/home/user/project", Purpose: "main dev"},
		}
		PrintListTable(&buf, workspaces)
		out := buf.String()

		if !strings.Contains(out, "NAME") {
			t.Error("expected header NAME")
		}
		if !strings.Contains(out, "PATH") {
			t.Error("expected header PATH")
		}
		if !strings.Contains(out, "PURPOSE") {
			t.Error("expected header PURPOSE")
		}
		if !strings.Contains(out, "default") {
			t.Error("expected workspace name 'default'")
		}
		if !strings.Contains(out, "/home/user/project") {
			t.Error("expected workspace path")
		}
		if !strings.Contains(out, "main dev") {
			t.Error("expected workspace purpose")
		}
	})

	t.Run("multiple workspaces", func(t *testing.T) {
		var buf bytes.Buffer
		workspaces := []WorkspaceInfo{
			{Name: "default", Path: "/home/user/project"},
			{Name: "feature", Path: "/home/user/project-feature"},
			{Name: "bugfix", Path: "/home/user/project-bugfix"},
		}
		PrintListTable(&buf, workspaces)
		out := buf.String()

		if !strings.Contains(out, "default") {
			t.Error("expected 'default'")
		}
		if !strings.Contains(out, "feature") {
			t.Error("expected 'feature'")
		}
		if !strings.Contains(out, "bugfix") {
			t.Error("expected 'bugfix'")
		}

		// Header + 3 data rows = 4 non-empty lines
		lines := nonEmptyLines(out)
		if len(lines) != 4 {
			t.Errorf("expected 4 lines (header + 3 rows), got %d", len(lines))
		}
	})

	t.Run("current workspace marker", func(t *testing.T) {
		var buf bytes.Buffer
		workspaces := []WorkspaceInfo{
			{Name: "default", Path: "/path/a", IsCurrent: false},
			{Name: "feature", Path: "/path/b", IsCurrent: true},
		}
		PrintListTable(&buf, workspaces)
		lines := nonEmptyLines(buf.String())

		// The current workspace line should contain "*"
		found := false
		for _, line := range lines {
			if strings.Contains(line, "feature") && strings.Contains(line, "*") {
				found = true
			}
		}
		if !found {
			t.Error("expected '*' marker on current workspace line")
		}

		// Non-current workspace line should NOT contain "*"
		for _, line := range lines {
			if strings.Contains(line, "default") && strings.Contains(line, "*") {
				t.Error("non-current workspace should not have '*' marker")
			}
		}
	})

	t.Run("empty list", func(t *testing.T) {
		var buf bytes.Buffer
		PrintListTable(&buf, []WorkspaceInfo{})
		lines := nonEmptyLines(buf.String())

		if len(lines) != 1 {
			t.Errorf("expected 1 header line, got %d", len(lines))
		}
		if !strings.Contains(lines[0], "NAME") {
			t.Error("expected header line")
		}
	})
}

func TestPrintStatusTable(t *testing.T) {
	t.Run("basic output", func(t *testing.T) {
		var buf bytes.Buffer
		workspaces := []WorkspaceInfo{
			{
				Name:        "default",
				ChangeID:    "abcdef1234567890",
				CommitID:    "1234567890abcdef",
				Description: "initial commit",
				Purpose:     "dev",
			},
		}
		PrintStatusTable(&buf, workspaces)
		out := buf.String()

		for _, col := range []string{"NAME", "CHANGE", "COMMIT", "DESCRIPTION", "PURPOSE"} {
			if !strings.Contains(out, col) {
				t.Errorf("expected header %q", col)
			}
		}
		if !strings.Contains(out, "default") {
			t.Error("expected workspace name")
		}
		if !strings.Contains(out, "initial commit") {
			t.Error("expected description")
		}
		if !strings.Contains(out, "dev") {
			t.Error("expected purpose")
		}
	})

	t.Run("ID truncation", func(t *testing.T) {
		var buf bytes.Buffer
		workspaces := []WorkspaceInfo{
			{
				Name:        "ws",
				ChangeID:    "abcdefghijklmnop",
				CommitID:    "1234567890abcdef",
				Description: "test",
			},
		}
		PrintStatusTable(&buf, workspaces)
		out := buf.String()

		if !strings.Contains(out, "abcdefgh") {
			t.Error("expected truncated change ID 'abcdefgh'")
		}
		if !strings.Contains(out, "12345678") {
			t.Error("expected truncated commit ID '12345678'")
		}
		// Full IDs should not appear
		if strings.Contains(out, "abcdefghijklmnop") {
			t.Error("full change ID should be truncated")
		}
		if strings.Contains(out, "1234567890abcdef") {
			t.Error("full commit ID should be truncated")
		}
	})

	t.Run("empty description", func(t *testing.T) {
		var buf bytes.Buffer
		workspaces := []WorkspaceInfo{
			{Name: "ws", Description: ""},
		}
		PrintStatusTable(&buf, workspaces)
		out := buf.String()

		if !strings.Contains(out, "(empty)") {
			t.Error("expected '(empty)' for empty description")
		}
	})

	t.Run("long description truncated", func(t *testing.T) {
		var buf bytes.Buffer
		longDesc := "this is a very long description that exceeds thirty characters"
		workspaces := []WorkspaceInfo{
			{Name: "ws", Description: longDesc},
		}
		PrintStatusTable(&buf, workspaces)
		out := buf.String()

		if !strings.Contains(out, "...") {
			t.Error("expected '...' for truncated description")
		}
		if strings.Contains(out, longDesc) {
			t.Error("full description should be truncated")
		}
		// First 27 chars should be present
		if !strings.Contains(out, longDesc[:27]) {
			t.Error("expected first 27 characters of description")
		}
	})

	t.Run("current marker", func(t *testing.T) {
		var buf bytes.Buffer
		workspaces := []WorkspaceInfo{
			{Name: "ws1", IsCurrent: true, Description: "d"},
		}
		PrintStatusTable(&buf, workspaces)
		lines := nonEmptyLines(buf.String())

		found := false
		for _, line := range lines {
			if strings.Contains(line, "ws1") && strings.Contains(line, "*") {
				found = true
			}
		}
		if !found {
			t.Error("expected '*' marker on current workspace")
		}
	})
}

func TestPrintShowDetail(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		var buf bytes.Buffer
		created := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
		updated := time.Date(2025, 1, 16, 14, 0, 0, 0, time.UTC)
		ws := WorkspaceInfo{
			Name:        "feature",
			Path:        "/home/user/project-feature",
			ChangeID:    "abcdef12",
			CommitID:    "12345678",
			Description: "add feature",
			AuthorName:  "Alice",
			AuthorEmail: "alice@example.com",
			Purpose:     "new feature",
			CreatedAt:   created,
			UpdatedAt:   updated,
			IsCurrent:   true,
		}
		PrintShowDetail(&buf, ws)
		out := buf.String()

		checks := map[string]string{
			"Name:":        "feature",
			"Path:":        "/home/user/project-feature",
			"Change ID:":   "abcdef12",
			"Commit ID:":   "12345678",
			"Description:": "add feature",
			"Author:":      "Alice <alice@example.com>",
			"Purpose:":     "new feature",
			"Created:":     "2025-01-15 10:30:00",
			"Updated:":     "2025-01-16 14:00:00",
			"Current:":     "yes",
		}
		for label, value := range checks {
			if !strings.Contains(out, label) {
				t.Errorf("expected label %q", label)
			}
			if !strings.Contains(out, value) {
				t.Errorf("expected value %q for label %q", value, label)
			}
		}
	})

	t.Run("empty path omitted", func(t *testing.T) {
		var buf bytes.Buffer
		ws := WorkspaceInfo{Name: "ws", Path: ""}
		PrintShowDetail(&buf, ws)
		if strings.Contains(buf.String(), "Path:") {
			t.Error("Path line should be omitted when path is empty")
		}
	})

	t.Run("empty purpose omitted", func(t *testing.T) {
		var buf bytes.Buffer
		ws := WorkspaceInfo{Name: "ws", Purpose: ""}
		PrintShowDetail(&buf, ws)
		if strings.Contains(buf.String(), "Purpose:") {
			t.Error("Purpose line should be omitted when purpose is empty")
		}
	})

	t.Run("empty description shows placeholder", func(t *testing.T) {
		var buf bytes.Buffer
		ws := WorkspaceInfo{Name: "ws", Description: ""}
		PrintShowDetail(&buf, ws)
		out := buf.String()

		if !strings.Contains(out, "(empty)") {
			t.Error("expected '(empty)' for empty description")
		}
	})

	t.Run("zero timestamps omitted", func(t *testing.T) {
		var buf bytes.Buffer
		ws := WorkspaceInfo{Name: "ws"}
		PrintShowDetail(&buf, ws)
		out := buf.String()

		if strings.Contains(out, "Created:") {
			t.Error("Created line should be omitted for zero time")
		}
		if strings.Contains(out, "Updated:") {
			t.Error("Updated line should be omitted for zero time")
		}
	})

	t.Run("non-current workspace omits current line", func(t *testing.T) {
		var buf bytes.Buffer
		ws := WorkspaceInfo{Name: "ws", IsCurrent: false}
		PrintShowDetail(&buf, ws)
		if strings.Contains(buf.String(), "Current:") {
			t.Error("Current line should be omitted for non-current workspace")
		}
	})
}

func TestShortID(t *testing.T) {
	tests := []struct {
		name   string
		id     string
		length int
		want   string
	}{
		{"long ID truncated", "abcdefghijklmnop", 8, "abcdefgh"},
		{"short ID unchanged", "abc", 8, "abc"},
		{"exact length unchanged", "abcdefgh", 8, "abcdefgh"},
		{"empty string", "", 8, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shortID(tt.id, tt.length)
			if got != tt.want {
				t.Errorf("shortID(%q, %d) = %q, want %q", tt.id, tt.length, got, tt.want)
			}
		})
	}
}

func TestPrintJSON(t *testing.T) {
	t.Run("normal output", func(t *testing.T) {
		var buf bytes.Buffer
		data := map[string]string{"name": "default", "purpose": "dev"}
		err := PrintJSON(&buf, data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out := buf.String()

		if !strings.Contains(out, `"name": "default"`) {
			t.Error("expected name field in JSON output")
		}
		if !strings.Contains(out, `"purpose": "dev"`) {
			t.Error("expected purpose field in JSON output")
		}
		// Should be indented
		if !strings.Contains(out, "  ") {
			t.Error("expected indented JSON output")
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		var buf bytes.Buffer
		err := PrintJSON(&buf, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out := strings.TrimSpace(buf.String())
		if out != "[]" {
			t.Errorf("expected '[]', got %q", out)
		}
	})

	t.Run("struct output", func(t *testing.T) {
		var buf bytes.Buffer
		type item struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}
		err := PrintJSON(&buf, item{Name: "test", Value: 42})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, `"name": "test"`) {
			t.Error("expected name field")
		}
		if !strings.Contains(out, `"value": 42`) {
			t.Error("expected value field")
		}
	})
}

// nonEmptyLines returns lines that are not blank after trimming.
func nonEmptyLines(s string) []string {
	var result []string
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}
	return result
}

package display

import (
	"testing"
	"time"

	"github.com/kosuke9809/ryoiki/internal/config"
	"github.com/kosuke9809/ryoiki/internal/jj"
)

func TestBuildWorkspaceInfos(t *testing.T) {
	t.Run("merges jj data with metadata", func(t *testing.T) {
		workspaces := []jj.Workspace{
			{
				Name: "default",
				Target: jj.Target{
					ChangeID:    "change123",
					CommitID:    "commit456",
					Description: "initial",
					Author:      jj.User{Name: "Alice", Email: "alice@example.com"},
				},
			},
			{
				Name: "feature",
				Target: jj.Target{
					ChangeID:    "change789",
					CommitID:    "commitabc",
					Description: "add feature",
					Author:      jj.User{Name: "Bob", Email: "bob@example.com"},
				},
			},
		}

		now := time.Now()
		metadataMap := map[string]config.WorkspaceMetadata{
			"feature": {
				Path:      "/home/user/project-feature",
				Purpose:   "new feature",
				CreatedAt: now,
				UpdatedAt: now,
			},
		}

		infos := BuildWorkspaceInfos(workspaces, metadataMap, "/home/user/project", "/home/user/project")

		if len(infos) != 2 {
			t.Fatalf("expected 2 infos, got %d", len(infos))
		}

		// default workspace
		def := infos[0]
		if def.Name != "default" {
			t.Errorf("expected name 'default', got %q", def.Name)
		}
		if def.Path != "/home/user/project" {
			t.Errorf("expected default path to be repo root, got %q", def.Path)
		}
		if !def.IsCurrent {
			t.Error("expected default to be current (path == currentRoot)")
		}
		if def.ChangeID != "change123" {
			t.Errorf("expected ChangeID 'change123', got %q", def.ChangeID)
		}
		if def.Description != "initial" {
			t.Errorf("expected Description 'initial', got %q", def.Description)
		}

		// feature workspace
		feat := infos[1]
		if feat.Name != "feature" {
			t.Errorf("expected name 'feature', got %q", feat.Name)
		}
		if feat.Path != "/home/user/project-feature" {
			t.Errorf("expected feature path from metadata, got %q", feat.Path)
		}
		if feat.Purpose != "new feature" {
			t.Errorf("expected purpose 'new feature', got %q", feat.Purpose)
		}
		if feat.IsCurrent {
			t.Error("expected feature to not be current")
		}
		if feat.AuthorName != "Bob" {
			t.Errorf("expected author 'Bob', got %q", feat.AuthorName)
		}
	})

	t.Run("default workspace uses repo root when no metadata path", func(t *testing.T) {
		workspaces := []jj.Workspace{
			{Name: "default", Target: jj.Target{}},
		}
		metadataMap := map[string]config.WorkspaceMetadata{}

		infos := BuildWorkspaceInfos(workspaces, metadataMap, "/repo", "/other")

		if infos[0].Path != "/repo" {
			t.Errorf("expected default path '/repo', got %q", infos[0].Path)
		}
	})

	t.Run("default workspace keeps metadata path if set", func(t *testing.T) {
		workspaces := []jj.Workspace{
			{Name: "default", Target: jj.Target{}},
		}
		metadataMap := map[string]config.WorkspaceMetadata{
			"default": {Path: "/custom/path"},
		}

		infos := BuildWorkspaceInfos(workspaces, metadataMap, "/repo", "/other")

		if infos[0].Path != "/custom/path" {
			t.Errorf("expected metadata path '/custom/path', got %q", infos[0].Path)
		}
	})

	t.Run("empty workspaces", func(t *testing.T) {
		infos := BuildWorkspaceInfos(nil, map[string]config.WorkspaceMetadata{}, "/repo", "/repo")
		if infos != nil {
			t.Errorf("expected nil, got %v", infos)
		}
	})

	t.Run("current workspace detection", func(t *testing.T) {
		workspaces := []jj.Workspace{
			{Name: "ws1", Target: jj.Target{}},
			{Name: "ws2", Target: jj.Target{}},
		}
		metadataMap := map[string]config.WorkspaceMetadata{
			"ws1": {Path: "/path/a"},
			"ws2": {Path: "/path/b"},
		}

		infos := BuildWorkspaceInfos(workspaces, metadataMap, "/repo", "/path/b")

		if infos[0].IsCurrent {
			t.Error("ws1 should not be current")
		}
		if !infos[1].IsCurrent {
			t.Error("ws2 should be current")
		}
	})
}

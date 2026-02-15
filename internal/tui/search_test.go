package tui

import (
	"reflect"
	"testing"

	"github.com/kosuke9809/ryoiki/internal/display"
)

func TestFuzzyMatcher_Search(t *testing.T) {
	matcher := NewFuzzyMatcher()

	workspaces := []display.WorkspaceInfo{
		{
			Name:        "feature-auth",
			Purpose:     "User authentication implementation",
			Description: "OAuth2 and JWT tokens",
			Path:        "/project/auth",
			AuthorName:  "developer",
		},
		{
			Name:        "bugfix-login",
			Purpose:     "Fix login issues",
			Description: "Login form validation",
			Path:        "/project/login",
			AuthorName:  "developer",
		},
		{
			Name:        "default",
			Purpose:     "Main development",
			Description: "Main branch work",
			Path:        "/project",
			AuthorName:  "developer",
		},
	}

	tests := []struct {
		name          string
		query         string
		expectedCount int
		expectedFirst string // Name of first result
	}{
		{
			name:          "empty query returns all",
			query:         "",
			expectedCount: 3,
			expectedFirst: "feature-auth", // First in original order
		},
		{
			name:          "exact name match",
			query:         "feature-auth",
			expectedCount: 1,
			expectedFirst: "feature-auth",
		},
		{
			name:          "fuzzy name match",
			query:         "feat",
			expectedCount: 1,
			expectedFirst: "feature-auth",
		},
		{
			name:          "purpose match",
			query:         "authentication",
			expectedCount: 1,
			expectedFirst: "feature-auth",
		},
		{
			name:          "fuzzy purpose match",
			query:         "auth",
			expectedCount: 1,
			expectedFirst: "feature-auth",
		},
		{
			name:          "partial match multiple fields",
			query:         "login",
			expectedCount: 1,
			expectedFirst: "bugfix-login",
		},
		{
			name:          "no matches",
			query:         "nonexistent",
			expectedCount: 0,
			expectedFirst: "",
		},
		{
			name:          "case insensitive",
			query:         "FEATURE",
			expectedCount: 1,
			expectedFirst: "feature-auth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := matcher.Search(tt.query, workspaces)

			if len(results) != tt.expectedCount {
				t.Errorf("expected %d results, got %d", tt.expectedCount, len(results))
				return
			}

			if tt.expectedCount > 0 && results[0].Workspace.Name != tt.expectedFirst {
				t.Errorf("expected first result to be %q, got %q", tt.expectedFirst, results[0].Workspace.Name)
			}

			// Verify results are sorted by score (descending)
			for i := 1; i < len(results); i++ {
				if results[i].Score > results[i-1].Score {
					t.Errorf("results not sorted by score: result[%d].Score=%d > result[%d].Score=%d",
						i, results[i].Score, i-1, results[i-1].Score)
				}
			}
		})
	}
}

func TestFuzzyMatcher_ScoreCalculation(t *testing.T) {
	matcher := NewFuzzyMatcher()

	workspace := display.WorkspaceInfo{
		Name:        "feature-auth",
		Purpose:     "Authentication feature",
		Description: "OAuth2 implementation",
	}

	workspaces := []display.WorkspaceInfo{workspace}

	// Test that exact matches score higher than fuzzy matches
	exactResults := matcher.Search("feature-auth", workspaces)
	fuzzyResults := matcher.Search("feat", workspaces)

	if len(exactResults) == 0 || len(fuzzyResults) == 0 {
		t.Fatal("expected both exact and fuzzy results")
	}

	if exactResults[0].Score <= fuzzyResults[0].Score {
		t.Errorf("exact match score (%d) should be higher than fuzzy match score (%d)",
			exactResults[0].Score, fuzzyResults[0].Score)
	}
}

func TestExtractWorkspaces(t *testing.T) {
	searchResults := []SearchResult{
		{
			Workspace: display.WorkspaceInfo{Name: "first", Purpose: "First workspace"},
			Score:     100,
		},
		{
			Workspace: display.WorkspaceInfo{Name: "second", Purpose: "Second workspace"},
			Score:     80,
		},
	}

	workspaces := ExtractWorkspaces(searchResults)

	if len(workspaces) != 2 {
		t.Errorf("expected 2 workspaces, got %d", len(workspaces))
	}

	if workspaces[0].Name != "first" {
		t.Errorf("expected first workspace name to be 'first', got %q", workspaces[0].Name)
	}

	if workspaces[1].Name != "second" {
		t.Errorf("expected second workspace name to be 'second', got %q", workspaces[1].Name)
	}
}

func TestApp_SearchIntegration(t *testing.T) {
	// Create a test app using the NewApp constructor
	app := NewApp(nil, nil, "")
	app.allWorkspaces = []display.WorkspaceInfo{
		{Name: "feature-auth", Purpose: "Authentication"},
		{Name: "bugfix-login", Purpose: "Login fixes"},
		{Name: "default", Purpose: "Main development"},
	}
	app.workspaces = app.allWorkspaces

	// Test starting search
	app.startSearch()
	if app.inputMode != InputSearch {
		t.Errorf("expected InputSearch mode, got %v", app.inputMode)
	}

	// Test updating search query
	app.updateSearchQuery("auth")
	if !app.searchActive {
		t.Error("expected search to be active")
	}

	if len(app.workspaces) != 1 {
		t.Errorf("expected 1 filtered workspace, got %d", len(app.workspaces))
	}

	if app.workspaces[0].Name != "feature-auth" {
		t.Errorf("expected filtered workspace to be 'feature-auth', got %q", app.workspaces[0].Name)
	}

	// Test clearing search
	app.clearSearch()
	if app.searchActive {
		t.Error("expected search to be inactive")
	}

	if len(app.workspaces) != len(app.allWorkspaces) {
		t.Errorf("expected all workspaces to be shown after clearing search, got %d", len(app.workspaces))
	}

	if app.inputMode != InputNone {
		t.Errorf("expected InputNone mode after clearing search, got %v", app.inputMode)
	}
}

func TestMatchPosition_Merge(t *testing.T) {
	matcher := &FuzzyMatcher{}

	tests := []struct {
		name     string
		matches  []MatchPosition
		expected []MatchPosition
	}{
		{
			name: "adjacent matches merged",
			matches: []MatchPosition{
				{Start: 0, End: 1},
				{Start: 1, End: 2},
				{Start: 2, End: 3},
			},
			expected: []MatchPosition{
				{Start: 0, End: 3},
			},
		},
		{
			name: "non-adjacent matches kept separate",
			matches: []MatchPosition{
				{Start: 0, End: 1},
				{Start: 3, End: 4},
			},
			expected: []MatchPosition{
				{Start: 0, End: 1},
				{Start: 3, End: 4},
			},
		},
		{
			name:     "empty matches",
			matches:  []MatchPosition{},
			expected: []MatchPosition{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.mergeAdjacentMatches(tt.matches)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}
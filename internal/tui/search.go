package tui

import (
	"sort"
	"strings"
	"unicode"

	"github.com/kosuke9809/ryoiki/internal/display"
)

// SearchResult represents a workspace that matches a search query with relevance score
type SearchResult struct {
	Workspace display.WorkspaceInfo
	Score     int
	Matches   []MatchPosition // Positions of matched characters for highlighting
}

// MatchPosition represents a range of characters that matched the search query
type MatchPosition struct {
	Start int
	End   int
	Field string // "name", "purpose", "description"
}

// FuzzyMatcher implements fuzzy string matching for workspace search
type FuzzyMatcher struct {
	caseSensitive bool
}

// NewFuzzyMatcher creates a new fuzzy matcher
func NewFuzzyMatcher() *FuzzyMatcher {
	return &FuzzyMatcher{
		caseSensitive: false,
	}
}

// Search performs fuzzy search on workspaces and returns sorted results
func (fm *FuzzyMatcher) Search(query string, workspaces []display.WorkspaceInfo) []SearchResult {
	if query == "" {
		// Return all workspaces with zero score when no query
		results := make([]SearchResult, len(workspaces))
		for i, ws := range workspaces {
			results[i] = SearchResult{
				Workspace: ws,
				Score:     0,
				Matches:   nil,
			}
		}
		return results
	}

	var results []SearchResult
	normalizedQuery := fm.normalizeString(query)

	for _, ws := range workspaces {
		if score, matches := fm.matchWorkspace(normalizedQuery, ws); score > 0 {
			results = append(results, SearchResult{
				Workspace: ws,
				Score:     score,
				Matches:   matches,
			})
		}
	}

	// Sort by score (descending), then by name (ascending) for tie-breaking
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].Workspace.Name < results[j].Workspace.Name
		}
		return results[i].Score > results[j].Score
	})

	return results
}

// matchWorkspace calculates match score and positions for a single workspace
func (fm *FuzzyMatcher) matchWorkspace(query string, ws display.WorkspaceInfo) (int, []MatchPosition) {
	var allMatches []MatchPosition
	totalScore := 0

	// Search in different fields with different weights
	fields := []struct {
		value  string
		field  string
		weight int
	}{
		{ws.Name, "name", 10},                    // Highest weight for name
		{ws.Purpose, "purpose", 7},               // High weight for purpose  
		{ws.Description, "description", 4},       // Medium weight for description
		{ws.Path, "path", 2},                     // Low weight for path
		{ws.AuthorName, "author", 1},             // Lowest weight for author
	}

	for _, field := range fields {
		if field.value == "" {
			continue
		}

		normalizedValue := fm.normalizeString(field.value)
		if score, matches := fm.fuzzyMatch(query, normalizedValue); score > 0 {
			// Apply field weight to score
			fieldScore := score * field.weight
			totalScore += fieldScore

			// Add field information to matches
			for _, match := range matches {
				match.Field = field.field
				allMatches = append(allMatches, match)
			}
		}
	}

	return totalScore, allMatches
}

// fuzzyMatch implements fuzzy matching algorithm similar to fzf
func (fm *FuzzyMatcher) fuzzyMatch(query, text string) (int, []MatchPosition) {
	if len(query) == 0 {
		return 0, nil
	}

	queryRunes := []rune(query)
	textRunes := []rune(text)
	
	if len(queryRunes) > len(textRunes) {
		return 0, nil
	}

	// Try to find all characters of query in text in order
	var matches []MatchPosition
	textPos := 0
	consecutiveMatches := 0
	lastMatchPos := -2
	
	for _, queryChar := range queryRunes {
		found := false
		startSearch := textPos
		
		// Look for the character in remaining text
		for i := startSearch; i < len(textRunes); i++ {
			if textRunes[i] == queryChar {
				// Found the character
				if i == lastMatchPos+1 {
					// Consecutive match - bonus points
					consecutiveMatches++
				} else {
					consecutiveMatches = 1
				}
				
				matches = append(matches, MatchPosition{
					Start: i,
					End:   i + 1,
				})
				
				textPos = i + 1
				lastMatchPos = i
				found = true
				break
			}
		}
		
		if !found {
			return 0, nil // Query character not found
		}
	}

	// Calculate score based on various factors
	score := fm.calculateScore(query, text, matches, consecutiveMatches)
	return score, fm.mergeAdjacentMatches(matches)
}

// calculateScore computes match score based on multiple factors
func (fm *FuzzyMatcher) calculateScore(query, text string, matches []MatchPosition, consecutiveMatches int) int {
	if len(matches) == 0 {
		return 0
	}

	baseScore := 100

	// Exact match bonus
	if strings.Contains(text, query) {
		baseScore += 200
	}

	// Prefix match bonus
	if strings.HasPrefix(text, query) {
		baseScore += 300
	}

	// Length ratio bonus (shorter strings with matches score higher)
	lengthRatio := float64(len(query)) / float64(len(text))
	baseScore += int(lengthRatio * 100)

	// Consecutive matches bonus
	baseScore += consecutiveMatches * 50

	// Early position bonus (matches near the beginning score higher)
	if len(matches) > 0 && matches[0].Start < 3 {
		baseScore += (3 - matches[0].Start) * 25
	}

	// Word boundary bonus
	wordBoundaryBonus := fm.calculateWordBoundaryBonus(query, text, matches)
	baseScore += wordBoundaryBonus

	return baseScore
}

// calculateWordBoundaryBonus gives bonus points for matches at word boundaries
func (fm *FuzzyMatcher) calculateWordBoundaryBonus(query, text string, matches []MatchPosition) int {
	bonus := 0
	textRunes := []rune(text)
	
	for _, match := range matches {
		if match.Start == 0 {
			// First character bonus
			bonus += 15
		} else if match.Start < len(textRunes) {
			prevChar := textRunes[match.Start-1]
			if !unicode.IsLetter(prevChar) && !unicode.IsDigit(prevChar) {
				// Word boundary bonus
				bonus += 10
			}
		}
	}
	
	return bonus
}

// mergeAdjacentMatches combines adjacent match positions for cleaner highlighting
func (fm *FuzzyMatcher) mergeAdjacentMatches(matches []MatchPosition) []MatchPosition {
	if len(matches) <= 1 {
		return matches
	}

	var merged []MatchPosition
	current := matches[0]

	for i := 1; i < len(matches); i++ {
		if matches[i].Start == current.End {
			// Adjacent match, extend current
			current.End = matches[i].End
		} else {
			// Non-adjacent, save current and start new
			merged = append(merged, current)
			current = matches[i]
		}
	}
	
	merged = append(merged, current)
	return merged
}

// normalizeString normalizes a string for case-insensitive matching
func (fm *FuzzyMatcher) normalizeString(s string) string {
	if fm.caseSensitive {
		return s
	}
	return strings.ToLower(s)
}

// ExtractWorkspaces extracts workspace list from search results
func ExtractWorkspaces(results []SearchResult) []display.WorkspaceInfo {
	workspaces := make([]display.WorkspaceInfo, len(results))
	for i, result := range results {
		workspaces[i] = result.Workspace
	}
	return workspaces
}
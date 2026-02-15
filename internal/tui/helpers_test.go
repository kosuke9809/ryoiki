package tui

import (
	"strings"
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"hello", 5, "hello"},
		{"hello", 3, "hel"},
		{"hi", 3, "hi"},
		{"", 5, ""},
		{"abcdef", 6, "abcdef"},
		{"abcdefg", 6, "abc..."},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}

func TestShortID(t *testing.T) {
	tests := []struct {
		input string
		n     int
		want  string
	}{
		{"abcdefghijklmnop", 8, "abcdefgh"},
		{"abcd", 8, "abcd"},
		{"", 8, ""},
		{"abcdefgh", 8, "abcdefgh"},
	}

	for _, tt := range tests {
		got := shortID(tt.input, tt.n)
		if got != tt.want {
			t.Errorf("shortID(%q, %d) = %q, want %q", tt.input, tt.n, got, tt.want)
		}
	}
}

func TestHighlightMatches_NoPositions(t *testing.T) {
	result := highlightMatches("hello", nil, 10)
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestHighlightMatches_WithPositions(t *testing.T) {
	positions := []MatchPosition{{Start: 0, End: 2, Field: "name"}}
	result := highlightMatches("hello", positions, 10)
	// Should contain the original chars
	if !strings.Contains(result, "h") || !strings.Contains(result, "llo") {
		t.Errorf("highlighted result should contain original characters: %q", result)
	}
	// Length should be at least as long as original (styling may or may not add ANSI codes)
	if len(result) < len("hello") {
		t.Errorf("highlighted result should be at least as long as original: %q", result)
	}
}

func TestHighlightMatches_Truncation(t *testing.T) {
	result := highlightMatches("very long string here", nil, 10)
	if result != "very lo..." {
		t.Errorf("expected truncated result, got %q", result)
	}
}

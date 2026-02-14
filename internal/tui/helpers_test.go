package tui

import "testing"

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

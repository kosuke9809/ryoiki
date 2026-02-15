package tui

import "strings"

// truncate shortens s to max characters, appending "..." if truncated.
func truncate(s string, max int) string {
	if max <= 3 {
		if len(s) > max {
			return s[:max]
		}
		return s
	}
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

// shortID returns the first n characters of id.
func shortID(id string, n int) string {
	if len(id) > n {
		return id[:n]
	}
	return id
}

// highlightMatches renders a string with match positions highlighted using highlightStyle.
// The string is truncated to maxLen visible characters before highlighting.
func highlightMatches(s string, positions []MatchPosition, maxLen int) string {
	if len(positions) == 0 {
		return truncate(s, maxLen)
	}

	// Truncate first (work with runes for proper unicode handling)
	runes := []rune(s)
	truncated := false
	if maxLen > 3 && len(runes) > maxLen {
		runes = runes[:maxLen-3]
		truncated = true
	}

	var b strings.Builder
	for i, r := range runes {
		highlighted := false
		for _, pos := range positions {
			if i >= pos.Start && i < pos.End {
				highlighted = true
				break
			}
		}
		if highlighted {
			b.WriteString(highlightStyle.Render(string(r)))
		} else {
			b.WriteRune(r)
		}
	}
	if truncated {
		b.WriteString("...")
	}
	return b.String()
}

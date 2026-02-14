package tui

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

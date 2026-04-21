package format

import (
	"strings"
)

// ProgressBar renders a text progress bar like [####----].
func ProgressBar(value, maxVal float64, width int) string {
	if maxVal == 0 {
		return "[" + strings.Repeat("-", width) + "]"
	}

	filled := int(value / maxVal * float64(width))
	if filled > width {
		filled = width
	}

	return "[" + strings.Repeat("#", filled) + strings.Repeat("-", width-filled) + "]"
}

// Truncate shortens a string to n runes, appending ellipsis if truncated.
func Truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}

	return string(runes[:n-1]) + "…"
}

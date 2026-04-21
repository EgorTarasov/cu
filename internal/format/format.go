package format

import (
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	hoursPerDay    = 24
	minutesPerHour = 60
)

// TimeLeft returns a human-readable duration until the given time.
func TimeLeft(t time.Time) string {
	d := time.Until(t)
	if d < 0 {
		return "OVERDUE"
	}

	days := int(d.Hours() / hoursPerDay)
	hours := int(math.Mod(d.Hours(), hoursPerDay))

	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, int(math.Mod(d.Minutes(), minutesPerHour)))
	}

	return fmt.Sprintf("%dm", int(d.Minutes()))
}

// StateLabel returns a short label for a task/deadline state.
func StateLabel(state string) string {
	switch state {
	case "backlog":
		return "TODO"
	case "inProgress":
		return "IN PROGRESS"
	case "submitted":
		return "SUBMITTED"
	case "evaluated":
		return "DONE"
	case "failed":
		return "FAILED"
	default:
		return strings.ToUpper(state)
	}
}

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

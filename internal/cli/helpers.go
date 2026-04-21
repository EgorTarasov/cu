package cli

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"cu-sync/internal/cu"
)

// mustClient creates an authenticated client or exits.
func mustClient() *cu.Client {
	client, err := cu.NewClientFromEnv()
	if err != nil {
		cookieRequiredError(err)
	}
	if err = client.ValidateCookie(); err != nil {
		fmt.Fprintf(os.Stderr, "Cookie expired: %v\nRun: cu login\n", err)
		os.Exit(1)
	}
	return client
}

// mustResolveCourse resolves a course query (ID or name substring) or exits.
func mustResolveCourse(ctx context.Context, client *cu.Client, query string) (int, string) {
	id, name, err := client.ResolveCourse(ctx, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	return id, name
}

// formatTimeLeft returns a human-readable duration until the given time.
func formatTimeLeft(t time.Time) string {
	d := time.Until(t)
	if d < 0 {
		return "OVERDUE"
	}
	days := int(d.Hours() / 24)
	hours := int(math.Mod(d.Hours(), 24))
	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, int(math.Mod(d.Minutes(), 60)))
	}
	return fmt.Sprintf("%dm", int(d.Minutes()))
}

// stateLabel returns a short colored label for a task/deadline state.
func stateLabel(state string) string {
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

// sanitizeFilename replaces invalid path characters.
func sanitizeFilename(name string) string {
	replacements := map[rune]rune{
		'/': '-', '\\': '-', ':': '-', '*': '-',
		'?': '-', '"': '-', '<': '-', '>': '-', '|': '-',
	}
	runes := []rune(name)
	for i, r := range runes {
		if replacement, ok := replacements[r]; ok {
			runes[i] = replacement
		}
	}
	return string(runes)
}

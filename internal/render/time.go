// Package render contains small writer-based helpers that both CLI (plain text)
// and MCP (markdown) consume. It is a transitional home for code that will
// move into internal/presenter/ when the project gains a third output format
// (e.g. JSON over HTTP).
package render

import (
	"fmt"
	"io"
	"time"

	"cu-sync/internal/model"
)

// TimeLine writes "<label>: <time>\n" to w if t != nil. No-op for nil.
// Time is formatted with model.DateTimeFormat for consistency across outputs.
func TimeLine(w io.Writer, label string, t *time.Time) {
	if t == nil {
		return
	}
	fmt.Fprintf(w, "%s: %s\n", label, t.Format(model.DateTimeFormat))
}

// MDTimeLine writes "**<label>:** <time>\n" to w if t != nil. No-op for nil.
func MDTimeLine(w io.Writer, label string, t *time.Time) {
	if t == nil {
		return
	}
	fmt.Fprintf(w, "**%s:** %s\n", label, t.Format(model.DateTimeFormat))
}

package recordings

import (
	"fmt"
	"strings"
)

// Format renders a single recording in the canonical bot-style block.
func Format(r Recording) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Пара: %s\n", r.Subject)
	if !r.Date.IsZero() {
		fmt.Fprintf(&b, "Дата: %s\n", r.Date.Format("02.01.2006"))
	}
	if r.StartTime != "" && r.EndTime != "" {
		fmt.Fprintf(&b, "Время: %s-%s\n", r.StartTime, r.EndTime)
	}
	if len(r.Links) > 0 {
		b.WriteString("Ссылки:\n")
		for _, l := range r.Links {
			if l.Size != "" {
				fmt.Fprintf(&b, "  - %s (размер: %s)\n", l.URL, l.Size)
			} else {
				fmt.Fprintf(&b, "  - %s\n", l.URL)
			}
		}
	}
	if r.PostURL != "" {
		fmt.Fprintf(&b, "Сообщение: %s\n", r.PostURL)
	}
	return b.String()
}

// FormatLine is a compact one-line summary used in pickers.
func FormatLine(r Recording) string {
	parts := []string{r.Subject}
	if !r.Date.IsZero() {
		parts = append(parts, r.Date.Format("02.01.2006"))
	}
	if r.StartTime != "" {
		parts = append(parts, r.StartTime+"-"+r.EndTime)
	}
	if len(r.Links) > 1 {
		parts = append(parts, fmt.Sprintf("%d links", len(r.Links)))
	}
	return strings.Join(parts, " · ")
}

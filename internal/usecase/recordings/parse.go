package recordings

import (
	"regexp"
	"strings"
	"time"
)

var (
	headerDateRe = regexp.MustCompile(`за\s+(\d{1,2}\.\d{1,2}\.\d{4})`)
	// itemStartRe matches "1. Пара: <subject>".
	itemStartRe = regexp.MustCompile(`(?m)^\s*\d+\.\s*Пара:\s*(.+)$`)
	timeRe      = regexp.MustCompile(`Время:\s*(\d{1,2}:\d{2})\s*-\s*(\d{1,2}:\d{2})`)
	// linkRe matches "- <URL> (размер: ...)".
	linkRe = regexp.MustCompile(`-\s*(https?://\S+?)(?:\s*\(размер:\s*([^)]+)\))?\s*$`)

	headerDateGroups = 2
	timeGroups       = 3
)

// Parse extracts zero or more recordings from a post message.
// `posted` is the create time of the post and is used as a fallback when
// the per-lesson date in the header is missing.
func Parse(message string, posted time.Time) []Recording {
	if !strings.Contains(message, RecordingHeader) {
		return nil
	}

	var date time.Time
	if m := headerDateRe.FindStringSubmatch(message); len(m) == headerDateGroups {
		if d, err := time.Parse("2.1.2006", m[1]); err == nil {
			date = d
		}
	}

	// Find all item start indices, then slice the message between them.
	idxs := itemStartRe.FindAllStringSubmatchIndex(message, -1)
	if len(idxs) == 0 {
		return nil
	}

	out := make([]Recording, 0, len(idxs))
	for i, m := range idxs {
		subject := strings.TrimSpace(message[m[2]:m[3]])
		start := m[0]
		end := len(message)
		if i+1 < len(idxs) {
			end = idxs[i+1][0]
		}
		block := message[start:end]

		rec := Recording{
			PostedAt: posted,
			Date:     date,
			Subject:  subject,
		}
		if tm := timeRe.FindStringSubmatch(block); len(tm) == timeGroups {
			rec.StartTime, rec.EndTime = tm[1], tm[2]
		}
		for _, lm := range linkRe.FindAllStringSubmatch(block, -1) {
			rec.Links = append(rec.Links, RecordingLink{
				URL:  strings.TrimRight(lm[1], "."),
				Size: strings.TrimSpace(lm[2]),
			})
		}
		// Skip items with no links — they're not actionable.
		if len(rec.Links) == 0 {
			continue
		}
		out = append(out, rec)
	}
	return out
}

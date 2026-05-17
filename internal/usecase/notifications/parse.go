package notifications

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// lmsURLRe matches the longread URL pattern emitted by the bot.
	lmsURLRe = regexp.MustCompile(`https?://[^\s)]*/learn/courses/view/[^/]+/(\d+)/themes/(\d+)/longreads/(\d+)`)

	// markdownLinkRe captures `[title](url)` pairs.
	markdownLinkRe = regexp.MustCompile(`\[([^\]]+)\]\((https?://[^)]+)\)`)
)

const (
	headerNewTask   = "🎓 Новые задачи"
	headerGraded    = "🎓 Задача оценена"
	headerRecording = "🎓 Записи занятий"

	lmsURLGroups       = 4
	markdownLinkGroups = 3
	excerptMaxLen      = 200
)

// Parse extracts a Notification from a single post. Returns nil for posts
// that don't reference any LMS resource we care about.
func Parse(postID, message string, postedAt time.Time) *Notification {
	kind := classify(message)
	if kind == KindOther {
		// Even non-recognized posts: only keep if they contain an LMS link.
		if !lmsURLRe.MatchString(message) {
			return nil
		}
	}

	n := &Notification{
		PostID:   postID,
		PostedAt: postedAt,
		Kind:     kind,
		Excerpt:  firstNonHeaderLine(message),
	}

	if m := lmsURLRe.FindStringSubmatch(message); len(m) == lmsURLGroups {
		n.LMSURL = m[0]
		n.CourseID, _ = strconv.Atoi(m[1])
		n.ThemeID, _ = strconv.Atoi(m[2])
		n.LongreadID, _ = strconv.Atoi(m[3])
	}

	if lm := markdownLinkRe.FindStringSubmatch(message); len(lm) == markdownLinkGroups {
		n.Title = strings.TrimSpace(lm[1])
	}
	if n.Title == "" {
		n.Title = n.Excerpt
	}

	return n
}

func classify(message string) string {
	switch {
	case strings.Contains(message, headerNewTask):
		return KindNewTask
	case strings.Contains(message, headerGraded):
		return KindGraded
	case strings.Contains(message, headerRecording):
		return KindRecording
	default:
		return KindOther
	}
}

func firstNonHeaderLine(message string) string {
	for _, line := range strings.Split(message, "\n") {
		l := strings.TrimSpace(line)
		if l == "" {
			continue
		}
		// Skip the bot header (bold/emoji-only line).
		if strings.HasPrefix(l, "**") && strings.HasSuffix(l, "**") {
			continue
		}
		if len(l) > excerptMaxLen {
			l = l[:excerptMaxLen] + "…"
		}
		return l
	}
	return ""
}

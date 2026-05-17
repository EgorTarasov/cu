package recordings

import "time"

// RecordingLink is a single playback/download URL.
type RecordingLink struct {
	URL  string
	Size string // raw size string from the bot, e.g. "24,9 МБ"; empty if missing
}

// Recording is one extracted lesson recording.
type Recording struct {
	PostID    string
	PostURL   string    // mattermost permalink to the source post, optional
	PostedAt  time.Time // post create time
	Date      time.Time // lesson date (day precision); zero if not parsed
	Subject   string    // "Пара:" value
	StartTime string    // "HH:MM" or empty
	EndTime   string    // "HH:MM" or empty
	Links     []RecordingLink
}

// RecordingHeader is the marker line the bot uses for recording posts.
const RecordingHeader = "🎓 Записи занятий"

// sortKey is used to order recordings newest-first; prefers Date when set.
func (r Recording) sortKey() time.Time {
	if !r.Date.IsZero() {
		return r.Date
	}
	return r.PostedAt
}

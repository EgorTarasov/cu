package timeclient

// User is a trimmed Mattermost user.
type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Nickname  string `json:"nickname"`
	IsBot     bool   `json:"is_bot"`
}

// Team is a trimmed Mattermost team.
type Team struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

// Channel represents a Mattermost channel. For direct messages the `name`
// field has the form "<userID1>__<userID2>" sorted lexicographically.
type Channel struct {
	ID          string `json:"id"`
	TeamID      string `json:"team_id"`
	Type        string `json:"type"` // "O" public, "P" private, "D" direct, "G" group
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Purpose     string `json:"purpose"`
	LastPostAt  int64  `json:"last_post_at"` // unix ms
	// TotalMsgCount minus ChannelMember.MsgCount gives the unread count.
	TotalMsgCount int64 `json:"total_msg_count"`
	DeleteAt      int64 `json:"delete_at"`
}

// ChannelMember is the user's read state in one channel.
type ChannelMember struct {
	ChannelID    string            `json:"channel_id"`
	UserID       string            `json:"user_id"`
	LastViewedAt int64             `json:"last_viewed_at"` // unix ms
	MsgCount     int64             `json:"msg_count"`
	MentionCount int64             `json:"mention_count"`
	NotifyProps  map[string]string `json:"notify_props"`
}

// Muted reports whether the channel only marks unread on mentions.
func (m ChannelMember) Muted() bool {
	return m.NotifyProps["mark_unread"] == "mention"
}

// SearchRequest is the body of POST /teams/{id}/posts/search.
type SearchRequest struct {
	Terms          string `json:"terms"`
	IsOrSearch     bool   `json:"is_or_search"`
	TimeZoneOffset int    `json:"time_zone_offset"` // seconds east of UTC
	Page           int    `json:"page"`
	PerPage        int    `json:"per_page"`
}

// Post is a chat message.
type Post struct {
	ID         string         `json:"id"`
	ChannelID  string         `json:"channel_id"`
	UserID     string         `json:"user_id"`
	RootID     string         `json:"root_id"`
	Type       string         `json:"type"`
	Message    string         `json:"message"`
	CreateAt   int64          `json:"create_at"` // unix ms
	UpdateAt   int64          `json:"update_at"`
	DeleteAt   int64          `json:"delete_at"`
	ReplyCount int            `json:"reply_count"`
	FileIDs    []string       `json:"file_ids,omitempty"`
	Props      map[string]any `json:"props,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// PostList is the standard Mattermost paginated post response.
type PostList struct {
	Order      []string         `json:"order"`
	Posts      map[string]*Post `json:"posts"`
	NextPostID string           `json:"next_post_id"`
	PrevPostID string           `json:"prev_post_id"`
}

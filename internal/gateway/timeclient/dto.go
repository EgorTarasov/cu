package timeclient

// User is a trimmed Mattermost user.
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
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
}

// Post is a chat message.
type Post struct {
	ID        string         `json:"id"`
	ChannelID string         `json:"channel_id"`
	UserID    string         `json:"user_id"`
	RootID    string         `json:"root_id"`
	Type      string         `json:"type"`
	Message   string         `json:"message"`
	CreateAt  int64          `json:"create_at"` // unix ms
	UpdateAt  int64          `json:"update_at"`
	DeleteAt  int64          `json:"delete_at"`
	FileIDs   []string       `json:"file_ids,omitempty"`
	Props     map[string]any `json:"props,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// PostList is the standard Mattermost paginated post response.
type PostList struct {
	Order      []string         `json:"order"`
	Posts      map[string]*Post `json:"posts"`
	NextPostID string           `json:"next_post_id"`
	PrevPostID string           `json:"prev_post_id"`
}

package clickstream

// EventType discriminates the payload variant sent to the /track endpoint.
type EventType string

const (
	EventTypeTrack       EventType = "track"
	EventTypeIdentify    EventType = "identify"
	EventTypeIncrement   EventType = "increment"
	EventTypeDecrement   EventType = "decrement"
	EventTypeGroup       EventType = "group"
	EventTypeAssignGroup EventType = "assign_group"
)

// Event is the envelope OpenPanel expects: {"type": "...", "payload": {...}}.
type Event struct {
	Type    EventType `json:"type"`
	Payload any       `json:"payload"`
}

// TrackPayload records a named event with optional properties.
type TrackPayload struct {
	Name       string         `json:"name"`
	Properties map[string]any `json:"properties,omitempty"`
	ProfileID  string         `json:"profileId,omitempty"`
	Groups     []string       `json:"groups,omitempty"`
}

// IdentifyPayload creates or updates a user profile.
type IdentifyPayload struct {
	ProfileID  string         `json:"profileId"`
	FirstName  string         `json:"firstName,omitempty"`
	LastName   string         `json:"lastName,omitempty"`
	Email      string         `json:"email,omitempty"`
	Avatar     string         `json:"avatar,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}

// CounterPayload increments or decrements a numeric profile property.
// Value must be positive; zero means "use server default of 1".
type CounterPayload struct {
	ProfileID string  `json:"profileId"`
	Property  string  `json:"property"`
	Value     float64 `json:"value,omitempty"`
}

// GroupPayload creates or updates a group.
type GroupPayload struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Name       string         `json:"name"`
	Properties map[string]any `json:"properties,omitempty"`
}

// AssignGroupPayload links a profile to one or more groups.
type AssignGroupPayload struct {
	GroupIDs  []string `json:"groupIds"`
	ProfileID string   `json:"profileId,omitempty"`
}

// TrackResponse is returned by /track; both fields may be empty for
// server-side events.
type TrackResponse struct {
	DeviceID  string `json:"deviceId,omitempty"`
	SessionID string `json:"sessionId,omitempty"`
}

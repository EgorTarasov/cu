package clickstream

import (
	"context"
	"errors"
)

// Track records a named event with optional properties.
func (c *Client) Track(ctx context.Context, payload TrackPayload) (*TrackResponse, error) {
	if payload.Name == "" {
		return nil, errors.New("event name is required")
	}
	return c.send(ctx, Event{Type: EventTypeTrack, Payload: payload})
}

// Identify creates or updates a user profile.
func (c *Client) Identify(ctx context.Context, payload IdentifyPayload) (*TrackResponse, error) {
	if payload.ProfileID == "" {
		return nil, errors.New("profileId is required")
	}
	return c.send(ctx, Event{Type: EventTypeIdentify, Payload: payload})
}

// Increment increases a numeric profile property.
func (c *Client) Increment(ctx context.Context, payload CounterPayload) (*TrackResponse, error) {
	if payload.ProfileID == "" || payload.Property == "" {
		return nil, errors.New("profileId and property are required")
	}
	return c.send(ctx, Event{Type: EventTypeIncrement, Payload: payload})
}

// Decrement decreases a numeric profile property.
func (c *Client) Decrement(ctx context.Context, payload CounterPayload) (*TrackResponse, error) {
	if payload.ProfileID == "" || payload.Property == "" {
		return nil, errors.New("profileId and property are required")
	}
	return c.send(ctx, Event{Type: EventTypeDecrement, Payload: payload})
}

// Group creates or updates a group.
func (c *Client) Group(ctx context.Context, payload GroupPayload) (*TrackResponse, error) {
	if payload.ID == "" || payload.Type == "" || payload.Name == "" {
		return nil, errors.New("id, type and name are required")
	}
	return c.send(ctx, Event{Type: EventTypeGroup, Payload: payload})
}

// AssignGroup links a profile to one or more groups. Note: groups are never
// auto-populated on events — pass Groups explicitly on each Track call.
func (c *Client) AssignGroup(ctx context.Context, payload AssignGroupPayload) (*TrackResponse, error) {
	if len(payload.GroupIDs) == 0 {
		return nil, errors.New("groupIds are required")
	}
	return c.send(ctx, Event{Type: EventTypeAssignGroup, Payload: payload})
}

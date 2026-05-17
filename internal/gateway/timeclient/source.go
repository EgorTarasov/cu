package timeclient

import (
	"context"
	"time"

	"github.com/EgorTarasov/cu/internal/usecase/notifications"
	"github.com/EgorTarasov/cu/internal/usecase/recordings"
)

// StoreSource adapts a local JSONL Store to the recordings.PostSource interface.
type StoreSource struct {
	Store *Store
}

func NewStoreSource(s *Store) *StoreSource { return &StoreSource{Store: s} }

// LoadPosts streams every post from the channel JSONL file.
// limit==0 means "all posts".
func (s *StoreSource) LoadPosts(_ context.Context) ([]recordings.RawPost, error) {
	posts, err := s.Store.ReadLastN(0)
	if err != nil {
		return nil, err
	}
	out := make([]recordings.RawPost, 0, len(posts))
	for _, p := range posts {
		out = append(out, recordings.RawPost{
			ID:       p.ID,
			Message:  p.Message,
			CreateAt: time.UnixMilli(p.CreateAt),
		})
	}
	return out, nil
}

// NotificationsSource adapts the JSONL Store to notifications.PostSource.
type NotificationsSource struct {
	Store *Store
}

func NewNotificationsSource(s *Store) *NotificationsSource { return &NotificationsSource{Store: s} }

func (s *NotificationsSource) LoadPosts(_ context.Context) ([]notifications.RawPost, error) {
	posts, err := s.Store.ReadLastN(0)
	if err != nil {
		return nil, err
	}
	out := make([]notifications.RawPost, 0, len(posts))
	for _, p := range posts {
		out = append(out, notifications.RawPost{
			ID:       p.ID,
			Message:  p.Message,
			CreateAt: time.UnixMilli(p.CreateAt),
		})
	}
	return out, nil
}

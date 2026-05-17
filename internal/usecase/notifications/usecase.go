package notifications

import (
	"context"
	"fmt"
	"sort"
)

// UseCase wraps a PostSource with notification extraction + indexing.
type UseCase struct {
	src       PostSource
	permalink Permalinker
}

func New(src PostSource) *UseCase {
	return &UseCase{src: src}
}

func (uc *UseCase) WithPermalinker(p Permalinker) *UseCase {
	uc.permalink = p
	return uc
}

// All returns every parsed notification, newest first.
func (uc *UseCase) All(ctx context.Context) ([]Notification, error) {
	posts, err := uc.src.LoadPosts(ctx)
	if err != nil {
		return nil, fmt.Errorf("load posts: %w", err)
	}
	var out []Notification
	for _, p := range posts {
		n := Parse(p.ID, p.Message, p.CreateAt)
		if n == nil {
			continue
		}
		if uc.permalink != nil {
			n.PostURL = uc.permalink.Permalink(p.ID)
		}
		out = append(out, *n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PostedAt.After(out[j].PostedAt) })
	return out, nil
}

// ForLongread filters notifications by LMS longread ID.
func (uc *UseCase) ForLongread(ctx context.Context, longreadID int) ([]Notification, error) {
	all, err := uc.All(ctx)
	if err != nil {
		return nil, err
	}
	var out []Notification
	for _, n := range all {
		if n.LongreadID == longreadID {
			out = append(out, n)
		}
	}
	return out, nil
}

// ForCourse filters notifications by LMS course ID.
func (uc *UseCase) ForCourse(ctx context.Context, courseID int) ([]Notification, error) {
	all, err := uc.All(ctx)
	if err != nil {
		return nil, err
	}
	var out []Notification
	for _, n := range all {
		if n.CourseID == courseID {
			out = append(out, n)
		}
	}
	return out, nil
}

// ByLongread indexes every notification by its LongreadID (skipping zero IDs).
func (uc *UseCase) ByLongread(ctx context.Context) (map[int][]Notification, error) {
	all, err := uc.All(ctx)
	if err != nil {
		return nil, err
	}
	out := map[int][]Notification{}
	for _, n := range all {
		if n.LongreadID == 0 {
			continue
		}
		out[n.LongreadID] = append(out[n.LongreadID], n)
	}
	return out, nil
}

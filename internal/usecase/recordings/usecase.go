package recordings

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// UseCase wraps a PostSource with parsing + search logic.
type UseCase struct {
	src       PostSource
	permalink Permalinker
}

func New(src PostSource) *UseCase {
	return &UseCase{src: src}
}

// WithPermalinker attaches a Permalinker so parsed recordings carry PostURL.
func (uc *UseCase) WithPermalinker(p Permalinker) *UseCase {
	uc.permalink = p
	return uc
}

// All returns every recording found in the source, newest first.
func (uc *UseCase) All(ctx context.Context) ([]Recording, error) {
	posts, err := uc.src.LoadPosts(ctx)
	if err != nil {
		return nil, fmt.Errorf("load posts: %w", err)
	}
	var recs []Recording
	for _, p := range posts {
		parsed := Parse(p.Message, p.CreateAt)
		for i := range parsed {
			parsed[i].PostID = p.ID
			if uc.permalink != nil {
				parsed[i].PostURL = uc.permalink.Permalink(p.ID)
			}
		}
		recs = append(recs, parsed...)
	}
	sort.Slice(recs, func(i, j int) bool {
		return recs[i].sortKey().After(recs[j].sortKey())
	})
	return recs, nil
}

// Search returns recordings whose subject contains query (case-insensitive).
// An empty query returns all recordings.
func (uc *UseCase) Search(ctx context.Context, query string) ([]Recording, error) {
	all, err := uc.All(ctx)
	if err != nil {
		return nil, err
	}
	q := strings.TrimSpace(strings.ToLower(query))
	if q == "" {
		return all, nil
	}
	var out []Recording
	for _, r := range all {
		if strings.Contains(strings.ToLower(r.Subject), q) {
			out = append(out, r)
		}
	}
	return out, nil
}

// ForCourse is a convenience wrapper over Search intended for course
// summaries: matches recordings whose subject contains the course name.
func (uc *UseCase) ForCourse(ctx context.Context, courseName string) ([]Recording, error) {
	return uc.Search(ctx, courseName)
}

package timechat

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	tc "github.com/EgorTarasov/cu/internal/gateway/timeclient"
	"github.com/EgorTarasov/cu/internal/model"
)

const (
	DefaultChannelLimit  = 30
	MaxChannelLimit      = 200
	DefaultSearchLimit   = 20
	MaxSearchLimit       = 100
	DefaultPerChannelCap = 20
	channelFetchParallel = 6
)

// ChannelPosts returns messages of one channel in chronological order: the last
// `limit` messages, or all messages after `since` (capped to the latest `limit`).
func (s *Service) ChannelPosts(
	ctx context.Context, query string, since *time.Time, limit int,
) (*model.TimeChannelPosts, error) {
	sess, err := s.session(ctx)
	if err != nil {
		return nil, err
	}
	all, err := s.channels(ctx, sess)
	if err != nil {
		return nil, err
	}
	ch, err := resolve(all, query)
	if err != nil {
		return nil, err
	}
	return s.fetchChannel(ctx, sess, *ch, since, clamp(limit, DefaultChannelLimit, MaxChannelLimit))
}

// Thread returns the root post and every reply of the thread containing postID.
func (s *Service) Thread(ctx context.Context, postID string) (*model.TimeChannelPosts, error) {
	sess, err := s.session(ctx)
	if err != nil {
		return nil, err
	}
	list, err := sess.client.GetPostThread(ctx, strings.TrimSpace(postID))
	if err != nil {
		return nil, s.wrap(err, "get thread "+postID)
	}
	msgs := s.messages(ctx, sess, list, nil)
	out := &model.TimeChannelPosts{Messages: msgs, Total: len(msgs)}
	if len(msgs) > 0 {
		if all, cerr := s.channels(ctx, sess); cerr == nil {
			if ch, rerr := resolve(all, msgs[0].ChannelID); rerr == nil {
				out.Channel = *ch
			}
		}
		if out.Channel.ID == "" {
			out.Channel = model.TimeChannel{ID: msgs[0].ChannelID, Title: msgs[0].ChannelID}
		}
	}
	return out, nil
}

// Search runs a Mattermost message search across the team. Terms support the
// usual modifiers: from:username, in:channel-name, after:YYYY-MM-DD, before:, on:, "exact phrase".
func (s *Service) Search(ctx context.Context, terms string, limit int) (*model.TimeSearchOutput, error) {
	terms = strings.TrimSpace(terms)
	if terms == "" {
		return nil, errors.New("terms are required")
	}
	sess, err := s.session(ctx)
	if err != nil {
		return nil, err
	}
	_, offset := time.Now().Zone()
	list, err := sess.client.SearchTeamPosts(ctx, sess.team.ID, tc.SearchRequest{
		Terms:          terms,
		TimeZoneOffset: offset,
		PerPage:        clamp(limit, DefaultSearchLimit, MaxSearchLimit),
	})
	if err != nil {
		return nil, s.wrap(err, "search posts")
	}

	titles := map[string]string{}
	if all, cerr := s.channels(ctx, sess); cerr == nil {
		for _, ch := range all {
			titles[ch.ID] = ch.Title
		}
	}
	msgs := s.messagesInOrder(ctx, sess, list, titles)
	return &model.TimeSearchOutput{Terms: terms, Messages: msgs}, nil
}

// Updates collects new messages across channels. With in.Since == nil it returns
// what is unread for the user; otherwise every message posted after Since in
// channels with activity, read or not. Muted channels are listed in Skipped
// unless they mention the user or IncludeMuted is set.
func (s *Service) Updates(ctx context.Context, in model.TimeUpdatesInput) (*model.TimeUpdatesOutput, error) {
	sess, err := s.session(ctx)
	if err != nil {
		return nil, err
	}
	all, err := s.channels(ctx, sess)
	if err != nil {
		return nil, err
	}
	sortChannels(all)

	out := &model.TimeUpdatesOutput{Since: in.Since}
	type job struct {
		ch    model.TimeChannel
		since time.Time
	}
	var jobs []job
	for _, ch := range all {
		if !kindAllowed(ch.Kind, in.Kinds) {
			continue
		}
		var from time.Time
		if in.Since == nil {
			if ch.Unread == 0 && ch.Mentions == 0 {
				continue
			}
			from = ch.LastViewedAt
		} else {
			if !ch.LastPostAt.After(*in.Since) {
				continue
			}
			from = *in.Since
		}
		if ch.Muted && ch.Mentions == 0 && !in.IncludeMuted {
			out.Skipped = append(out.Skipped, ch)
			continue
		}
		jobs = append(jobs, job{ch: ch, since: from})
	}

	limit := clamp(in.MaxPerChannel, DefaultPerChannelCap, MaxChannelLimit)
	results := make([]*model.TimeChannelPosts, len(jobs))
	errs := make([]error, len(jobs))
	sem := make(chan struct{}, channelFetchParallel)
	var wg sync.WaitGroup
	for i, j := range jobs {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			since := j.since
			results[i], errs[i] = s.fetchChannel(ctx, sess, j.ch, &since, limit)
		}()
	}
	wg.Wait()

	for i, r := range results {
		if errs[i] != nil {
			return nil, errs[i]
		}
		if r.Total > 0 {
			out.Channels = append(out.Channels, *r)
		}
	}
	return out, nil
}

func (s *Service) fetchChannel(
	ctx context.Context, sess *session, ch model.TimeChannel, since *time.Time, limit int,
) (*model.TimeChannelPosts, error) {
	q := tc.PostsQuery{PerPage: limit}
	if since != nil && !since.IsZero() {
		q = tc.PostsQuery{Since: since.UnixMilli()}
	}
	list, err := sess.client.GetChannelPosts(ctx, ch.ID, q)
	if err != nil {
		return nil, s.wrap(err, "read channel "+ch.Title)
	}

	msgs := s.messages(ctx, sess, list, since)
	for i := range msgs {
		msgs[i].ChannelTitle = ch.Title
	}
	out := &model.TimeChannelPosts{Channel: ch, Since: since, Total: len(msgs)}
	if len(msgs) > limit {
		msgs = msgs[len(msgs)-limit:]
	}
	out.Messages = msgs
	return out, nil
}

// messages converts a post list to chronological messages, dropping deleted
// posts and, when since is set, posts created before it (the API also returns edits).
func (s *Service) messages(
	ctx context.Context,
	sess *session,
	list *tc.PostList,
	since *time.Time,
) []model.TimeMessage {
	posts := make([]*tc.Post, 0, len(list.Posts))
	for _, p := range list.Posts {
		if p == nil || p.DeleteAt != 0 {
			continue
		}
		if since != nil && !since.IsZero() && p.CreateAt <= since.UnixMilli() {
			continue
		}
		posts = append(posts, p)
	}
	sort.Slice(posts, func(i, j int) bool { return posts[i].CreateAt < posts[j].CreateAt })
	return s.convert(ctx, sess, posts, nil)
}

// messagesInOrder keeps the API order (search relevance / recency).
func (s *Service) messagesInOrder(
	ctx context.Context, sess *session, list *tc.PostList, titles map[string]string,
) []model.TimeMessage {
	posts := make([]*tc.Post, 0, len(list.Order))
	for _, id := range list.Order {
		if p := list.Posts[id]; p != nil && p.DeleteAt == 0 {
			posts = append(posts, p)
		}
	}
	return s.convert(ctx, sess, posts, titles)
}

func (s *Service) convert(
	ctx context.Context, sess *session, posts []*tc.Post, titles map[string]string,
) []model.TimeMessage {
	ids := make([]string, 0, len(posts))
	seen := map[string]bool{}
	for _, p := range posts {
		if !seen[p.UserID] {
			seen[p.UserID] = true
			ids = append(ids, p.UserID)
		}
	}
	s.prefetchUsers(ctx, sess.client, ids)

	out := make([]model.TimeMessage, 0, len(posts))
	for _, p := range posts {
		m := model.TimeMessage{
			ID:        p.ID,
			RootID:    p.RootID,
			ChannelID: p.ChannelID,
			Author:    displayName(s.user(ctx, sess.client, p.UserID)),
			FromMe:    p.UserID == sess.me.ID,
			At:        fromMillis(p.CreateAt),
			Text:      p.Message,
			Replies:   p.ReplyCount,
			Files:     len(p.FileIDs),
			System:    strings.HasPrefix(p.Type, "system_"),
		}
		if titles != nil {
			m.ChannelTitle = titles[p.ChannelID]
		}
		out = append(out, m)
	}
	return out
}

// ParseSince accepts a relative duration ("24h", "90m", "3d"), a date
// ("2026-10-29"), a local "2026-10-29 18:00", or RFC 3339. Empty means nil.
func ParseSince(raw string, now time.Time) (*time.Time, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return nil, nil //nolint:nilnil // empty input means no lower bound
	}
	if days, ok := strings.CutSuffix(v, "d"); ok {
		if n, err := strconv.Atoi(days); err == nil && n > 0 {
			t := now.Add(-time.Duration(n) * 24 * time.Hour)
			return &t, nil
		}
	}
	if d, err := time.ParseDuration(v); err == nil && d > 0 {
		t := now.Add(-d)
		return &t, nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04", "2006-01-02T15:04", "2006-01-02"} {
		if t, err := time.ParseInLocation(layout, v, now.Location()); err == nil {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("cannot parse since %q: use 24h, 3d, 2026-10-29 or 2026-10-29 18:00", raw)
}

func clamp(v, def, maxV int) int {
	if v <= 0 {
		return def
	}
	return min(v, maxV)
}

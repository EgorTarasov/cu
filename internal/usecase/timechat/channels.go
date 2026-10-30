package timechat

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	tc "github.com/EgorTarasov/cu/internal/gateway/timeclient"
	"github.com/EgorTarasov/cu/internal/model"
)

// ChannelsFilter narrows ListChannels.
type ChannelsFilter struct {
	Kinds      []model.TimeChannelKind
	UnreadOnly bool
}

// ListChannels returns the user's channels, DMs and group chats in the team
// with unread and mention counters, most relevant first.
func (s *Service) ListChannels(ctx context.Context, f ChannelsFilter) ([]model.TimeChannel, error) {
	sess, err := s.session(ctx)
	if err != nil {
		return nil, err
	}
	all, err := s.channels(ctx, sess)
	if err != nil {
		return nil, err
	}

	out := make([]model.TimeChannel, 0, len(all))
	for _, ch := range all {
		if !kindAllowed(ch.Kind, f.Kinds) {
			continue
		}
		if f.UnreadOnly && ch.Unread == 0 && ch.Mentions == 0 {
			continue
		}
		out = append(out, ch)
	}
	sortChannels(out)
	return out, nil
}

// ResolveChannel finds a channel by ID, slug, display name, or DM partner
// username ("@ivanov" or "ivanov"). A unique substring of the title also matches.
func (s *Service) ResolveChannel(ctx context.Context, query string) (*model.TimeChannel, error) {
	sess, err := s.session(ctx)
	if err != nil {
		return nil, err
	}
	all, err := s.channels(ctx, sess)
	if err != nil {
		return nil, err
	}
	return resolve(all, query)
}

func resolve(all []model.TimeChannel, query string) (*model.TimeChannel, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, errors.New("channel is required")
	}
	lq := strings.ToLower(strings.TrimPrefix(q, "@"))

	for i := range all {
		ch := &all[i]
		if ch.ID == q || strings.EqualFold(ch.Name, q) || strings.EqualFold(ch.Title, q) {
			return ch, nil
		}
	}
	for i := range all {
		ch := &all[i]
		if ch.Kind == model.TimeChannelDM && strings.Contains(strings.ToLower(ch.Title), "(@"+lq+")") {
			return ch, nil
		}
	}

	var matches []*model.TimeChannel
	for i := range all {
		if strings.Contains(strings.ToLower(all[i].Title), lq) {
			matches = append(matches, &all[i])
		}
	}
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("channel %q not found — list channels via list_time_channels", query)
	case 1:
		return matches[0], nil
	}
	names := make([]string, 0, len(matches))
	for _, m := range matches {
		names = append(names, fmt.Sprintf("%q (%s)", m.Title, m.ID))
	}
	return nil, fmt.Errorf("channel %q is ambiguous: %s", query, strings.Join(names, ", "))
}

func (s *Service) channels(ctx context.Context, sess *session) ([]model.TimeChannel, error) {
	raw, err := sess.client.GetUserChannelsForTeam(ctx, sess.me.ID, sess.team.ID)
	if err != nil {
		return nil, s.wrap(err, "list channels")
	}
	members, err := sess.client.GetChannelMembersForUser(ctx, sess.me.ID, sess.team.ID)
	if err != nil {
		return nil, s.wrap(err, "list channel read state")
	}
	byChannel := make(map[string]tc.ChannelMember, len(members))
	for _, m := range members {
		byChannel[m.ChannelID] = m
	}

	var partnerIDs []string
	for _, ch := range raw {
		if ch.Type == "D" {
			partnerIDs = append(partnerIDs, dmPartner(ch.Name, sess.me.ID))
		}
	}
	s.prefetchUsers(ctx, sess.client, partnerIDs)

	out := make([]model.TimeChannel, 0, len(raw))
	for _, ch := range raw {
		if ch.DeleteAt != 0 {
			continue
		}
		m := byChannel[ch.ID]
		item := model.TimeChannel{
			ID:           ch.ID,
			Name:         ch.Name,
			Title:        ch.DisplayName,
			Kind:         kindOf(ch.Type),
			Mentions:     m.MentionCount,
			Muted:        m.Muted(),
			LastPostAt:   fromMillis(ch.LastPostAt),
			LastViewedAt: fromMillis(m.LastViewedAt),
		}
		if unread := ch.TotalMsgCount - m.MsgCount; unread > 0 {
			item.Unread = unread
		}
		if ch.Type == "D" {
			item.Title = displayName(s.user(ctx, sess.client, dmPartner(ch.Name, sess.me.ID)))
		}
		if item.Title == "" {
			item.Title = ch.Name
		}
		out = append(out, item)
	}
	return out, nil
}

func dmPartner(name, meID string) string {
	a, b, ok := strings.Cut(name, "__")
	if !ok {
		return name
	}
	if a == meID {
		return b
	}
	return a
}

func kindOf(t string) model.TimeChannelKind {
	switch t {
	case "D":
		return model.TimeChannelDM
	case "G":
		return model.TimeChannelGroup
	case "P":
		return model.TimeChannelPrivate
	default:
		return model.TimeChannelPublic
	}
}

// ParseKind maps a user-facing kind ("dm", "group", "public", "private", "all", "")
// to a filter; nil means every kind.
func ParseKind(s string) ([]model.TimeChannelKind, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "all":
		return nil, nil
	case "dm", "direct":
		return []model.TimeChannelKind{model.TimeChannelDM, model.TimeChannelGroup}, nil
	case "group":
		return []model.TimeChannelKind{model.TimeChannelGroup}, nil
	case "public":
		return []model.TimeChannelKind{model.TimeChannelPublic}, nil
	case "private":
		return []model.TimeChannelKind{model.TimeChannelPrivate}, nil
	case "channels":
		return []model.TimeChannelKind{model.TimeChannelPublic, model.TimeChannelPrivate}, nil
	}
	return nil, fmt.Errorf("unknown kind %q: use dm, group, channels, public, private or all", s)
}

func kindAllowed(k model.TimeChannelKind, kinds []model.TimeChannelKind) bool {
	return len(kinds) == 0 || slices.Contains(kinds, k)
}

const (
	rankMentioned = iota
	rankUnread
	rankMutedUnread
	rankRead
)

func sortChannels(chs []model.TimeChannel) {
	rank := func(c model.TimeChannel) int {
		switch {
		case c.Mentions > 0:
			return rankMentioned
		case c.Unread > 0 && !c.Muted:
			return rankUnread
		case c.Unread > 0:
			return rankMutedUnread
		default:
			return rankRead
		}
	}
	sort.SliceStable(chs, func(i, j int) bool {
		ri, rj := rank(chs[i]), rank(chs[j])
		if ri != rj {
			return ri < rj
		}
		return chs[i].LastPostAt.After(chs[j].LastPostAt)
	})
}

func fromMillis(ms int64) time.Time {
	if ms <= 0 {
		return time.Time{}
	}
	return time.UnixMilli(ms)
}

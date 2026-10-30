// Package timechat reads time.cu.ru (Mattermost) channels, direct messages and threads.
// It never posts, edits or marks anything as read.
package timechat

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	tc "github.com/EgorTarasov/cu/internal/gateway/timeclient"
)

// ErrNotLoggedIn is returned when there is no saved token or the session has expired.
var ErrNotLoggedIn = errors.New(
	"нет входа в time.cu.ru или сессия истекла — войдите: `cu login --time`")

// Client is the read-only subset of the time.cu.ru gateway used here.
type Client interface {
	GetMe(ctx context.Context) (*tc.User, error)
	GetTeamByName(ctx context.Context, name string) (*tc.Team, error)
	GetUser(ctx context.Context, userID string) (*tc.User, error)
	GetUserChannelsForTeam(ctx context.Context, userID, teamID string) ([]tc.Channel, error)
	GetChannelMembersForUser(ctx context.Context, userID, teamID string) ([]tc.ChannelMember, error)
	GetChannelPosts(ctx context.Context, channelID string, q tc.PostsQuery) (*tc.PostList, error)
	GetPostThread(ctx context.Context, postID string) (*tc.PostList, error)
	SearchTeamPosts(ctx context.Context, teamID string, req tc.SearchRequest) (*tc.PostList, error)
}

// Connector builds a client; it is called lazily so a fresh `cu login --time`
// is picked up without restarting the MCP server.
type Connector func() (Client, error)

type Service struct {
	connect  Connector
	teamName string

	mu     sync.Mutex
	client Client
	me     *tc.User
	team   *tc.Team
	users  map[string]*tc.User
}

func New(connect Connector, teamName string) *Service {
	return &Service{connect: connect, teamName: teamName, users: map[string]*tc.User{}}
}

type session struct {
	client Client
	me     *tc.User
	team   *tc.Team
}

func (s *Service) session(ctx context.Context) (*session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client != nil && s.me != nil && s.team != nil {
		return &session{client: s.client, me: s.me, team: s.team}, nil
	}

	client, err := s.connect()
	if err != nil {
		if errors.Is(err, tc.ErrNoToken) || errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotLoggedIn
		}
		return nil, fmt.Errorf("connect to time.cu.ru: %w", err)
	}

	me, err := client.GetMe(ctx)
	if err != nil {
		return nil, s.wrapLocked(err, "get current user")
	}
	team, err := client.GetTeamByName(ctx, s.teamName)
	if err != nil {
		return nil, s.wrapLocked(err, "get team "+s.teamName)
	}

	s.client, s.me, s.team = client, me, team
	return &session{client: client, me: me, team: team}, nil
}

func (s *Service) wrap(err error, op string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.wrapLocked(err, op)
}

func (s *Service) wrapLocked(err error, op string) error {
	if errors.Is(err, tc.ErrUnauthorized) || errors.Is(err, tc.ErrNoToken) {
		s.client, s.me, s.team = nil, nil, nil
		return ErrNotLoggedIn
	}
	return fmt.Errorf("%s: %w", op, err)
}

func (s *Service) user(ctx context.Context, client Client, id string) *tc.User {
	s.mu.Lock()
	u, ok := s.users[id]
	s.mu.Unlock()
	if ok {
		return u
	}

	fetched, err := client.GetUser(ctx, id)
	if err != nil {
		return &tc.User{ID: id}
	}
	s.mu.Lock()
	s.users[id] = fetched
	s.mu.Unlock()
	return fetched
}

const userFetchParallelism = 8

func (s *Service) prefetchUsers(ctx context.Context, client Client, ids []string) {
	sem := make(chan struct{}, userFetchParallelism)
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		sem <- struct{}{}
		go func(id string) {
			defer wg.Done()
			defer func() { <-sem }()
			s.user(ctx, client, id)
		}(id)
	}
	wg.Wait()
}

func displayName(u *tc.User) string {
	full := u.FirstName
	if u.LastName != "" {
		if full != "" {
			full += " "
		}
		full += u.LastName
	}
	switch {
	case u.Username == "":
		return u.ID
	case full == "":
		return "@" + u.Username
	default:
		return full + " (@" + u.Username + ")"
	}
}

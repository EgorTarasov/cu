package timechat

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tc "github.com/EgorTarasov/cu/internal/gateway/timeclient"
	"github.com/EgorTarasov/cu/internal/model"
)

var base = time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC)

func ms(t time.Time) int64 { return t.UnixMilli() }

type fakeClient struct {
	mu       sync.Mutex
	channels []tc.Channel
	members  []tc.ChannelMember
	users    map[string]tc.User
	posts    map[string][]tc.Post
	queries  map[string]tc.PostsQuery
	meErr    error
}

func newFake() *fakeClient {
	return &fakeClient{
		channels: []tc.Channel{
			{ID: "dm1", Type: "D", Name: "me__u2", TotalMsgCount: 5, LastPostAt: ms(base.Add(-time.Hour))},
			{ID: "dbms", Type: "O", Name: "msc-dbms", DisplayName: "MSc-DBMS-autumn-2026",
				TotalMsgCount: 40, LastPostAt: ms(base.Add(-2 * time.Hour))},
			{ID: "offtopic", Type: "O", Name: "off-topic", DisplayName: "off-topic",
				TotalMsgCount: 900, LastPostAt: ms(base.Add(-10 * time.Minute))},
			{ID: "net", Type: "O", Name: "msc-net", DisplayName: "MSc-NetworkTech-autumn-2026",
				TotalMsgCount: 3, LastPostAt: ms(base.Add(-72 * time.Hour))},
		},
		members: []tc.ChannelMember{
			{ChannelID: "dm1", MsgCount: 3, MentionCount: 2, LastViewedAt: ms(base.Add(-3 * time.Hour))},
			{ChannelID: "dbms", MsgCount: 39, LastViewedAt: ms(base.Add(-5 * time.Hour))},
			{ChannelID: "offtopic", MsgCount: 100, NotifyProps: map[string]string{"mark_unread": "mention"}},
			{ChannelID: "net", MsgCount: 3, LastViewedAt: ms(base.Add(-70 * time.Hour))},
		},
		users: map[string]tc.User{
			"me": {ID: "me", Username: "e.tarasov"},
			"u2": {ID: "u2", Username: "ivanov", FirstName: "Иван", LastName: "Иванов"},
		},
		posts: map[string][]tc.Post{
			"dm1": {
				{ID: "a", ChannelID: "dm1", UserID: "u2", Message: "old", CreateAt: ms(base.Add(-4 * time.Hour))},
				{
					ID:        "b",
					ChannelID: "dm1",
					UserID:    "u2",
					Message:   "пришли отчёт",
					CreateAt:  ms(base.Add(-2 * time.Hour)),
				},
				{ID: "c", ChannelID: "dm1", UserID: "me", Message: "ок", CreateAt: ms(base.Add(-time.Hour))},
				{
					ID:        "d",
					ChannelID: "dm1",
					UserID:    "u2",
					Message:   "deleted",
					CreateAt:  ms(base.Add(-time.Hour)),
					DeleteAt:  1,
				},
			},
			"dbms": {
				{
					ID:        "e",
					ChannelID: "dbms",
					UserID:    "u2",
					Message:   "перенос защиты",
					CreateAt:  ms(base.Add(-2 * time.Hour)),
				},
			},
		},
		queries: map[string]tc.PostsQuery{},
	}
}

func (f *fakeClient) GetMe(context.Context) (*tc.User, error) {
	if f.meErr != nil {
		return nil, f.meErr
	}
	u := f.users["me"]
	return &u, nil
}

func (f *fakeClient) GetTeamByName(context.Context, string) (*tc.Team, error) {
	return &tc.Team{ID: "t1", Name: "team"}, nil
}

func (f *fakeClient) GetUser(_ context.Context, id string) (*tc.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return &u, nil
}

func (f *fakeClient) GetUserChannelsForTeam(context.Context, string, string) ([]tc.Channel, error) {
	return f.channels, nil
}

func (f *fakeClient) GetChannelMembersForUser(context.Context, string, string) ([]tc.ChannelMember, error) {
	return f.members, nil
}

func (f *fakeClient) GetChannelPosts(_ context.Context, id string, q tc.PostsQuery) (*tc.PostList, error) {
	f.mu.Lock()
	f.queries[id] = q
	f.mu.Unlock()
	list := &tc.PostList{Posts: map[string]*tc.Post{}}
	for i := range f.posts[id] {
		p := f.posts[id][i]
		if q.Since > 0 && p.CreateAt < q.Since {
			continue
		}
		list.Posts[p.ID] = &p
		list.Order = append([]string{p.ID}, list.Order...)
	}
	return list, nil
}

func (f *fakeClient) GetPostThread(context.Context, string) (*tc.PostList, error) {
	return &tc.PostList{}, nil
}

func (f *fakeClient) SearchTeamPosts(context.Context, string, tc.SearchRequest) (*tc.PostList, error) {
	return nil, tc.ErrUnauthorized
}

func newService(f *fakeClient) *Service {
	return New(func() (Client, error) { return f, nil }, "team")
}

func TestListChannels_UnreadCountsTitlesAndOrder(t *testing.T) {
	svc := newService(newFake())

	chs, err := svc.ListChannels(context.Background(), ChannelsFilter{UnreadOnly: true})
	require.NoError(t, err)
	require.Len(t, chs, 3)

	require.Equal(t, "Иван Иванов (@ivanov)", chs[0].Title)
	require.Equal(t, model.TimeChannelDM, chs[0].Kind)
	require.Equal(t, int64(2), chs[0].Unread)
	require.Equal(t, int64(2), chs[0].Mentions)

	require.Equal(t, "MSc-DBMS-autumn-2026", chs[1].Title)
	require.Equal(t, "off-topic", chs[2].Title)
	require.True(t, chs[2].Muted, "muted channels sort after unmuted unread ones")
}

func TestResolve(t *testing.T) {
	svc := newService(newFake())
	sess, err := svc.session(context.Background())
	require.NoError(t, err)
	all, err := svc.channels(context.Background(), sess)
	require.NoError(t, err)

	cases := []struct {
		query, wantID, wantErr string
	}{
		{query: "dbms", wantID: "dbms"},
		{query: "MSc-DBMS-autumn-2026", wantID: "dbms"},
		{query: "@ivanov", wantID: "dm1"},
		{query: "ivanov", wantID: "dm1"},
		{query: "network", wantID: "net"},
		{query: "MSc", wantErr: "ambiguous"},
		{query: "nothing", wantErr: "not found"},
	}
	for _, tt := range cases {
		t.Run(tt.query, func(t *testing.T) {
			ch, err := resolve(all, tt.query)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantID, ch.ID)
		})
	}
}

func TestUpdates_UnreadSkipsMutedAndReadMessages(t *testing.T) {
	f := newFake()
	svc := newService(f)

	out, err := svc.Updates(context.Background(), model.TimeUpdatesInput{})
	require.NoError(t, err)
	require.Nil(t, out.Since)
	require.Len(t, out.Channels, 2)

	dm := out.Channels[0]
	require.Equal(t, "dm1", dm.Channel.ID)
	require.Equal(t, []string{"b", "c"}, messageIDs(dm.Messages), "read and deleted messages are dropped")
	require.True(t, dm.Messages[1].FromMe)
	require.Equal(t, "Иван Иванов (@ivanov)", dm.Messages[0].Author)
	require.Equal(t, ms(base.Add(-3*time.Hour)), f.queries["dm1"].Since)

	require.Len(t, out.Skipped, 1)
	require.Equal(t, "offtopic", out.Skipped[0].ID)
	_, fetched := f.queries["offtopic"]
	require.False(t, fetched, "muted channel without mentions must not be read")
}

func TestUpdates_SinceReturnsReadMessagesToo(t *testing.T) {
	f := newFake()
	svc := newService(f)
	since := base.Add(-150 * time.Minute)

	out, err := svc.Updates(context.Background(), model.TimeUpdatesInput{
		Since: &since, Kinds: []model.TimeChannelKind{model.TimeChannelDM}, MaxPerChannel: 1,
	})
	require.NoError(t, err)
	require.Len(t, out.Channels, 1)
	require.Equal(t, 2, out.Channels[0].Total)
	require.Equal(t, []string{"c"}, messageIDs(out.Channels[0].Messages), "keeps the newest when capped")
}

func TestNotLoggedIn(t *testing.T) {
	t.Run("no token", func(t *testing.T) {
		svc := New(func() (Client, error) { return nil, tc.ErrNoToken }, "team")
		_, err := svc.ListChannels(context.Background(), ChannelsFilter{})
		require.ErrorIs(t, err, ErrNotLoggedIn)
	})
	t.Run("expired session", func(t *testing.T) {
		f := newFake()
		f.meErr = tc.ErrUnauthorized
		_, err := newService(f).ListChannels(context.Background(), ChannelsFilter{})
		require.ErrorIs(t, err, ErrNotLoggedIn)
	})
	t.Run("401 mid-session drops the cached client", func(t *testing.T) {
		svc := newService(newFake())
		_, err := svc.Search(context.Background(), "дедлайн", 5)
		require.ErrorIs(t, err, ErrNotLoggedIn)
		require.Nil(t, svc.client)
	})
}

func TestParseSince(t *testing.T) {
	now := time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC)
	cases := []struct {
		in      string
		want    time.Time
		wantNil bool
		wantErr bool
	}{
		{in: "", wantNil: true},
		{in: "24h", want: now.Add(-24 * time.Hour)},
		{in: "3d", want: now.Add(-72 * time.Hour)},
		{in: "2026-10-29", want: time.Date(2026, 10, 29, 0, 0, 0, 0, time.UTC)},
		{in: "2026-10-29 18:00", want: time.Date(2026, 10, 29, 18, 0, 0, 0, time.UTC)},
		{in: "2026-10-29T18:00:00+03:00", want: time.Date(2026, 10, 29, 15, 0, 0, 0, time.UTC)},
		{in: "вчера", wantErr: true},
	}
	for _, tt := range cases {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParseSince(tt.in, now)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.wantNil {
				require.Nil(t, got)
				return
			}
			require.True(t, tt.want.Equal(*got), "got %s", got)
		})
	}
}

func messageIDs(msgs []model.TimeMessage) []string {
	ids := make([]string, 0, len(msgs))
	for _, m := range msgs {
		ids = append(ids, m.ID)
	}
	return ids
}

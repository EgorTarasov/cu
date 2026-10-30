package timeclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type APITestSuite struct {
	suite.Suite
	server *httptest.Server
	client *Client
	search SearchRequest
}

func TestAPITestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}

const testToken = "test-token"

func (s *APITestSuite) SetupTest() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/users/u1/teams/t1/channels/members", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"channel_id":"c1","user_id":"u1","last_viewed_at":1761800000000,
			"msg_count":10,"mention_count":2,"notify_props":{"mark_unread":"mention"}}]`))
	})
	mux.HandleFunc("GET /api/v4/users/u2", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"u2","username":"ivanov","first_name":"Иван","last_name":"Иванов"}`))
	})
	mux.HandleFunc("GET /api/v4/posts/p1/thread", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"order":["p1","p2"],"posts":{
			"p1":{"id":"p1","channel_id":"c1","user_id":"u2","message":"root","create_at":1761800000000,"reply_count":1},
			"p2":{"id":"p2","channel_id":"c1","user_id":"u1","root_id":"p1","message":"reply","create_at":1761800001000}}}`))
	})
	mux.HandleFunc("POST /api/v4/teams/t1/posts/search", func(w http.ResponseWriter, r *http.Request) {
		if _, err := r.Cookie(CookieName); err == nil {
			http.Error(w, `{"message":"csrf"}`, http.StatusUnauthorized)
			return
		}
		s.Equal("application/json", r.Header.Get("Content-Type"))
		if !s.NoError(json.NewDecoder(r.Body).Decode(&s.search)) {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"order":["p1"],"posts":{"p1":{"id":"p1","channel_id":"c1","message":"дедлайн"}}}`))
	})

	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+testToken {
			http.Error(w, `{"message":"Invalid or expired session"}`, http.StatusUnauthorized)
			return
		}
		mux.ServeHTTP(w, r)
	}))
	s.client = NewClientWithBase(testToken, s.server.URL)
	s.search = SearchRequest{}
}

func (s *APITestSuite) TearDownTest() {
	s.server.Close()
}

func (s *APITestSuite) TestGetChannelMembersForUser_Success() {
	members, err := s.client.GetChannelMembersForUser(context.Background(), "u1", "t1")
	s.Require().NoError(err)
	s.Require().Len(members, 1)
	s.Equal("c1", members[0].ChannelID)
	s.Equal(int64(10), members[0].MsgCount)
	s.Equal(int64(2), members[0].MentionCount)
	s.True(members[0].Muted())
}

func (s *APITestSuite) TestGetUser_Success() {
	u, err := s.client.GetUser(context.Background(), "u2")
	s.Require().NoError(err)
	s.Equal("ivanov", u.Username)
	s.Equal("Иван", u.FirstName)
	s.Equal("Иванов", u.LastName)
}

func (s *APITestSuite) TestGetPostThread_Success() {
	list, err := s.client.GetPostThread(context.Background(), "p1")
	s.Require().NoError(err)
	s.Equal([]string{"p1", "p2"}, list.Order)
	s.Equal(1, list.Posts["p1"].ReplyCount)
	s.Equal("p1", list.Posts["p2"].RootID)
}

func (s *APITestSuite) TestSearchTeamPosts_SendsBodyWithoutCookie() {
	list, err := s.client.SearchTeamPosts(context.Background(), "t1", SearchRequest{
		Terms: "дедлайн from:ivanov", TimeZoneOffset: 10800, PerPage: 5,
	})
	s.Require().NoError(err)
	s.Equal([]string{"p1"}, list.Order)
	s.Equal("дедлайн from:ivanov", s.search.Terms)
	s.Equal(10800, s.search.TimeZoneOffset)
	s.Equal(5, s.search.PerPage)
}

func (s *APITestSuite) TestExpiredToken_ReturnsErrUnauthorized() {
	c := NewClientWithBase("expired", s.server.URL)
	_, err := c.GetUser(context.Background(), "u2")
	s.ErrorIs(err, ErrUnauthorized)
}

func (s *APITestSuite) TestNoToken_ReturnsErrNoToken() {
	c := NewClientWithBase("", s.server.URL)
	_, err := c.GetPostThread(context.Background(), "p1")
	s.ErrorIs(err, ErrNoToken)
}

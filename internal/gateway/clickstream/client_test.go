package clickstream

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ClientSuite struct {
	suite.Suite

	server   *httptest.Server
	client   *Client
	received struct {
		headers http.Header
		body    map[string]any
	}
	status int
}

func (s *ClientSuite) SetupTest() {
	s.status = http.StatusOK
	mux := http.NewServeMux()
	mux.HandleFunc("POST /track", func(w http.ResponseWriter, r *http.Request) {
		s.received.headers = r.Header.Clone()
		s.NoError(json.NewDecoder(r.Body).Decode(&s.received.body))
		w.WriteHeader(s.status)
		if s.status == http.StatusUnauthorized {
			_, _ = w.Write([]byte(`{"error":"Unauthorized"}`))
		}
	})
	s.server = httptest.NewServer(mux)
	s.client = NewClient(s.server.URL, "test-client-id", "test-secret", "v0.0.1")
}

func (s *ClientSuite) TearDownTest() {
	s.server.Close()
}

func (s *ClientSuite) TestTrack_Success() {
	resp, err := s.client.Track(context.Background(), TrackPayload{
		Name:       "command_executed",
		ProfileID:  "student-42",
		Properties: map[string]any{"command": "fetch theme"},
	})

	s.Require().NoError(err)
	s.Require().NotNil(resp)

	s.Equal("test-client-id", s.received.headers.Get(HeaderClientID))
	s.Equal("test-secret", s.received.headers.Get(HeaderClientSecret))
	s.Equal("cu-cli", s.received.headers.Get(HeaderSDKName))
	s.Equal("v0.0.1", s.received.headers.Get(HeaderSDKVersion))
	s.Equal("application/json", s.received.headers.Get("Content-Type"))

	s.Equal("track", s.received.body["type"])
	payload, ok := s.received.body["payload"].(map[string]any)
	s.Require().True(ok)
	s.Equal("command_executed", payload["name"])
	s.Equal("student-42", payload["profileId"])
	s.Equal(map[string]any{"command": "fetch theme"}, payload["properties"])
	s.NotContains(payload, "groups")
}

func (s *ClientSuite) TestTrack_UserAgent() {
	s.client.SetUserAgent("cu-cli/v0.0.1 (X11; Linux x86_64)")

	_, err := s.client.Track(context.Background(), TrackPayload{Name: "x"})
	s.Require().NoError(err)
	s.Equal("cu-cli/v0.0.1 (X11; Linux x86_64)", s.received.headers.Get("User-Agent"))
}

func (s *ClientSuite) TestTrack_EmptyName() {
	_, err := s.client.Track(context.Background(), TrackPayload{})
	s.Require().ErrorContains(err, "event name is required")
}

func (s *ClientSuite) TestTrack_Unauthorized() {
	s.status = http.StatusUnauthorized

	_, err := s.client.Track(context.Background(), TrackPayload{Name: "x"})
	s.Require().ErrorContains(err, "HTTP 401")
	s.Require().ErrorContains(err, "Unauthorized")
}

func (s *ClientSuite) TestTrack_MissingCredentials() {
	client := NewClient(s.server.URL, "", "", "")
	_, err := client.Track(context.Background(), TrackPayload{Name: "x"})
	s.Require().ErrorContains(err, "client id and secret are required")
}

func (s *ClientSuite) TestIdentify_Success() {
	resp, err := s.client.Identify(context.Background(), IdentifyPayload{
		ProfileID: "student-42",
		FirstName: "Egor",
		Email:     "e@example.com",
	})

	s.Require().NoError(err)
	s.Require().NotNil(resp)

	s.Equal("identify", s.received.body["type"])
	payload, ok := s.received.body["payload"].(map[string]any)
	s.Require().True(ok)
	s.Equal("student-42", payload["profileId"])
	s.Equal("Egor", payload["firstName"])
	s.Equal("e@example.com", payload["email"])
	s.NotContains(payload, "lastName")
	s.NotContains(payload, "avatar")
}

func (s *ClientSuite) TestIdentify_MissingProfileID() {
	_, err := s.client.Identify(context.Background(), IdentifyPayload{})
	s.Require().ErrorContains(err, "profileId is required")
}

func (s *ClientSuite) TestIncrementDecrement() {
	tests := []struct {
		name     string
		call     func() (*TrackResponse, error)
		wantType string
	}{
		{
			name: "increment",
			call: func() (*TrackResponse, error) {
				return s.client.Increment(context.Background(), CounterPayload{
					ProfileID: "student-42", Property: "commands_run", Value: 2,
				})
			},
			wantType: "increment",
		},
		{
			name: "decrement",
			call: func() (*TrackResponse, error) {
				return s.client.Decrement(context.Background(), CounterPayload{
					ProfileID: "student-42", Property: "credits", Value: 1,
				})
			},
			wantType: "decrement",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, err := tt.call()
			s.Require().NoError(err)
			s.Equal(tt.wantType, s.received.body["type"])
			payload, ok := s.received.body["payload"].(map[string]any)
			s.Require().True(ok)
			s.Equal("student-42", payload["profileId"])
		})
	}
}

func (s *ClientSuite) TestCounter_MissingFields() {
	_, err := s.client.Increment(context.Background(), CounterPayload{ProfileID: "x"})
	s.Require().ErrorContains(err, "profileId and property are required")

	_, err = s.client.Decrement(context.Background(), CounterPayload{Property: "x"})
	s.Require().ErrorContains(err, "profileId and property are required")
}

func (s *ClientSuite) TestAssignGroup() {
	_, err := s.client.AssignGroup(context.Background(), AssignGroupPayload{
		GroupIDs:  []string{"g1", "g2"},
		ProfileID: "student-42",
	})
	s.Require().NoError(err)
	s.Equal("assign_group", s.received.body["type"])

	_, err = s.client.AssignGroup(context.Background(), AssignGroupPayload{})
	s.Require().ErrorContains(err, "groupIds are required")
}

func (s *ClientSuite) TestGroup_MissingFields() {
	_, err := s.client.Group(context.Background(), GroupPayload{ID: "g1"})
	s.Require().ErrorContains(err, "id, type and name are required")
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}

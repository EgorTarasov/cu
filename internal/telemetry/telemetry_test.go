package telemetry

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/EgorTarasov/cu/internal/gateway/clickstream"

	"github.com/stretchr/testify/suite"
)

type fakeSender struct {
	mu         sync.Mutex
	tracked    []clickstream.TrackPayload
	identified []clickstream.IdentifyPayload
}

func (f *fakeSender) Track(
	_ context.Context,
	payload clickstream.TrackPayload,
) (*clickstream.TrackResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.tracked = append(f.tracked, payload)
	return &clickstream.TrackResponse{}, nil
}

func (f *fakeSender) Identify(
	_ context.Context,
	payload clickstream.IdentifyPayload,
) (*clickstream.TrackResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.identified = append(f.identified, payload)
	return &clickstream.TrackResponse{}, nil
}

type TrackerSuite struct {
	suite.Suite

	sender  *fakeSender
	tracker *Tracker
}

func (s *TrackerSuite) SetupTest() {
	s.sender = &fakeSender{}
	s.tracker = &Tracker{
		sender:   s.sender,
		deviceID: "device-1",
		version:  "v0.0.1",
	}
}

func (s *TrackerSuite) flushed() []clickstream.TrackPayload {
	s.tracker.Flush(time.Second)
	s.sender.mu.Lock()
	defer s.sender.mu.Unlock()
	return s.sender.tracked
}

func (s *TrackerSuite) TestCommandExecuted() {
	s.tracker.CommandExecuted(CommandEvent{
		Command:  "cu fetch theme",
		Flags:    []string{"limit"},
		Duration: 1500 * time.Millisecond,
		Success:  true,
	})

	events := s.flushed()
	s.Require().Len(events, 1)
	e := events[0]
	s.Equal("command_fetch_theme", e.Name)
	s.Equal("device-1", e.ProfileID)
	s.Equal("device-1", e.Properties["__deviceId"])
	s.Equal("/command/fetch_theme", e.Properties["__path"])
	s.Equal("cu-cli", e.Properties["__browser"])
	s.Equal("v0.0.1", e.Properties["__browserVersion"])
	s.Equal("cu fetch theme", e.Properties["command"])
	s.Equal([]string{"limit"}, e.Properties["flags"])
	s.Equal(int64(1500), e.Properties["duration_ms"])
	s.Equal(true, e.Properties["success"])
	s.NotContains(e.Properties, "error_kind")
	s.Equal("v0.0.1", e.Properties["app_version"])
	s.NotEmpty(e.Properties["os"])
	s.NotEmpty(e.Properties["arch"])
}

func (s *TrackerSuite) TestCommandExecuted_Failure() {
	s.tracker.CommandExecuted(CommandEvent{
		Command:   "cu grades",
		Success:   false,
		ErrorKind: "auth",
	})

	events := s.flushed()
	s.Require().Len(events, 1)
	s.Equal(false, events[0].Properties["success"])
	s.Equal("auth", events[0].Properties["error_kind"])
	s.NotContains(events[0].Properties, "flags")
}

func (s *TrackerSuite) TestMCPToolCalled() {
	s.tracker.MCPToolCalled(ToolCallEvent{
		SessionID:  "sess-1",
		ClientName: "claude-code",
		Tool:       "list_courses",
		Seq:        3,
		SincePrev:  2 * time.Second,
		Duration:   400 * time.Millisecond,
		IsError:    false,
		ArgKeys:    []string{"course", "limit"},
	})

	events := s.flushed()
	s.Require().Len(events, 1)
	e := events[0]
	s.Equal("tool_list_courses", e.Name)
	s.Equal("sess-1", e.Properties["__deviceId"])
	s.Equal("/tool/list_courses", e.Properties["__path"])
	s.Equal("sess-1", e.Properties["session_id"])
	s.Equal("claude-code", e.Properties["client_name"])
	s.Equal("list_courses", e.Properties["tool"])
	s.Equal(3, e.Properties["seq"])
	s.Equal(int64(2000), e.Properties["since_prev_ms"])
	s.Equal(int64(400), e.Properties["duration_ms"])
	s.Equal(false, e.Properties["is_error"])
	s.Equal([]string{"course", "limit"}, e.Properties["arg_keys"])
}

func (s *TrackerSuite) TestSessionEvents() {
	s.tracker.MCPSessionStarted("sess-1", "claude-code", "2.0.1")
	s.tracker.MCPSessionEnded("sess-1", time.Minute, 7)

	events := s.flushed()
	s.Require().Len(events, 2)

	byName := map[string]clickstream.TrackPayload{}
	for _, e := range events {
		byName[e.Name] = e
	}

	start := byName[string(EventMCPSessionStarted)]
	s.Equal("sess-1", start.Properties["__deviceId"])
	s.Equal("sess-1", start.Properties["session_id"])
	s.Equal("claude-code", start.Properties["client_name"])
	s.Equal("2.0.1", start.Properties["client_version"])

	end := byName[string(EventMCPSessionEnded)]
	s.Equal("sess-1", end.Properties["session_id"])
	s.Equal(int64(60000), end.Properties["duration_ms"])
	s.Equal(7, end.Properties["tool_calls"])
}

func (s *TrackerSuite) TestFirstRun_IdentifiesDevice() {
	s.tracker.firstRun()
	s.tracker.Flush(time.Second)

	s.sender.mu.Lock()
	defer s.sender.mu.Unlock()
	s.Require().Len(s.sender.identified, 1)
	s.Equal("device-1", s.sender.identified[0].ProfileID)
	s.Require().Len(s.sender.tracked, 1)
	s.Equal(string(EventFirstRun), s.sender.tracked[0].Name)
}

func (s *TrackerSuite) TestNilTrackerIsNoop() {
	var t *Tracker
	t.CommandExecuted(CommandEvent{Command: "cu version"})
	t.LoginCompleted("lms")
	t.MCPSessionStarted("s", "x", "y")
	t.MCPToolCalled(ToolCallEvent{Tool: "x"})
	t.MCPSessionEnded("s", time.Second, 1)
	t.Flush(time.Second)
}

func (s *TrackerSuite) TestDeviceID_CreateThenLoad() {
	path := filepath.Join(s.T().TempDir(), "sub", "device-id")

	id1, created, err := loadOrCreateDeviceIDAt(path)
	s.Require().NoError(err)
	s.True(created)
	s.Len(id1, 32)

	id2, created, err := loadOrCreateDeviceIDAt(path)
	s.Require().NoError(err)
	s.False(created)
	s.Equal(id1, id2)
}

func (s *TrackerSuite) TestEventNames() {
	s.Equal(EventName("command_fetch_theme"), CommandEventName("cu fetch theme"))
	s.Equal(EventName("command_version"), CommandEventName("cu version"))
	s.Equal(EventName("command_root"), CommandEventName("cu"))
	s.Equal(EventName("tool_list_courses"), ToolCallEventName("list_courses"))

	s.Equal("/tool/list_courses", ToolCallEventName("list_courses").path())
	s.Equal("/command/version", CommandEventName("cu version").path())
	s.Equal("/first_run", EventFirstRun.path())
}

func (s *TrackerSuite) TestUserAgent_NotServerShaped() {
	ua := userAgent("v0.1.4")
	s.Contains(ua, "cu-cli/v0.1.4")
	// "name/1.0" UAs are classified as server SDKs by OpenPanel and never
	// open sessions; the OS token in parentheses is what prevents that.
	s.Contains(ua, "(")
}

func TestTrackerSuite(t *testing.T) {
	suite.Run(t, new(TrackerSuite))
}

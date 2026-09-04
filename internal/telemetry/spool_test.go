package telemetry

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/EgorTarasov/cu/internal/gateway/clickstream"

	"github.com/stretchr/testify/suite"
)

// failSender rejects every send, standing in for an unreachable endpoint.
type failSender struct {
	calls int
}

func (f *failSender) Track(
	_ context.Context,
	_ clickstream.TrackPayload,
) (*clickstream.TrackResponse, error) {
	f.calls++
	return nil, errors.New("unreachable")
}

func (f *failSender) Identify(
	_ context.Context,
	_ clickstream.IdentifyPayload,
) (*clickstream.TrackResponse, error) {
	return nil, errors.New("unreachable")
}

type SpoolSuite struct {
	suite.Suite

	queue  string
	sender *fakeSender
}

func (s *SpoolSuite) SetupTest() {
	s.queue = filepath.Join(s.T().TempDir(), "telemetry-queue.jsonl")
	s.sender = &fakeSender{}
}

func (s *SpoolSuite) tracker(snd sender) *Tracker {
	return &Tracker{sender: snd, deviceID: "device-1", version: "v0.0.1", spool: s.queue}
}

func (s *SpoolSuite) event(name string) clickstream.TrackPayload {
	return clickstream.TrackPayload{
		Name:       name,
		ProfileID:  "device-1",
		Properties: map[string]any{"k": "v"},
	}
}

func (s *SpoolSuite) TestAppendThenRead() {
	s.Require().NoError(appendSpool(s.queue, s.event("a")))
	s.Require().NoError(appendSpool(s.queue, s.event("b")))

	got := readSpool(s.queue)
	s.Require().Len(got, 2)
	s.Equal("a", got[0].Name)
	s.Equal("b", got[1].Name)
	s.Equal("v", got[1].Properties["k"])
}

func (s *SpoolSuite) TestReadSkipsCorruptLine() {
	s.Require().NoError(appendSpool(s.queue, s.event("a")))
	f, err := os.OpenFile(s.queue, os.O_WRONLY|os.O_APPEND, filePerm)
	s.Require().NoError(err)
	_, err = f.WriteString("{not json\n")
	s.Require().NoError(err)
	s.Require().NoError(f.Close())
	s.Require().NoError(appendSpool(s.queue, s.event("b")))

	// A torn write must not strand the events written after it.
	got := readSpool(s.queue)
	s.Require().Len(got, 2)
	s.Equal("b", got[1].Name)
}

func (s *SpoolSuite) TestCommandExecutedQueuesInsteadOfSending() {
	t := s.tracker(s.sender)
	t.CommandExecuted(CommandEvent{Command: "cuni task", Duration: 0, Success: true})

	// Nothing may go out on the wire: the CLI process is about to exit.
	t.Flush(0)
	s.sender.mu.Lock()
	sent := len(s.sender.tracked)
	s.sender.mu.Unlock()
	s.Zero(sent)

	queued := readSpool(s.queue)
	s.Require().Len(queued, 1)
	s.Equal("command_task", queued[0].Name)
	s.Equal("device-1", queued[0].ProfileID)
}

func (s *SpoolSuite) TestDrainShipsEverythingAndClearsFiles() {
	s.Require().NoError(appendSpool(s.queue, s.event("a")))
	s.Require().NoError(appendSpool(s.queue, s.event("b")))

	s.tracker(s.sender).drainSpool(context.Background())

	s.sender.mu.Lock()
	names := []string{s.sender.tracked[0].Name, s.sender.tracked[1].Name}
	s.sender.mu.Unlock()
	s.Equal([]string{"a", "b"}, names)

	s.NoFileExists(s.queue)
	s.NoFileExists(s.queue + ".sending")
}

func (s *SpoolSuite) TestDrainKeepsEventsWhenSendFails() {
	s.Require().NoError(appendSpool(s.queue, s.event("a")))
	fail := &failSender{}

	s.tracker(fail).drainSpool(context.Background())

	s.Equal(1, fail.calls)
	// The event must survive an outage rather than be dropped on the floor.
	s.Require().Len(readSpool(s.queue+".sending"), 1)
}

func (s *SpoolSuite) TestDrainAdoptsOrphanedBatch() {
	// A previous run claimed events and died before shipping them.
	s.Require().NoError(writeSpool(s.queue+".sending", []clickstream.TrackPayload{s.event("orphan")}))
	s.Require().NoError(appendSpool(s.queue, s.event("fresh")))

	s.tracker(s.sender).drainSpool(context.Background())

	s.sender.mu.Lock()
	names := []string{s.sender.tracked[0].Name, s.sender.tracked[1].Name}
	s.sender.mu.Unlock()
	s.Equal([]string{"orphan", "fresh"}, names)
	s.NoFileExists(s.queue + ".sending")
}

func (s *SpoolSuite) TestDrainDropsOldestBeyondCap() {
	for i := range maxSpoolEvents + 5 {
		s.Require().NoError(appendSpool(s.queue, s.event(fmt.Sprintf("e%d", i))))
	}

	s.tracker(s.sender).drainSpool(context.Background())

	s.sender.mu.Lock()
	defer s.sender.mu.Unlock()
	// Capped at maxDrainPerRun per run, and the oldest 5 are discarded first.
	s.Len(s.sender.tracked, maxDrainPerRun)
	s.Equal("e5", s.sender.tracked[0].Name)
}

func (s *SpoolSuite) TestFallsBackToDirectSendWithoutSpoolPath() {
	t := &Tracker{sender: s.sender, deviceID: "device-1", version: "v0.0.1"}
	t.CommandExecuted(CommandEvent{Command: "cuni task", Success: true})
	t.Flush(time.Second)

	s.sender.mu.Lock()
	defer s.sender.mu.Unlock()
	s.Require().Len(s.sender.tracked, 1)
	s.Equal("command_task", s.sender.tracked[0].Name)
}

func TestSpoolSuite(t *testing.T) {
	suite.Run(t, new(SpoolSuite))
}

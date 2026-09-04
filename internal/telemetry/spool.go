package telemetry

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/EgorTarasov/cu/internal/gateway/clickstream"
)

const (
	// maxSpoolEvents caps the queue so an endpoint that stays unreachable
	// cannot grow the file without bound; the oldest events are dropped.
	maxSpoolEvents = 100
	// maxDrainPerRun bounds how much one process tries to ship, so a large
	// backlog cannot keep a long-lived MCP session sending forever.
	maxDrainPerRun = 20
	// maxSpoolLineBytes rejects an implausibly long line rather than letting
	// a corrupted file allocate without limit.
	maxSpoolLineBytes = 64 * 1024
)

// spoolPath returns ~/.cu-cli/telemetry-queue.jsonl — next to device-id.
func spoolPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cu-cli", "telemetry-queue.jsonl"), nil
}

// readSpool loads queued payloads. A malformed line is skipped rather than
// failing the batch: a truncated write must not strand every later event.
func readSpool(path string) []clickstream.TrackPayload {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var out []clickstream.TrackPayload
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), maxSpoolLineBytes)
	for sc.Scan() {
		var p clickstream.TrackPayload
		if json.Unmarshal(sc.Bytes(), &p) == nil && p.Name != "" {
			out = append(out, p)
		}
	}
	return out
}

// writeSpool replaces the file atomically; an empty batch removes it.
func writeSpool(path string, events []clickstream.TrackPayload) error {
	if len(events) == 0 {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}

	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, filePerm)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(f)
	for i := range events {
		if err = enc.Encode(events[i]); err != nil {
			_ = f.Close()
			_ = os.Remove(tmp)
			return err
		}
	}
	if err = f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

// appendSpool queues one payload. It is a single O_APPEND write so that the
// CLI hot path costs a syscall rather than a network round trip.
func appendSpool(path string, p clickstream.TrackPayload) error {
	line, err := json.Marshal(p)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, filePerm)
	if err != nil {
		return err
	}
	if _, err = f.Write(append(line, '\n')); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// drainSpool ships queued events. It runs detached from the tracker's
// WaitGroup on purpose: Flush must not wait for it, otherwise the CLI would
// pay the very round trip the queue exists to avoid. A process that exits
// mid-drain simply leaves the rest for the next run.
//
// The queue is claimed by renaming it, which is atomic — two concurrent cuni
// processes can never ship the same event. The claimed file is rewritten
// after every success, so a killed process resends at most the one event
// that was in flight.
func (t *Tracker) drainSpool(ctx context.Context) {
	if t == nil || t.spool == "" {
		return
	}
	inFlight := t.spool + ".sending"

	// Adopt an orphaned batch from a run that died mid-drain before the
	// rename overwrites it, then merge in whatever the queue holds now.
	batch := readSpool(inFlight)
	if err := os.Rename(t.spool, inFlight); err == nil {
		batch = append(batch, readSpool(inFlight)...)
	}
	if len(batch) > maxSpoolEvents {
		batch = batch[len(batch)-maxSpoolEvents:]
	}
	if len(batch) == 0 {
		_ = os.Remove(inFlight)
		return
	}
	if err := writeSpool(inFlight, batch); err != nil {
		return
	}

	for sent := 0; len(batch) > 0 && sent < maxDrainPerRun; sent++ {
		if _, err := t.sender.Track(ctx, batch[0]); err != nil {
			_ = writeSpool(inFlight, batch)
			return
		}
		batch = batch[1:]
		if err := writeSpool(inFlight, batch); err != nil {
			return
		}
	}
}

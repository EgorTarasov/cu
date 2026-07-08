package telemetry

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	randomIDBytes = 16
	dirPerm       = 0700
	filePerm      = 0600
)

// newRandomID returns 32 hex chars of crypto randomness — used for the
// persistent device ID and per-run MCP session IDs.
func newRandomID() string {
	buf := make([]byte, randomIDBytes)
	if _, err := rand.Read(buf); err != nil {
		// crypto/rand never fails on supported platforms; fall back to a
		// constant so telemetry stays functional rather than crashing.
		return "rand-unavailable"
	}
	return hex.EncodeToString(buf)
}

// deviceIDPath returns ~/.cu-cli/device-id — next to the saved cookies.
func deviceIDPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cu-cli", "device-id"), nil
}

// loadOrCreateDeviceID returns the persistent anonymous device ID, generating
// and saving one on the very first run. The bool reports whether this call
// generated the ID (i.e. this is the first run on the device).
func loadOrCreateDeviceID() (string, bool, error) {
	path, err := deviceIDPath()
	if err != nil {
		return "", false, err
	}
	return loadOrCreateDeviceIDAt(path)
}

func loadOrCreateDeviceIDAt(path string) (string, bool, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		if id := strings.TrimSpace(string(data)); id != "" {
			return id, false, nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", false, err
	}

	id := newRandomID()

	if err = os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return "", false, err
	}
	if err = os.WriteFile(path, []byte(id), filePerm); err != nil {
		return "", false, err
	}

	return id, true, nil
}

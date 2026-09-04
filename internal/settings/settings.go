// Package settings persists user choices that are not secrets, alongside the
// cookie files in ~/.cu-cli.
package settings

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// LoginMethod records how the user prefers to obtain the LMS cookie.
type LoginMethod string

const (
	// LoginMethodUnset means the first-run prompt has not happened yet.
	LoginMethodUnset LoginMethod = ""
	// LoginMethodManual copies bff.cookie out of an already-open browser.
	LoginMethodManual LoginMethod = "manual"
	// LoginMethodBrowser drives Chrome over CDP and captures the cookie.
	LoginMethodBrowser LoginMethod = "browser"
)

// Settings is the on-disk shape of ~/.cu-cli/config.json.
type Settings struct {
	LoginMethod LoginMethod `json:"login_method,omitempty"`
}

func FilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cu-cli", "config.json"), nil
}

// Load returns zero-valued Settings when the file is absent, so a fresh
// install is indistinguishable from "nothing chosen yet".
func Load() (Settings, error) {
	path, err := FilePath()
	if err != nil {
		return Settings{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Settings{}, nil
		}
		return Settings{}, err
	}

	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return Settings{}, err
	}
	return s, nil
}

func Save(s Settings) error {
	path, err := FilePath()
	if err != nil {
		return err
	}

	if mkErr := os.MkdirAll(filepath.Dir(path), 0o700); mkErr != nil {
		return mkErr
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(data, '\n'), 0o600)
}

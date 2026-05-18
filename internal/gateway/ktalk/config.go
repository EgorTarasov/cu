package ktalk

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultBaseURL = "https://centraluniversity.ktalk.ru"
	HostSuffix     = "ktalk.ru"

	CookieNgToken      = "ngtoken"
	CookieSessionToken = "sessionToken"

	EnvBaseURL      = "CU_KTALK_BASE_URL"
	EnvNgToken      = "CU_KTALK_NG_TOKEN"      // #nosec G101 -- env var name, not a credential
	EnvSessionToken = "CU_KTALK_SESSION_TOKEN" // #nosec G101 -- env var name, not a credential
)

// RequiredCookies are the readiness markers: login is considered complete
// once every one of them is present. Everything else on the domain is
// captured opportunistically too.
var RequiredCookies = []string{CookieNgToken}

// Tokens is everything we captured from a Ktalk browser session.
type Tokens struct {
	Cookies        map[string]string `json:"cookies"`
	LocalStorage   map[string]string `json:"local_storage,omitempty"`
	SessionStorage map[string]string `json:"session_storage,omitempty"`
}

// Get looks up a key across cookies, localStorage and sessionStorage in that order.
func (t Tokens) Get(name string) string {
	if v, ok := t.Cookies[name]; ok && v != "" {
		return v
	}
	if v, ok := t.LocalStorage[name]; ok && v != "" {
		return v
	}
	if v, ok := t.SessionStorage[name]; ok && v != "" {
		return v
	}
	return ""
}

func (t Tokens) NgToken() string      { return t.Get(CookieNgToken) }
func (t Tokens) SessionToken() string { return t.Get(CookieSessionToken) }

func (t Tokens) Complete() bool {
	for _, name := range RequiredCookies {
		if t.Cookies[name] == "" {
			return false
		}
	}
	return true
}

type Config struct {
	BaseURL string
}

func LoadConfig() Config {
	cfg := Config{BaseURL: DefaultBaseURL}
	if v := os.Getenv(EnvBaseURL); v != "" {
		cfg.BaseURL = strings.TrimRight(v, "/")
	}
	return cfg
}

func TokensFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cu-cli", "ktalk-tokens.json"), nil
}

func SaveTokens(t Tokens) error {
	path, err := TokensFilePath()
	if err != nil {
		return err
	}
	if mkErr := os.MkdirAll(filepath.Dir(path), 0700); mkErr != nil {
		return mkErr
	}
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func LoadTokens() (Tokens, error) {
	t := Tokens{Cookies: map[string]string{}}

	path, err := TokensFilePath()
	if err != nil {
		return t, err
	}
	data, readErr := os.ReadFile(path)
	switch {
	case readErr == nil:
		if err := json.Unmarshal(data, &t); err != nil {
			return t, err
		}
		if t.Cookies == nil {
			t.Cookies = map[string]string{}
		}
	case errors.Is(readErr, os.ErrNotExist):
		// no saved file yet — fall through to env overrides
	default:
		return t, readErr
	}

	if v := os.Getenv(EnvNgToken); v != "" {
		t.Cookies[CookieNgToken] = v
	}
	if v := os.Getenv(EnvSessionToken); v != "" {
		t.Cookies[CookieSessionToken] = v
	}
	return t, nil
}

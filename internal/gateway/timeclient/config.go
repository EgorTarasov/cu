package timeclient

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Defaults for the Mattermost instance run by Central University.
// Override via environment variables if the deployment changes.
const (
	DefaultBaseURL     = "https://time.cu.ru"
	DefaultTeamName    = "tsentralnyy-universitet"
	DefaultBotUsername = "cu_notification_bot"
	CookieName         = "MMAUTHTOKEN"

	EnvBaseURL     = "CU_MM_BASE_URL"
	EnvTeamName    = "CU_MM_TEAM"
	EnvBotUsername = "CU_MM_BOT_USERNAME"
	EnvToken       = "CU_MM_TOKEN" // #nosec G101 -- env var name, not a credential
)

// Config groups the tunable knobs for the Mattermost gateway.
type Config struct {
	BaseURL     string
	TeamName    string
	BotUsername string
}

func LoadConfig() Config {
	cfg := Config{
		BaseURL:     DefaultBaseURL,
		TeamName:    DefaultTeamName,
		BotUsername: DefaultBotUsername,
	}
	if v := os.Getenv(EnvBaseURL); v != "" {
		cfg.BaseURL = strings.TrimRight(v, "/")
	}
	if v := os.Getenv(EnvTeamName); v != "" {
		cfg.TeamName = v
	}
	if v := os.Getenv(EnvBotUsername); v != "" {
		cfg.BotUsername = v
	}
	return cfg
}

func CookieFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cu-cli", "mm-cookie"), nil
}

func SaveCookie(cookie string) error {
	path, err := CookieFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(cookie), 0600)
}

func LoadCookie() (string, error) {
	path, err := CookieFilePath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// StorageDir returns the per-channel directory under ~/.cu-cli/mm/<channelID>.
func StorageDir(channelID string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cu-cli", "mm", channelID), nil
}

// Permalink builds the in-app URL for a post: <base>/<team>/pl/<postID>.
func (c Config) Permalink(postID string) string {
	if postID == "" {
		return ""
	}
	return strings.TrimRight(c.BaseURL, "/") + "/" + c.TeamName + "/pl/" + postID
}

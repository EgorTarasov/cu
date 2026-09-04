package timeclient

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	APIPrefix      = "/api/v4"
	DefaultTimeout = 30 * time.Second

	MeEndpoint             = APIPrefix + "/users/me"
	MyTeamsEndpoint        = APIPrefix + "/users/me/teams"
	TeamByNameEndpoint     = APIPrefix + "/teams/name/%s"
	UserChannelsEndpoint   = APIPrefix + "/users/%s/teams/%s/channels"
	UserByUsernameEndpoint = APIPrefix + "/users/username/%s"
	ChannelPostsEndpoint   = APIPrefix + "/channels/%s/posts"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
	userAgent  string
}

func NewClient(token string) *Client {
	return NewClientWithBase(token, DefaultBaseURL)
}

func NewClientWithBase(token, baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: DefaultTimeout},
		baseURL:    baseURL,
		token:      token,
		userAgent:  "cu-cli/mm",
	}
}

// NewClientFromEnv loads a token from CU_MM_TOKEN or the saved cookie file
// and uses CU_MM_BASE_URL when set, otherwise DefaultBaseURL.
func NewClientFromEnv() (*Client, error) {
	cfg := LoadConfig()
	token := os.Getenv(EnvToken)
	if token == "" {
		saved, err := LoadCookie()
		if err != nil {
			return nil, fmt.Errorf("failed to load mm cookie: %w", err)
		}
		token = saved
	}
	if token == "" {
		return nil, errors.New("no Mattermost auth found. Run 'cuni login --mattermost' or set CU_MM_TOKEN")
	}
	return NewClientWithBase(token, cfg.BaseURL), nil
}

func (c *Client) Token() string   { return c.token }
func (c *Client) BaseURL() string { return c.baseURL }

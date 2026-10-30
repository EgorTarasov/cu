package timeclient

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	APIPrefix      = "/api/v4"
	DefaultTimeout = 60 * time.Second

	// time.cu.ru sometimes takes ~10s to finish a TLS handshake, which is
	// exactly Go's default limit.
	tlsHandshakeTimeout = 25 * time.Second
	maxConnsPerHost     = 16

	MeEndpoint                 = APIPrefix + "/users/me"
	MyTeamsEndpoint            = APIPrefix + "/users/me/teams"
	TeamByNameEndpoint         = APIPrefix + "/teams/name/%s"
	UserChannelsEndpoint       = APIPrefix + "/users/%s/teams/%s/channels"
	UserByUsernameEndpoint     = APIPrefix + "/users/username/%s"
	ChannelPostsEndpoint       = APIPrefix + "/channels/%s/posts"
	UserEndpoint               = APIPrefix + "/users/%s"
	UserChannelMembersEndpoint = APIPrefix + "/users/%s/teams/%s/channels/members"
	PostThreadEndpoint         = APIPrefix + "/posts/%s/thread"
	TeamPostsSearchEndpoint    = APIPrefix + "/teams/%s/posts/search"
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
		httpClient: &http.Client{Timeout: DefaultTimeout, Transport: newTransport()},
		baseURL:    baseURL,
		token:      token,
		userAgent:  "cu-cli/mm",
	}
}

func newTransport() *http.Transport {
	t := &http.Transport{Proxy: http.ProxyFromEnvironment}
	if def, ok := http.DefaultTransport.(*http.Transport); ok {
		t = def.Clone()
	}
	t.TLSHandshakeTimeout = tlsHandshakeTimeout
	t.MaxIdleConnsPerHost = maxConnsPerHost
	return t
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
		return nil, fmt.Errorf("%w: run 'cu login --time' or set CU_MM_TOKEN", ErrNoToken)
	}
	return NewClientWithBase(token, cfg.BaseURL), nil
}

func (c *Client) Token() string   { return c.token }
func (c *Client) BaseURL() string { return c.baseURL }

package cu

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	BaseURL = "https://my.centraluniversity.ru"

	CourseOverviewEndpoint = "/api/micro-lms/courses/%d/overview"

	DefaultTimeout = 30 * time.Second
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	bffCookie  string
	userAgent  string
}

func NewClient(bffCookie string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		baseURL:   BaseURL,
		bffCookie: bffCookie,
		userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36",
	}
}

func NewClientFromEnv() (*Client, error) {
	bffCookie := os.Getenv("CU_BFF_COOKIE")
	if bffCookie == "" {
		return nil, fmt.Errorf("CU_BFF_COOKIE environment variable is required")
	}

	return NewClient(bffCookie), nil
}

func NewClientWithOptions(bffCookie string, timeout time.Duration, userAgent string) *Client {
	client := NewClient(bffCookie)
	client.httpClient.Timeout = timeout
	if userAgent != "" {
		client.userAgent = userAgent
	}
	return client
}

func (c *Client) SetBaseURL(baseURL string) {

	if _, err := url.Parse(baseURL); err != nil {
		return
	}
	c.baseURL = baseURL
}

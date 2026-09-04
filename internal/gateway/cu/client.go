package cu

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	BaseURL       = "https://my.centraluniversity.ru"
	GitLabBaseURL = "https://git.culab.ru"

	CourseEndpoint                = "/api/micro-lms/courses/%d"
	CourseOverviewEndpoint        = "/api/micro-lms/courses/%d/overview"
	CourseProgressEndpoint        = "/api/micro-lms/courses/%d/student/progress"
	StudentPerformanceEndpoint    = "/api/micro-lms/courses/%d/student-performance"
	ActivitiesPerformanceEndpoint = "/api/micro-lms/courses/%d/activities-performance"
	CourseExercisesEndpoint       = "/api/micro-lms/courses/%d/exercises"
	StudentCoursesEndpoint        = "/api/micro-lms/courses/student"
	ThemeEndpoint                 = "/api/micro-lms/themes/%d"
	LongreadEndpoint              = "/api/micro-lms/longreads/%d"
	LongreadMaterialsEndpoint     = "/api/micro-lms/longreads/%d/materials"
	CurrentStudentEndpoint        = "/api/micro-lms/students/me"
	TaskEndpoint                  = "/api/micro-lms/tasks/%d"
	DeadlinesEndpoint             = "/api/micro-lms/deadlines"
	DownloadLinkEndpoint          = "/api/micro-lms/content/download-link"

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
		userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
			"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36",
	}
}

func NewClientFromEnv() (*Client, error) {
	bffCookie := os.Getenv("CU_BFF_COOKIE")
	if bffCookie == "" {
		saved, err := LoadCookie()
		if err != nil {
			return nil, fmt.Errorf("failed to load saved cookie: %w", err)
		}
		bffCookie = saved
	}
	if bffCookie == "" {
		return nil, errors.New("no authentication found. Run 'cuni login' or set CU_BFF_COOKIE")
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

var allowedHosts = map[string]bool{
	"my.centraluniversity.ru": true,
	"git.culab.ru":            true,
}

func (c *Client) SetBaseURL(baseURL string) bool {
	u, err := url.Parse(baseURL)
	if err != nil || !allowedHosts[u.Hostname()] {
		return false
	}
	c.baseURL = baseURL
	return true
}

func (c *Client) setBaseURL(baseURL string) {
	c.baseURL = baseURL
}

package cu

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

func NewClientWithOptions(bffCookie string, timeout time.Duration, userAgent string) *Client {
	client := NewClient(bffCookie)
	client.httpClient.Timeout = timeout
	if userAgent != "" {
		client.userAgent = userAgent
	}
	return client
}

func (c *Client) SetBffCookie(cookie string) {
	c.bffCookie = cookie
}

func (c *Client) GetBffCookie() string {
	return c.bffCookie
}

func (c *Client) prepareRequest(method, endpoint string) (*http.Request, error) {
	fullURL := c.baseURL + endpoint

	req, err := http.NewRequest(method, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("DNT", "1")
	req.Header.Set("Priority", "u=1, i")
	req.Header.Set("Referer", "https://my.centraluniversity.ru/learn/courses/view/actual")
	req.Header.Set("Sec-Ch-Ua", `"Not=A?Brand";v="24", "Chromium";v="140"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"macOS"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", c.userAgent)

	if c.bffCookie != "" {
		cookie := &http.Cookie{
			Name:  "bff.cookie",
			Value: c.bffCookie,
		}
		req.AddCookie(cookie)
	}

	return req, nil
}

// executeRequest executes an HTTP request and returns the response
func (c *Client) executeRequest(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

func (c *Client) GetStudentCourses(limit int, state string) (*StudentCoursesResponse, error) {
	if c.bffCookie == "" {
		return nil, fmt.Errorf("bff.cookie is required for authentication")
	}

	endpoint := "/api/micro-lms/courses/student"
	params := url.Values{}

	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if state != "" {
		params.Set("state", state)
	}

	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	req, err := c.prepareRequest("GET", endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	resp, err := c.executeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("HTTP %d: failed to decode error response", resp.StatusCode)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, apiErr.Error())
	}

	var courses StudentCoursesResponse
	if err := json.NewDecoder(resp.Body).Decode(&courses); err != nil {
		return nil, fmt.Errorf("failed to decode student courses response: %w", err)
	}

	return &courses, nil
}

func (c *Client) GetCourseOverview(courseID int) (*CourseOverview, error) {
	if c.bffCookie == "" {
		return nil, fmt.Errorf("bff.cookie is required for authentication")
	}

	endpoint := fmt.Sprintf(CourseOverviewEndpoint, courseID)

	req, err := c.prepareRequest("GET", endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	resp, err := c.executeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("HTTP %d: failed to decode error response", resp.StatusCode)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, apiErr.Error())
	}

	var courseOverview CourseOverview
	if err := json.NewDecoder(resp.Body).Decode(&courseOverview); err != nil {
		return nil, fmt.Errorf("failed to decode course overview response: %w", err)
	}

	return &courseOverview, nil
}

func (c *Client) ValidateCookie() error {
	if c.bffCookie == "" {
		return fmt.Errorf("no bff.cookie set")
	}

	req, err := c.prepareRequest("GET", "/api/account/me")
	if err != nil {
		return fmt.Errorf("failed to prepare validation request: %w", err)
	}

	resp, err := c.executeRequest(req)
	if err != nil {
		return fmt.Errorf("validation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("bff.cookie is invalid or expired: %v", resp.StatusCode)
	}

	return nil
}

func (c *Client) SetBaseURL(baseURL string) {

	if _, err := url.Parse(baseURL); err != nil {
		return
	}
	c.baseURL = baseURL
}

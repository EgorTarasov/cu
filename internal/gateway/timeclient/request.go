package timeclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// APIError mirrors the Mattermost error envelope.
type APIError struct {
	ID            string `json:"id"`
	Message       string `json:"message"`
	DetailedError string `json:"detailed_error"`
	StatusCode    int    `json:"status_code"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.DetailedError != "" {
		return e.DetailedError
	}
	return fmt.Sprintf("mattermost api error (status %d)", e.StatusCode)
}

func (c *Client) prepareRequest(ctx context.Context, method, endpoint string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.AddCookie(&http.Cookie{Name: CookieName, Value: c.token})
	}
	return req, nil
}

// doJSON performs an authenticated GET and decodes the JSON body into *T.
func doJSON[T any](ctx context.Context, c *Client, endpoint string) (*T, error) {
	if c.token == "" {
		return nil, errors.New("MMAUTHTOKEN is required for authentication")
	}

	req, err := c.prepareRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var apiErr APIError
		if jsonErr := json.Unmarshal(body, &apiErr); jsonErr == nil && apiErr.Message != "" {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, apiErr.Error())
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var out T
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &out, nil
}

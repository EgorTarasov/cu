package timeclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	ErrNoToken      = errors.New("MMAUTHTOKEN is required for authentication")
	ErrUnauthorized = errors.New("time.cu.ru session is missing or expired")
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

func (c *Client) prepareRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
		// With the session cookie present Mattermost enforces CSRF on non-GET
		// requests and answers 401, so the cookie goes only with GETs.
		if method == http.MethodGet {
			req.AddCookie(&http.Cookie{Name: CookieName, Value: c.token})
		}
	}
	return req, nil
}

// doJSON performs an authenticated GET and decodes the JSON body into *T.
func doJSON[T any](ctx context.Context, c *Client, endpoint string) (*T, error) {
	return doJSONRequest[T](ctx, c, http.MethodGet, endpoint, nil)
}

// doJSONRequest sends payload (if non-nil) as a JSON body and decodes the response into *T.
func doJSONRequest[T any](ctx context.Context, c *Client, method, endpoint string, payload any) (*T, error) {
	if c.token == "" {
		return nil, ErrNoToken
	}

	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to encode request: %w", err)
		}
		body = bytes.NewReader(raw)
	}

	req, err := c.prepareRequest(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		var apiErr APIError
		if jsonErr := json.Unmarshal(raw, &apiErr); jsonErr == nil && apiErr.Message != "" {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, apiErr.Error())
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(raw))
	}

	var out T
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &out, nil
}

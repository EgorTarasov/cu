package cu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func (c *Client) prepareRequest(ctx context.Context, method, endpoint string) (*http.Request, error) {
	fullURL := c.baseURL + endpoint

	req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
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

func (c *Client) executeRequest(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req) //nolint:gosec // G704: URL is built from validated base URL
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// doJSON performs an authenticated GET request and decodes the JSON response into *T.
// Returns an error if bff.cookie is missing, the request fails, or the response is non-2xx.
func doJSON[T any](ctx context.Context, c *Client, endpoint string) (*T, error) {
	if c.bffCookie == "" {
		return nil, errors.New("bff.cookie is required for authentication")
	}

	req, err := c.prepareRequest(ctx, http.MethodGet, endpoint)
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
		if err = json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("HTTP %d: failed to decode error response", resp.StatusCode)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, apiErr.Error())
	}

	var out T
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &out, nil
}

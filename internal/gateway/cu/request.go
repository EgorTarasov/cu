package cu

import (
	"context"
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
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

package cu

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

func (c *Client) ValidateCookie() error {
	return c.ValidateCookieWithContext(context.Background())
}

func (c *Client) ValidateCookieWithContext(ctx context.Context) error {
	if c.bffCookie == "" {
		return errors.New("no bff.cookie set")
	}

	req, err := c.prepareRequest(ctx, http.MethodGet, "/api/account/me")
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

func (c *Client) SetBffCookie(cookie string) {
	c.bffCookie = cookie
}

func (c *Client) GetBffCookie() string {
	return c.bffCookie
}

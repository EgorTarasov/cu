package timeclient

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"

	"github.com/EgorTarasov/cu/internal/browser"
)

const tickerInterval = 500 * time.Millisecond

// LoginWithBrowser opens Chrome for Mattermost SSO login and captures MMAUTHTOKEN.
func LoginWithBrowser(ctx context.Context, timeout time.Duration) (string, error) {
	cfg := LoadConfig()
	return loginViaBrowser(ctx, timeout, cfg.BaseURL, CookieName)
}

func loginViaBrowser(ctx context.Context, timeout time.Duration, url, cookieName string) (string, error) {
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, browser.AllocatorOptions()...)
	defer allocCancel()

	chromeCtx, chromeCancel := chromedp.NewContext(allocCtx)
	defer chromeCancel()

	timeoutCtx, timeoutCancel := context.WithTimeout(chromeCtx, timeout)
	defer timeoutCancel()

	if err := chromedp.Run(timeoutCtx, chromedp.Navigate(url)); err != nil {
		return "", fmt.Errorf("failed to open browser: %w", wrapChromeError(err))
	}

	ticker := time.NewTicker(tickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			if timeoutCtx.Err() == context.DeadlineExceeded {
				return "", fmt.Errorf("login timed out after %s", timeout)
			}
			return "", errors.New("browser was closed before login completed")
		case <-ticker.C:
			cookie, err := extractCookieByName(timeoutCtx, cookieName)
			if err != nil {
				if strings.Contains(err.Error(), "target closed") {
					return "", errors.New("browser was closed before login completed")
				}
				continue
			}
			if cookie != "" {
				return cookie, nil
			}
		}
	}
}

func extractCookieByName(ctx context.Context, name string) (string, error) {
	var cookies []*network.Cookie
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		cookies, err = network.GetCookies().Do(ctx)
		return err
	}))
	if err != nil {
		return "", err
	}
	for _, c := range cookies {
		if c.Name == name {
			return c.Value, nil
		}
	}
	return "", nil
}

func wrapChromeError(err error) error {
	if strings.Contains(err.Error(), "exec") || strings.Contains(err.Error(), "not found") {
		return errors.New("chrome not found: install Google Chrome or set CHROME_PATH environment variable")
	}
	return err
}

// ValidateToken does a cheap call to /users/me to confirm the token works.
func (c *Client) ValidateToken(ctx context.Context) error {
	if c.token == "" {
		return errors.New("no MMAUTHTOKEN set")
	}
	_, err := doJSON[User](ctx, c, MeEndpoint)
	return err
}

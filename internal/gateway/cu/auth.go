package cu

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

const tickerInterval = 500 * time.Millisecond

// LoginWithBrowser opens Chrome for LMS login and captures bff.cookie.
func LoginWithBrowser(ctx context.Context, timeout time.Duration) (string, error) {
	return loginViaBrowser(ctx, timeout, BaseURL, "bff.cookie")
}

// LoginGitLabWithBrowser opens Chrome for GitLab SSO login and captures _gitlab_session cookie.
// It waits until the user completes SSO and is redirected away from the sign-in page.
func LoginGitLabWithBrowser(ctx context.Context, timeout time.Duration) (string, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCancel()

	chromeCtx, chromeCancel := chromedp.NewContext(allocCtx)
	defer chromeCancel()

	timeoutCtx, timeoutCancel := context.WithTimeout(chromeCtx, timeout)
	defer timeoutCancel()

	if err := chromedp.Run(timeoutCtx, chromedp.Navigate(GitLabBaseURL)); err != nil {
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
			// Check current URL — wait until we leave the sign-in page.
			var currentURL string
			if err := chromedp.Run(timeoutCtx, chromedp.Location(&currentURL)); err != nil {
				if strings.Contains(err.Error(), "target closed") {
					return "", errors.New("browser was closed before login completed")
				}
				continue
			}

			// Still on sign-in or Keycloak auth page — keep waiting.
			if strings.Contains(currentURL, "/users/sign_in") ||
				strings.Contains(currentURL, "id.centraluniversity.ru") {
				continue
			}

			// Redirected away from sign-in — SSO completed, grab cookie.
			cookie, err := extractCookieByName(timeoutCtx, "_gitlab_session")
			if err != nil {
				continue
			}
			if cookie != "" {
				return cookie, nil
			}
		}
	}
}

func loginViaBrowser(ctx context.Context, timeout time.Duration, url, cookieName string) (string, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
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

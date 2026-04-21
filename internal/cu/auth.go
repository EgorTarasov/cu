package cu

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func LoginWithBrowser(ctx context.Context, timeout time.Duration) (string, error) {
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

	if err := chromedp.Run(timeoutCtx, chromedp.Navigate(BaseURL)); err != nil {
		return "", fmt.Errorf("failed to open browser: %w", wrapChromeError(err))
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			if timeoutCtx.Err() == context.DeadlineExceeded {
				return "", fmt.Errorf("login timed out after %s", timeout)
			}
			return "", fmt.Errorf("browser was closed before login completed")
		case <-ticker.C:
			cookie, err := extractBffCookie(timeoutCtx)
			if err != nil {
				if strings.Contains(err.Error(), "target closed") {
					return "", fmt.Errorf("browser was closed before login completed")
				}
				continue
			}
			if cookie != "" {
				return cookie, nil
			}
		}
	}
}

func extractBffCookie(ctx context.Context) (string, error) {
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
		if c.Name == "bff.cookie" {
			return c.Value, nil
		}
	}
	return "", nil
}

func wrapChromeError(err error) error {
	if strings.Contains(err.Error(), "exec") || strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("Chrome not found. Install Google Chrome or set CHROME_PATH environment variable")
	}
	return err
}

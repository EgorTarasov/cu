package ktalk

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"

	"github.com/EgorTarasov/cu/internal/browser"
)

const tickerInterval = 500 * time.Millisecond

// LoginWithBrowser opens Chrome and waits until the browser is back on the
// Ktalk domain with the required cookies present, then snapshots cookies,
// localStorage and sessionStorage.
func LoginWithBrowser(ctx context.Context, timeout time.Duration) (Tokens, error) {
	cfg := LoadConfig()
	return capture(ctx, timeout, cfg.BaseURL, HostSuffix, RequiredCookies)
}

func capture(
	ctx context.Context,
	timeout time.Duration,
	startURL, hostSuffix string,
	required []string,
) (Tokens, error) {
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, browser.AllocatorOptions()...)
	defer allocCancel()

	chromeCtx, chromeCancel := chromedp.NewContext(allocCtx)
	defer chromeCancel()

	timeoutCtx, timeoutCancel := context.WithTimeout(chromeCtx, timeout)
	defer timeoutCancel()

	if err := chromedp.Run(timeoutCtx, chromedp.Navigate(startURL)); err != nil {
		return Tokens{}, fmt.Errorf("failed to open browser: %w", wrapChromeError(err))
	}

	ticker := time.NewTicker(tickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			if timeoutCtx.Err() == context.DeadlineExceeded {
				return Tokens{}, fmt.Errorf("login timed out after %s", timeout)
			}
			return Tokens{}, errors.New("browser was closed before login completed")
		case <-ticker.C:
			ok, err := onTargetHost(timeoutCtx, hostSuffix)
			if err != nil {
				if strings.Contains(err.Error(), "target closed") {
					return Tokens{}, errors.New("browser was closed before login completed")
				}
				continue
			}
			if !ok {
				// still on Keycloak / identity provider — keep waiting
				continue
			}
			cookies, err := snapshotCookies(timeoutCtx)
			if err != nil {
				continue
			}
			if !hasAll(cookies, required) {
				continue
			}
			local, _ := snapshotStorage(timeoutCtx, "localStorage")
			session, _ := snapshotStorage(timeoutCtx, "sessionStorage")
			return Tokens{
				Cookies:        cookies,
				LocalStorage:   local,
				SessionStorage: session,
			}, nil
		}
	}
}

func onTargetHost(ctx context.Context, suffix string) (bool, error) {
	var current string
	if err := chromedp.Run(ctx, chromedp.Location(&current)); err != nil {
		return false, err
	}
	u, err := url.Parse(current)
	if err != nil {
		return false, err
	}
	return strings.HasSuffix(u.Hostname(), suffix), nil
}

func snapshotCookies(ctx context.Context) (map[string]string, error) {
	var cookies []*network.Cookie
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		cookies, err = network.GetCookies().Do(ctx)
		return err
	}))
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, c := range cookies {
		if c.Value != "" {
			out[c.Name] = c.Value
		}
	}
	return out, nil
}

// snapshotStorage reads window.<storage> as a flat string map.
func snapshotStorage(ctx context.Context, store string) (map[string]string, error) {
	out := map[string]string{}
	script := fmt.Sprintf(`(() => {
		const s = window.%s;
		const o = {};
		for (let i = 0; i < s.length; i++) {
			const k = s.key(i);
			o[k] = s.getItem(k);
		}
		return o;
	})()`, store)
	if err := chromedp.Run(ctx, chromedp.Evaluate(script, &out)); err != nil {
		return nil, err
	}
	return out, nil
}

func hasAll(m map[string]string, required []string) bool {
	for _, name := range required {
		if m[name] == "" {
			return false
		}
	}
	return true
}

func wrapChromeError(err error) error {
	if strings.Contains(err.Error(), "exec") || strings.Contains(err.Error(), "not found") {
		return errors.New("chrome not found: install Google Chrome or set CHROME_PATH environment variable")
	}
	return err
}

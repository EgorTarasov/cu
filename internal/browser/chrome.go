// Package browser resolves which Chrome binary cu should drive over CDP and,
// when the host has none, fetches a pinned Chrome for Testing build.
package browser

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/chromedp/chromedp"
)

const (
	osDarwin  = "darwin"
	osWindows = "windows"
)

// systemCandidates lists absolute paths worth probing before falling back to
// a PATH lookup.
func systemCandidates() []string {
	switch runtime.GOOS {
	case osDarwin:
		return []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
		}
	case osWindows:
		return []string{
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		}
	default:
		return nil
	}
}

var pathNames = []string{
	"google-chrome",
	"google-chrome-stable",
	"chromium",
	"chromium-browser",
}

// Detect returns a system Chrome path, or "" when the host has none.
func Detect() string {
	if p := os.Getenv("CHROME_PATH"); p != "" {
		if isExecutable(p) {
			return p
		}
	}

	for _, p := range systemCandidates() {
		if isExecutable(p) {
			return p
		}
	}

	for _, name := range pathNames {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}

	return ""
}

// ManagedDir is where an opt-in Chrome for Testing download is unpacked.
func ManagedDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cu-cli", "chrome"), nil
}

// Managed returns the previously downloaded Chrome, or "" when absent.
func Managed() string {
	dir, err := ManagedDir()
	if err != nil {
		return ""
	}
	p := filepath.Join(dir, managedBinaryRelPath())
	if isExecutable(p) {
		return p
	}
	return ""
}

// Resolve prefers a system Chrome and falls back to the managed download.
// An empty result means the caller should offer to download one.
func Resolve() string {
	if p := Detect(); p != "" {
		return p
	}
	return Managed()
}

// AllocatorOptions returns exec-allocator options with a resolved Chrome
// applied. With no resolved path chromedp keeps its own default search.
func AllocatorOptions() []chromedp.ExecAllocatorOption {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
	)
	if p := Resolve(); p != "" {
		opts = append(opts, chromedp.ExecPath(p))
	}
	return opts
}

func isExecutable(path string) bool {
	info, err := os.Stat(path) //nolint:gosec // G703: path is CHROME_PATH, a fixed candidate, or ManagedDir
	if err != nil || info.IsDir() {
		return false
	}
	if runtime.GOOS == osWindows {
		return true
	}
	return info.Mode().Perm()&0o111 != 0
}

package browser

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDetect_PrefersChromePathEnv(t *testing.T) {
	if runtime.GOOS == osWindows {
		t.Skip("permission bits are not meaningful on windows")
	}

	fake := filepath.Join(t.TempDir(), "chrome")
	require.NoError(t, os.WriteFile(fake, []byte("#!/bin/sh\n"), 0o700))
	t.Setenv("CHROME_PATH", fake)

	require.Equal(t, fake, Detect())
}

func TestDetect_IgnoresNonExecutableChromePath(t *testing.T) {
	if runtime.GOOS == osWindows {
		t.Skip("permission bits are not meaningful on windows")
	}

	fake := filepath.Join(t.TempDir(), "not-runnable")
	require.NoError(t, os.WriteFile(fake, []byte("x"), 0o600))
	t.Setenv("CHROME_PATH", fake)

	require.NotEqual(t, fake, Detect())
}

func TestManagedBinaryRelPath_MatchesArchiveLayout(t *testing.T) {
	rel := managedBinaryRelPath()
	require.NotEmpty(t, rel)

	key, err := platformKey()
	require.NoError(t, err)
	require.Equal(t, "chrome-"+key, filepath.SplitList(rel)[0][:len("chrome-"+key)])
}

func TestCheckHost_RejectsUntrustedAndPlainHTTP(t *testing.T) {
	require.NoError(t, checkHost("https://storage.googleapis.com/chrome/x.zip"))
	require.Error(t, checkHost("https://evil.example.com/chrome.zip"))
	require.Error(t, checkHost("http://storage.googleapis.com/chrome.zip"))
}

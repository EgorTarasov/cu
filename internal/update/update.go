package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/EgorTarasov/cu/internal/version"
)

const (
	repoAPI       = "https://api.github.com/repos/EgorTarasov/cu/releases/latest"
	checkInterval = 7 * 24 * time.Hour
	stateFile     = "update-check"
	httpTimeout   = 3 * time.Second
	semverParts   = 3
	dirPerm       = 0o750
	filePerm      = 0o600
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

func stateFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cu-cli", stateFile), nil
}

func lastCheckTime() time.Time {
	path, err := stateFilePath()
	if err != nil {
		return time.Time{}
	}
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

func touchStateFile() {
	path, err := stateFilePath()
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(path), dirPerm)
	_ = os.WriteFile(path, nil, filePerm)
}

// CheckForUpdate fetches the latest release from GitHub at most once per
// checkInterval and prints a hint to stderr when a newer version is available.
// It is best-effort: any network/parse error silently no-ops.
func CheckForUpdate() {
	if version.Version == "dev" {
		return
	}

	if time.Since(lastCheckTime()) < checkInterval {
		return
	}

	touchStateFile()

	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, repoAPI, http.NoBody)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(version.Version, "v")

	if latest == current || !isNewer(latest, current) {
		return
	}

	fmt.Fprintf(os.Stderr, "\nДоступна новая версия: %s (текущая: %s)\n",
		release.TagName, version.Version)

	if runtime.GOOS == "windows" {
		fmt.Fprintf(os.Stderr, "Скачайте обновление: %s\n\n", release.HTMLURL)
		return
	}
	fmt.Fprintln(os.Stderr,
		"Обновите командой:\n"+
			"  curl -fsSL https://raw.githubusercontent.com/EgorTarasov/cu/main/install.sh | sh")
}

func isNewer(a, b string) bool {
	pa := splitVersion(a)
	pb := splitVersion(b)
	for i := range semverParts {
		if pa[i] != pb[i] {
			return pa[i] > pb[i]
		}
	}
	return false
}

func splitVersion(v string) [semverParts]int {
	var parts [semverParts]int
	segments := strings.SplitN(v, ".", semverParts)
	for i := 0; i < len(segments) && i < semverParts; i++ {
		// strip any pre-release/build suffix (e.g. "3-rc1") by taking the leading digits.
		s := segments[i]
		for j, c := range s {
			if c < '0' || c > '9' {
				s = s[:j]
				break
			}
		}
		n, _ := strconv.Atoi(s)
		parts[i] = n
	}
	return parts
}

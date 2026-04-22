package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"cu-sync/internal/version"
)

const (
	versionSegments = 3
	httpTimeout     = 3 * time.Second
)

const (
	repoAPI       = "https://api.github.com/repos/EgorTarasov/cu/releases/latest"
	checkInterval = 7 * 24 * time.Hour
	stateFile     = "update-check"
	updateCmd     = "Обновите командой:\n" +
		"  curl -fsSL https://raw.githubusercontent.com/EgorTarasov/cu/main/install.sh | sh\n\n"
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
	_ = os.MkdirAll(filepath.Dir(path), 0o750)
	_ = os.WriteFile(path, nil, 0o600)
}

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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, repoAPI, nil)
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

	if latest == current {
		return
	}

	if !isNewer(latest, current) {
		return
	}

	fmt.Fprintf(os.Stderr, "\nДоступна новая версия: %s (текущая: %s)\n", release.TagName, version.Version)

	if runtime.GOOS == "windows" {
		fmt.Fprintf(os.Stderr, "Скачайте обновление: %s\n\n", release.HTMLURL)
	} else {
		fmt.Fprint(os.Stderr, updateCmd)
	}
}

func isNewer(a, b string) bool {
	pa := splitVersion(a)
	pb := splitVersion(b)
	for i := range versionSegments {
		if pa[i] != pb[i] {
			return pa[i] > pb[i]
		}
	}
	return false
}

func splitVersion(v string) [versionSegments]int {
	var parts [versionSegments]int
	segments := strings.SplitN(v, ".", versionSegments)
	for i, s := range segments {
		if i >= versionSegments {
			break
		}
		for _, c := range s {
			if c >= '0' && c <= '9' {
				parts[i] = parts[i]*10 + int(c-'0') //nolint:gosec // simple digit parsing, no overflow risk
			} else {
				break
			}
		}
	}
	return parts
}

package update

import (
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
	repoAPI       = "https://api.github.com/repos/EgorTarasov/cu/releases/latest"
	checkInterval = 7 * 24 * time.Hour
	stateFile     = "update-check"
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
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, nil, 0o644)
}

func CheckForUpdate() {
	if version.Version == "dev" {
		return
	}

	if time.Since(lastCheckTime()) < checkInterval {
		return
	}

	touchStateFile()

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(repoAPI)
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
		fmt.Fprintf(os.Stderr, "Обновите командой:\n  curl -fsSL https://raw.githubusercontent.com/EgorTarasov/cu/main/install.sh | sh\n\n")
	}
}

func isNewer(a, b string) bool {
	pa := splitVersion(a)
	pb := splitVersion(b)
	for i := range 3 {
		if pa[i] != pb[i] {
			return pa[i] > pb[i]
		}
	}
	return false
}

func splitVersion(v string) [3]int {
	var parts [3]int
	segments := strings.SplitN(v, ".", 3)
	for i, s := range segments {
		if i >= 3 {
			break
		}
		for _, c := range s {
			if c >= '0' && c <= '9' {
				parts[i] = parts[i]*10 + int(c-'0')
			} else {
				break
			}
		}
	}
	return parts
}

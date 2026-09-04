package browser

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// versionsURL is the Chrome for Testing manifest maintained by the Chrome team.
	versionsURL = "https://googlechromelabs.github.io/chrome-for-testing/" +
		"last-known-good-versions-with-downloads.json"

	downloadTimeout = 15 * time.Minute

	// maxExtractBytes bounds zip expansion; a Chrome build unpacks to ~700MB.
	maxExtractBytes = 2 << 30

	// ownerReadWrite is OR-ed into archive modes so extracted files stay readable.
	ownerReadWrite = 0o600
)

// allowedDownloadHosts pins where a build may be fetched from, so a tampered
// manifest cannot redirect the download to an arbitrary host.
var allowedDownloadHosts = map[string]bool{
	"googlechromelabs.github.io": true,
	"storage.googleapis.com":     true,
	"edgedl.me.gvt1.com":         true,
}

type versionsManifest struct {
	Channels map[string]struct {
		Version   string `json:"version"`
		Downloads struct {
			Chrome []struct {
				Platform string `json:"platform"`
				URL      string `json:"url"`
			} `json:"chrome"`
		} `json:"downloads"`
	} `json:"channels"`
}

// platformKey maps the running host to a Chrome for Testing platform string.
func platformKey() (string, error) {
	switch runtime.GOOS {
	case osDarwin:
		if runtime.GOARCH == "arm64" {
			return "mac-arm64", nil
		}
		return "mac-x64", nil
	case "linux":
		if runtime.GOARCH != "amd64" {
			return "", fmt.Errorf("chrome for testing has no linux/%s build", runtime.GOARCH)
		}
		return "linux64", nil
	case osWindows:
		return "win64", nil
	default:
		return "", fmt.Errorf("unsupported platform %s/%s", runtime.GOOS, runtime.GOARCH)
	}
}

// managedBinaryRelPath is the executable's location inside the unpacked archive.
func managedBinaryRelPath() string {
	key, err := platformKey()
	if err != nil {
		return ""
	}
	switch runtime.GOOS {
	case osDarwin:
		return filepath.Join("chrome-"+key,
			"Google Chrome for Testing.app", "Contents", "MacOS",
			"Google Chrome for Testing")
	case osWindows:
		return filepath.Join("chrome-"+key, "chrome.exe")
	default:
		return filepath.Join("chrome-"+key, "chrome")
	}
}

// Download fetches the stable Chrome for Testing build for this host and
// unpacks it under ManagedDir. It reports progress to w and returns the
// executable path.
func Download(ctx context.Context, w io.Writer) (string, error) {
	key, err := platformKey()
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(ctx, downloadTimeout)
	defer cancel()

	fmt.Fprintln(w, "Определяем последнюю стабильную сборку Chrome for Testing...")
	buildURL, version, err := resolveBuild(ctx, key)
	if err != nil {
		return "", err
	}

	dir, err := ManagedDir()
	if err != nil {
		return "", err
	}
	if err := os.RemoveAll(dir); err != nil {
		return "", fmt.Errorf("не удалось очистить %s: %w", dir, err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}

	fmt.Fprintf(w, "Скачиваем Chrome %s (~170 МБ)...\n", version)
	archivePath := filepath.Join(dir, "chrome.zip")
	if err := fetchTo(ctx, buildURL, archivePath); err != nil {
		return "", err
	}

	fmt.Fprintln(w, "Распаковываем...")
	if err := unzip(archivePath, dir); err != nil {
		return "", err
	}
	if err := os.Remove(archivePath); err != nil {
		return "", err
	}

	bin := filepath.Join(dir, managedBinaryRelPath())
	if !isExecutable(bin) {
		return "", fmt.Errorf("после распаковки не найден исполняемый файл: %s", bin)
	}
	return bin, nil
}

func resolveBuild(ctx context.Context, platform string) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, versionsURL, nil)
	if err != nil {
		return "", "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("не удалось получить список версий Chrome: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("список версий Chrome вернул %s", resp.Status)
	}

	var manifest versionsManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return "", "", fmt.Errorf("не удалось разобрать список версий Chrome: %w", err)
	}

	stable, ok := manifest.Channels["Stable"]
	if !ok {
		return "", "", errors.New("в списке версий нет канала Stable")
	}

	for _, d := range stable.Downloads.Chrome {
		if d.Platform != platform {
			continue
		}
		if err := checkHost(d.URL); err != nil {
			return "", "", err
		}
		return d.URL, stable.Version, nil
	}

	return "", "", fmt.Errorf("для платформы %s сборка не найдена", platform)
}

func checkHost(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("некорректный URL сборки: %w", err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("URL сборки должен быть https, получен %q", u.Scheme)
	}
	if !allowedDownloadHosts[u.Hostname()] {
		return fmt.Errorf("URL сборки ведёт на недоверенный хост %q", u.Hostname())
	}
	return nil
}

func fetchTo(ctx context.Context, rawURL, dest string) error {
	if err := checkHost(rawURL); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("загрузка Chrome не удалась: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("загрузка Chrome вернула %s", resp.Status)
	}

	f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, io.LimitReader(resp.Body, maxExtractBytes)); err != nil {
		return fmt.Errorf("загрузка Chrome прервана: %w", err)
	}
	return nil
}

// unzip extracts src into dir, rejecting entries that escape dir and stopping
// if the archive expands past maxExtractBytes.
func unzip(src, dir string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	root, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	var written int64
	for _, f := range r.File {
		target := filepath.Join(root, filepath.Clean(f.Name))
		if target != root && !strings.HasPrefix(target, root+string(os.PathSeparator)) {
			return fmt.Errorf("архив содержит путь за пределами каталога: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o700); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}

		n, err := writeZipEntry(f, target, maxExtractBytes-written)
		if err != nil {
			return err
		}
		written += n
		if written >= maxExtractBytes {
			return errors.New("архив превышает допустимый размер распаковки")
		}
	}

	return nil
}

func writeZipEntry(f *zip.File, target string, budget int64) (int64, error) {
	rc, err := f.Open()
	if err != nil {
		return 0, err
	}
	defer rc.Close()

	mode := f.Mode().Perm() | ownerReadWrite
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	n, err := io.Copy(out, io.LimitReader(rc, budget))
	if err != nil {
		return n, err
	}
	return n, nil
}

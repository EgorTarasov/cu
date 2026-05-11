package cu

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func (c *Client) GetLongReadContent(ctx context.Context, longReadID int) (*MaterialsResponse, error) {
	endpoint := fmt.Sprintf(LongreadMaterialsEndpoint, longReadID) + "?limit=10000"
	return doJSON[MaterialsResponse](ctx, c, endpoint)
}

func (c *Client) GetLongread(ctx context.Context, longreadID int) (*Longread, error) {
	return doJSON[Longread](ctx, c, fmt.Sprintf(LongreadEndpoint, longreadID))
}

func (c *Client) GetDownloadLink(ctx context.Context, filename, version string) (string, error) {
	params := url.Values{}
	params.Add("filename", filename)
	params.Add("version", version)

	endpoint := DownloadLinkEndpoint + "?" + params.Encode()

	link, err := doJSON[DownloadLinkResponse](ctx, c, endpoint)
	if err != nil {
		return "", err
	}
	return link.URL, nil
}

func (c *Client) DownloadFile(ctx context.Context, material Material, destDir string) (string, error) {
	if material.Discriminator != "file" {
		return "", fmt.Errorf("material is not a file, got discriminator: %s", material.Discriminator)
	}

	downloadURL, err := c.GetDownloadLink(ctx, material.Filename, material.Version)
	if err != nil {
		return "", fmt.Errorf("failed to get download link: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create download request: %w", err)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status code: %d", res.StatusCode)
	}

	if err = os.MkdirAll(destDir, 0o750); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	filename := material.Content.Name
	if filename == "" {
		filename = filepath.Base(material.Filename)
	}

	destPath := filepath.Join(destDir, filename)

	outFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return destPath, nil
}

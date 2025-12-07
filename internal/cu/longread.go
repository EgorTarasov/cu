package cu

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)


// GetLongReadContent retrieves all materials for a longread
func (c *Client) GetLongReadContent(ctx context.Context, longReadID int) (*MaterialsResponse, error) {
	path := fmt.Sprintf("/api/micro-lms/longreads/%d/materials?limit=10000", longReadID)

	req, err := c.prepareRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	res, err := c.executeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var materials MaterialsResponse
	if err := json.NewDecoder(res.Body).Decode(&materials); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &materials, nil
}

func (c *Client) GetDownloadLink(ctx context.Context, filename, version string) (string, error) {
	params := url.Values{}
	params.Add("filename", filename)
	params.Add("version", version)

	path := fmt.Sprintf("/api/micro-lms/content/download-link?%s", params.Encode())

	req, err := c.prepareRequest(ctx, http.MethodGet, path)
	if err != nil {
		return "", fmt.Errorf("failed to prepare request: %w", err)
	}

	res, err := c.executeRequest(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var downloadLink DownloadLinkResponse
	if err := json.NewDecoder(res.Body).Decode(&downloadLink); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return downloadLink.URL, nil
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

	if err := os.MkdirAll(destDir, 0755); err != nil {
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

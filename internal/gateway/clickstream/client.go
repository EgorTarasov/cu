package clickstream

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	TrackEndpoint = "/track"

	// HeaderClientID etc. are spelled canonically: HTTP header names are
	// case-insensitive and Go canonicalizes them on Set anyway.
	HeaderClientID     = "Openpanel-Client-Id"
	HeaderClientSecret = "Openpanel-Client-Secret"
	HeaderSDKName      = "Openpanel-Sdk-Name"
	HeaderSDKVersion   = "Openpanel-Sdk-Version"

	sdkName = "cu-cli"

	DefaultTimeout   = 10 * time.Second
	maxErrorBodySize = 4096
)

// Client sends server-side analytics events to an OpenPanel instance.
type Client struct {
	httpClient   *http.Client
	baseURL      string
	clientID     string
	clientSecret string
	sdkVersion   string
	userAgent    string
}

func NewClient(baseURL, clientID, clientSecret, sdkVersion string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		baseURL:      baseURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		sdkVersion:   sdkVersion,
	}
}

// SetUserAgent overrides the User-Agent header. OpenPanel derives device,
// OS and session handling from it: UAs shaped like "name/1.0" are treated
// as server SDKs and never open sessions.
func (c *Client) SetUserAgent(userAgent string) {
	c.userAgent = userAgent
}

// send posts an Event to /track and decodes the response if any.
// OpenPanel replies 200 or 202 on success; server-side events usually
// get an empty body.
func (c *Client) send(ctx context.Context, event Event) (*TrackResponse, error) {
	if c.clientID == "" || c.clientSecret == "" {
		return nil, errors.New("client id and secret are required for server events")
	}

	body, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+TrackEndpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(HeaderClientID, c.clientID)
	req.Header.Set(HeaderClientSecret, c.clientSecret)
	req.Header.Set(HeaderSDKName, sdkName)
	if c.sdkVersion != "" {
		req.Header.Set(HeaderSDKVersion, c.sdkVersion)
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, bytes.TrimSpace(msg))
	}

	out := &TrackResponse{}
	if err = json.NewDecoder(resp.Body).Decode(out); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return out, nil
}

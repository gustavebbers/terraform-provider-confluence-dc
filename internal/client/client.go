// Package client is a minimal Confluence Data Center REST API client covering
// only the endpoints this provider needs: spaces, groups, and space permissions.
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client is a Confluence Data Center REST API client.
type Client struct {
	// baseURL is the Confluence base URL without a trailing slash, e.g.
	// https://confluence.example.com or https://confluence.example.com/confluence.
	baseURL string

	httpClient *http.Client

	// Exactly one of the following auth modes is configured.
	token              string // Personal Access Token, sent as an Authorization: Bearer header.
	username, password string // Basic auth.
}

// Config configures a new Client.
type Config struct {
	Host     string
	Token    string
	Username string
	Password string

	// SkipTLSVerify disables TLS certificate verification. Intended for DC
	// instances behind self-signed certificates in trusted networks.
	SkipTLSVerify bool
}

// New creates a Client from Config. Exactly one of Token or
// Username+Password must be set; callers are expected to have validated this
// already (see provider-level schema validation).
func New(cfg Config) (*Client, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("host must not be empty")
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.SkipTLSVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // opt-in via provider config
	}

	return &Client{
		baseURL:    strings.TrimRight(cfg.Host, "/"),
		httpClient: &http.Client{Transport: transport},
		token:      cfg.Token,
		username:   cfg.Username,
		password:   cfg.Password,
	}, nil
}

// APIError represents a non-2xx response from the Confluence REST API.
type APIError struct {
	StatusCode int
	Method     string
	Path       string
	Message    string
	Body       string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("confluence API error: %s %s returned %d: %s", e.Method, e.Path, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("confluence API error: %s %s returned %d: %s", e.Method, e.Path, e.StatusCode, e.Body)
}

// IsNotFound reports whether err is an APIError with a 404 status.
func IsNotFound(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.StatusCode == http.StatusNotFound
}

// confluenceErrorBody is the typical error envelope returned by the
// Confluence REST API.
type confluenceErrorBody struct {
	Message    string `json:"message"`
	Reason     string `json:"reason"`
	Data       map[string]any
	StatusCode int `json:"statusCode"`
}

// do executes an authenticated JSON request against the Confluence REST API.
// path must start with "/rest/api". If body is non-nil it is marshaled as
// the JSON request body. If out is non-nil, a 2xx response body is
// unmarshaled into it.
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	c.authenticate(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			Method:     method,
			Path:       path,
			Body:       string(respBody),
		}
		var errBody confluenceErrorBody
		if json.Unmarshal(respBody, &errBody) == nil {
			if errBody.Message != "" {
				apiErr.Message = errBody.Message
			} else if errBody.Reason != "" {
				apiErr.Message = errBody.Reason
			}
		}
		return apiErr
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decoding response body: %w", err)
		}
	}

	return nil
}

func (c *Client) authenticate(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
		return
	}
	req.SetBasicAuth(c.username, c.password)
}

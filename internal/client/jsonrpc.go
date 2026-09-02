package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Confluence Data Center's REST API does not implement write operations for
// groups or space permissions (verified empirically: POST /rest/api/group
// and POST/DELETE /rest/api/space/{key}/permission return 404/405 even on
// current Data Center releases). The only way to create/delete groups and
// grant/revoke space permissions is Confluence's legacy JSON-RPC API
// (confluenceservice-v2), deprecated by Atlassian since Confluence 5.5 but
// still present and functional. It authenticates through the same request
// filter chain as the REST API, so both Personal Access Tokens and HTTP
// Basic auth work against it unchanged.
const jsonRPCPath = "/rpc/json-rpc/confluenceservice-v2"

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type jsonRPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *jsonRPCError   `json:"error"`
}

// rpcCall invokes a method on Confluence's legacy JSON-RPC API. If out is
// non-nil, the "result" field of a successful response is unmarshaled into
// it. A JSON-RPC-level error (the "error" field being set) is returned as an
// *APIError so callers can handle it the same way as REST API errors.
func (c *Client) rpcCall(ctx context.Context, method string, params []any, out any) error {
	reqBody, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      1,
	})
	if err != nil {
		return fmt.Errorf("marshaling JSON-RPC request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+jsonRPCPath, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("building JSON-RPC request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	// Bypasses Confluence's XSRF check, which otherwise rejects JSON-RPC
	// POSTs made outside of a browser session.
	req.Header.Set("X-Atlassian-Token", "no-check")
	c.authenticate(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("performing JSON-RPC request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading JSON-RPC response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{StatusCode: resp.StatusCode, Method: "JSON-RPC", Path: method, Body: string(body)}
	}

	var rpcResp jsonRPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return fmt.Errorf("decoding JSON-RPC response: %w", err)
	}
	if rpcResp.Error != nil {
		return &APIError{StatusCode: rpcResp.Error.Code, Method: "JSON-RPC", Path: method, Message: rpcResp.Error.Message}
	}

	if out != nil && len(rpcResp.Result) > 0 {
		if err := json.Unmarshal(rpcResp.Result, out); err != nil {
			return fmt.Errorf("decoding JSON-RPC result: %w", err)
		}
	}

	return nil
}

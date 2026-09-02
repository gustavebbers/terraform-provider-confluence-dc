package client

import (
	"context"
	"net/url"
)

// Group is a Confluence group.
type Group struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// CreateGroup creates a new group via Confluence's legacy JSON-RPC API (see
// jsonrpc.go for why: the REST API's POST /rest/api/group returns 405 on
// real Data Center instances).
//
// Group creation delegates to Confluence's configured user directory. If the
// active directory is read-only (e.g. an LDAP or Crowd directory configured
// for read-only sync), this call fails with the error Confluence's API
// returns for that case; that is an environment/configuration concern, not a
// provider bug.
func (c *Client) CreateGroup(ctx context.Context, name string) (*Group, error) {
	if err := c.rpcCall(ctx, "addGroup", []any{name}, nil); err != nil {
		return nil, err
	}
	return &Group{Name: name, Type: "group"}, nil
}

// GetGroup fetches a single group by name. It returns an *APIError with
// StatusCode 404 (check with IsNotFound) if the group does not exist.
func (c *Client) GetGroup(ctx context.Context, name string) (*Group, error) {
	path := "/rest/api/group/" + url.PathEscape(name)
	var group Group
	if err := c.do(ctx, "GET", path, nil, &group); err != nil {
		return nil, err
	}
	return &group, nil
}

// DeleteGroup deletes a group by name via Confluence's legacy JSON-RPC API
// (see CreateGroup). It is idempotent: deleting an already-absent group is
// not an error, matching the behavior callers would get from a REST DELETE
// endpoint returning 404.
func (c *Client) DeleteGroup(ctx context.Context, name string) error {
	if _, err := c.GetGroup(ctx, name); err != nil {
		if IsNotFound(err) {
			return nil
		}
		return err
	}
	return c.rpcCall(ctx, "removeGroup", []any{name, nil}, nil)
}

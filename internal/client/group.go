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

// CreateGroup creates a new group via POST /rest/api/admin/group, falling
// back to Confluence's legacy JSON-RPC API (see jsonrpc.go) if that endpoint
// isn't found - the write endpoint under /rest/api/group itself returns 405,
// but /rest/api/admin/group is a separate, working endpoint on current Data
// Center releases.
//
// Group creation delegates to Confluence's configured user directory. If the
// active directory is read-only (e.g. an LDAP or Crowd directory configured
// for read-only sync), this call fails with the error Confluence's API
// returns for that case; that is an environment/configuration concern, not a
// provider bug.
func (c *Client) CreateGroup(ctx context.Context, name string) (*Group, error) {
	var group Group
	// The "type" field is required even though it's constant: Confluence
	// deserializes the request body as the polymorphic
	// com.atlassian.confluence.api.model.people.Group type via Jackson,
	// which needs it as a type-id discriminator and 400s with "missing
	// type id property 'type'" otherwise.
	body := map[string]string{"name": name, "type": "group"}
	err := c.do(ctx, "POST", "/rest/api/admin/group", body, &group)
	if err == nil {
		return &group, nil
	}
	if !isRouteNotFound(err) {
		return nil, err
	}

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

// DeleteGroup deletes a group by name via DELETE /rest/api/admin/group/{name}
// (see CreateGroup for why, and for the JSON-RPC fallback this also uses). It
// is idempotent: deleting an already-absent group is not an error, matching
// the behavior callers would get from a REST DELETE endpoint returning 404.
func (c *Client) DeleteGroup(ctx context.Context, name string) error {
	if _, err := c.GetGroup(ctx, name); err != nil {
		if IsNotFound(err) {
			return nil
		}
		return err
	}

	err := c.do(ctx, "DELETE", "/rest/api/admin/group/"+url.PathEscape(name), nil, nil)
	if err == nil || !isRouteNotFound(err) {
		return err
	}

	return c.rpcCall(ctx, "removeGroup", []any{name, nil}, nil)
}

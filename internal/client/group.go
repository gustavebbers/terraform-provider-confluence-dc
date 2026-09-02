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

// CreateGroup creates a new group.
//
// Group creation delegates to Confluence's configured user directory. If the
// active directory is read-only (e.g. an LDAP or Crowd directory configured
// for read-only sync), this call fails with the error Confluence's REST API
// returns for that case; that is an environment/configuration concern, not a
// provider bug.
func (c *Client) CreateGroup(ctx context.Context, name string) (*Group, error) {
	var group Group
	if err := c.do(ctx, "POST", "/rest/api/group", map[string]string{"name": name}, &group); err != nil {
		return nil, err
	}
	if group.Name == "" {
		group.Name = name
	}
	return &group, nil
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

// DeleteGroup deletes a group by name.
func (c *Client) DeleteGroup(ctx context.Context, name string) error {
	path := "/rest/api/group/" + url.PathEscape(name)
	return c.do(ctx, "DELETE", path, nil, nil)
}

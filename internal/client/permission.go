package client

import (
	"context"
	"fmt"
	"net/url"
)

// PermissionSubject identifies who a space permission applies to.
type PermissionSubject struct {
	Type       string `json:"type"` // "group" or "user"
	Identifier string `json:"identifier"`
}

// PermissionOperation identifies what a space permission grants.
//
// Valid Key/Target combinations (Confluence Data Center 9.1+), per the
// Confluence REST API:
//
//	read              / space
//	create            / page, blogpost, comment, attachment
//	delete            / page, blogpost, comment, attachment, space
//	export            / space
//	administer        / space
//	restrict_content  / space, page (add/remove restrictions)
//	archive           / page
type PermissionOperation struct {
	Key    string `json:"key"`
	Target string `json:"target"`
}

// SpacePermission is a single permission grant on a space.
type SpacePermission struct {
	ID        int64               `json:"id"`
	Subject   PermissionSubject   `json:"subject"`
	Operation PermissionOperation `json:"operation"`
}

type addSpacePermissionResponse struct {
	ID int64 `json:"id"`
}

// AddSpacePermission grants a permission on a space to a subject (typically
// a group). It requires Confluence Data Center 9.1+, which is the first
// version to expose this as a REST endpoint; earlier versions only support
// this through the deprecated JSON-RPC API, which this client does not
// implement.
func (c *Client) AddSpacePermission(ctx context.Context, spaceKey string, subject PermissionSubject, operation PermissionOperation) (*SpacePermission, error) {
	path := fmt.Sprintf("/rest/api/space/%s/permission", url.PathEscape(spaceKey))

	reqBody := map[string]any{
		"subject":   subject,
		"operation": operation,
	}

	var resp addSpacePermissionResponse
	if err := c.do(ctx, "POST", path, reqBody, &resp); err != nil {
		return nil, err
	}

	return &SpacePermission{
		ID:        resp.ID,
		Subject:   subject,
		Operation: operation,
	}, nil
}

// GetSpacePermission fetches a single space permission by ID by listing all
// permissions on the space and finding the matching ID. The Confluence REST
// API does not expose a get-by-id endpoint for individual space
// permissions, only list and create/delete.
func (c *Client) GetSpacePermission(ctx context.Context, spaceKey string, id int64) (*SpacePermission, error) {
	perms, err := c.ListSpacePermissions(ctx, spaceKey)
	if err != nil {
		return nil, err
	}
	for _, p := range perms {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, &APIError{
		StatusCode: 404,
		Method:     "GET",
		Path:       fmt.Sprintf("/rest/api/space/%s/permission", spaceKey),
		Message:    fmt.Sprintf("permission %d not found on space %s", id, spaceKey),
	}
}

type listSpacePermissionsResponse struct {
	Results []SpacePermission `json:"results"`
}

// ListSpacePermissions lists all permissions granted on a space.
func (c *Client) ListSpacePermissions(ctx context.Context, spaceKey string) ([]SpacePermission, error) {
	path := fmt.Sprintf("/rest/api/space/%s/permission", url.PathEscape(spaceKey))

	var resp listSpacePermissionsResponse
	if err := c.do(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Results, nil
}

// RemoveSpacePermission revokes a previously granted space permission by ID.
func (c *Client) RemoveSpacePermission(ctx context.Context, spaceKey string, id int64) error {
	path := fmt.Sprintf("/rest/api/space/%s/permission/%d", url.PathEscape(spaceKey), id)
	return c.do(ctx, "DELETE", path, nil, nil)
}

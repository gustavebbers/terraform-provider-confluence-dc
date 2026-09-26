package client

import (
	"context"
	"fmt"
	"net/url"
)

// GlobalPermissionOperation identifies a global (instance-wide) permission,
// as opposed to a permission scoped to one space (see PermissionOperation).
//
// Granting/revoking goes through
// PUT /rest/api/permissions/group/{groupName}/grant (and .../revoke), using
// the same operationKey/targetType vocabulary as space permissions. There is
// no legacy JSON-RPC fallback for these: global/system permission
// management was never exposed by the legacy ConfluenceService API, so on
// instances old enough to lack this REST endpoint there is simply no way to
// manage global permissions through this provider. The pairs below are the
// complete set of grantable global permissions, per Confluence's REST API
// documentation for grantPermissionsToGroup:
//
//	Key         Target
//	use         application
//	administer  application
//	administer  system
//	create      personal_space
//	create      space
type GlobalPermissionOperation struct {
	Key    string
	Target string
}

var validGlobalPermissionOperations = map[GlobalPermissionOperation]bool{
	{Key: "use", Target: "application"}:        true,
	{Key: "administer", Target: "application"}: true,
	{Key: "administer", Target: "system"}:      true,
	{Key: "create", Target: "personal_space"}:  true,
	{Key: "create", Target: "space"}:           true,
}

func validateGlobalPermissionOperation(op GlobalPermissionOperation) error {
	if !validGlobalPermissionOperations[op] {
		return fmt.Errorf("unsupported operation_key/operation_target combination: %q/%q", op.Key, op.Target)
	}
	return nil
}

func globalPermissionGrantRevokePath(groupName, action string) string {
	return fmt.Sprintf("/rest/api/permissions/group/%s/%s", url.PathEscape(groupName), action)
}

// AddGlobalPermission grants a global permission to a group via
// PUT /rest/api/permissions/group/{groupName}/grant.
func (c *Client) AddGlobalPermission(ctx context.Context, groupName string, operation GlobalPermissionOperation) error {
	if err := validateGlobalPermissionOperation(operation); err != nil {
		return err
	}
	body := []restOperationDescription{{TargetType: operation.Target, OperationKey: operation.Key}}
	return c.do(ctx, "PUT", globalPermissionGrantRevokePath(groupName, "grant"), body, nil)
}

// RemoveGlobalPermission revokes a global permission from a group via
// PUT /rest/api/permissions/group/{groupName}/revoke. Revoking a permission
// that is not currently granted is not an error (the endpoint is
// idempotent).
func (c *Client) RemoveGlobalPermission(ctx context.Context, groupName string, operation GlobalPermissionOperation) error {
	if err := validateGlobalPermissionOperation(operation); err != nil {
		return err
	}
	body := []restOperationDescription{{TargetType: operation.Target, OperationKey: operation.Key}}
	return c.do(ctx, "PUT", globalPermissionGrantRevokePath(groupName, "revoke"), body, nil)
}

// globalPermissionEntry is the wire shape of one element returned by
// GET /rest/api/permissions/group/{groupName}.
type globalPermissionEntry struct {
	Operation struct {
		Key    string `json:"operationKey"`
		Target string `json:"targetType"`
	} `json:"operation"`
}

// ListGlobalPermissions lists all global permissions granted to a group.
func (c *Client) ListGlobalPermissions(ctx context.Context, groupName string) ([]GlobalPermissionOperation, error) {
	path := fmt.Sprintf("/rest/api/permissions/group/%s", url.PathEscape(groupName))

	var entries []globalPermissionEntry
	if err := c.do(ctx, "GET", path, nil, &entries); err != nil {
		return nil, err
	}

	ops := make([]GlobalPermissionOperation, 0, len(entries))
	for _, e := range entries {
		ops = append(ops, GlobalPermissionOperation{Key: e.Operation.Key, Target: e.Operation.Target})
	}
	return ops, nil
}

// GetGlobalPermission fetches a single global permission grant by looking it
// up in the full list of permissions granted to the group. It returns an
// *APIError with StatusCode 404 (check with IsNotFound) if no matching grant
// exists.
func (c *Client) GetGlobalPermission(ctx context.Context, groupName string, operation GlobalPermissionOperation) (*GlobalPermissionOperation, error) {
	ops, err := c.ListGlobalPermissions(ctx, groupName)
	if err != nil {
		return nil, err
	}
	for _, op := range ops {
		if op == operation {
			return &op, nil
		}
	}
	return nil, &APIError{
		StatusCode: 404,
		Method:     "GET",
		Path:       fmt.Sprintf("/rest/api/permissions/group/%s", groupName),
		Message: fmt.Sprintf("no %s/%s global permission granted to group %q",
			operation.Key, operation.Target, groupName),
	}
}

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
// Confluence Data Center's REST API only supports *reading* space
// permissions; granting and revoking them requires the legacy JSON-RPC API
// (see jsonrpc.go), whose permission model predates and differs from the
// REST API's operation-key/target vocabulary. The pairs below are the
// complete, verified set of REST (Key, Target) pairs and the legacy
// JSON-RPC permission type each one corresponds to; this is the full set of
// grantable space permissions, empirically confirmed against a live
// Confluence Data Center 9.2 instance (there is no way to enumerate this
// mapping from documentation, since Atlassian does not publish it):
//
//	Key          Target       legacy JSON-RPC type
//	read         space        VIEWSPACE
//	create       page         EDITSPACE
//	delete       page         REMOVEPAGE
//	create       blogpost     EDITBLOG
//	delete       blogpost     REMOVEBLOG
//	create       comment      COMMENT
//	delete       comment      REMOVECOMMENT
//	create       attachment   CREATEATTACHMENT
//	delete       attachment   REMOVEATTACHMENT
//	delete_own   space        REMOVEOWNCONTENT
//	delete_mail  space        REMOVEMAIL
//	export       space        EXPORTSPACE
//	restrict     space        SETPAGEPERMISSIONS
//	administer   space        SETSPACEPERMISSIONS
type PermissionOperation struct {
	Key    string `json:"key"`
	Target string `json:"target"`
}

// legacyPermissionType maps a (Key, Target) pair to the legacy JSON-RPC
// permission type name required by addPermissionToSpace /
// removePermissionFromSpace. See the PermissionOperation doc comment for
// how this table was derived.
var legacyPermissionType = map[PermissionOperation]string{
	{Key: "read", Target: "space"}:        "VIEWSPACE",
	{Key: "create", Target: "page"}:       "EDITSPACE",
	{Key: "delete", Target: "page"}:       "REMOVEPAGE",
	{Key: "create", Target: "blogpost"}:   "EDITBLOG",
	{Key: "delete", Target: "blogpost"}:   "REMOVEBLOG",
	{Key: "create", Target: "comment"}:    "COMMENT",
	{Key: "delete", Target: "comment"}:    "REMOVECOMMENT",
	{Key: "create", Target: "attachment"}: "CREATEATTACHMENT",
	{Key: "delete", Target: "attachment"}: "REMOVEATTACHMENT",
	{Key: "delete_own", Target: "space"}:  "REMOVEOWNCONTENT",
	{Key: "delete_mail", Target: "space"}: "REMOVEMAIL",
	{Key: "export", Target: "space"}:      "EXPORTSPACE",
	{Key: "restrict", Target: "space"}:    "SETPAGEPERMISSIONS",
	{Key: "administer", Target: "space"}:  "SETSPACEPERMISSIONS",
}

func toLegacyPermissionType(op PermissionOperation) (string, error) {
	t, ok := legacyPermissionType[op]
	if !ok {
		return "", fmt.Errorf("unsupported operation_key/operation_target combination: %q/%q", op.Key, op.Target)
	}
	return t, nil
}

// SpacePermission is a single permission grant on a space.
type SpacePermission struct {
	Subject   PermissionSubject   `json:"subject"`
	Operation PermissionOperation `json:"operation"`
}

// AddSpacePermission grants a permission on a space to a subject (only
// "group" subjects are exercised by this provider, though the underlying
// JSON-RPC method also accepts usernames).
func (c *Client) AddSpacePermission(ctx context.Context, spaceKey string, subject PermissionSubject, operation PermissionOperation) (*SpacePermission, error) {
	legacyType, err := toLegacyPermissionType(operation)
	if err != nil {
		return nil, err
	}

	if err := c.rpcCall(ctx, "addPermissionToSpace", []any{legacyType, subject.Identifier, spaceKey}, nil); err != nil {
		return nil, err
	}

	return &SpacePermission{Subject: subject, Operation: operation}, nil
}

// RemoveSpacePermission revokes a previously granted space permission.
// Revoking a permission that is not currently granted is not an error
// (Confluence's underlying JSON-RPC method is idempotent).
func (c *Client) RemoveSpacePermission(ctx context.Context, spaceKey string, subject PermissionSubject, operation PermissionOperation) error {
	legacyType, err := toLegacyPermissionType(operation)
	if err != nil {
		return err
	}

	return c.rpcCall(ctx, "removePermissionFromSpace", []any{legacyType, subject.Identifier, spaceKey}, nil)
}

// GetSpacePermission fetches a single space permission grant by looking it
// up in the full list of permissions on the space. It returns an *APIError
// with StatusCode 404 (check with IsNotFound) if no matching grant exists.
func (c *Client) GetSpacePermission(ctx context.Context, spaceKey string, subject PermissionSubject, operation PermissionOperation) (*SpacePermission, error) {
	perms, err := c.ListSpacePermissions(ctx, spaceKey)
	if err != nil {
		return nil, err
	}
	for _, p := range perms {
		if p.Subject.Type == subject.Type && p.Subject.Identifier == subject.Identifier && p.Operation == operation {
			return &p, nil
		}
	}
	return nil, &APIError{
		StatusCode: 404,
		Method:     "GET",
		Path:       fmt.Sprintf("/rest/api/space/%s/permissions", spaceKey),
		Message: fmt.Sprintf("no %s/%s permission granted to %s %q on space %s",
			operation.Key, operation.Target, subject.Type, subject.Identifier, spaceKey),
	}
}

// restSpacePermissionEntry is the wire shape returned by
// GET /rest/api/space/{spaceKey}/permissions (note: plural; the singular
// /permission sub-resource used by Confluence Cloud does not exist on Data
// Center). It uses different field names than the request bodies accepted
// by the (nonfunctional, on Data Center) POST/DELETE variants of this
// endpoint.
type restSpacePermissionEntry struct {
	Operation struct {
		Key    string `json:"operationKey"`
		Target string `json:"targetType"`
	} `json:"operation"`
	Subject struct {
		Type string `json:"type"`
		Name string `json:"name"`
	} `json:"subject"`
}

// ListSpacePermissions lists all permissions granted on a space.
func (c *Client) ListSpacePermissions(ctx context.Context, spaceKey string) ([]SpacePermission, error) {
	path := fmt.Sprintf("/rest/api/space/%s/permissions", url.PathEscape(spaceKey))

	var entries []restSpacePermissionEntry
	if err := c.do(ctx, "GET", path, nil, &entries); err != nil {
		return nil, err
	}

	perms := make([]SpacePermission, 0, len(entries))
	for _, e := range entries {
		perms = append(perms, SpacePermission{
			Subject: PermissionSubject{
				Type:       e.Subject.Type,
				Identifier: e.Subject.Name,
			},
			Operation: PermissionOperation{
				Key:    e.Operation.Key,
				Target: e.Operation.Target,
			},
		})
	}
	return perms, nil
}

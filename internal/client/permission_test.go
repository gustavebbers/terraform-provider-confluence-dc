package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddSpacePermission_Success(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result": true, "jsonrpc": "2.0", "id": 1}`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	perm, err := c.AddSpacePermission(context.Background(), "ENG",
		PermissionSubject{Type: "group", Identifier: "developers"},
		PermissionOperation{Key: "read", Target: "space"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodPost)
	}
	if gotPath != jsonRPCPath {
		t.Errorf("path = %q, want %q", gotPath, jsonRPCPath)
	}
	if gotBody["method"] != "addPermissionToSpace" {
		t.Errorf("RPC method = %v, want %q", gotBody["method"], "addPermissionToSpace")
	}
	params, ok := gotBody["params"].([]any)
	if !ok || len(params) != 3 {
		t.Fatalf("params = %+v, want 3 elements", gotBody["params"])
	}
	if params[0] != "VIEWSPACE" || params[1] != "developers" || params[2] != "ENG" {
		t.Errorf("params = %+v, want [VIEWSPACE developers ENG]", params)
	}

	if perm.Subject.Identifier != "developers" {
		t.Errorf("perm.Subject = %+v, want identifier=developers", perm.Subject)
	}
	if perm.Operation.Key != "read" || perm.Operation.Target != "space" {
		t.Errorf("perm.Operation = %+v, want key=read target=space", perm.Operation)
	}
}

func TestAddSpacePermission_UnsupportedCombination(t *testing.T) {
	c, err := New(Config{Host: "http://example.invalid", Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.AddSpacePermission(context.Background(), "ENG",
		PermissionSubject{Type: "group", Identifier: "developers"},
		PermissionOperation{Key: "bogus", Target: "space"},
	)
	if err == nil {
		t.Fatal("expected an error for an unsupported operation_key/operation_target combination, got nil")
	}
}

func TestRemoveSpacePermission_Success(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result": true, "jsonrpc": "2.0", "id": 1}`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = c.RemoveSpacePermission(context.Background(), "ENG",
		PermissionSubject{Type: "group", Identifier: "developers"},
		PermissionOperation{Key: "administer", Target: "space"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotBody["method"] != "removePermissionFromSpace" {
		t.Errorf("RPC method = %v, want %q", gotBody["method"], "removePermissionFromSpace")
	}
	params, ok := gotBody["params"].([]any)
	if !ok || len(params) != 3 || params[0] != "SETSPACEPERMISSIONS" || params[1] != "developers" || params[2] != "ENG" {
		t.Errorf("params = %+v, want [SETSPACEPERMISSIONS developers ENG]", gotBody["params"])
	}
}

func TestListSpacePermissions_Success(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{"operation": {"operationKey": "read", "targetType": "space"}, "subject": {"type": "group", "name": "developers"}, "spaceKey": "ENG"},
			{"operation": {"operationKey": "administer", "targetType": "space"}, "subject": {"type": "group", "name": "admins"}, "spaceKey": "ENG"}
		]`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	perms, err := c.ListSpacePermissions(context.Background(), "ENG")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodGet)
	}
	if gotPath != "/rest/api/space/ENG/permissions" {
		t.Errorf("path = %q, want %q", gotPath, "/rest/api/space/ENG/permissions")
	}

	if len(perms) != 2 {
		t.Fatalf("len(perms) = %d, want 2", len(perms))
	}
	if perms[0].Subject.Identifier != "developers" || perms[0].Operation.Key != "read" {
		t.Errorf("perms[0] = %+v, unexpected", perms[0])
	}
	if perms[1].Subject.Identifier != "admins" || perms[1].Operation.Key != "administer" {
		t.Errorf("perms[1] = %+v, unexpected", perms[1])
	}
}

func TestListSpacePermissions_SpaceKeyIsURLEscaped(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := c.ListSpacePermissions(context.Background(), "A B"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/rest/api/space/A%20B/permissions" {
		t.Errorf("path = %q, want %q", gotPath, "/rest/api/space/A%20B/permissions")
	}
}

func TestGetSpacePermission_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{"operation": {"operationKey": "read", "targetType": "space"}, "subject": {"type": "group", "name": "developers"}, "spaceKey": "ENG"},
			{"operation": {"operationKey": "administer", "targetType": "space"}, "subject": {"type": "group", "name": "admins"}, "spaceKey": "ENG"}
		]`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	perm, err := c.GetSpacePermission(context.Background(), "ENG",
		PermissionSubject{Type: "group", Identifier: "admins"},
		PermissionOperation{Key: "administer", Target: "space"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if perm.Subject.Identifier != "admins" {
		t.Errorf("perm = %+v, want identifier=admins", perm)
	}
}

func TestGetSpacePermission_NotFoundInList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{"operation": {"operationKey": "read", "targetType": "space"}, "subject": {"type": "group", "name": "developers"}, "spaceKey": "ENG"}
		]`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	perm, err := c.GetSpacePermission(context.Background(), "ENG",
		PermissionSubject{Type: "group", Identifier: "admins"},
		PermissionOperation{Key: "administer", Target: "space"},
	)
	if perm != nil {
		t.Errorf("expected nil permission, got %+v", perm)
	}
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound(err) to be true (synthesized 404), got err=%v", err)
	}
}

func TestGetSpacePermission_PropagatesListError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.GetSpacePermission(context.Background(), "ENG",
		PermissionSubject{Type: "group", Identifier: "admins"},
		PermissionOperation{Key: "administer", Target: "space"},
	)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusInternalServerError)
	}
}

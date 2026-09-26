package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddGlobalPermission_Success(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody []restOperationDescription
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = c.AddGlobalPermission(context.Background(), "confluence-license",
		GlobalPermissionOperation{Key: "use", Target: "application"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPut {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodPut)
	}
	if gotPath != "/rest/api/permissions/group/confluence-license/grant" {
		t.Errorf("path = %q, want %q", gotPath, "/rest/api/permissions/group/confluence-license/grant")
	}
	if len(gotBody) != 1 || gotBody[0].TargetType != "application" || gotBody[0].OperationKey != "use" {
		t.Errorf("body = %+v, want [{targetType:application operationKey:use}]", gotBody)
	}
}

func TestAddGlobalPermission_UnsupportedCombination(t *testing.T) {
	c, err := New(Config{Host: "http://example.invalid", Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = c.AddGlobalPermission(context.Background(), "confluence-license",
		GlobalPermissionOperation{Key: "bogus", Target: "application"},
	)
	if err == nil {
		t.Fatal("expected an error for an unsupported operation_key/operation_target combination, got nil")
	}
}

func TestAddGlobalPermission_RESTError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"group not found"}`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = c.AddGlobalPermission(context.Background(), "nonexistent",
		GlobalPermissionOperation{Key: "use", Target: "application"},
	)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Message != "group not found" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "group not found")
	}
}

func TestRemoveGlobalPermission_Success(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody []restOperationDescription
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = c.RemoveGlobalPermission(context.Background(), "cugs-admin",
		GlobalPermissionOperation{Key: "administer", Target: "system"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPut {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodPut)
	}
	if gotPath != "/rest/api/permissions/group/cugs-admin/revoke" {
		t.Errorf("path = %q, want %q", gotPath, "/rest/api/permissions/group/cugs-admin/revoke")
	}
	if len(gotBody) != 1 || gotBody[0].TargetType != "system" || gotBody[0].OperationKey != "administer" {
		t.Errorf("body = %+v, want [{targetType:system operationKey:administer}]", gotBody)
	}
}

func TestListGlobalPermissions_Success(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{"operation": {"operationKey": "use", "targetType": "application"}, "subject": {"type": "group"}},
			{"operation": {"operationKey": "administer", "targetType": "system"}, "subject": {"type": "group"}}
		]`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ops, err := c.ListGlobalPermissions(context.Background(), "confluence-license")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodGet)
	}
	if gotPath != "/rest/api/permissions/group/confluence-license" {
		t.Errorf("path = %q, want %q", gotPath, "/rest/api/permissions/group/confluence-license")
	}
	if len(ops) != 2 {
		t.Fatalf("len(ops) = %d, want 2", len(ops))
	}
	if ops[0].Key != "use" || ops[0].Target != "application" {
		t.Errorf("ops[0] = %+v, unexpected", ops[0])
	}
	if ops[1].Key != "administer" || ops[1].Target != "system" {
		t.Errorf("ops[1] = %+v, unexpected", ops[1])
	}
}

func TestGetGlobalPermission_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{"operation": {"operationKey": "use", "targetType": "application"}, "subject": {"type": "group"}}
		]`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	op, err := c.GetGlobalPermission(context.Background(), "confluence-license",
		GlobalPermissionOperation{Key: "use", Target: "application"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if op.Key != "use" || op.Target != "application" {
		t.Errorf("op = %+v, want key=use target=application", op)
	}
}

func TestGetGlobalPermission_NotFoundInList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.GetGlobalPermission(context.Background(), "confluence-license",
		GlobalPermissionOperation{Key: "use", Target: "application"},
	)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound(err) to be true (synthesized 404), got err=%v", err)
	}
}

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
		_, _ = w.Write([]byte(`{"id": 42}`))
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
	if gotPath != "/rest/api/space/ENG/permission" {
		t.Errorf("path = %q, want %q", gotPath, "/rest/api/space/ENG/permission")
	}

	subject, ok := gotBody["subject"].(map[string]any)
	if !ok {
		t.Fatalf("request body missing subject object: %+v", gotBody)
	}
	if subject["type"] != "group" || subject["identifier"] != "developers" {
		t.Errorf("subject = %+v, want type=group identifier=developers", subject)
	}

	operation, ok := gotBody["operation"].(map[string]any)
	if !ok {
		t.Fatalf("request body missing operation object: %+v", gotBody)
	}
	if operation["key"] != "read" || operation["target"] != "space" {
		t.Errorf("operation = %+v, want key=read target=space", operation)
	}

	if perm.ID != 42 {
		t.Errorf("perm.ID = %d, want %d", perm.ID, 42)
	}
	if perm.Subject.Type != "group" || perm.Subject.Identifier != "developers" {
		t.Errorf("perm.Subject = %+v, want type=group identifier=developers", perm.Subject)
	}
	if perm.Operation.Key != "read" || perm.Operation.Target != "space" {
		t.Errorf("perm.Operation = %+v, want key=read target=space", perm.Operation)
	}
}

func TestAddSpacePermission_SpaceKeyIsURLEscaped(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id": 1}`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.AddSpacePermission(context.Background(), "A B",
		PermissionSubject{Type: "group", Identifier: "developers"},
		PermissionOperation{Key: "read", Target: "space"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/rest/api/space/A%20B/permission" {
		t.Errorf("path = %q, want %q", gotPath, "/rest/api/space/A%20B/permission")
	}
}

func TestListSpacePermissions_Success(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"results": [
				{"id": 1, "subject": {"type": "group", "identifier": "developers"}, "operation": {"key": "read", "target": "space"}},
				{"id": 2, "subject": {"type": "group", "identifier": "admins"}, "operation": {"key": "administer", "target": "space"}}
			]
		}`))
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
	if gotPath != "/rest/api/space/ENG/permission" {
		t.Errorf("path = %q, want %q", gotPath, "/rest/api/space/ENG/permission")
	}

	if len(perms) != 2 {
		t.Fatalf("len(perms) = %d, want 2", len(perms))
	}
	if perms[0].ID != 1 || perms[0].Subject.Identifier != "developers" {
		t.Errorf("perms[0] = %+v, unexpected", perms[0])
	}
	if perms[1].ID != 2 || perms[1].Subject.Identifier != "admins" {
		t.Errorf("perms[1] = %+v, unexpected", perms[1])
	}
}

func TestGetSpacePermission_FoundByID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"results": [
				{"id": 1, "subject": {"type": "group", "identifier": "developers"}, "operation": {"key": "read", "target": "space"}},
				{"id": 2, "subject": {"type": "group", "identifier": "admins"}, "operation": {"key": "administer", "target": "space"}}
			]
		}`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	perm, err := c.GetSpacePermission(context.Background(), "ENG", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if perm.ID != 2 || perm.Subject.Identifier != "admins" {
		t.Errorf("perm = %+v, want id=2 identifier=admins", perm)
	}
}

func TestGetSpacePermission_NotFoundInList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"results": [
				{"id": 1, "subject": {"type": "group", "identifier": "developers"}, "operation": {"key": "read", "target": "space"}}
			]
		}`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	perm, err := c.GetSpacePermission(context.Background(), "ENG", 999)
	if perm != nil {
		t.Errorf("expected nil permission, got %+v", perm)
	}
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound(err) to be true (synthesized 404), got err=%v", err)
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
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

	_, err = c.GetSpacePermission(context.Background(), "ENG", 1)
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

func TestRemoveSpacePermission_Success(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.RemoveSpacePermission(context.Background(), "ENG", 42); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodDelete)
	}
	if gotPath != "/rest/api/space/ENG/permission/42" {
		t.Errorf("path = %q, want %q", gotPath, "/rest/api/space/ENG/permission/42")
	}
}

func TestRemoveSpacePermission_SpaceKeyIsURLEscaped(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.RemoveSpacePermission(context.Background(), "A B", 7); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/rest/api/space/A%20B/permission/7" {
		t.Errorf("path = %q, want %q", gotPath, "/rest/api/space/A%20B/permission/7")
	}
}

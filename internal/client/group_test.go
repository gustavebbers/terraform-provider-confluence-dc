package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateGroup_Success(t *testing.T) {
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

	group, err := c.CreateGroup(context.Background(), "developers")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodPost)
	}
	if gotPath != jsonRPCPath {
		t.Errorf("path = %q, want %q", gotPath, jsonRPCPath)
	}
	if gotBody["method"] != "addGroup" {
		t.Errorf("request method = %v, want %q", gotBody["method"], "addGroup")
	}
	params, ok := gotBody["params"].([]any)
	if !ok || len(params) != 1 || params[0] != "developers" {
		t.Errorf("request params = %+v, want [\"developers\"]", gotBody["params"])
	}
	if group.Name != "developers" {
		t.Errorf("group.Name = %q, want %q", group.Name, "developers")
	}
}

func TestCreateGroup_RPCError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"error": {"code": 500, "message": "directory is read-only"}, "jsonrpc": "2.0", "id": 1}`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.CreateGroup(context.Background(), "developers")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Message != "directory is read-only" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "directory is read-only")
	}
}

func TestGetGroup_Success(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"confluence-users","type":"group"}`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	group, err := c.GetGroup(context.Background(), "confluence-users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodGet)
	}
	if gotPath != "/rest/api/group/confluence-users" {
		t.Errorf("path = %q, want %q", gotPath, "/rest/api/group/confluence-users")
	}
	if group.Name != "confluence-users" {
		t.Errorf("group.Name = %q, want %q", group.Name, "confluence-users")
	}
}

func TestGetGroup_NameIsURLEscaped(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"a b/c","type":"group"}`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := c.GetGroup(context.Background(), "a b/c"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/rest/api/group/a%20b%2Fc" {
		t.Errorf("path = %q, want %q", gotPath, "/rest/api/group/a%20b%2Fc")
	}
}

func TestGetGroup_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Group does not exist"}`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.GetGroup(context.Background(), "nope")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound(err) to be true, got err=%v", err)
	}
}

func TestDeleteGroup_Success(t *testing.T) {
	var gotRPCMethod string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/rest/api/group/developers":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"name":"developers","type":"group"}`))
		case r.URL.Path == jsonRPCPath:
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &gotBody)
			gotRPCMethod, _ = gotBody["method"].(string)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result": true, "jsonrpc": "2.0", "id": 1}`))
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.DeleteGroup(context.Background(), "developers"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotRPCMethod != "removeGroup" {
		t.Errorf("RPC method = %q, want %q", gotRPCMethod, "removeGroup")
	}
	params, ok := gotBody["params"].([]any)
	if !ok || len(params) != 2 || params[0] != "developers" {
		t.Errorf("params = %+v, want [\"developers\", nil]", gotBody["params"])
	}
}

func TestDeleteGroup_AlreadyAbsentIsNotAnError(t *testing.T) {
	rpcCalled := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/rest/api/group/developers":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"Group does not exist"}`))
		case r.URL.Path == jsonRPCPath:
			rpcCalled = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result": true, "jsonrpc": "2.0", "id": 1}`))
		}
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.DeleteGroup(context.Background(), "developers"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rpcCalled {
		t.Error("removeGroup should not have been called for an already-absent group")
	}
}

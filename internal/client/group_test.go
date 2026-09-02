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
	var gotBody map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"developers","type":"group"}`))
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
	if gotPath != "/rest/api/group" {
		t.Errorf("path = %q, want %q", gotPath, "/rest/api/group")
	}
	if gotBody["name"] != "developers" {
		t.Errorf("request body name = %q, want %q", gotBody["name"], "developers")
	}
	if group.Name != "developers" {
		t.Errorf("group.Name = %q, want %q", group.Name, "developers")
	}
	if group.Type != "group" {
		t.Errorf("group.Type = %q, want %q", group.Type, "group")
	}
}

func TestCreateGroup_FallsBackToRequestedNameWhenResponseOmitsIt(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
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
	if group.Name != "developers" {
		t.Errorf("group.Name = %q, want fallback %q", group.Name, "developers")
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

	if err := c.DeleteGroup(context.Background(), "developers"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodDelete)
	}
	if gotPath != "/rest/api/group/developers" {
		t.Errorf("path = %q, want %q", gotPath, "/rest/api/group/developers")
	}
}

func TestDeleteGroup_ToleratesEmpty2xxBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// no body written at all
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.DeleteGroup(context.Background(), "developers"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

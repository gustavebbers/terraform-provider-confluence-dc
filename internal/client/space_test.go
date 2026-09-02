package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSpace_Success(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": 12345,
			"key": "ENG",
			"name": "Engineering",
			"type": "global",
			"description": {
				"plain": {
					"value": "The engineering team space.",
					"representation": "plain"
				}
			}
		}`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	space, err := c.GetSpace(context.Background(), "ENG")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/rest/api/space/ENG" {
		t.Errorf("request path = %q, want %q", gotPath, "/rest/api/space/ENG")
	}
	if gotQuery != "expand=description.plain" {
		t.Errorf("request query = %q, want %q", gotQuery, "expand=description.plain")
	}

	if space.ID != 12345 {
		t.Errorf("ID = %d, want %d", space.ID, 12345)
	}
	if space.Key != "ENG" {
		t.Errorf("Key = %q, want %q", space.Key, "ENG")
	}
	if space.Name != "Engineering" {
		t.Errorf("Name = %q, want %q", space.Name, "Engineering")
	}
	if space.Type != "global" {
		t.Errorf("Type = %q, want %q", space.Type, "global")
	}
	if space.Description.Plain.Value != "The engineering team space." {
		t.Errorf("Description.Plain.Value = %q, want %q", space.Description.Plain.Value, "The engineering team space.")
	}
}

func TestGetSpace_SpaceKeyIsURLEscaped(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":1,"key":"A B","name":"n","type":"global"}`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := c.GetSpace(context.Background(), "A B"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/rest/api/space/A%20B" {
		t.Errorf("request path = %q, want %q", gotPath, "/rest/api/space/A%20B")
	}
}

func TestGetSpace_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Space does not exist"}`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	space, err := c.GetSpace(context.Background(), "NOPE")
	if space != nil {
		t.Errorf("expected nil space on error, got %+v", space)
	}
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound(err) to be true, got err=%v", err)
	}
}

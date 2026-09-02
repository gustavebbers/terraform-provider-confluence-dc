package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNew_EmptyHost(t *testing.T) {
	_, err := New(Config{Host: ""})
	if err == nil {
		t.Fatal("expected an error for empty host, got nil")
	}
	if !strings.Contains(err.Error(), "host") {
		t.Errorf("expected error to mention host, got: %v", err)
	}
}

func TestNew_TrimsTrailingSlash(t *testing.T) {
	c, err := New(Config{Host: "https://confluence.example.com/", Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.baseURL != "https://confluence.example.com" {
		t.Errorf("expected trailing slash to be trimmed, got baseURL=%q", c.baseURL)
	}
}

func TestAuthenticate_TokenSendsBearer(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "my-pat-token"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.do(context.Background(), "GET", "/rest/api/space/FOO", nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "Bearer my-pat-token"
	if gotAuth != want {
		t.Errorf("Authorization header = %q, want %q", gotAuth, want)
	}
}

func TestAuthenticate_BasicAuth(t *testing.T) {
	var gotUser, gotPass string
	var gotOK bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPass, gotOK = r.BasicAuth()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Username: "alice", Password: "s3cret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.do(context.Background(), "GET", "/rest/api/space/FOO", nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !gotOK {
		t.Fatal("expected a valid Basic auth header, got none")
	}
	if gotUser != "alice" || gotPass != "s3cret" {
		t.Errorf("BasicAuth = (%q, %q), want (%q, %q)", gotUser, gotPass, "alice", "s3cret")
	}
}

func TestDo_NonSuccessStatusReturnsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Space does not exist"}`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = c.do(context.Background(), "GET", "/rest/api/space/NOPE", nil, nil)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusNotFound)
	}
	if apiErr.Method != "GET" {
		t.Errorf("Method = %q, want %q", apiErr.Method, "GET")
	}
	if apiErr.Path != "/rest/api/space/NOPE" {
		t.Errorf("Path = %q, want %q", apiErr.Path, "/rest/api/space/NOPE")
	}
	if apiErr.Message != "Space does not exist" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Space does not exist")
	}
}

func TestDo_ErrorBodyWithoutMessageFallsBackToReason(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"reason":"Forbidden"}`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = c.do(context.Background(), "GET", "/rest/api/space/FOO", nil, nil)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Message != "Forbidden" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Forbidden")
	}
}

func TestDo_NonJSONErrorBodyKeepsRawBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal server error, not json"))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = c.do(context.Background(), "GET", "/rest/api/space/FOO", nil, nil)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Message != "" {
		t.Errorf("Message = %q, want empty (non-JSON body)", apiErr.Message)
	}
	if apiErr.Body != "internal server error, not json" {
		t.Errorf("Body = %q, want raw response body", apiErr.Body)
	}
	if !strings.Contains(apiErr.Error(), "internal server error, not json") {
		t.Errorf("Error() = %q, want it to contain the raw body", apiErr.Error())
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "404 APIError",
			err:  &APIError{StatusCode: http.StatusNotFound},
			want: true,
		},
		{
			name: "500 APIError",
			err:  &APIError{StatusCode: http.StatusInternalServerError},
			want: false,
		},
		{
			name: "400 APIError",
			err:  &APIError{StatusCode: http.StatusBadRequest},
			want: false,
		},
		{
			name: "plain non-APIError error",
			err:  context.DeadlineExceeded,
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

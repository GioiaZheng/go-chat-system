package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadJSONAcceptsJSONContentTypes(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
	}{
		{name: "missing content type"},
		{name: "json", contentType: "application/json"},
		{name: "json with charset", contentType: "application/json; charset=utf-8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/session", strings.NewReader(`{"name":"Diana"}`))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			var got loginRequest
			if err := readJSON(req, &got); err != nil {
				t.Fatalf("readJSON returned error: %v", err)
			}
			if got.Name != "Diana" {
				t.Fatalf("name = %q, want Diana", got.Name)
			}
		})
	}
}

func TestReadJSONRejectsUnsupportedContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/session", strings.NewReader(`{"name":"Diana"}`))
	req.Header.Set("Content-Type", "application/json-bad")

	var got loginRequest
	err := readJSON(req, &got)
	if !errors.Is(err, http.ErrNotSupported) {
		t.Fatalf("readJSON error = %v, want %v", err, http.ErrNotSupported)
	}
}

func TestReadJSONRejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/session", strings.NewReader(`{"name":"Diana","role":"admin"}`))
	req.Header.Set("Content-Type", "application/json")

	var got loginRequest
	if err := readJSON(req, &got); err == nil {
		t.Fatal("expected readJSON to reject unknown field")
	}
}

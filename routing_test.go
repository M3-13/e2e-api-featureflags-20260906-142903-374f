package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoutesAreWired(t *testing.T) {
	handler := newHandler()

	// Create the flag with a valid body so the keyed routes below exercise the
	// registered handlers with an existing flag: GET answers 200 and DELETE 204
	// instead of a 404 for an unknown key (AC-03).
	createReq := httptest.NewRequest(http.MethodPost, "/flags",
		strings.NewReader(`{"key":"myfeature","enabled":true}`))
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: POST /flags want 201, got %d", createRec.Code)
	}

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "create flag", method: http.MethodPost, path: "/flags"},
		{name: "list flags", method: http.MethodGet, path: "/flags"},
		{name: "get flag", method: http.MethodGet, path: "/flags/myfeature"},
		{name: "update flag", method: http.MethodPut, path: "/flags/myfeature"},
		{name: "delete flag", method: http.MethodDelete, path: "/flags/myfeature"},
		{name: "evaluate flag", method: http.MethodGet, path: "/flags/myfeature/evaluate"},
		{name: "get unknown flag", method: http.MethodGet, path: "/flags/unknown-key"},
		{name: "delete unknown flag", method: http.MethodDelete, path: "/flags/unknown-key"},
		{name: "healthz", method: http.MethodGet, path: "/healthz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code == http.StatusMethodNotAllowed {
				t.Fatalf("%s %s should accept method %s, got 405", tt.method, tt.path, tt.method)
			}
			if rec.Code == http.StatusNotFound && !hasJSONErrorBody(rec) {
				// A registered handler answers 404 for an unknown key with a
				// JSON error object {"error":...}; a missing ServeMux route
				// answers 404 with a plain-text/empty body. Only the latter is
				// "not wired".
				t.Fatalf("%s %s should be wired, got 404 without a JSON error body", tt.method, tt.path)
			}
		})
	}
}

func TestUnknownRouteReturnsNotFound(t *testing.T) {
	handler := newHandler()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "unknown top-level path", method: http.MethodGet, path: "/nope"},
		{name: "missing key subpath", method: http.MethodGet, path: "/flags/myfeature/evaluate/extra"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("%s %s want 404, got %d", tt.method, tt.path, rec.Code)
			}
		})
	}
}

// hasJSONErrorBody reports whether the response body is a JSON object carrying
// an "error" field, i.e. it came from one of our writeError handlers rather
// than the mux's default plain-text 404.
func hasJSONErrorBody(rec *httptest.ResponseRecorder) bool {
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		return false
	}
	_, ok := body["error"]
	return ok
}

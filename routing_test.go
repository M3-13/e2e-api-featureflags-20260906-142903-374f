package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newHandlerWithFlag provisions a flag so the keyed routes can exercise their
// success path (AC-03): for an existing key GET answers 200 and DELETE 204,
// never a 404 for an unknown key.
func newHandlerWithFlag(t *testing.T) http.Handler {
	t.Helper()
	t.Setenv("FEATUREFLAGS_API_TOKEN", "test-token")
	handler := newHandler()
	req := httptest.NewRequest(http.MethodPost, "/flags",
		strings.NewReader(`{"key":"myfeature","enabled":true}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup: POST /flags want 201, got %d", rec.Code)
	}
	return handler
}

func TestRoutesAreWired(t *testing.T) {
	t.Run("create flag", func(t *testing.T) {
		t.Setenv("FEATUREFLAGS_API_TOKEN", "test-token")
		handler := newHandler()
		req := httptest.NewRequest(http.MethodPost, "/flags",
			strings.NewReader(`{"key":"other","enabled":true}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("POST /flags want 201, got %d", rec.Code)
		}
	})

	t.Run("list flags", func(t *testing.T) {
		handler := newHandler()
		req := httptest.NewRequest(http.MethodGet, "/flags", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /flags want 200, got %d", rec.Code)
		}
	})

	t.Run("get_flag", func(t *testing.T) {
		handler := newHandlerWithFlag(t)

		// (a) An existing, previously POST-created key answers 200.
		req := httptest.NewRequest(http.MethodGet, "/flags/myfeature", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /flags/myfeature want 200, got %d", rec.Code)
		}

		// (b) An unknown key answers 404 with a JSON error object. A registered
		// handler produces {"error":...}; a missing ServeMux route would answer
		// 404 with an empty/text body. Only the latter is "not wired" (AC-03).
		req = httptest.NewRequest(http.MethodGet, "/flags/unknown-key", nil)
		rec = httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("GET /flags/unknown-key want 404, got %d", rec.Code)
		}
		if !hasJSONErrorBody(rec) {
			t.Fatalf("GET /flags/unknown-key not wired: 404 response has no JSON error body")
		}
	})

	t.Run("update flag", func(t *testing.T) {
		handler := newHandlerWithFlag(t)
		req := httptest.NewRequest(http.MethodPut, "/flags/myfeature",
			strings.NewReader(`{"enabled":false}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("PUT /flags/myfeature want 200, got %d", rec.Code)
		}
	})

	t.Run("delete_flag", func(t *testing.T) {
		handler := newHandlerWithFlag(t)

		// (a) An existing, previously POST-created key answers 204.
		req := httptest.NewRequest(http.MethodDelete, "/flags/myfeature", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("DELETE /flags/myfeature want 204, got %d", rec.Code)
		}

		// (b) An unknown key answers 404 with a JSON error object, i.e. the
		// handler is wired (AC-03).
		req = httptest.NewRequest(http.MethodDelete, "/flags/unknown-key", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		rec = httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("DELETE /flags/unknown-key want 404, got %d", rec.Code)
		}
		if !hasJSONErrorBody(rec) {
			t.Fatalf("DELETE /flags/unknown-key not wired: 404 response has no JSON error body")
		}
	})

	t.Run("evaluate flag", func(t *testing.T) {
		handler := newHandlerWithFlag(t)
		req := httptest.NewRequest(http.MethodGet, "/flags/myfeature/evaluate?user=u1", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /flags/myfeature/evaluate want 200, got %d", rec.Code)
		}
	})

	t.Run("healthz", func(t *testing.T) {
		handler := newHandler()
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /healthz want 200, got %d", rec.Code)
		}
	})
}

func TestMutatingRoutesRequireAuth(t *testing.T) {
	t.Setenv("FEATUREFLAGS_API_TOKEN", "test-token")
	handler := newHandler()

	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "create", method: http.MethodPost, path: "/flags", body: `{"key":"k","enabled":true}`},
		{name: "update", method: http.MethodPut, path: "/flags/k", body: `{"enabled":true}`},
		{name: "delete", method: http.MethodDelete, path: "/flags/k", body: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/missing", func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("%s %s without token want 401, got %d", tc.method, tc.path, rec.Code)
			}
		})
		t.Run(tc.name+"/wrong", func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Authorization", "Bearer wrong-token")
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("%s %s with wrong token want 401, got %d", tc.method, tc.path, rec.Code)
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

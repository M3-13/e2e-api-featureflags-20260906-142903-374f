package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"featureflags/internal/store"
)

func newDeleteRequest(path string) *http.Request {
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	req.SetPathValue("key", "test-key")
	return req
}

func doDelete(t *testing.T, s *store.Store, path string) *httptest.ResponseRecorder {
	t.Helper()
	t.Setenv("FEATUREFLAGS_API_TOKEN", "test-token")
	req := newDeleteRequest(path)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	RequireAuth(DeleteFlag(s)).ServeHTTP(rec, req)
	return rec
}

func TestDeleteRemovesExistingFlag(t *testing.T) {
	s := store.NewStore()
	_ = s.Create(store.Flag{Key: "test-key", Enabled: true})

	rec := doDelete(t, s, "/flags/test-key")

	if rec.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("want empty body, got %q", rec.Body.String())
	}

	if _, err := s.Get("test-key"); err != store.ErrNotFound {
		t.Fatalf("want flag removed, got err %v", err)
	}
}

func TestDeleteUnknownKey(t *testing.T) {
	s := store.NewStore()

	rec := doDelete(t, s, "/flags/missing")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestDeleteRequiresAuth(t *testing.T) {
	t.Setenv("FEATUREFLAGS_API_TOKEN", "test-token")
	s := store.NewStore()

	t.Run("missing token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		RequireAuth(DeleteFlag(s)).ServeHTTP(rec, newDeleteRequest("/flags/test-key"))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", rec.Code)
		}
	})

	t.Run("wrong token", func(t *testing.T) {
		req := newDeleteRequest("/flags/test-key")
		req.Header.Set("Authorization", "Bearer wrong-token")
		rec := httptest.NewRecorder()
		RequireAuth(DeleteFlag(s)).ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", rec.Code)
		}
	})
}

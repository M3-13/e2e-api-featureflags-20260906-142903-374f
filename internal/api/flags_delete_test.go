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

func TestDeleteRemovesExistingFlag(t *testing.T) {
	s := store.NewStore()
	_ = s.Create(store.Flag{Key: "test-key", Enabled: true})

	rec := httptest.NewRecorder()
	DeleteFlag(s)(rec, newDeleteRequest("/flags/test-key"))

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

	rec := httptest.NewRecorder()
	DeleteFlag(s)(rec, newDeleteRequest("/flags/missing"))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

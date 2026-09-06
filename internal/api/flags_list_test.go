package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"featureflags/internal/store"
)

func TestListEmpty(t *testing.T) {
	s := store.NewStore()
	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rec := httptest.NewRecorder()
	List(s)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "[]\n" {
		t.Fatalf("want empty JSON array, got %q", got)
	}
}

func TestListAfterCreate(t *testing.T) {
	s := store.NewStore()
	f := store.Flag{
		Key:            "feature-x",
		Enabled:        true,
		Description:    "example",
		RolloutPercent: 50,
	}
	if err := s.Create(f); err != nil {
		t.Fatalf("create: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rec := httptest.NewRecorder()
	List(s)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}

	var flags []store.Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &flags); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(flags) != 1 {
		t.Fatalf("want 1 flag, got %d", len(flags))
	}
	if flags[0] != f {
		t.Fatalf("want flag %+v, got %+v", f, flags[0])
	}
}

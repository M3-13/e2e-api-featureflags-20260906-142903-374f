package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"featureflags/internal/store"
)

func TestGetFlagFound(t *testing.T) {
	s := store.NewStore()
	flag := store.Flag{
		Key:            "my-key",
		Enabled:        true,
		Description:    "my description",
		RolloutPercent: 42,
	}
	if err := s.Create(flag); err != nil {
		t.Fatalf("create: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/flags/my-key", nil)
	req.SetPathValue("key", "my-key")
	rec := httptest.NewRecorder()
	GetFlag(s)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}

	var got store.Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got != flag {
		t.Fatalf("flag changed: got %+v, want %+v", got, flag)
	}
}

func TestGetFlagNotFound(t *testing.T) {
	s := store.NewStore()

	req := httptest.NewRequest(http.MethodGet, "/flags/missing", nil)
	req.SetPathValue("key", "missing")
	rec := httptest.NewRecorder()
	GetFlag(s)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Fatalf("want error field in JSON body, got %q", rec.Body.String())
	}
}

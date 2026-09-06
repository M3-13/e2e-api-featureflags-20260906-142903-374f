package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"featureflags/internal/store"
)

func updateReq(t *testing.T, s *store.Store, key, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, "/flags/"+key, strings.NewReader(body))
	req.SetPathValue("key", key)
	rec := httptest.NewRecorder()
	UpdateFlag(s)(rec, req)
	return rec
}

func TestUpdateFlagAllFields(t *testing.T) {
	s := store.NewStore()
	if err := s.Create(store.Flag{Key: "flag1", Enabled: false, Description: "old", RolloutPercent: 10}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	rec := updateReq(t, s, "flag1", `{"enabled":true,"description":"new","rollout_percent":75}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var got store.Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got.Key != "flag1" {
		t.Fatalf("key changed: got %q", got.Key)
	}
	if !got.Enabled {
		t.Fatalf("enabled not updated: %v", got.Enabled)
	}
	if got.Description != "new" {
		t.Fatalf("description not updated: %q", got.Description)
	}
	if got.RolloutPercent != 75 {
		t.Fatalf("rollout_percent not updated: %d", got.RolloutPercent)
	}
}

func TestUpdateFlagOmitsOptionalsKeepsValues(t *testing.T) {
	s := store.NewStore()
	if err := s.Create(store.Flag{Key: "flag1", Enabled: true, Description: "keep", RolloutPercent: 30}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	rec := updateReq(t, s, "flag1", `{"enabled":false}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var got store.Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got.Enabled {
		t.Fatalf("enabled not updated")
	}
	if got.Description != "keep" {
		t.Fatalf("description lost: %q", got.Description)
	}
	if got.RolloutPercent != 30 {
		t.Fatalf("rollout_percent lost: %d", got.RolloutPercent)
	}
}

func TestUpdateFlagMissingEnabled(t *testing.T) {
	s := store.NewStore()
	if err := s.Create(store.Flag{Key: "flag1", Enabled: false}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	rec := updateReq(t, s, "flag1", `{"description":"x"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestUpdateFlagRolloutPercentOutOfRange(t *testing.T) {
	s := store.NewStore()
	if err := s.Create(store.Flag{Key: "flag1", Enabled: false}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	rec := updateReq(t, s, "flag1", `{"enabled":true,"rollout_percent":101}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestUpdateFlagUnknownKey(t *testing.T) {
	s := store.NewStore()

	rec := updateReq(t, s, "missing", `{"enabled":true}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestUpdateFlagBodyTooLarge(t *testing.T) {
	s := store.NewStore()
	if err := s.Create(store.Flag{Key: "flag1", Enabled: false}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	body := `{"enabled":true,"description":"` + strings.Repeat("a", maxBodyBytes) + `"}`
	rec := updateReq(t, s, "flag1", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

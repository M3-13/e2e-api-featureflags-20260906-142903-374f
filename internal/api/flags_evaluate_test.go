package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"featureflags/internal/store"
)

func evaluateRequest(t *testing.T, h http.HandlerFunc, key, query string) *httptest.ResponseRecorder {
	t.Helper()
	target := "/flags/" + key + "/evaluate"
	if query != "" {
		target += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, target, nil)
	req.SetPathValue("key", key)
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

func TestEvaluateDeterministicSameUser(t *testing.T) {
	s := store.NewStore()
	if err := s.Create(store.Flag{Key: "f", Enabled: true, RolloutPercent: 50}); err != nil {
		t.Fatalf("create: %v", err)
	}
	h := EvaluateFlag(s)

	var first bool
	for i := 0; i < 5; i++ {
		rec := evaluateRequest(t, h, "f", "user=alice")
		if rec.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", rec.Code)
		}
		var body evaluateResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if i == 0 {
			first = body.Enabled
		} else if body.Enabled != first {
			t.Fatalf("non-deterministic: call %d got %v, want %v", i, body.Enabled, first)
		}
		if body.Key != "f" {
			t.Fatalf("want key f, got %q", body.Key)
		}
		if body.User != "alice" {
			t.Fatalf("want user alice, got %q", body.User)
		}
	}
}

func TestEvaluateMissingUserBadRequest(t *testing.T) {
	s := store.NewStore()
	if err := s.Create(store.Flag{Key: "f", Enabled: true, RolloutPercent: 50}); err != nil {
		t.Fatalf("create: %v", err)
	}
	h := EvaluateFlag(s)

	rec := evaluateRequest(t, h, "f", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestEvaluateUnknownKeyNotFound(t *testing.T) {
	s := store.NewStore()
	h := EvaluateFlag(s)

	rec := evaluateRequest(t, h, "missing", "user=alice")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestEvaluateDoesNotMutateStore(t *testing.T) {
	s := store.NewStore()
	if err := s.Create(store.Flag{Key: "f", Enabled: true, RolloutPercent: 50}); err != nil {
		t.Fatalf("create: %v", err)
	}
	h := EvaluateFlag(s)

	evaluateRequest(t, h, "f", "user=alice")

	if got := len(s.List()); got != 1 {
		t.Fatalf("store mutated: want 1 flag, got %d", got)
	}
}

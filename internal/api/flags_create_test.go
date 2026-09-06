package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"featureflags/internal/store"
)

func doCreate(t *testing.T, s *store.Store, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	CreateFlag(s)(rec, req)
	return rec
}

func TestCreateValid(t *testing.T) {
	s := store.NewStore()
	rec := doCreate(t, s, `{"key":"myflag","enabled":true,"description":"desc","rollout_percent":42}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var got store.Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	want := store.Flag{Key: "myflag", Enabled: true, Description: "desc", RolloutPercent: 42}
	if got != want {
		t.Fatalf("want %+v, got %+v", want, got)
	}
}

func TestCreateDuplicateKey(t *testing.T) {
	s := store.NewStore()
	if rec := doCreate(t, s, `{"key":"dup","enabled":true}`); rec.Code != http.StatusCreated {
		t.Fatalf("first create want 201, got %d", rec.Code)
	}
	rec := doCreate(t, s, `{"key":"dup","enabled":false}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("want 409, got %d", rec.Code)
	}
}

func TestCreateMissingKey(t *testing.T) {
	s := store.NewStore()
	rec := doCreate(t, s, `{"enabled":true}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestCreateEmptyKey(t *testing.T) {
	s := store.NewStore()
	rec := doCreate(t, s, `{"key":"","enabled":true}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestCreateMissingEnabled(t *testing.T) {
	s := store.NewStore()
	rec := doCreate(t, s, `{"key":"noenabled"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestCreateRolloutPercentOutOfRange(t *testing.T) {
	for _, rp := range []int{-1, 101} {
		s := store.NewStore()
		body, _ := json.Marshal(map[string]any{"key": "rp", "enabled": true, "rollout_percent": rp})
		rec := doCreate(t, s, string(body))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("rollout_percent=%d want 400, got %d", rp, rec.Code)
		}
	}
}

func TestCreateBodyTooLarge(t *testing.T) {
	s := store.NewStore()
	var b bytes.Buffer
	b.WriteString(`{"key":"big","enabled":true,"description":"`)
	b.Write(bytes.Repeat([]byte("a"), maxBodyBytes))
	b.WriteString(`"}`)
	rec := doCreate(t, s, b.String())
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

package api

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggingLogsOnlyMethodPathStatus(t *testing.T) {
	var buf bytes.Buffer
	orig := accessLog
	accessLog = log.New(&buf, "", 0)
	defer func() { accessLog = orig }()

	handler := Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("response body secret"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/flags?user=secretuser", strings.NewReader(`{"key":"secret-key"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	out := buf.String()
	t.Logf("log output: %q", out)

	if !strings.Contains(out, http.MethodPost) {
		t.Fatalf("log should contain method, got %q", out)
	}
	if !strings.Contains(out, "/flags") {
		t.Fatalf("log should contain path, got %q", out)
	}
	if !strings.Contains(out, "201") {
		t.Fatalf("log should contain status code, got %q", out)
	}
	for _, secret := range []string{"secretuser", "secret-key", "response body secret"} {
		if strings.Contains(out, secret) {
			t.Fatalf("log must not contain %q, got %q", secret, out)
		}
	}
}

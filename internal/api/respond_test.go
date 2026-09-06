package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONMissingContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"x","enabled":true}`))
	var dst map[string]any
	if err := decodeJSON(req, &dst); err == nil {
		t.Fatalf("want error for missing Content-Type, got nil")
	}
}

func TestDecodeJSONWrongContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"x","enabled":true}`))
	req.Header.Set("Content-Type", "text/plain")
	var dst map[string]any
	if err := decodeJSON(req, &dst); err == nil {
		t.Fatalf("want error for text/plain Content-Type, got nil")
	}
}

func TestDecodeJSONTrailingData(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"x"}{"enabled":true}`))
	req.Header.Set("Content-Type", "application/json")
	var dst map[string]any
	if err := decodeJSON(req, &dst); err == nil {
		t.Fatalf("want error for trailing JSON data, got nil")
	}
}

func TestDecodeJSONValid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"x","enabled":true}`))
	req.Header.Set("Content-Type", "application/json")
	var dst map[string]any
	if err := decodeJSON(req, &dst); err != nil {
		t.Fatalf("want nil for valid JSON, got %v", err)
	}
	if dst["key"] != "x" {
		t.Fatalf("key not decoded: %v", dst)
	}
}

func TestDecodeJSONContentTypeWithCharset(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"x","enabled":true}`))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	var dst map[string]any
	if err := decodeJSON(req, &dst); err != nil {
		t.Fatalf("want nil for application/json with charset, got %v", err)
	}
}

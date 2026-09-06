package api

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
)

const maxBodyBytes = 1 << 20

var (
	errBodyTooLarge       = errors.New("request body too large")
	errMissingContentType = errors.New("missing Content-Type")
	errInvalidContentType = errors.New("Content-Type must be application/json")
	errTrailingData       = errors.New("request body must contain exactly one JSON value")
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decodeJSON(r *http.Request, dst any) error {
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		return errMissingContentType
	}
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil || mt != "application/json" {
		return errInvalidContentType
	}

	lr := &io.LimitedReader{R: r.Body, N: maxBodyBytes + 1}
	dec := json.NewDecoder(lr)
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errTrailingData
	}
	if lr.N <= 0 {
		return errBodyTooLarge
	}
	return nil
}

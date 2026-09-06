package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const maxBodyBytes = 1 << 20

var errBodyTooLarge = errors.New("request body too large")

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decodeJSON(r *http.Request, dst any) error {
	lr := &io.LimitedReader{R: r.Body, N: maxBodyBytes + 1}
	if err := json.NewDecoder(lr).Decode(dst); err != nil {
		return err
	}
	if lr.N <= 0 {
		return errBodyTooLarge
	}
	return nil
}

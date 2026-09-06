package api

import (
	"net/http"

	"featureflags/internal/store"
)

func Get(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}

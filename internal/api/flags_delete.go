package api

import (
	"net/http"

	"featureflags/internal/store"
)

func DeleteFlag(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}

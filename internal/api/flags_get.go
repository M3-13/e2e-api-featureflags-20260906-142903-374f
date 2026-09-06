package api

import (
	"errors"
	"net/http"

	"featureflags/internal/store"
)

func GetFlag(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		flag, err := s.Get(key)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(w, http.StatusNotFound, "flag not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		writeJSON(w, http.StatusOK, flag)
	}
}

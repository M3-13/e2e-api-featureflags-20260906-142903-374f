package api

import (
	"errors"
	"net/http"

	"featureflags/internal/evaluate"
	"featureflags/internal/store"
)

type evaluateResponse struct {
	Enabled bool   `json:"enabled"`
	Key     string `json:"key"`
	User    string `json:"user"`
}

func EvaluateFlag(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		user := r.URL.Query().Get("user")
		if user == "" {
			writeError(w, http.StatusBadRequest, "missing user parameter")
			return
		}

		f, err := s.Get(key)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(w, http.StatusNotFound, "flag not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		enabled := evaluate.Decide(key, user, f.RolloutPercent, f.Enabled)
		writeJSON(w, http.StatusOK, evaluateResponse{
			Enabled: enabled,
			Key:     key,
			User:    user,
		})
	}
}

package api

import (
	"errors"
	"net/http"
	"unicode/utf8"

	"featureflags/internal/store"
)

type createFlagRequest struct {
	Key            string `json:"key"`
	Enabled        *bool  `json:"enabled"`
	Description    string `json:"description"`
	RolloutPercent *int   `json:"rollout_percent"`
}

func CreateFlag(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createFlagRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Key == "" {
			writeError(w, http.StatusBadRequest, "key is required")
			return
		}
		if req.Enabled == nil {
			writeError(w, http.StatusBadRequest, "enabled is required")
			return
		}
		rollout := 0
		if req.RolloutPercent != nil {
			rollout = *req.RolloutPercent
			if rollout < 0 || rollout > 100 {
				writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
				return
			}
		}

		if utf8.RuneCountInString(req.Key) > 128 {
			writeError(w, http.StatusBadRequest, "key is too long")
			return
		}
		if len(req.Description) > 2000 {
			writeError(w, http.StatusBadRequest, "description too long")
			return
		}

		flag := store.Flag{
			Key:            req.Key,
			Enabled:        *req.Enabled,
			Description:    req.Description,
			RolloutPercent: rollout,
		}

		if err := s.Create(flag); err != nil {
			if errors.Is(err, store.ErrExists) {
				writeError(w, http.StatusConflict, "flag already exists")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusCreated, flag)
	}
}

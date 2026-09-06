package api

import (
	"errors"
	"net/http"

	"featureflags/internal/store"
)

type updateFlagRequest struct {
	Enabled        *bool   `json:"enabled"`
	Description    *string `json:"description"`
	RolloutPercent *int    `json:"rollout_percent"`
}

func UpdateFlag(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")

		var req updateFlagRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Enabled == nil {
			writeError(w, http.StatusBadRequest, "enabled is required")
			return
		}

		if req.RolloutPercent != nil && (*req.RolloutPercent < 0 || *req.RolloutPercent > 100) {
			writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
			return
		}

		if req.Description != nil && len(*req.Description) > 2000 {
			writeError(w, http.StatusBadRequest, "description too long")
			return
		}

		existing, err := s.Get(key)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(w, http.StatusNotFound, "flag not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		existing.Enabled = *req.Enabled
		if req.Description != nil {
			existing.Description = *req.Description
		}
		if req.RolloutPercent != nil {
			existing.RolloutPercent = *req.RolloutPercent
		}

		updated, err := s.Update(key, existing)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, updated)
	}
}

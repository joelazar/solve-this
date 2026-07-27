package api

import (
	"encoding/json"
	"net/http"
)

type bulkCompleteRequest struct {
	IDs []string `json:"ids"`
}

func (h *Handler) bulkComplete(w http.ResponseWriter, r *http.Request) {
	var req bulkCompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if len(req.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "ids must not be empty")
		return
	}
	completed, err := h.store.CompleteAll(req.IDs)
	if err != nil {
		failed(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"completed": len(completed),
		"items":     tasksToDTO(completed),
	})
}

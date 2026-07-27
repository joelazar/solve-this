package api

import (
	"encoding/json"
	"net/http"

	"github.com/joelazar/solve-this/internal/domain"
)

type createListRequest struct {
	Name string `json:"name"`
}

func (h *Handler) createList(w http.ResponseWriter, r *http.Request) {
	var req createListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if invalid := domain.ValidateListName(req.Name); invalid != nil {
		writeJSON(w, http.StatusBadRequest, errorDTO{Error: invalid.Error(), Fields: invalid.Fields})
		return
	}
	writeJSON(w, http.StatusCreated, listToDTO(h.store.CreateList(req.Name)))
}

func (h *Handler) getLists(w http.ResponseWriter, r *http.Request) {
	lists := h.store.Lists()
	dtos := make([]listDTO, 0, len(lists))
	for _, list := range lists {
		dtos = append(dtos, listToDTO(list))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": dtos})
}

func (h *Handler) getList(w http.ResponseWriter, r *http.Request) {
	list, err := h.store.List(r.PathValue("id"))
	if err != nil {
		failed(w, err)
		return
	}
	writeJSON(w, http.StatusOK, listToDTO(list))
}

func (h *Handler) deleteList(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	list, err := h.store.List(id)
	if err != nil {
		failed(w, err)
		return
	}
	if err := h.store.DeleteList(list.ID); err != nil {
		failed(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

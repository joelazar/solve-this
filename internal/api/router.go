package api

import (
	"net/http"

	"github.com/joelazar/solve-this/internal/store"
)

type Handler struct {
	store *store.Store
}

func NewRouter(s *store.Store) http.Handler {
	h := &Handler{store: s}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.health)

	mux.HandleFunc("POST /lists", h.createList)
	mux.HandleFunc("GET /lists", h.getLists)
	mux.HandleFunc("GET /lists/{id}", h.getList)
	mux.HandleFunc("DELETE /lists/{id}", h.deleteList)
	mux.HandleFunc("POST /lists/{id}/tasks", h.createTask)

	mux.HandleFunc("GET /tasks", h.getTasks)
	mux.HandleFunc("POST /tasks/bulk/complete", h.bulkComplete)
	mux.HandleFunc("GET /tasks/{id}", h.getTask)
	mux.HandleFunc("PATCH /tasks/{id}", h.patchTask)
	mux.HandleFunc("DELETE /tasks/{id}", h.deleteTask)
	mux.HandleFunc("POST /tasks/{id}/complete", h.completeTask)
	mux.HandleFunc("POST /tasks/{id}/reopen", h.reopenTask)
	mux.HandleFunc("POST /tasks/{id}/tags", h.addTag)
	mux.HandleFunc("DELETE /tasks/{id}/tags/{tag}", h.removeTag)

	mux.HandleFunc("GET /stats", h.getStats)
	mux.HandleFunc("GET /export.csv", h.exportCSV)

	return recoverPanic(requestID(logRequests(mux)))
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

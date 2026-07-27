package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/joelazar/solve-this/internal/domain"
	"github.com/joelazar/solve-this/internal/store"
)

type createTaskRequest struct {
	Title    string   `json:"title"`
	Notes    string   `json:"notes"`
	Priority string   `json:"priority"`
	Tags     []string `json:"tags"`
	Due      string   `json:"due"`
}

type patchTaskRequest struct {
	Title    *string   `json:"title"`
	Notes    *string   `json:"notes"`
	Done     *bool     `json:"done"`
	Priority *string   `json:"priority"`
	Tags     *[]string `json:"tags"`
	Due      *string   `json:"due"`
}

type tagRequest struct {
	Tag string `json:"tag"`
}

func parseDue(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	due, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, err
	}
	return &due, nil
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	due, err := parseDue(req.Due)
	if err != nil {
		writeError(w, http.StatusBadRequest, "due must be an RFC3339 timestamp")
		return
	}
	priority := domain.Priority(req.Priority)
	if req.Priority == "" {
		priority = domain.PriorityMedium
	}
	input := store.TaskInput{
		ListID:   r.PathValue("id"),
		Title:    req.Title,
		Notes:    req.Notes,
		Priority: priority,
		Tags:     domain.NormalizeTags(req.Tags),
		Due:      due,
	}
	candidate := domain.Task{Title: input.Title, Notes: input.Notes, Priority: input.Priority}
	candidate.SetTags(input.Tags)
	if err := domain.ValidateTask(candidate); err != nil {
		failed(w, err)
		return
	}
	task, err := h.store.CreateTask(input)
	if err != nil {
		failed(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, taskToDTO(task))
}

func (h *Handler) getTask(w http.ResponseWriter, r *http.Request) {
	task, err := h.store.Task(r.PathValue("id"))
	if err != nil {
		failed(w, err)
		return
	}
	writeJSON(w, http.StatusOK, taskToDTO(task))
}

func (h *Handler) patchTask(w http.ResponseWriter, r *http.Request) {
	var req patchTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	task, err := h.store.UpdateTask(r.PathValue("id"), func(task *domain.Task) error {
		if req.Title != nil {
			task.Title = *req.Title
		}
		if req.Notes != nil {
			task.Notes = *req.Notes
		}
		if req.Done != nil {
			task.Done = *req.Done
		}
		if req.Priority != nil {
			task.Priority = domain.Priority(*req.Priority)
		}
		if req.Tags != nil {
			task.SetTags(domain.NormalizeTags(*req.Tags))
		}
		if req.Due != nil {
			due, err := parseDue(*req.Due)
			if err != nil {
				return &domain.ValidationError{Fields: []domain.FieldError{{Field: "due", Message: "must be an RFC3339 timestamp"}}}
			}
			task.Due = due
		}
		return domain.ValidateTask(*task)
	})
	if err != nil {
		failed(w, err)
		return
	}
	writeJSON(w, http.StatusOK, taskToDTO(task))
}

func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request) {
	if err := h.store.DeleteTask(r.PathValue("id")); err != nil {
		failed(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) setDone(w http.ResponseWriter, r *http.Request, done bool) {
	task, err := h.store.UpdateTask(r.PathValue("id"), func(task *domain.Task) error {
		task.Done = done
		return nil
	})
	if err != nil {
		failed(w, err)
		return
	}
	writeJSON(w, http.StatusOK, taskToDTO(task))
}

func (h *Handler) completeTask(w http.ResponseWriter, r *http.Request) {
	h.setDone(w, r, true)
}

func (h *Handler) reopenTask(w http.ResponseWriter, r *http.Request) {
	h.setDone(w, r, false)
}

func (h *Handler) addTag(w http.ResponseWriter, r *http.Request) {
	var req tagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	tag := domain.NormalizeTag(req.Tag)
	if err := domain.ValidateTag(tag); err != nil {
		failed(w, err)
		return
	}
	task, err := h.store.UpdateTask(r.PathValue("id"), func(task *domain.Task) error {
		task.AddTag(tag)
		return domain.ValidateTask(*task)
	})
	if err != nil {
		failed(w, err)
		return
	}
	writeJSON(w, http.StatusOK, taskToDTO(task))
}

func (h *Handler) removeTag(w http.ResponseWriter, r *http.Request) {
	tag := domain.NormalizeTag(r.PathValue("tag"))
	var removed bool
	task, err := h.store.UpdateTask(r.PathValue("id"), func(task *domain.Task) error {
		removed = task.RemoveTag(tag)
		return nil
	})
	if err != nil {
		failed(w, err)
		return
	}
	if !removed {
		writeError(w, http.StatusNotFound, "tag "+tag+" not found")
		return
	}
	writeJSON(w, http.StatusOK, taskToDTO(task))
}

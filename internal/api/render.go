package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/joelazar/solve-this/internal/domain"
)

const timeFormat = time.RFC3339Nano

type taskDTO struct {
	ID        string   `json:"id"`
	ListID    string   `json:"list_id"`
	Title     string   `json:"title"`
	Notes     string   `json:"notes"`
	Done      bool     `json:"done"`
	Priority  string   `json:"priority"`
	Tags      []string `json:"tags"`
	Due       *string  `json:"due"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type listDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type pageDTO struct {
	Items      []taskDTO `json:"items"`
	Page       int       `json:"page"`
	PerPage    int       `json:"per_page"`
	Total      int       `json:"total"`
	TotalPages int       `json:"total_pages"`
}

func taskToDTO(task domain.Task) taskDTO {
	dto := taskDTO{
		ID:        task.ID,
		ListID:    task.ListID,
		Title:     task.Title,
		Notes:     task.Notes,
		Done:      task.Done,
		Priority:  string(task.Priority),
		Tags:      task.TagList(),
		CreatedAt: task.CreatedAt.Format(timeFormat),
		UpdatedAt: task.UpdatedAt.Format(timeFormat),
	}
	if task.Due != nil {
		due := task.Due.Format(timeFormat)
		dto.Due = &due
	}
	return dto
}

func tasksToDTO(tasks []domain.Task) []taskDTO {
	dtos := make([]taskDTO, 0, len(tasks))
	for _, task := range tasks {
		dtos = append(dtos, taskToDTO(task))
	}
	return dtos
}

func listToDTO(list domain.List) listDTO {
	return listDTO{ID: list.ID, Name: list.Name, CreatedAt: list.CreatedAt.Format(timeFormat)}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body != nil {
		json.NewEncoder(w).Encode(body)
	}
}

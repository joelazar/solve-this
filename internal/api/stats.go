package api

import (
	"encoding/csv"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/joelazar/solve-this/internal/domain"
)

type tagCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

type statsDTO struct {
	Total             int            `json:"total"`
	Done              int            `json:"done"`
	Open              int            `json:"open"`
	Overdue           int            `json:"overdue"`
	CompletionPercent int            `json:"completion_percent"`
	ByPriority        map[string]int `json:"by_priority"`
	ByTag             []tagCount     `json:"by_tag"`
}

func tagCounts(tasks []domain.Task) []tagCount {
	counts := make(map[string]int)
	for _, task := range tasks {
		for _, tag := range task.TagList() {
			counts[tag]++
		}
	}
	ranked := make([]tagCount, 0, len(counts))
	for tag, count := range counts {
		ranked = append(ranked, tagCount{Tag: tag, Count: count})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Count != ranked[j].Count {
			return ranked[i].Count > ranked[j].Count
		}
		return ranked[i].Tag < ranked[j].Tag
	})
	return ranked
}

func (h *Handler) getStats(w http.ResponseWriter, r *http.Request) {
	tasks := h.store.Tasks()
	now := domain.Now()

	stats := statsDTO{
		Total: len(tasks),
		ByPriority: map[string]int{
			string(domain.PriorityHigh):   0,
			string(domain.PriorityMedium): 0,
			string(domain.PriorityLow):    0,
		},
		ByTag: tagCounts(tasks),
	}
	for _, task := range tasks {
		if task.Done {
			stats.Done++
		}
		if task.Overdue(now) {
			stats.Overdue++
		}
		stats.ByPriority[string(task.Priority)]++
	}
	stats.Open = stats.Total - stats.Done
	if stats.Total > 0 {
		stats.CompletionPercent = stats.Done * 100 / stats.Total
	}

	writeJSON(w, http.StatusOK, stats)
}

func (h *Handler) exportCSV(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv")
	writer := csv.NewWriter(w)
	defer writer.Flush()

	writer.Write([]string{"id", "list_id", "title", "done", "priority", "tags", "due", "created_at"})
	for _, task := range h.store.Tasks() {
		due := ""
		if task.Due != nil {
			due = task.Due.Format(timeFormat)
		}
		writer.Write([]string{
			task.ID,
			task.ListID,
			task.Title,
			strconv.FormatBool(task.Done),
			string(task.Priority),
			strings.Join(task.TagList(), " "),
			due,
			task.CreatedAt.Format(timeFormat),
		})
	}
}

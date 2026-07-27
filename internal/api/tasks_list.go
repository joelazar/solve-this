package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/joelazar/solve-this/internal/domain"
	"github.com/joelazar/solve-this/internal/store"
)

const (
	defaultPerPage = 20
	maxPerPage     = 100
	defaultSort    = "created"
)

func intParam(values url.Values, key string, fallback int) (int, error) {
	raw := values.Get(key)
	if raw == "" {
		return fallback, nil
	}
	return strconv.Atoi(raw)
}

func (h *Handler) getTasks(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()

	page, err := intParam(values, "page", 1)
	if err != nil || page < 1 {
		writeError(w, http.StatusBadRequest, "page must be an integer greater than 0")
		return
	}

	perPage, err := intParam(values, "per_page", defaultPerPage)
	if err != nil || perPage < 1 || perPage > maxPerPage {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("per_page must be an integer between 1 and %d", maxPerPage))
		return
	}

	query := store.Query{
		ListID: values.Get("list"),
		Tag:    domain.NormalizeTag(values.Get("tag")),
		Search: values.Get("q"),
	}
	if raw := values.Get("done"); raw != "" {
		done, err := strconv.ParseBool(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "done must be true or false")
			return
		}
		query.Done = &done
	}

	sortKey := values.Get("sort")
	if sortKey == "" {
		sortKey = defaultSort
	}

	tasks := h.store.Tasks()
	if err := store.Sort(tasks, sortKey); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result := store.Paginate(store.Filter(tasks, query), page, perPage)
	writeJSON(w, http.StatusOK, pageDTO{
		Items:      tasksToDTO(result.Items),
		Page:       result.Number,
		PerPage:    result.Size,
		Total:      result.Total,
		TotalPages: (result.Total + result.Size - 1) / result.Size,
	})
}

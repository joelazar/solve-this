package api

import (
	"errors"
	"net/http"

	"github.com/joelazar/solve-this/internal/domain"
	"github.com/joelazar/solve-this/internal/store"
)

type errorDTO struct {
	Error  string              `json:"error"`
	Fields []domain.FieldError `json:"fields,omitempty"`
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorDTO{Error: message})
}

func failed(w http.ResponseWriter, err error) {
	var invalid *domain.ValidationError
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorDTO{Error: err.Error()})
	case errors.As(err, &invalid):
		writeJSON(w, http.StatusBadRequest, errorDTO{Error: invalid.Error(), Fields: invalid.Fields})
	default:
		writeJSON(w, http.StatusInternalServerError, errorDTO{Error: err.Error()})
	}
}

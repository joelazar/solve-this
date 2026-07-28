package domain

import (
	"fmt"
	"strings"
)

const (
	MaxTitleLen = 200
	MaxNotesLen = 2000
	MaxNameLen  = 100
	MaxTagLen   = 40
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	parts := make([]string, 0, len(e.Fields))
	for _, f := range e.Fields {
		parts = append(parts, fmt.Sprintf("%s %s", f.Field, f.Message))
	}
	return "validation failed: " + strings.Join(parts, ", ")
}

func ValidateTask(t Task) error {
	var fields []FieldError
	if strings.TrimSpace(t.Title) == "" {
		fields = append(fields, FieldError{Field: "title", Message: "must not be empty"})
	}
	if len(t.Title) > MaxTitleLen {
		fields = append(fields, FieldError{Field: "title", Message: fmt.Sprintf("must be at most %d characters", MaxTitleLen)})
	}
	if len(t.Notes) > MaxNotesLen {
		fields = append(fields, FieldError{Field: "notes", Message: fmt.Sprintf("must be at most %d characters", MaxNotesLen)})
	}
	if !t.Priority.Valid() {
		fields = append(fields, FieldError{Field: "priority", Message: "must be one of high, medium, low"})
	}
	for tag := range t.Tags {
		if message := tagProblem(tag); message != "" {
			fields = append(fields, FieldError{Field: "tags", Message: message})
			break
		}
	}
	if len(fields) == 0 {
		return nil
	}
	return &ValidationError{Fields: fields}
}

func tagProblem(tag string) string {
	switch {
	case tag == "":
		return "must not be empty"
	case len(tag) > MaxTagLen:
		return fmt.Sprintf("must be at most %d characters", MaxTagLen)
	}
	return ""
}

func ValidateTag(tag string) error {
	if message := tagProblem(tag); message != "" {
		return &ValidationError{Fields: []FieldError{{Field: "tag", Message: message}}}
	}
	return nil
}

func ValidateListName(name string) *ValidationError {
	var invalid *ValidationError
	trimmed := strings.TrimSpace(name)
	switch {
	case trimmed == "":
		invalid = &ValidationError{Fields: []FieldError{{Field: "name", Message: "must not be empty"}}}
	case len(trimmed) > MaxNameLen:
		invalid = &ValidationError{Fields: []FieldError{{Field: "name", Message: fmt.Sprintf("must be at most %d characters", MaxNameLen)}}}
	}
	return invalid
}

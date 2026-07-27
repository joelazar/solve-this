package store

import (
	"fmt"
	"sort"
	"strings"

	"github.com/joelazar/solve-this/internal/domain"
)

type Query struct {
	ListID string
	Tag    string
	Search string
	Done   *bool
}

func Filter(tasks []domain.Task, q Query) []domain.Task {
	out := make([]domain.Task, 0, len(tasks))
	for _, task := range tasks {
		if q.ListID != "" && task.ListID != q.ListID {
			continue
		}
		if q.Tag != "" && !task.HasTag(q.Tag) {
			continue
		}
		if q.Done != nil && task.Done != *q.Done {
			continue
		}
		if q.Search != "" && !matches(task, q.Search) {
			continue
		}
		out = append(out, task)
	}
	return out
}

func matches(task domain.Task, term string) bool {
	term = strings.ToLower(term)
	return strings.Contains(strings.ToLower(task.Title), term) ||
		strings.Contains(strings.ToLower(task.Notes), term)
}

var comparators = map[string]func(a, b domain.Task) int{
	"created": func(a, b domain.Task) int {
		return a.CreatedAt.Compare(b.CreatedAt)
	},
	"title": func(a, b domain.Task) int {
		return strings.Compare(strings.ToLower(a.Title), strings.ToLower(b.Title))
	},
	"priority": func(a, b domain.Task) int {
		return a.Priority.Rank() - b.Priority.Rank()
	},
	"due": func(a, b domain.Task) int {
		switch {
		case a.Due == nil && b.Due == nil:
			return 0
		case a.Due == nil:
			return 1
		case b.Due == nil:
			return -1
		default:
			return a.Due.Compare(*b.Due)
		}
	},
}

func Sort(tasks []domain.Task, key string) error {
	descending := strings.HasPrefix(key, "-")
	compare, ok := comparators[strings.TrimPrefix(key, "-")]
	if !ok {
		return fmt.Errorf("unknown sort key %q", key)
	}
	sort.SliceStable(tasks, func(i, j int) bool {
		order := compare(tasks[i], tasks[j])
		if order == 0 {
			return tasks[i].ID < tasks[j].ID
		}
		if descending {
			return order > 0
		}
		return order < 0
	})
	return nil
}

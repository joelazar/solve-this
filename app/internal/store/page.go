package store

import "github.com/joelazar/solve-this/internal/domain"

type Page struct {
	Items  []domain.Task
	Total  int
	Number int
	Size   int
}

func Paginate(tasks []domain.Task, number, size int) Page {
	total := len(tasks)
	start := (number - 1) * size
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}
	return Page{Items: tasks[start:end], Total: total, Number: number, Size: size}
}

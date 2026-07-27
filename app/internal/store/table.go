package store

import "fmt"

type table[T any] struct {
	rows   []T
	index  map[string]int
	seq    int
	prefix string
	idOf   func(T) string
}

func newTable[T any](prefix string, idOf func(T) string) *table[T] {
	return &table[T]{index: make(map[string]int), prefix: prefix, idOf: idOf}
}

func (t *table[T]) nextID() string {
	t.seq++
	return fmt.Sprintf("%s_%04d", t.prefix, t.seq)
}

func (t *table[T]) add(row T) {
	t.index[t.idOf(row)] = len(t.rows)
	t.rows = append(t.rows, row)
}

func (t *table[T]) at(id string) (int, bool) {
	i, ok := t.index[id]
	return i, ok
}

func (t *table[T]) remove(id string) bool {
	i, ok := t.index[id]
	if !ok {
		return false
	}
	t.rows = append(t.rows[:i], t.rows[i+1:]...)
	delete(t.index, id)
	for j := i; j < len(t.rows); j++ {
		t.index[t.idOf(t.rows[j])] = j
	}
	return true
}

func (t *table[T]) removeWhere(match func(T) bool) int {
	kept := t.rows[:0]
	removed := 0
	for _, row := range t.rows {
		if match(row) {
			delete(t.index, t.idOf(row))
			removed++
			continue
		}
		kept = append(kept, row)
	}
	t.rows = kept
	for i, row := range t.rows {
		t.index[t.idOf(row)] = i
	}
	return removed
}

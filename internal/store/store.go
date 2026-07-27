package store

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/joelazar/solve-this/internal/domain"
)

var ErrNotFound = errors.New("not found")

type TaskInput struct {
	ListID   string
	Title    string
	Notes    string
	Priority domain.Priority
	Tags     []string
	Due      *time.Time
}

type Store struct {
	mu    sync.RWMutex
	lists *table[domain.List]
	tasks *table[domain.Task]
}

func New() *Store {
	return &Store{
		lists: newTable("list", func(l domain.List) string { return l.ID }),
		tasks: newTable("task", func(t domain.Task) string { return t.ID }),
	}
}

func (s *Store) CreateList(name string) domain.List {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := domain.List{ID: s.lists.nextID(), Name: name, CreatedAt: domain.Now()}
	s.lists.add(list)
	return list
}

func (s *Store) Lists() []domain.List {
	s.mu.RLock()
	defer s.mu.RUnlock()
	lists := make([]domain.List, len(s.lists.rows))
	copy(lists, s.lists.rows)
	return lists
}

func (s *Store) List(id string) (domain.List, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	i, ok := s.lists.at(id)
	if !ok {
		return domain.List{}, fmt.Errorf("list %s: %w", id, ErrNotFound)
	}
	return s.lists.rows[i], nil
}

func (s *Store) DeleteList(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.lists.remove(id) {
		return fmt.Errorf("list %s: %w", id, ErrNotFound)
	}
	s.tasks.removeWhere(func(t domain.Task) bool { return t.ListID == id })
	return nil
}

func (s *Store) CreateTask(in TaskInput) (domain.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.lists.at(in.ListID); !ok {
		return domain.Task{}, fmt.Errorf("list %s: %w", in.ListID, ErrNotFound)
	}
	now := domain.Now()
	task := domain.Task{
		ID:        s.tasks.nextID(),
		ListID:    in.ListID,
		Title:     in.Title,
		Notes:     in.Notes,
		Priority:  in.Priority,
		Due:       in.Due,
		CreatedAt: now,
		UpdatedAt: now,
	}
	for _, tag := range in.Tags {
		task.AddTag(tag)
	}
	s.tasks.add(task)
	return task.Clone(), nil
}

func (s *Store) Task(id string) (domain.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	i, ok := s.tasks.at(id)
	if !ok {
		return domain.Task{}, fmt.Errorf("task %s: %w", id, ErrNotFound)
	}
	return s.tasks.rows[i].Clone(), nil
}

func (s *Store) Tasks() []domain.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tasks := make([]domain.Task, len(s.tasks.rows))
	for i, task := range s.tasks.rows {
		tasks[i] = task.Clone()
	}
	return tasks
}

func (s *Store) UpdateTask(id string, mutate func(*domain.Task) error) (domain.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i, ok := s.tasks.at(id)
	if !ok {
		return domain.Task{}, fmt.Errorf("task %s: %w", id, ErrNotFound)
	}
	draft := s.tasks.rows[i].Clone()
	if err := mutate(&draft); err != nil {
		return domain.Task{}, err
	}
	draft.Touch()
	s.tasks.rows[i] = draft
	return draft.Clone(), nil
}

func (s *Store) DeleteTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.tasks.remove(id) {
		return fmt.Errorf("task %s: %w", id, ErrNotFound)
	}
	return nil
}

func (s *Store) CompleteAll(ids []string) ([]domain.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	completed := make([]domain.Task, 0, len(ids))
	for _, id := range ids {
		i, ok := s.tasks.at(id)
		if !ok {
			return nil, fmt.Errorf("task %s: %w", id, ErrNotFound)
		}
		s.tasks.rows[i].Done = true
		s.tasks.rows[i].Touch()
		completed = append(completed, s.tasks.rows[i].Clone())
	}
	return completed, nil
}

# t3-race-tasks

Tier 3. Planted in `internal/store/store.go`, caught by `TestT3RaceTasks`.

The read lock disappears from `Tasks`:

```go
func (s *Store) Tasks() []domain.Task {
	tasks := make([]domain.Task, len(s.tasks.rows))
	for i, task := range s.tasks.rows {
		tasks[i] = task.Clone()
	}
	return tasks
}
```

The copy loop still runs, so the method keeps its correct shape and single-threaded
behavior. But it now reads `s.tasks.rows` while writers append and mutate under the
write lock, and the reader no longer participates in that exclusion. Under concurrent
load `GET /tasks` can observe a torn slice header, clone half-written rows, or panic
when the length changes mid-loop.

The baseline brackets the loop:

```go
s.mu.RLock()
defer s.mu.RUnlock()
```

The hidden test runs concurrent `GET /tasks` and task creations against a `-race`
build and fails when the log contains a race report naming `store.(*Store).Tasks`.

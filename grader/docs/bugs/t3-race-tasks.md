# t3-race-tasks

Tier 3. Planted in `internal/store/store.go`, caught by `TestT3RaceTasks`.

The read lock disappears from `Task`:

```go
func (s *Store) Task(id string) (domain.Task, error) {
	i, ok := s.tasks.at(id)
	if !ok {
		return domain.Task{}, fmt.Errorf("task %s: %w", id, ErrNotFound)
	}
	return s.tasks.rows[i].Clone(), nil
}
```

The lookup still runs, so the method keeps its correct shape and single-threaded
behavior. But it now reads the id index and `s.tasks.rows` while writers insert into
the map and append to the slice under the write lock, and the reader no longer
participates in that exclusion. Under concurrent load `GET /tasks/{id}` can read the
index mid-write, index a stale slice header, or crash the process with a concurrent
map access.

The baseline brackets the lookup:

```go
s.mu.RLock()
defer s.mu.RUnlock()
```

The bug lives in `Task` and not in `Tasks` so it never shares a function with
[t2-slice-alias](t2-slice-alias.md). When both sat in `Tasks`, the planted code read
as one defect and agents reported a single fix for two bugs.

The hidden test runs concurrent `GET /tasks/{id}` and task creations against a `-race`
build and fails when the log contains a race report naming `store.(*Store).Task()`.

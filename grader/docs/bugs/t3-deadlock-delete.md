# t3-deadlock-delete

Tier 3. Planted in `internal/store/store.go`, caught by `TestT3DeadlockDelete`.

`DeleteTask` trades its `defer` for explicit unlocks and misses the error path:

```go
func (s *Store) DeleteTask(id string) error {
	s.mu.Lock()
	if !s.tasks.remove(id) {
		return fmt.Errorf("task %s: %w", id, ErrNotFound)
	}
	s.mu.Unlock()
	return nil
}
```

Deleting an existing task unlocks and works. Deleting an unknown id returns with the
write lock still held. The 404 for that request goes out normally, and then the whole
store is frozen: every later handler blocks on the mutex and the API stops responding
until the process is restarted.

Unlike the other tier 3 bugs this one is fully deterministic and needs no race
detector. The hidden test deletes an unknown task id, then issues `GET /lists` with a
3 second client timeout and fails if the request never returns.

The baseline:

```go
s.mu.Lock()
defer s.mu.Unlock()
```

# t3-mutex-copy

Tier 3. Planted in `internal/store/store.go`, caught by `TestT3MutexCopy`. The one bug
that is not vet-clean.

A single character goes missing from the receiver:

```go
func (s Store) CompleteAll(ids []string) ([]domain.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	...
}
```

Every call copies the `Store`, including its `sync.RWMutex`, and locks the copy. The
tables inside are pointers, so the writes still land in shared state; the exclusion is
what's gone. Two concurrent bulk completes mutate the same rows with no lock between
them. Single-threaded everything works, which is why the bug survives a read-through.

`go vet` flags it directly:

```
CompleteAll passes lock by value: store.Store contains sync.RWMutex
```

so an agent that runs vet gets the answer for free. The hidden test also hammers
concurrent `POST /tasks/bulk/complete` requests on a `-race` build and fails when the
log contains a race report naming `store.Store.CompleteAll` (the value-receiver
symbol; the healthy pointer receiver shows up as `store.(*Store).CompleteAll`, which
that substring does not match).

The baseline receiver is `func (s *Store) CompleteAll(`.

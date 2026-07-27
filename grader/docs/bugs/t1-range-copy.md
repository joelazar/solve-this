# t1-range-copy

Tier 1. Planted in `internal/store/store.go`, caught by `TestT1RangeCopy`.

The planted version of `CompleteAll` walks the rows with a value-range loop:

```go
for _, task := range s.tasks.rows {
	if _, ok := wanted[task.ID]; !ok {
		continue
	}
	task.Done = true
	task.Touch()
	completed = append(completed, task.Clone())
}
```

`task` is a copy of the slice element, so `task.Done = true` mutates the copy and the
row in the store never changes. The response is built from the mutated copies, which
makes the endpoint look correct: `POST /tasks/bulk/complete` answers with `done: true`
while a later `GET /tasks/{id}` still reports `done: false`.

The baseline writes through the index instead:

```go
s.tasks.rows[i].Done = true
s.tasks.rows[i].Touch()
```

The spec sentence it breaks: bulk complete persists, a later GET reports `done: true`.

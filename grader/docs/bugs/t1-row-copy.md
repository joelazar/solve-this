# t1-row-copy

Tier 1. Planted in `internal/store/store.go`, caught by `TestT1RowCopy`.

The planted version of `CompleteAll` hoists the row into a local before mutating it:

```go
for _, i := range rows {
	task := s.tasks.rows[i]
	task.Done = true
	task.Touch()
	completed = append(completed, task.Clone())
}
```

`domain.Task` is a struct, so `task` is a copy of the row: `task.Done = true` mutates the
copy and the store never changes. The response is built from the mutated copies, which
makes the endpoint look correct: `POST /tasks/bulk/complete` answers with `done: true`
while a later `GET /tasks/{id}` still reports `done: false`.

The baseline writes through the index:

```go
for _, i := range rows {
	s.tasks.rows[i].Done = true
	s.tasks.rows[i].Touch()
	completed = append(completed, s.tasks.rows[i].Clone())
}
```

The edit touches nothing else: the ids are resolved to row indexes in a first pass, so
the response order, the duplicate handling and the all-or-nothing behavior are the same
in both versions and the only difference left is persistence. `TestRegressionBulkResponse`
and `TestRegressionBulkAtomic` pin those, which is why the fix has to be the write-back
and not a rewrite of the loop.

The spec sentence it breaks: bulk complete persists, a later GET reports `done: true`.

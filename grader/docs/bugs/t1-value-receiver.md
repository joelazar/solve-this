# t1-value-receiver

Tier 1. Planted in `internal/domain/task.go`, caught by `TestT1ValueReceiver`.

The receiver loses its pointer:

```go
func (t Task) Touch() {
	t.UpdatedAt = Now()
}
```

`Touch` runs on a copy of the task, so the write to `UpdatedAt` is thrown away when the
method returns. Every mutation still calls `Touch`, the code compiles, and nothing
fails. The only trace is that `updated_at` never moves: after a PATCH it still equals
`created_at`.

The baseline declares `func (t *Task) Touch()`.

The spec sentence it breaks: `updated_at` is refreshed on every mutation of the task.

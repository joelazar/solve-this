# t2-errors-is

Tier 2. Planted in `internal/api/errors.go`, caught by `TestT2ErrorsIs`.

The error translator compares with `==` instead of `errors.Is`:

```go
case err == store.ErrNotFound:
	writeJSON(w, http.StatusNotFound, ...)
```

The store never returns the sentinel bare. It wraps it with context:

```go
return fmt.Errorf("task %s: %w", id, ErrNotFound)
```

A wrapped error is never `==` to its sentinel, so the not-found branch is dead code and
every miss falls through to the default 500. Requests for unknown task or list ids
return 500 instead of 404.

The baseline unwraps:

```go
case errors.Is(err, store.ErrNotFound):
```

The spec sentence it breaks: an unknown task id is 404.

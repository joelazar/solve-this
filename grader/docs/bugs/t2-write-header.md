# t2-write-header

Tier 2. Planted in `internal/api/tasks_write.go`, caught by `TestT2WriteHeader`.

The create handler writes the body before the status:

```go
w.Header().Set("Content-Type", "application/json")
w.Write(body)
w.WriteHeader(http.StatusCreated)
```

The first `Write` commits the response with an implicit 200, and the late
`WriteHeader(201)` is a no-op (the server logs a "superfluous WriteHeader" warning).
Creating a task returns 200 with a correct body, so clients that only look at the
payload never notice; clients that check the status per the spec do.

The baseline goes through the shared helper, which sets the status first:

```go
writeJSON(w, http.StatusCreated, taskToDTO(task))
```

The spec sentence it breaks: a valid create request returns 201.

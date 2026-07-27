# t1-make-append

Tier 1. Planted in `internal/api/render.go`, caught by `TestT1MakeAppend`.

The DTO slice is made with a length instead of a capacity:

```go
dtos := make([]taskDTO, len(tasks))
for _, task := range tasks {
	dtos = append(dtos, taskToDTO(task))
}
```

`make([]taskDTO, n)` prefills the slice with n zero values, and the appends add the
real DTOs after them. `GET /tasks` returns twice the expected count, with empty
placeholder items (blank ids, blank titles) in front of the real ones.

The baseline reserves capacity without length:

```go
dtos := make([]taskDTO, 0, len(tasks))
```

The spec sentence it breaks: `items` holds at most `per_page` tasks.

# t1-atoi-ignored

Tier 1. Planted in `internal/api/tasks_list.go`, caught by `TestT1AtoiIgnored`.

The planted code parses `per_page` with a discarded error:

```go
perPage := defaultPerPage
if raw := values.Get("per_page"); raw != "" {
	perPage, _ = strconv.Atoi(raw)
}
if perPage > maxPerPage {
	perPage = maxPerPage
}
```

`per_page=abc` fails to parse, the error vanishes into `_`, and `perPage` becomes 0.
The only guard left caps the upper bound, so zero sails through and later blows up the
`total_pages` division by `result.Size`. The client sees a 500.

The baseline validates and rejects:

```go
perPage, err := intParam(values, "per_page", defaultPerPage)
if err != nil || perPage < 1 || perPage > maxPerPage {
	writeError(w, http.StatusBadRequest, ...)
	return
}
```

The spec requires `per_page` to be an integer between 1 and 100, anything else is 400.

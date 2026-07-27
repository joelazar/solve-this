# t2-total-pages

Tier 2. Planted in `internal/api/tasks_list.go`, caught by `TestT2TotalPages`.

The page count drops its rounding term:

```go
TotalPages: result.Total / result.Size,
```

Integer division truncates. With 3 tasks and `per_page=2`, `total_pages` reports 1,
yet `page=2` happily returns the third task. Anything paginating by `total_pages`
silently skips the final partial page.

The baseline rounds up:

```go
TotalPages: (result.Total + result.Size - 1) / result.Size,
```

The spec sentence it breaks: `total_pages` is `total` divided by `per_page`, rounded
up. Compare the decoy [d1-percent-truncation](d1-percent-truncation.md), where integer
division is exactly what the spec asks for.

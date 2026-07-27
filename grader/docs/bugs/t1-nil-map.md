# t1-nil-map

Tier 1. Planted in `internal/api/stats.go`, caught by `TestT1NilMap`.

One line changes:

```go
var counts map[string]int
```

instead of

```go
counts := make(map[string]int)
```

A declared-but-never-made map is nil. Reading from a nil map is fine, which is why the
code looks plausible, but the first write panics:

```go
counts[tag]++ // panic: assignment to entry in nil map
```

As soon as any task carries a tag, `GET /stats` hits this line and returns 500. With no
tagged tasks the endpoint works, which can make the bug look intermittent.

The spec sentence it breaks: `by_tag` counts how many tasks carry each tag.

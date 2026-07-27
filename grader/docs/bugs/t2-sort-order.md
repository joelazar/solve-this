# t2-sort-order

Tier 2. Planted in `internal/store/query.go`, caught by `TestT2SortOrder`.

The priority comparator sorts the enum as text:

```go
"priority": func(a, b domain.Task) int {
	return strings.Compare(string(a.Priority), string(b.Priority))
},
```

Alphabetically that is `high`, `low`, `medium`. The order looks right at a glance
because `high` still comes first; the swap of the other two is the tell.

The baseline compares ranks:

```go
return a.Priority.Rank() - b.Priority.Rank()
```

The spec sentence it breaks: priority orders high, then medium, then low.

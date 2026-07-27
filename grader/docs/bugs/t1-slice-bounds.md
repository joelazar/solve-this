# t1-slice-bounds

Tier 1. Planted in `internal/store/page.go`, caught by `TestT1SliceBounds`.

The bug deletes the clamp on `start`:

```go
start := (number - 1) * size
end := start + size
if end > total {
	end = total
}
return Page{Items: tasks[start:end], Total: total, Number: number, Size: size}
```

For a page past the end, `start` is larger than `total` while `end` still clamps down
to `total`, so the slice expression runs with `start > end` and panics. The recovery
middleware turns the panic into a 500.

The baseline keeps both clamps:

```go
if start > total {
	start = total
}
```

which makes `tasks[start:end]` an empty slice for any page past the last one. The spec
says a page beyond the final page returns an empty `items` array, not an error.

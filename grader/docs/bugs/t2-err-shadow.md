# t2-err-shadow

Tier 2. Planted in `internal/api/lists.go`, caught by `TestT2ErrShadow`.

The delete handler grows an outer error that nothing ever assigns:

```go
var err error
if list, err := h.store.List(id); err == nil {
	err = h.store.DeleteList(list.ID)
}
if err != nil {
	failed(w, err)
	return
}
```

The `err` in the if statement's init is a new variable that shadows the outer one. The
assignment inside the block writes to the inner `err` too, so the outer `err` is nil on
every path and the handler always falls through to 204. `DELETE /lists/{id}` reports
success for list ids that never existed.

The baseline uses two plain checks with no shadowing:

```go
list, err := h.store.List(id)
if err != nil {
	failed(w, err)
	return
}
if err := h.store.DeleteList(list.ID); err != nil {
	failed(w, err)
	return
}
```

The spec sentence it breaks: deleting a list that does not exist is 404.

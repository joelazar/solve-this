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

The test asserts that the delete is reported as a failure, and that deleting the same
list twice fails the second time. It deliberately does not pin the exact status: with
[t2-errors-is](t2-errors-is.md) planted the very same endpoint answers 500, and every bug
has to fail its own test only. The 404 mapping of a wrapped not-found belongs to that bug
and is pinned by `TestT2ErrorsIs`; fixing both is what makes
`DELETE /lists/{unknown}` answer 404.

The spec sentence it breaks: deleting a list that does not exist is 404.

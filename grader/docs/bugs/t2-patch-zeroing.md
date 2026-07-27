# t2-patch-zeroing

Tier 2. Planted in `internal/api/tasks_write.go`, caught by `TestT2PatchZeroing`.

The PATCH request type loses one pointer:

```go
Done bool `json:"done"` // baseline: Done *bool
```

and the handler assigns it unconditionally:

```go
task.Done = req.Done // baseline: if req.Done != nil { task.Done = *req.Done }
```

With a plain `bool` there is no way to tell "field absent" from "field set to false":
both decode to `false`. Every other field in the request keeps its pointer, so the
partial-update machinery looks intact. PATCHing only the title of a completed task
silently reopens it.

The spec sentence it breaks: PATCH updates the fields present in the body and leaves
every other field untouched.

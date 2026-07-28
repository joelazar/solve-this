# t2-typed-nil

Tier 2. Planted in `internal/domain/validate.go`, caught by `TestT2TypedNil`.

The planted `ValidateTag` funnels its result through a typed pointer:

```go
func ValidateTag(tag string) error {
	var invalid *ValidationError
	if message := tagProblem(tag); message != "" {
		invalid = &ValidationError{Fields: []FieldError{{Field: "tag", Message: message}}}
	}
	return invalid
}
```

For a valid tag, `invalid` stays nil, but returning it boxes the nil pointer into a
non-nil `error` interface. The caller's `if err != nil` fires on every input, valid or
not, and the error path dereferences the nil pointer, so `POST /tasks/{id}/tags`
answers 500 for perfectly good tags.

The baseline returns a literal `nil` when there is nothing to report:

```go
if message := tagProblem(tag); message != "" {
	return &ValidationError{Fields: []FieldError{{Field: "tag", Message: message}}}
}
return nil
```

Only the tag endpoint goes through `ValidateTag`; the tags of a whole task are checked by
`ValidateTask` against the same `tagProblem`, so this bug cannot be mistaken for the
validation of `POST /lists/{id}/tasks`.

Note the decoy: `ValidateListName` has the same shape but declares its return type as
`*ValidationError`, so no interface boxing happens and the nil check works. See
[d2-concrete-error](d2-concrete-error.md).

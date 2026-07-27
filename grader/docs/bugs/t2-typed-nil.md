# t2-typed-nil

Tier 2. Planted in `internal/domain/validate.go`, caught by `TestT2TypedNil`.

The planted `ValidateTag` funnels its result through a typed pointer:

```go
func ValidateTag(tag string) error {
	var invalid *ValidationError
	switch {
	case tag == "":
		invalid = &ValidationError{...}
	case len(tag) > MaxTagLen:
		invalid = &ValidationError{...}
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
if len(fields) == 0 {
	return nil
}
return &ValidationError{Fields: fields}
```

Note the decoy: `ValidateListName` has the same shape but declares its return type as
`*ValidationError`, so no interface boxing happens and the nil check works. See
[d2-concrete-error](d2-concrete-error.md).

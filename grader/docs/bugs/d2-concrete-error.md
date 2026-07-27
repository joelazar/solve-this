# d2-concrete-error

Decoy. Baked into the baseline in `internal/domain/validate.go`, guarded by
`TestDecoyListNameValidation`.

```go
func ValidateListName(name string) *ValidationError {
	var invalid *ValidationError
	...
	return invalid
}
```

This has the exact shape of the typed-nil trap ([t2-typed-nil](t2-typed-nil.md)): a
nil pointer declared up front and returned at the end. The difference is the return
type. `ValidateListName` returns the concrete `*ValidationError`, not the `error`
interface, and its caller keeps the concrete type:

```go
if invalid := domain.ValidateListName(req.Name); invalid != nil {
```

No interface boxing happens, the nil comparison behaves, and valid list names pass. An
agent that pattern-matches on the shape and "fixes" it breaks the decoy test.

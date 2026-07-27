# t2-trim-cutset

Tier 2. Planted in `internal/domain/tag.go`, caught by `TestT2TrimCutset`.

Tag normalisation swaps `TrimPrefix` for `TrimLeft`:

```go
return strings.TrimLeft(strings.ToLower(strings.TrimSpace(tag)), TagPrefix)
```

`TrimLeft` does not remove a prefix. Its second argument is a set of characters, so
`"tag:"` means "strip any leading run of `t`, `a`, `g` or `:`". Tags that never had the
prefix lose their leading letters: `groceries` comes back as `roceries` and `attic`
as `ic`.

The baseline removes exactly one literal prefix:

```go
return strings.TrimPrefix(strings.ToLower(strings.TrimSpace(tag)), TagPrefix)
```

The spec sentence it breaks: a single leading `tag:` prefix is removed, nothing else.

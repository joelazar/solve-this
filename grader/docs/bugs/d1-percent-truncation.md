# d1-percent-truncation

Decoy. Baked into the baseline in `internal/api/stats.go`, guarded by
`TestDecoyCompletionPercentFloor`.

```go
if stats.Total > 0 {
	stats.CompletionPercent = stats.Done * 100 / stats.Total
}
```

Integer division here looks like a rounding bug, especially next to
[t2-total-pages](t2-total-pages.md) where the same pattern really is one. But the spec
says `completion_percent` is `done / total` as a percentage, rounded down. One done
task out of three is 33, not 33.33 and not 34.

An agent that "fixes" this to `math.Round` or float formatting breaks the decoy test
and shows up in the scorecard under `decoys_broken`.

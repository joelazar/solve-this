# Bug catalog

One page per bug: the planted code, why it is wrong, and how it shows up from the
outside. The machine-readable answer key lives in
[`grader/internal/bugs/bugs.go`](../../internal/bugs/bugs.go); these pages explain it.
`mise run patches` regenerates raw diffs into `bugs/` at the repository root if you
want to see the exact edits.

Tier 1 bugs are mechanical Go mistakes. A single request against the right endpoint
exposes each of them.

| bug | class |
| --- | --- |
| [t1-range-copy](t1-range-copy.md) | range over values mutates a copy |
| [t1-slice-bounds](t1-slice-bounds.md) | missing bounds check |
| [t1-atoi-ignored](t1-atoi-ignored.md) | discarded conversion error |
| [t1-nil-map](t1-nil-map.md) | write to a nil map |
| [t1-value-receiver](t1-value-receiver.md) | value receiver on a mutating method |
| [t1-make-append](t1-make-append.md) | make with length instead of capacity |

Tier 2 bugs are semantic. The code reads plausibly and often works for the common
case; you need the spec in `docs/api.md` to see what is wrong.

| bug | class |
| --- | --- |
| [t2-typed-nil](t2-typed-nil.md) | nil pointer boxed in a non-nil error interface |
| [t2-err-shadow](t2-err-shadow.md) | error shadowed by an inner declaration |
| [t2-errors-is](t2-errors-is.md) | wrapped error compared with == |
| [t2-slice-alias](t2-slice-alias.md) | internal slice handed to the caller |
| [t2-sort-order](t2-sort-order.md) | ordering by string instead of rank |
| [t2-trim-cutset](t2-trim-cutset.md) | TrimLeft used as TrimPrefix |
| [t2-total-pages](t2-total-pages.md) | integer division where rounding up is required |
| [t2-write-header](t2-write-header.md) | WriteHeader after Write |
| [t2-patch-zeroing](t2-patch-zeroing.md) | non pointer field in a partial update |

Tier 3 bugs are concurrency defects. Each behaves correctly on a single request; the
tools (`go vet`, a `-race` build under load) or sustained traffic expose them.

| bug | class |
| --- | --- |
| [t3-race-request-id](t3-race-request-id.md) | unsynchronized shared counter |
| [t3-race-tasks](t3-race-tasks.md) | shared state read without the lock |
| [t3-mutex-copy](t3-mutex-copy.md) | mutex locked on a copy of the receiver |
| [t3-deadlock-delete](t3-deadlock-delete.md) | lock not released on the error path |

The two decoys are correct code that looks buggy. They are part of the baseline, not
planted, and each has a test pinning the correct behavior.

| decoy | looks like |
| --- | --- |
| [d1-percent-truncation](d1-percent-truncation.md) | a rounding bug in completion_percent |
| [d2-concrete-error](d2-concrete-error.md) | the typed nil trap, minus the trap |

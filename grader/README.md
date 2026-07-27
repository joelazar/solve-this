# grader

Plants bugs into [app](../app) and scores agent runs against hidden tests. Everything
below runs from the repository root; run trees are created outside the repository so
agents never see this directory.

`grader/internal/bugs/bugs.go` is the answer key: 19 bugs as exact text replacements against the
clean baseline, each tied to the spec sentence it breaks and the hidden test that covers
it. Two decoys (weird-looking but correct code baked into the baseline) are listed there
too. Tests in `tests/` are black-box HTTP checks against the built binary; the
`TestRegression*` tests guard behavior no bug touches.

## Bug tiers

Every bug is explained one by one, with code, in the [bug catalog](docs/bugs/README.md).
The tiers grade how much work detection takes:

- Tier 1 (6 bugs): mechanical Go mistakes, like a write to a nil map or a discarded
  `Atoi` error. One request against the right endpoint exposes each of them, so these
  mostly measure whether an agent exercises the API at all.
- Tier 2 (9 bugs): semantic traps. The code compiles, often works for the common case,
  and reads plausibly; the defect only shows against the spec in `docs/api.md`. A
  wrapped error compared with `==`, or a PATCH that zeroes a field the request never
  mentioned.
- Tier 3 (4 bugs): concurrency. Correct on every single request, broken under load: two
  data races, a mutex locked on a copy of its struct, and a lock leaked on an error
  path that freezes the whole API. The hidden tests hammer a `-race` build and grep its
  log for reports naming the buggy site, so they reward agents who run `go vet` and the
  race detector instead of only reading code. `t3-mutex-copy` is the one bug that is
  not vet-clean; `verify` expects vet to fail on exactly that variant.

Random selection defaults to tiers 1 and 2, so tier 3 is opt-in via `-tiers`:
`-tiers 1,2,3 -n 19` plants everything, and an explicit `-bugs` list bypasses the tier
filter entirely.

## Workflow

One run = one buggy tree + one prompt + one agent + one scorecard, keyed by `-id`.
The same `-seed`/`-n` (or `-bugs`) always plants the same set, so give each agent its
own `-id` and reuse the seed for a fair comparison.

```sh
# headless agents, end to end
go run ./grader/cmd/bench -id s42-claude -seed 42 -n 10 -mode hunt \
  -agent 'claude --dangerously-skip-permissions -p "$PROMPT"'
go run ./grader/cmd/bench -id s42-codex -seed 42 -n 10 -mode hunt \
  -agent 'codex exec --full-auto "$PROMPT"'

# interactive agents: generate the tree, paste the printed prompt, score afterwards
go run ./grader/cmd/bench -id s42-pi -seed 42 -n 10 -mode count
go run ./grader/cmd/score -id s42-pi

# hand the agent a bug report instead of a hunt
go run ./grader/cmd/bench -id repro1 -bugs t2-slice-alias -mode report -agent '...'
```

The agent command runs through `sh -c` inside the run tree with `PROMPT` (rendered
text) and `PROMPT_FILE` (absolute path) in its environment. Bench exports the baseline
at a pinned revision, applies the bugs, verifies the tree builds, and squashes it into a
single `initial import` commit so the history leaks nothing.

Everything lands in `runs/`: `<id>.json` manifest, `<id>-prompt.md`, `<id>-agent.log`,
`<id>-score.json`. The scorecard reports fixed bugs per tier, broken decoys, collateral
damage (unplanted bugs whose tests fail), regressions, diff size from the initial
commit, and wall time.

## Prompts

| template | mode |
| --- | --- |
| `grader/prompts/vague.md` | two sentences, no spec pointer, no rules, no count |
| `grader/prompts/hunt.md` | find and fix everything, count unknown |
| `grader/prompts/count.md` | states the exact count via `{{.N}}` |
| `grader/prompts/report.md` | user-style bug reports via `{{.Symptoms}}` |

They are Go templates rendered against the planted set; edit them freely or pass
`-prompt my-prompt.md`. Symptoms live next to each bug in `grader/internal/bugs/bugs.go`.

## Maintenance

```sh
go run ./grader/cmd/verify       # every bug in isolation: builds, fails exactly its own test, vet-clean except t3-mutex-copy
go run ./grader/cmd/genpatches   # human-readable diffs in bugs/
```

Run `verify` after any change to the baseline or to `bugs.go`. If a replacement anchor no
longer matches the baseline, apply fails loudly instead of planting garbage.

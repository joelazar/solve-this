# grader

Plants bugs into [app](../app) and scores agent runs against hidden tests. Everything
below runs from the repository root; run trees are created outside the repository so
agents never see this directory.

`grader/internal/bugs/bugs.go` is the answer key: 19 bugs as exact text replacements against the
clean baseline, each tied to the spec sentence it breaks and the hidden test that covers
it. Two decoys (weird-looking but correct code baked into the baseline) are listed there
too. Tests in `tests/` are black-box HTTP checks against the built binary; the
`TestRegression*` tests guard behavior no bug touches.

Tier 3 is opt-in via `-tiers` and covers concurrency: two data races, a mutex copied by a
value receiver, and a lock leaked on an error path. Random selection defaults to tiers 1
and 2, so existing runs are unaffected; `-tiers 1,2,3 -n 19` plants everything. The race
tests hammer a `-race` build of the binary and grep its log for reports naming the buggy
site, so they reward agents who run the race detector under load rather than only reading
code. `t3-mutex-copy` is the one bug that is not vet-clean: `go vet` flags the copied
lock, and `verify` expects exactly that.

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

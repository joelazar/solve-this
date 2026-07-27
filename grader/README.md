# solve-this-grader

Plants bugs into [solve-this](../solve-this) and scores agent runs against hidden tests.
Keep this repo away from the agent.

`internal/bugs/bugs.go` is the answer key: 15 bugs as exact text replacements against the
clean baseline, each tied to the spec sentence it breaks and the hidden test that covers
it. Two decoys (weird-looking but correct code baked into the baseline) are listed there
too. Tests in `tests/` are black-box HTTP checks against the built binary; the
`TestRegression*` tests guard behavior no bug touches.

## Workflow

One run = one buggy tree + one prompt + one agent + one scorecard, keyed by `-id`.
The same `-seed`/`-n` (or `-bugs`) always plants the same set, so give each agent its
own `-id` and reuse the seed for a fair comparison.

```sh
# headless agents, end to end
go run ./cmd/bench -id s42-claude -seed 42 -n 10 -mode hunt \
  -agent 'claude --dangerously-skip-permissions -p "$PROMPT"'
go run ./cmd/bench -id s42-codex -seed 42 -n 10 -mode hunt \
  -agent 'codex exec --full-auto "$PROMPT"'

# interactive agents: generate the tree, paste the printed prompt, score afterwards
go run ./cmd/bench -id s42-pi -seed 42 -n 10 -mode count
go run ./cmd/score -id s42-pi

# hand the agent a bug report instead of a hunt
go run ./cmd/bench -id repro1 -bugs t2-slice-alias -mode report -agent '...'
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
| `prompts/hunt.md` | find and fix everything, count unknown |
| `prompts/count.md` | states the exact count via `{{.N}}` |
| `prompts/report.md` | user-style bug reports via `{{.Symptoms}}` |

They are Go templates rendered against the planted set; edit them freely or pass
`-prompt my-prompt.md`. Symptoms live next to each bug in `internal/bugs/bugs.go`.

## Maintenance

```sh
go run ./cmd/verify       # every bug in isolation: builds, vet-clean, fails exactly its own test
go run ./cmd/genpatches   # human-readable diffs in bugs/
```

Run `verify` after any change to the baseline or to `bugs.go`. If a replacement anchor no
longer matches the baseline, apply fails loudly instead of planting garbage.

## Prompt modes

- blind hunt: "Find and fix all bugs."
- bug report: plant one bug with `-bugs`, hand the agent its symptom.
- known count: "There are exactly N bugs." with `-n N`.

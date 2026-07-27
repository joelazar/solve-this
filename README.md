# solve-this

Coding agent benchmark: a Go todo list HTTP API with a corpus of plantable bugs,
hidden black-box tests and a scorer.

```
app/      clean, correct baseline; exported into a fresh tree for every run
grader/   answer key, hidden tests, bench and scoring tools
runs/     per-run manifests, prompts, agent logs, scorecards (gitignored)
```

Agents never see this repository. Each run gets a standalone copy of `app/` with bugs
planted and history squashed, created outside this repo in
`~/Code/joelazar/solve-this-runs/`, so an agent exploring its filesystem finds nothing.

## Quickstart

```sh
mise run bench -- -id s42-claude -seed 42 -n 10 -mode hunt \
  -agent 'claude --dangerously-skip-permissions -p "$PROMPT"'

mise run bench -- -id s42-pi -seed 42 -n 10 -mode count   # interactive: paste the prompt
mise run score -- -id s42-pi                              # score when the agent is done
```

| task | purpose |
| --- | --- |
| `bench` | plant bugs, run an agent, score it |
| `score` | (re)score an existing run |
| `verify` | corpus health: every bug fails exactly its own hidden test |
| `test` | hidden tests against the working-tree baseline |
| `check` | gofmt, build and vet both modules |
| `patches` | regenerate reviewable bug diffs into `bugs/` |
| `api` | run the baseline API on :8080 |

The full workflow lives in [grader/README.md](grader/README.md).

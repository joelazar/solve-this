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
go run ./grader/cmd/bench -id s42-claude -seed 42 -n 10 -mode hunt \
  -agent 'claude --dangerously-skip-permissions -p "$PROMPT"'
```

The full workflow lives in [grader/README.md](grader/README.md).

# solve-this

A benchmark for coding agents. The subject is a small Go todo list HTTP API; the
harness plants known bugs into a copy of it, hands the copy to an agent with a prompt,
and scores the result against hidden tests.

```
app/      clean, correct baseline; exported into a fresh tree for every run
grader/   answer key, hidden tests, bench and scoring tools
runs/     per-run manifests, prompts, agent logs, scorecards (gitignored)
```

The corpus holds 19 bugs in three tiers (mechanical mistakes, semantic traps,
concurrency defects) plus two decoys: correct code that looks buggy and costs points
when "fixed". Each bug breaks a documented behavior and is covered by exactly one
hidden test. The [bug catalog](grader/docs/bugs/README.md) explains every one with
code.

Agents never see this repository. Each run gets a standalone copy of `app/` with bugs
planted and git history squashed to a single commit, created outside this repo, so an
agent exploring its filesystem finds neither the answer key nor the tests.

## Setup

You need `git` and Go. The repo pins its Go version with [mise](https://mise.jdx.dev),
which is the easy path:

```sh
git clone <this-repo>
cd solve-this
mise install
```

Without mise, any Go matching the version in `mise.toml` works; run the commands from
the task definitions there directly (every task is a plain `go run` or `go test`).

Check that the baseline is healthy before benchmarking anything:

```sh
mise run test     # hidden tests against the clean app, should be all green
mise run verify   # every bug in isolation fails exactly its own test (takes a while)
```

## Running a benchmark

One run = one buggy tree + one prompt + one agent + one scorecard. For an agent with a
headless CLI, one command does everything:

```sh
mise run bench -- -id s42-claude -seed 42 -n 10 -mode hunt \
  -agent 'claude --dangerously-skip-permissions -p "$PROMPT"'
```

This plants 10 seed-42 bugs into a fresh tree, runs the agent command inside it with
the prompt in `$PROMPT`, then builds, vets and tests the result and prints a
scorecard. The same `-seed` and `-n` always plant the same set, so give each agent its
own `-id` and reuse the seed to compare them fairly.

For an interactive agent (an editor session, a chat window), leave `-agent` off: bench
prints the prompt and the run directory, you drive the agent yourself, then score
afterwards:

```sh
mise run bench -- -id s42-pi -seed 42 -n 10 -mode count
# work in the printed run dir, then:
mise run score -- -id s42-pi
```

Prompt modes range from `vague` (two sentences, no hints) through `hunt` and `count`
to `report`, which turns the planted bugs into user-style bug reports. Concurrency
bugs are opt-in: add `-tiers 1,2,3`. Details, prompt templates and maintenance
commands live in [grader/README.md](grader/README.md).

## Where things land

Two directories are involved per run, both configurable:

- The buggy tree the agent works in. Defaults to `../solve-this-runs/<id>` next to
  your checkout, so agents cannot stumble into this repo; override with `-out`.
- The run artifacts: `<id>.json` manifest, `<id>-prompt.md`, `<id>-agent.log` and
  `<id>-score.json`. Default `runs/` inside the repo; override with `-manifests`
  (score needs the same value to find the manifest).

```sh
mise run bench -- -id demo -out /tmp/trees -manifests /tmp/artifacts ...
mise run score -- -id demo -manifests /tmp/artifacts
```

## Tasks

| task | purpose |
| --- | --- |
| `bench` | plant bugs, run an agent, score it |
| `score` | (re)score an existing run |
| `verify` | corpus health: every bug fails exactly its own hidden test |
| `test` | hidden tests against the working-tree baseline |
| `check` | gofmt, build and vet both modules |
| `patches` | regenerate reviewable bug diffs into `bugs/` |
| `api` | run the baseline API on :8080 |

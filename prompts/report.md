You are working on a Go HTTP service, a todo list API. This repository
contains its full source. docs/api.md specifies the intended behavior and
is the source of truth; README.md explains how to run the service.

Users filed the following bug reports:

{{range .Symptoms}}- {{.}}
{{end}}
Find the root cause of each report and fix it.

Rules:
- Fix the code, never the spec.
- Keep each fix minimal: no refactoring, no renaming, no new features,
  no new dependencies.
- You may write temporary scripts or tests to investigate, but delete
  them before you finish.
- The project must still build with: go build ./...

When you are done, explain each root cause in one line.

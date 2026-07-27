You are working on a Go HTTP service, a todo list API. This repository
contains its full source. docs/api.md specifies the intended behavior and
is the source of truth; README.md explains how to run the service.

The implementation contains exactly {{.N}} bugs: places where the code
deviates from the spec. Find all {{.N}} and fix them. Everything else,
however odd it may look, is intentional and must keep working.

Rules:
- Fix the code, never the spec.
- Keep each fix minimal: no refactoring, no renaming, no new features,
  no new dependencies.
- You may write temporary scripts or tests to investigate, but delete
  them before you finish.
- The project must still build with: go build ./...

When you are done, list all {{.N}} bugs, one line each: what was wrong
and where.

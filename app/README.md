# solve-this

A todo list HTTP API: lists hold tasks, tasks carry tags, priorities and due dates.
Everything is kept in memory, the standard library is the only dependency.

## Run

```sh
go run ./cmd/api            # listens on :8080
go run ./cmd/api -addr :9000
```

```sh
curl -X POST localhost:8080/lists -d '{"name":"Home"}'
curl -X POST localhost:8080/lists/list_0001/tasks -d '{"title":"Buy milk","tags":["groceries"]}'
curl 'localhost:8080/tasks?sort=priority&per_page=10'
```

The full contract lives in [docs/api.md](docs/api.md).

## Layout

```
cmd/api            entry point
internal/domain    tasks, lists, tags, validation
internal/store     in-memory tables, filtering, sorting, pagination
internal/api       routing, handlers, middleware, JSON rendering
```

Ids are sequential (`list_0001`, `task_0001`) and the store keeps tasks in creation order.

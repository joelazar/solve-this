# t2-slice-alias

Tier 2. Planted in `internal/store/store.go`, caught by `TestT2SliceAlias`.

`Tasks` hands the caller the store's internal slice:

```go
func (s *Store) Tasks() []domain.Task {
	return s.tasks.rows
}
```

The `GET /tasks` handler sorts whatever `Tasks` returns, in place. With the alias, a
single `GET /tasks?sort=title` reorders the store's own rows. The table keeps an
`index` map from id to row position, and that map still points at the old positions,
so `GET /tasks/{id}` starts resolving to the wrong task for ids that exist. The CSV
export shows the creation order is gone too.

The damage is done by an earlier request and persists, which makes this one confusing
to reproduce: reads look fine until someone sorts.

The baseline returns deep copies:

```go
tasks := make([]domain.Task, len(s.tasks.rows))
for i, task := range s.tasks.rows {
	tasks[i] = task.Clone()
}
return tasks
```

The test pins what the spec pins: the sorted page comes back sorted, ids keep resolving to
their own task, and the CSV export still lists the creation order. A shallow
`copy(tasks, s.tasks.rows)` satisfies all three and counts as fixed. It leaves the tag
maps shared with the stored rows, which is invisible from outside because nothing mutates
the tasks handed out by `Tasks`: every write path clones the row first. The baseline clones
anyway, so the shape of the fix stays a matter of taste and not of score.

The spec sentence it breaks: the stored order of tasks is their creation order and no
request can change it.

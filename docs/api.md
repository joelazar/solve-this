# API reference

Base URL `http://localhost:8080`. All request and response bodies are JSON unless stated
otherwise. Timestamps are RFC3339 with nanosecond precision, in the server's local zone.

## Conventions

| Situation | Status |
| --- | --- |
| Resource created | `201` |
| Resource read or updated | `200` |
| Resource deleted | `204`, empty body |
| Malformed body or query parameter | `400` |
| Unknown list, task or tag | `404` |

Error bodies look like `{"error": "...", "fields": [{"field": "title", "message": "..."}]}`.
The `fields` array is present only for validation failures.

Every response carries an `X-Request-Id` header.

## Objects

### List

```json
{ "id": "list_0001", "name": "Home", "created_at": "2024-05-01T09:00:00Z" }
```

### Task

```json
{
  "id": "task_0001",
  "list_id": "list_0001",
  "title": "Buy milk",
  "notes": "",
  "done": false,
  "priority": "medium",
  "tags": ["groceries"],
  "due": null,
  "created_at": "2024-05-01T09:00:00Z",
  "updated_at": "2024-05-01T09:00:00Z"
}
```

`priority` is one of `high`, `medium`, `low` and defaults to `medium`.
`tags` is always sorted alphabetically and never contains duplicates.
`due` is `null` or an RFC3339 timestamp.
`updated_at` is refreshed on every mutation of the task and equals `created_at` until the
first mutation.

### Tag normalisation

A tag is normalised before it is stored or compared: surrounding whitespace is trimmed, the
tag is lowercased, and a single leading `tag:` prefix is removed. `" Tag:Groceries "` and
`"groceries"` are the same tag. A tag that normalises to the empty string is rejected with
`400`; a tag longer than 40 characters is rejected with `400`.

## Lists

### `POST /lists`

Body `{"name": "Home"}`. The name must be non-empty after trimming and at most 100
characters. Returns `201` with the list.

### `GET /lists`

Returns `200` with `{"items": [...]}` in creation order.

### `GET /lists/{id}`

Returns `200` with the list, or `404`.

### `DELETE /lists/{id}`

Deletes the list and every task belonging to it. Returns `204`, or `404` if the list does
not exist.

### `POST /lists/{id}/tasks`

Creates a task in the list. Returns `404` if the list does not exist.

```json
{
  "title": "Buy milk",
  "notes": "",
  "priority": "low",
  "tags": ["groceries"],
  "due": "2024-06-01T12:00:00Z"
}
```

Only `title` is required. It must be non-empty after trimming and at most 200 characters;
`notes` is at most 2000 characters; `priority` must be a valid priority; `due` must parse as
RFC3339. A valid request returns `201` with the created task.

## Tasks

### `GET /tasks`

Returns a page of tasks.

| Parameter | Default | Meaning |
| --- | --- | --- |
| `list` | – | keep tasks belonging to this list id |
| `tag` | – | keep tasks carrying this tag, normalised before comparison |
| `done` | – | `true` or `false` |
| `q` | – | case-insensitive substring of `title` or `notes` |
| `sort` | `created` | sort key, see below |
| `page` | `1` | 1-based page number |
| `per_page` | `20` | page size, between 1 and 100 |

`page` and `per_page` must be integers within their range; anything else is `400`.
`done` must be `true` or `false`; anything else is `400`. An unknown `sort` key is `400`.

Sort keys are `created`, `title`, `priority` and `due`, each optionally prefixed with `-`
to reverse the direction. Ties are broken by ascending `id`, so the order of a page is
fully determined.

- `created` orders by creation time, oldest first. This is the default order.
- `title` orders case-insensitively.
- `priority` orders `high`, then `medium`, then `low`.
- `due` orders earliest first; tasks without a due date come last.

Sorting affects the response only. The stored order of tasks is their creation order and no
request can change it.

Response:

```json
{ "items": [], "page": 1, "per_page": 20, "total": 0, "total_pages": 0 }
```

`total` counts every task matching the filters, ignoring pagination. `total_pages` is
`total` divided by `per_page`, rounded up. `items` holds at most `per_page` tasks and holds
exactly the remaining tasks on the final page. A `page` beyond the final page returns an
empty `items` array with the same `total`.

### `GET /tasks/{id}`

Returns `200` with the task, or `404`.

### `PATCH /tasks/{id}`

Updates the fields present in the body and leaves every other field untouched. Accepts
`title`, `notes`, `done`, `priority`, `tags` and `due`. `tags` replaces the whole tag set.
Sending `"due": ""` clears the due date. The updated task is validated with the same rules
as creation. Returns `200` with the task, `400` on a validation failure, or `404`.

### `DELETE /tasks/{id}`

Returns `204`, or `404`.

### `POST /tasks/{id}/complete`, `POST /tasks/{id}/reopen`

Sets `done` to `true` or `false`. Empty body. Returns `200` with the task, or `404`.

### `POST /tasks/{id}/tags`

Body `{"tag": "groceries"}`. Adds the normalised tag to the task; adding a tag the task
already carries is a no-op. Returns `200` with the task, or `404` if the task does not
exist.

### `DELETE /tasks/{id}/tags/{tag}`

Removes the normalised tag. Returns `200` with the task, `404` if the task does not exist
or does not carry the tag.

### `POST /tasks/bulk/complete`

Body `{"ids": ["task_0001", "task_0002"]}`. Marks every listed task as done and persists
the change, so a later `GET /tasks/{id}` reports `done: true`. The list must not be empty.
If any id is unknown the request fails with `404` and no task is modified.

```json
{ "completed": 2, "items": [] }
```

## Reports

### `GET /stats`

```json
{
  "total": 3,
  "done": 1,
  "open": 2,
  "overdue": 1,
  "completion_percent": 33,
  "by_priority": { "high": 1, "medium": 1, "low": 1 },
  "by_tag": [{ "tag": "groceries", "count": 2 }]
}
```

A task is overdue when it is not done and its due date is in the past.
`completion_percent` is `done / total` as a percentage, rounded down, and `0` when there
are no tasks. `by_tag` counts how many tasks carry each tag, ordered by descending count
and then by ascending tag name. Tasks without tags contribute nothing to `by_tag`.

### `GET /export.csv`

Returns `text/csv` with the header row
`id,list_id,title,done,priority,tags,due,created_at` followed by one row per task in
creation order. Tags are space separated.

### `GET /health`

Returns `200` with `{"status": "ok"}`.

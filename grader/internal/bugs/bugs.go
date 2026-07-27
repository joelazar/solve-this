package bugs

type Edit struct {
	File string
	Old  string
	New  string
}

type Bug struct {
	ID      string
	Tier    int
	Class   string
	Spec    string
	Symptom string
	Test    string
	Edits   []Edit
}

type Decoy struct {
	ID    string
	Looks string
	Test  string
}

var Decoys = []Decoy{
	{
		ID:    "d1-percent-truncation",
		Looks: "integer division in completion_percent looks like a rounding bug",
		Test:  "TestDecoyCompletionPercentFloor",
	},
	{
		ID:    "d2-concrete-error",
		Looks: "ValidateListName has the exact shape of the typed nil trap but returns a concrete pointer",
		Test:  "TestDecoyListNameValidation",
	},
}

var All = []Bug{
	{
		ID:      "t1-range-copy",
		Tier:    1,
		Class:   "range over values mutates a copy",
		Spec:    "bulk complete persists, a later GET reports done: true",
		Symptom: "POST /tasks/bulk/complete reports the tasks as completed, but fetching them afterwards still shows done: false.",
		Test:    "TestT1RangeCopy",
		Edits: []Edit{{
			File: "internal/store/store.go",
			Old: `	completed := make([]domain.Task, 0, len(ids))
	for _, id := range ids {
		i, ok := s.tasks.at(id)
		if !ok {
			return nil, fmt.Errorf("task %s: %w", id, ErrNotFound)
		}
		s.tasks.rows[i].Done = true
		s.tasks.rows[i].Touch()
		completed = append(completed, s.tasks.rows[i].Clone())
	}
	return completed, nil`,
			New: `	wanted := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if _, ok := s.tasks.at(id); !ok {
			return nil, fmt.Errorf("task %s: %w", id, ErrNotFound)
		}
		wanted[id] = struct{}{}
	}
	completed := make([]domain.Task, 0, len(ids))
	for _, task := range s.tasks.rows {
		if _, ok := wanted[task.ID]; !ok {
			continue
		}
		task.Done = true
		task.Touch()
		completed = append(completed, task.Clone())
	}
	return completed, nil`,
		}},
	},
	{
		ID:      "t1-slice-bounds",
		Tier:    1,
		Class:   "missing bounds check",
		Spec:    "a page beyond the final page returns an empty items array",
		Symptom: "GET /tasks?page=99 returns 500; the spec says a page past the end is an empty list.",
		Test:    "TestT1SliceBounds",
		Edits: []Edit{{
			File: "internal/store/page.go",
			Old: `	start := (number - 1) * size
	if start > total {
		start = total
	}
	end := start + size`,
			New: `	start := (number - 1) * size
	end := start + size`,
		}},
	},
	{
		ID:      "t1-atoi-ignored",
		Tier:    1,
		Class:   "discarded conversion error",
		Spec:    "per_page must be an integer between 1 and 100, anything else is 400",
		Symptom: "GET /tasks?per_page=abc returns 500 instead of 400.",
		Test:    "TestT1AtoiIgnored",
		Edits: []Edit{
			{
				File: "internal/api/tasks_list.go",
				Old: `import (
	"fmt"
	"net/http"`,
				New: `import (
	"net/http"`,
			},
			{
				File: "internal/api/tasks_list.go",
				Old: `	perPage, err := intParam(values, "per_page", defaultPerPage)
	if err != nil || perPage < 1 || perPage > maxPerPage {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("per_page must be an integer between 1 and %d", maxPerPage))
		return
	}`,
				New: `	perPage := defaultPerPage
	if raw := values.Get("per_page"); raw != "" {
		perPage, _ = strconv.Atoi(raw)
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}`,
			},
		},
	},
	{
		ID:      "t1-nil-map",
		Tier:    1,
		Class:   "write to a nil map",
		Spec:    "by_tag counts how many tasks carry each tag",
		Symptom: "GET /stats returns 500 as soon as any task carries a tag.",
		Test:    "TestT1NilMap",
		Edits: []Edit{{
			File: "internal/api/stats.go",
			Old:  `	counts := make(map[string]int)`,
			New:  `	var counts map[string]int`,
		}},
	},
	{
		ID:      "t1-value-receiver",
		Tier:    1,
		Class:   "value receiver on a mutating method",
		Spec:    "updated_at is refreshed on every mutation of the task",
		Symptom: "updated_at never changes; after a PATCH it still equals created_at.",
		Test:    "TestT1ValueReceiver",
		Edits: []Edit{{
			File: "internal/domain/task.go",
			Old:  `func (t *Task) Touch() {`,
			New:  `func (t Task) Touch() {`,
		}},
	},
	{
		ID:      "t1-make-append",
		Tier:    1,
		Class:   "make with length instead of capacity",
		Spec:    "items holds at most per_page tasks",
		Symptom: "GET /tasks returns empty placeholder items in front of the real ones, twice the expected count.",
		Test:    "TestT1MakeAppend",
		Edits: []Edit{{
			File: "internal/api/render.go",
			Old:  `	dtos := make([]taskDTO, 0, len(tasks))`,
			New:  `	dtos := make([]taskDTO, len(tasks))`,
		}},
	},
	{
		ID:      "t2-typed-nil",
		Tier:    2,
		Class:   "nil pointer boxed in a non-nil error interface",
		Spec:    "adding a valid tag returns 200 with the task",
		Symptom: "POST /tasks/{id}/tags returns 500 for perfectly valid tags.",
		Test:    "TestT2TypedNil",
		Edits: []Edit{{
			File: "internal/domain/validate.go",
			Old: `func ValidateTag(tag string) error {
	var fields []FieldError
	switch {
	case tag == "":
		fields = append(fields, FieldError{Field: "tag", Message: "must not be empty"})
	case len(tag) > MaxTagLen:
		fields = append(fields, FieldError{Field: "tag", Message: fmt.Sprintf("must be at most %d characters", MaxTagLen)})
	}
	if len(fields) == 0 {
		return nil
	}
	return &ValidationError{Fields: fields}
}`,
			New: `func ValidateTag(tag string) error {
	var invalid *ValidationError
	switch {
	case tag == "":
		invalid = &ValidationError{Fields: []FieldError{{Field: "tag", Message: "must not be empty"}}}
	case len(tag) > MaxTagLen:
		invalid = &ValidationError{Fields: []FieldError{{Field: "tag", Message: fmt.Sprintf("must be at most %d characters", MaxTagLen)}}}
	}
	return invalid
}`,
		}},
	},
	{
		ID:      "t2-err-shadow",
		Tier:    2,
		Class:   "error shadowed by an inner declaration",
		Spec:    "deleting a list that does not exist is 404",
		Symptom: "DELETE /lists/{id} answers 204 for list ids that never existed.",
		Test:    "TestT2ErrShadow",
		Edits: []Edit{{
			File: "internal/api/lists.go",
			Old: `	list, err := h.store.List(id)
	if err != nil {
		failed(w, err)
		return
	}
	if err := h.store.DeleteList(list.ID); err != nil {
		failed(w, err)
		return
	}`,
			New: `	var err error
	if list, err := h.store.List(id); err == nil {
		err = h.store.DeleteList(list.ID)
	}
	if err != nil {
		failed(w, err)
		return
	}`,
		}},
	},
	{
		ID:      "t2-errors-is",
		Tier:    2,
		Class:   "wrapped error compared with ==",
		Spec:    "an unknown task id is 404",
		Symptom: "Requests for unknown task or list ids return 500 instead of 404.",
		Test:    "TestT2ErrorsIs",
		Edits: []Edit{{
			File: "internal/api/errors.go",
			Old:  `	case errors.Is(err, store.ErrNotFound):`,
			New:  `	case err == store.ErrNotFound:`,
		}},
	},
	{
		ID:      "t2-slice-alias",
		Tier:    2,
		Class:   "internal slice handed to the caller",
		Spec:    "the stored order of tasks is their creation order and no request can change it",
		Symptom: "After anyone calls GET /tasks?sort=title, GET /tasks/{id} starts returning the wrong task for existing ids.",
		Test:    "TestT2SliceAlias",
		Edits: []Edit{{
			File: "internal/store/store.go",
			Old: `	tasks := make([]domain.Task, len(s.tasks.rows))
	for i, task := range s.tasks.rows {
		tasks[i] = task.Clone()
	}
	return tasks`,
			New: `	return s.tasks.rows`,
		}},
	},
	{
		ID:      "t2-sort-order",
		Tier:    2,
		Class:   "ordering by string instead of rank",
		Spec:    "priority orders high, then medium, then low",
		Symptom: "GET /tasks?sort=priority orders tasks high, low, medium instead of high, medium, low.",
		Test:    "TestT2SortOrder",
		Edits: []Edit{{
			File: "internal/store/query.go",
			Old:  `		return a.Priority.Rank() - b.Priority.Rank()`,
			New:  `		return strings.Compare(string(a.Priority), string(b.Priority))`,
		}},
	},
	{
		ID:      "t2-trim-cutset",
		Tier:    2,
		Class:   "TrimLeft used as TrimPrefix",
		Spec:    "a single leading tag: prefix is removed, nothing else",
		Symptom: "Some tags lose their leading letters: 'groceries' comes back as 'roceries'.",
		Test:    "TestT2TrimCutset",
		Edits: []Edit{{
			File: "internal/domain/tag.go",
			Old:  `	return strings.TrimPrefix(strings.ToLower(strings.TrimSpace(tag)), TagPrefix)`,
			New:  `	return strings.TrimLeft(strings.ToLower(strings.TrimSpace(tag)), TagPrefix)`,
		}},
	},
	{
		ID:      "t2-total-pages",
		Tier:    2,
		Class:   "integer division where rounding up is required",
		Spec:    "total_pages is total divided by per_page, rounded up",
		Symptom: "With 3 tasks and per_page=2, total_pages reports 1 although page=2 returns the third task.",
		Test:    "TestT2TotalPages",
		Edits: []Edit{{
			File: "internal/api/tasks_list.go",
			Old:  `		TotalPages: (result.Total + result.Size - 1) / result.Size,`,
			New:  `		TotalPages: result.Total / result.Size,`,
		}},
	},
	{
		ID:      "t2-write-header",
		Tier:    2,
		Class:   "WriteHeader after Write",
		Spec:    "a valid create request returns 201",
		Symptom: "Creating a task returns 200 instead of the documented 201.",
		Test:    "TestT2WriteHeader",
		Edits: []Edit{{
			File: "internal/api/tasks_write.go",
			Old:  `	writeJSON(w, http.StatusCreated, taskToDTO(task))`,
			New: `	body, err := json.Marshal(taskToDTO(task))
	if err != nil {
		failed(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
	w.WriteHeader(http.StatusCreated)`,
		}},
	},
	{
		ID:      "t2-patch-zeroing",
		Tier:    2,
		Class:   "non pointer field in a partial update",
		Spec:    "PATCH updates the fields present in the body and leaves every other field untouched",
		Symptom: "PATCHing only the title of a completed task silently reopens it: done flips back to false.",
		Test:    "TestT2PatchZeroing",
		Edits: []Edit{
			{
				File: "internal/api/tasks_write.go",
				Old:  "	Done     *bool     `json:\"done\"`",
				New:  "	Done     bool      `json:\"done\"`",
			},
			{
				File: "internal/api/tasks_write.go",
				Old: `		if req.Done != nil {
			task.Done = *req.Done
		}`,
				New: `		task.Done = req.Done`,
			},
		},
	},
}

func ByID(id string) (Bug, bool) {
	for _, bug := range All {
		if bug.ID == id {
			return bug, true
		}
	}
	return Bug{}, false
}

func TestOwner(test string) (string, bool) {
	for _, bug := range All {
		if bug.Test == test {
			return bug.ID, true
		}
	}
	for _, decoy := range Decoys {
		if decoy.Test == test {
			return decoy.ID, true
		}
	}
	return "", false
}

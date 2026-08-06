package bugs

type Edit struct {
	File string
	Old  string
	New  string
}

type Bug struct {
	ID       string
	Tier     int
	Class    string
	Spec     string
	Symptom  string
	Test     string
	TripsVet bool
	Edits    []Edit
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
		ID:      "t1-row-copy",
		Tier:    1,
		Class:   "a copy of the stored row is mutated",
		Spec:    "bulk complete persists, a later GET reports done: true",
		Symptom: "POST /tasks/bulk/complete reports the tasks as completed, but fetching them afterwards still shows done: false.",
		Test:    "TestT1RowCopy",
		Edits: []Edit{{
			File: "internal/store/store.go",
			Old: `	for _, i := range rows {
		s.tasks.rows[i].Done = true
		s.tasks.rows[i].Touch()
		completed = append(completed, s.tasks.rows[i].Clone())
	}`,
			New: `	for _, i := range rows {
		task := s.tasks.rows[i]
		task.Done = true
		task.Touch()
		completed = append(completed, task.Clone())
	}`,
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
	if message := tagProblem(tag); message != "" {
		return &ValidationError{Fields: []FieldError{{Field: "tag", Message: message}}}
	}
	return nil
}`,
			New: `func ValidateTag(tag string) error {
	var invalid *ValidationError
	if message := tagProblem(tag); message != "" {
		invalid = &ValidationError{Fields: []FieldError{{Field: "tag", Message: message}}}
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
	{
		ID:      "t3-race-request-id",
		Tier:    3,
		Class:   "unsynchronized shared counter",
		Spec:    "every response carries its own X-Request-Id header",
		Symptom: "Under concurrent load some responses share the same X-Request-Id.",
		Test:    "TestT3RaceRequestID",
		Edits: []Edit{
			{
				File: "internal/api/middleware.go",
				Old: `	"strconv"
	"sync/atomic"
	"time"`,
				New: `	"strconv"
	"time"`,
			},
			{
				File: "internal/api/middleware.go",
				Old: `var requestCounter atomic.Uint64

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := "req-" + strconv.FormatUint(requestCounter.Add(1), 10)`,
				New: `var requestCounter uint64

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCounter++
		id := "req-" + strconv.FormatUint(requestCounter, 10)`,
			},
		},
	},
	{
		ID:      "t3-race-tasks",
		Tier:    3,
		Class:   "shared state read without the lock",
		Spec:    "GET /tasks returns a consistent snapshot while other requests mutate the store",
		Symptom: "Under load, GET /tasks sometimes returns corrupted results or crashes.",
		Test:    "TestT3RaceTasks",
		Edits: []Edit{{
			File: "internal/store/store.go",
			Old: `func (s *Store) Tasks() []domain.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
`,
			New: `func (s *Store) Tasks() []domain.Task {
`,
		}},
	},
	{
		ID:       "t3-mutex-copy",
		Tier:     3,
		Class:    "mutex locked on a copy of the receiver",
		Spec:     "bulk complete mutates the store under mutual exclusion; go vet is clean",
		Symptom:  "go vet fails on the store package, and bulk completes corrupt data under load.",
		Test:     "TestT3MutexCopy",
		TripsVet: true,
		Edits: []Edit{{
			File: "internal/store/store.go",
			Old:  `func (s *Store) CompleteAll(`,
			New:  `func (s Store) CompleteAll(`,
		}},
	},
	{
		ID:      "t3-deadlock-delete",
		Tier:    3,
		Class:   "lock not released on the error path",
		Spec:    "deleting an unknown task returns 404 and the API keeps serving",
		Symptom: "After a DELETE of a task id that does not exist, the whole API stops responding.",
		Test:    "TestT3DeadlockDelete",
		Edits: []Edit{{
			File: "internal/store/store.go",
			Old: `func (s *Store) DeleteTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.tasks.remove(id) {
		return fmt.Errorf("task %s: %w", id, ErrNotFound)
	}
	return nil
}`,
			New: `func (s *Store) DeleteTask(id string) error {
	s.mu.Lock()
	if !s.tasks.remove(id) {
		return fmt.Errorf("task %s: %w", id, ErrNotFound)
	}
	s.mu.Unlock()
	return nil
}`,
		}},
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

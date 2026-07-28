package tests

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"
)

func TestRegressionHealth(t *testing.T) {
	s := start(t)

	status, body := s.call(http.MethodGet, "/health", "")
	if status != http.StatusOK || text(t, body, "status") != "ok" {
		t.Fatalf("health: %d %v", status, body)
	}
}

func TestRegressionFilters(t *testing.T) {
	s := start(t)
	home := s.list("Home")
	work := s.list("Work")
	groceries := text(t, s.task(home, `{"title":"buy milk","tags":["errand"]}`), "id")
	report := text(t, s.task(work, `{"title":"write report","notes":"quarterly"}`), "id")
	s.call(http.MethodPost, "/tasks/"+report+"/complete", "")

	for _, tc := range []struct {
		query   string
		present string
		absent  string
	}{
		{"?list=" + home, groceries, report},
		{"?done=true", report, groceries},
		{"?done=false", groceries, report},
		{"?q=QUARTERLY", report, groceries},
		{"?q=milk", groceries, report},
		{"?tag=errand", groceries, report},
	} {
		_, body := s.call(http.MethodGet, "/tasks"+tc.query, "")
		ids := field(t, filled(t, body), "id")
		if !slices.Contains(ids, tc.present) {
			t.Errorf("%s: %s missing from %v", tc.query, tc.present, ids)
		}
		if slices.Contains(ids, tc.absent) {
			t.Errorf("%s: %s should not match, got %v", tc.query, tc.absent, ids)
		}
	}
}

func TestRegressionValidation(t *testing.T) {
	s := start(t)
	list := s.list("Home")

	for _, body := range []string{
		`{"title":"   "}`,
		`{"title":"one","priority":"urgent"}`,
		`{"title":"one","due":"tomorrow"}`,
		`{"title":"one","tags":[""]}`,
		`{"title":"one","tags":["   "]}`,
		`{"title":"one","tags":["tag:"]}`,
		`{"title":"one","tags":["groceries","  "]}`,
		`{"title":"one","tags":["` + strings.Repeat("x", 41) + `"]}`,
		`{"title":"` + strings.Repeat("x", 201) + `"}`,
		`{"title":"one","notes":"` + strings.Repeat("x", 2001) + `"}`,
	} {
		status, response := s.call(http.MethodPost, "/lists/"+list+"/tasks", body)
		if status != http.StatusBadRequest {
			t.Errorf("%s: %d %v", body, status, response)
		}
	}

	id := s.titled(list, "one")
	if status, response := s.call(http.MethodPatch, "/tasks/"+id, `{"tags":["tag:"]}`); status != http.StatusBadRequest {
		t.Errorf("patching an empty tag: %d %v", status, response)
	}
}

func TestRegressionQueryParams(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	s.titled(list, "one")

	for _, query := range []string{"done=1", "done=0", "done=t", "done=TRUE", "done=yes", "page=abc", "page=0", "page=-1"} {
		status, body := s.call(http.MethodGet, "/tasks?"+query, "")
		if status != http.StatusBadRequest {
			t.Errorf("%s: %d %v", query, status, body)
		}
	}
	for _, query := range []string{"done=true", "done=false", "page=1", "per_page=100"} {
		status, body := s.call(http.MethodGet, "/tasks?"+query, "")
		if status != http.StatusOK {
			t.Errorf("%s: %d %v", query, status, body)
		}
	}
}

func TestRegressionBulkResponse(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	first := s.titled(list, "one")
	second := s.titled(list, "two")

	status, body := s.call(http.MethodPost, "/tasks/bulk/complete", fmt.Sprintf(`{"ids":[%q,%q]}`, second, first))
	if status != http.StatusOK {
		t.Fatalf("bulk complete: %d %v", status, body)
	}
	if completed := number(t, body, "completed"); completed != 2 {
		t.Errorf("completed is %d", completed)
	}
	if ids := field(t, filled(t, body), "id"); !slices.Equal(ids, []string{second, first}) {
		t.Errorf("items are %v, want the ids in the order given", ids)
	}
}

func TestRegressionBulkAtomic(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	id := s.titled(list, "one")

	status, body := s.call(http.MethodPost, "/tasks/bulk/complete", fmt.Sprintf(`{"ids":[%q,"task_9999"]}`, id))
	if status < 400 {
		t.Fatalf("bulk complete with an unknown id reported success: %d %v", status, body)
	}
	_, task := s.call(http.MethodGet, "/tasks/"+id, "")
	if boolean(t, task, "done") {
		t.Fatal("a failed bulk complete modified a task")
	}
	if status, body := s.call(http.MethodPost, "/tasks/bulk/complete", `{"ids":[]}`); status != http.StatusBadRequest {
		t.Fatalf("empty ids: %d %v", status, body)
	}
}

func TestRegressionCascadeDelete(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	id := s.titled(list, "one")

	if status, body := s.call(http.MethodDelete, "/lists/"+list, ""); status != http.StatusNoContent {
		t.Fatalf("delete list: %d %v", status, body)
	}
	if status, _ := s.call(http.MethodGet, "/tasks/"+id, ""); status < 400 {
		t.Fatalf("task survived its list: %d", status)
	}
}

func TestRegressionTagRemoval(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	id := text(t, s.task(list, `{"title":"one","tags":["home"]}`), "id")

	status, task := s.call(http.MethodDelete, "/tasks/"+id+"/tags/home", "")
	if status != http.StatusOK {
		t.Fatalf("remove tag: %d %v", status, task)
	}
	if tags := stringList(t, task, "tags"); len(tags) != 0 {
		t.Fatalf("tags are %v", tags)
	}
	touched := text(t, task, "updated_at")

	if status, body := s.call(http.MethodDelete, "/tasks/"+id+"/tags/home", ""); status < 400 {
		t.Fatalf("removing a missing tag reported success: %d %v", status, body)
	}
	_, task = s.call(http.MethodGet, "/tasks/"+id, "")
	if now := text(t, task, "updated_at"); now != touched {
		t.Fatalf("a failed tag removal moved updated_at from %s to %s", touched, now)
	}
}

func TestRegressionLists(t *testing.T) {
	s := start(t)
	home := s.list("Home")
	work := s.list("Work")

	status, body := s.call(http.MethodGet, "/lists", "")
	if status != http.StatusOK {
		t.Fatalf("lists: %d %v", status, body)
	}
	if ids := field(t, items(t, body), "id"); !slices.Equal(ids, []string{home, work}) {
		t.Fatalf("lists are %v, want creation order", ids)
	}

	status, list := s.call(http.MethodGet, "/lists/"+work, "")
	if status != http.StatusOK || text(t, list, "name") != "Work" {
		t.Fatalf("list: %d %v", status, list)
	}
	if status, body := s.call(http.MethodGet, "/lists/list_9999", ""); status < 400 {
		t.Fatalf("unknown list reported success: %d %v", status, body)
	}
}

func TestRegressionTaskDefaults(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	task := s.task(list, `{"title":"one"}`)

	if priority := text(t, task, "priority"); priority != "medium" {
		t.Errorf("default priority is %q", priority)
	}
	if notes := text(t, task, "notes"); notes != "" {
		t.Errorf("default notes are %q", notes)
	}
	if boolean(t, task, "done") {
		t.Error("a new task is done")
	}
	if task["due"] != nil {
		t.Errorf("default due is %v", task["due"])
	}
	if text(t, task, "list_id") != list {
		t.Errorf("list_id is %q", task["list_id"])
	}
	if text(t, task, "updated_at") != text(t, task, "created_at") {
		t.Error("updated_at does not equal created_at before the first mutation")
	}
	if status, body := s.call(http.MethodPost, "/lists/list_9999/tasks", `{"title":"one"}`); status < 400 {
		t.Errorf("creating a task in an unknown list reported success: %d %v", status, body)
	}
}

func TestRegressionTagSet(t *testing.T) {
	s := start(t)
	list := s.list("Home")

	task := s.task(list, `{"title":"one","tags":["Zebra","milk","MILK"," milk "]}`)
	if tags := stringList(t, task, "tags"); !slices.Equal(tags, []string{"milk", "zebra"}) {
		t.Fatalf("tags are %v, want them sorted and deduplicated", tags)
	}
}

func TestRegressionPatchFields(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	id := text(t, s.task(list, `{"title":"one","tags":["milk"],"due":"2030-01-01T00:00:00Z"}`), "id")

	status, task := s.call(http.MethodPatch, "/tasks/"+id, `{"tags":["bread"]}`)
	if status != http.StatusOK {
		t.Fatalf("patch tags: %d %v", status, task)
	}
	if tags := stringList(t, task, "tags"); !slices.Equal(tags, []string{"bread"}) {
		t.Fatalf("tags are %v, want the whole set replaced", tags)
	}

	status, task = s.call(http.MethodPatch, "/tasks/"+id, `{"due":""}`)
	if status != http.StatusOK {
		t.Fatalf("clear due: %d %v", status, task)
	}
	if task["due"] != nil {
		t.Fatalf(`due is %v after patching ""`, task["due"])
	}

	if status, body := s.call(http.MethodPatch, "/tasks/task_9999", `{"title":"two"}`); status < 400 {
		t.Fatalf("patching an unknown task reported success: %d %v", status, body)
	}
}

func TestRegressionTaskLifecycle(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	id := s.titled(list, "one")

	status, task := s.call(http.MethodPost, "/tasks/"+id+"/complete", "")
	if status != http.StatusOK || !boolean(t, task, "done") {
		t.Fatalf("complete: %d %v", status, task)
	}
	status, task = s.call(http.MethodPost, "/tasks/"+id+"/reopen", "")
	if status != http.StatusOK || boolean(t, task, "done") {
		t.Fatalf("reopen: %d %v", status, task)
	}
	if status, body := s.call(http.MethodDelete, "/tasks/"+id, ""); status != http.StatusNoContent {
		t.Fatalf("delete task: %d %v", status, body)
	}
	if status, body := s.call(http.MethodGet, "/tasks/"+id, ""); status < 400 {
		t.Fatalf("a deleted task is still readable: %d %v", status, body)
	}
}

func TestRegressionSortKeys(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	s.task(list, `{"title":"bravo","due":"2030-01-01T00:00:00Z"}`)
	s.task(list, `{"title":"alpha","due":"2020-01-01T00:00:00Z"}`)
	s.task(list, `{"title":"charlie"}`)

	for query, want := range map[string][]string{
		"?sort=due":     {"alpha", "bravo", "charlie"},
		"?sort=title":   {"alpha", "bravo", "charlie"},
		"?sort=-title":  {"charlie", "bravo", "alpha"},
		"?sort=created": {"bravo", "alpha", "charlie"},
		"":              {"bravo", "alpha", "charlie"},
	} {
		_, body := s.call(http.MethodGet, "/tasks"+query, "")
		if titles := field(t, filled(t, body), "title"); !slices.Equal(titles, want) {
			t.Errorf("%q sorts to %v, want %v", query, titles, want)
		}
	}
	if status, body := s.call(http.MethodGet, "/tasks?sort=nonsense", ""); status != http.StatusBadRequest {
		t.Errorf("unknown sort key: %d %v", status, body)
	}
}

func TestRegressionPagination(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	s.titled(list, "one")
	s.titled(list, "two")
	third := s.titled(list, "three")

	_, body := s.call(http.MethodGet, "/tasks?per_page=2&page=2", "")
	if ids := field(t, filled(t, body), "id"); !slices.Equal(ids, []string{third}) {
		t.Fatalf("the last page holds %v, want the remaining task", ids)
	}
	if total := number(t, body, "total"); total != 3 {
		t.Fatalf("total is %d, want every match ignoring pagination", total)
	}
	if size := number(t, body, "per_page"); size != 2 {
		t.Fatalf("per_page is %d", size)
	}
}

func TestRegressionStats(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	s.task(list, `{"title":"one","priority":"high","due":"2020-01-01T00:00:00Z"}`)
	s.task(list, `{"title":"two","priority":"low","due":"2030-01-01T00:00:00Z"}`)
	done := text(t, s.task(list, `{"title":"three","due":"2020-01-01T00:00:00Z"}`), "id")
	s.call(http.MethodPost, "/tasks/"+done+"/complete", "")

	status, body := s.call(http.MethodGet, "/stats", "")
	if status != http.StatusOK {
		t.Fatalf("stats: %d %v", status, body)
	}
	for key, want := range map[string]int{"total": 3, "done": 1, "open": 2, "overdue": 1} {
		if got := number(t, body, key); got != want {
			t.Errorf("%s is %d, want %d", key, got, want)
		}
	}
	priorities := nested(t, body, "by_priority")
	for key, want := range map[string]int{"high": 1, "medium": 1, "low": 1} {
		if got := number(t, priorities, key); got != want {
			t.Errorf("by_priority[%s] is %d, want %d", key, got, want)
		}
	}
}

func TestRegressionErrorBody(t *testing.T) {
	s := start(t)
	list := s.list("Home")

	status, body := s.call(http.MethodPost, "/lists/"+list+"/tasks", `{"title":""}`)
	if status != http.StatusBadRequest {
		t.Fatalf("empty title: %d %v", status, body)
	}
	if text(t, body, "error") == "" {
		t.Error("a validation failure carries no error message")
	}
	fields, ok := body["fields"].([]any)
	if !ok || len(fields) == 0 {
		t.Fatalf("a validation failure carries no fields: %v", body)
	}
	for _, entry := range fields {
		row, ok := entry.(map[string]any)
		if !ok || text(t, row, "field") == "" || text(t, row, "message") == "" {
			t.Errorf("field error is %v", entry)
		}
	}

	status, body = s.call(http.MethodGet, "/tasks/task_9999", "")
	if status < 400 {
		t.Fatalf("unknown task reported success: %d %v", status, body)
	}
	if _, ok := body["fields"]; ok {
		t.Errorf("a lookup failure carries validation fields: %v", body)
	}
}

func TestRegressionRequestID(t *testing.T) {
	s := start(t)

	for _, path := range []string{"/health", "/tasks", "/tasks/task_9999"} {
		if id := s.header(http.MethodGet, path, "X-Request-Id"); id == "" {
			t.Errorf("%s carries no X-Request-Id", path)
		}
	}
}

func TestRegressionExportCSV(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	id := s.titled(list, "buy milk")

	status, payload := s.raw(http.MethodGet, "/export.csv", "")
	if status != http.StatusOK {
		t.Fatalf("export: %d", status)
	}
	rows, err := csv.NewReader(strings.NewReader(string(payload))).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected a header and one task, got %d rows", len(rows))
	}
	if !slices.Equal(rows[0], []string{"id", "list_id", "title", "done", "priority", "tags", "due", "created_at"}) {
		t.Fatalf("header is %v", rows[0])
	}
	if rows[1][0] != id || rows[1][2] != "buy milk" {
		t.Fatalf("row is %v", rows[1])
	}
}

package tests

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestT1RowCopy(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	first := s.titled(list, "one")
	second := s.titled(list, "two")

	status, body := s.call(http.MethodPost, "/tasks/bulk/complete", fmt.Sprintf(`{"ids":[%q,%q]}`, first, second))
	if status != http.StatusOK {
		t.Fatalf("bulk complete: %d %v", status, body)
	}
	for _, id := range []string{first, second} {
		_, task := s.call(http.MethodGet, "/tasks/"+id, "")
		if !boolean(t, task, "done") {
			t.Fatalf("%s was not persisted as done", id)
		}
	}
}

func TestT1SliceBounds(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	s.titled(list, "one")
	s.titled(list, "two")

	status, body := s.call(http.MethodGet, "/tasks?page=99", "")
	if status != http.StatusOK {
		t.Fatalf("page beyond the end: %d %v", status, body)
	}
	if rows := items(t, body); len(rows) != 0 {
		t.Fatalf("expected no items, got %d", len(rows))
	}
	if total := number(t, body, "total"); total != 2 {
		t.Fatalf("total on a page past the end is %d", total)
	}
}

func TestT1AtoiIgnored(t *testing.T) {
	s := start(t)
	s.list("Home")

	for _, raw := range []string{"abc", "0", "-1", "101"} {
		status, body := s.call(http.MethodGet, "/tasks?per_page="+raw, "")
		if status != http.StatusBadRequest {
			t.Errorf("per_page=%s: %d %v", raw, status, body)
		}
	}
}

func TestT1NilMap(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	s.task(list, `{"title":"one","tags":["home","errand"]}`)
	s.task(list, `{"title":"two","tags":["home"]}`)
	s.task(list, `{"title":"three","tags":["zzz","errand"]}`)

	status, body := s.call(http.MethodGet, "/stats", "")
	if status != http.StatusOK {
		t.Fatalf("stats: %d %v", status, body)
	}
	if counts := ranking(t, body, "by_tag"); !slices.Equal(counts, []string{"errand=2", "home=2", "zzz=1"}) {
		t.Fatalf("by_tag is %v, want descending count then ascending tag", counts)
	}
}

func TestT1ValueReceiver(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	id := s.titled(list, "one")

	status, task := s.call(http.MethodPatch, "/tasks/"+id, `{"title":"renamed"}`)
	if status != http.StatusOK {
		t.Fatalf("patch: %d %v", status, task)
	}
	if text(t, task, "updated_at") == text(t, task, "created_at") {
		t.Fatal("updated_at was not refreshed")
	}
}

func TestT1MakeAppend(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	s.titled(list, "one")
	s.titled(list, "two")

	_, body := s.call(http.MethodGet, "/tasks", "")
	rows := items(t, body)
	if len(rows) != 2 {
		t.Fatalf("expected two items, got %d", len(rows))
	}
	for _, row := range rows {
		if text(t, row, "id") == "" {
			t.Fatalf("empty task in the page: %v", rows)
		}
	}
}

func TestT2TypedNil(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	id := s.titled(list, "one")

	status, task := s.call(http.MethodPost, "/tasks/"+id+"/tags", `{"tag":"home"}`)
	if status != http.StatusOK {
		t.Fatalf("add tag: %d %v", status, task)
	}
	if tags := stringList(t, task, "tags"); len(tags) != 1 {
		t.Fatalf("expected one tag, got %v", tags)
	}
	status, task = s.call(http.MethodPost, "/tasks/"+id+"/tags", `{"tag":" Tag:Home "}`)
	if status != http.StatusOK {
		t.Fatalf("adding a tag the task already carries: %d %v", status, task)
	}
	if tags := stringList(t, task, "tags"); !slices.Equal(tags, []string{"home"}) {
		t.Fatalf("tags are %v after re-adding the same tag", tags)
	}
}

func TestT2ErrShadow(t *testing.T) {
	s := start(t)
	list := s.list("Home")

	status, body := s.call(http.MethodDelete, "/lists/list_9999", "")
	if status < 400 {
		t.Fatalf("deleting an unknown list reported success: %d %v", status, body)
	}
	if status, body := s.call(http.MethodDelete, "/lists/"+list, ""); status != http.StatusNoContent {
		t.Fatalf("delete list: %d %v", status, body)
	}
	if status, body := s.call(http.MethodDelete, "/lists/"+list, ""); status < 400 {
		t.Fatalf("deleting the same list twice reported success: %d %v", status, body)
	}
}

func TestT2ErrorsIs(t *testing.T) {
	s := start(t)
	s.list("Home")

	status, body := s.call(http.MethodGet, "/tasks/task_9999", "")
	if status != http.StatusNotFound {
		t.Fatalf("unknown task: %d %v", status, body)
	}
}

func TestT2SliceAlias(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	created := []string{s.titled(list, "zulu"), s.titled(list, "alpha"), s.titled(list, "mike")}

	_, page := s.call(http.MethodGet, "/tasks?sort=title", "")
	if titles := field(t, filled(t, page), "title"); !slices.Equal(titles, []string{"alpha", "mike", "zulu"}) {
		t.Fatalf("sorted page is %v", titles)
	}

	_, task := s.call(http.MethodGet, "/tasks/"+created[0], "")
	if title := text(t, task, "title"); title != "zulu" {
		t.Fatalf("%s resolves to %q", created[0], title)
	}
	if order := exported(t, s); !slices.Equal(order, created) {
		t.Fatalf("stored order changed to %v", order)
	}
}

func TestT2SortOrder(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	s.task(list, `{"title":"one","priority":"low"}`)
	s.task(list, `{"title":"two","priority":"high"}`)
	s.task(list, `{"title":"three","priority":"medium"}`)

	_, body := s.call(http.MethodGet, "/tasks?sort=priority", "")
	order := field(t, filled(t, body), "priority")
	if !slices.Equal(order, []string{"high", "medium", "low"}) {
		t.Fatalf("priority order is %v", order)
	}
}

func TestT2TrimCutset(t *testing.T) {
	s := start(t)
	list := s.list("Home")

	task := s.task(list, `{"title":"one","tags":["Groceries"," tag:Urgent "]}`)
	if tags := stringList(t, task, "tags"); !slices.Equal(tags, []string{"groceries", "urgent"}) {
		t.Fatalf("normalised tags are %v", tags)
	}
}

func TestT2TotalPages(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	s.titled(list, "one")
	s.titled(list, "two")
	s.titled(list, "three")

	_, body := s.call(http.MethodGet, "/tasks?per_page=2", "")
	if pages := number(t, body, "total_pages"); pages != 2 {
		t.Fatalf("total_pages is %d", pages)
	}
}

func TestT2WriteHeader(t *testing.T) {
	s := start(t)
	list := s.list("Home")

	status, body := s.call(http.MethodPost, "/lists/"+list+"/tasks", `{"title":"one"}`)
	if status != http.StatusCreated {
		t.Fatalf("create task: %d %v", status, body)
	}
}

func TestT2PatchZeroing(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	id := s.titled(list, "one")
	s.call(http.MethodPost, "/tasks/"+id+"/complete", "")

	status, task := s.call(http.MethodPatch, "/tasks/"+id, `{"title":"renamed"}`)
	if status != http.StatusOK {
		t.Fatalf("patch: %d %v", status, task)
	}
	if !boolean(t, task, "done") {
		t.Fatal("patching the title cleared done")
	}
}

func TestT3RaceRequestID(t *testing.T) {
	s := startRace(t)

	hammer(s, 20, 25, func(worker, i int) {
		s.hit(http.MethodGet, "/health", "")
	})
	if s.raceReport("requestID") {
		t.Fatal("data race on the request id counter")
	}
}

func TestT3RaceTasks(t *testing.T) {
	s := startRace(t)
	list := s.list("Home")

	hammer(s, 20, 25, func(worker, i int) {
		if worker%2 == 0 {
			s.hit(http.MethodGet, "/tasks", "")
			return
		}
		s.hit(http.MethodPost, "/lists/"+list+"/tasks", fmt.Sprintf(`{"title":"t%d-%d"}`, worker, i))
	})
	if s.raceReport("store.(*Store).Tasks") {
		t.Fatal("data race reading tasks without the lock")
	}
}

func TestT3MutexCopy(t *testing.T) {
	s := startRace(t)
	list := s.list("Home")
	ids := make([]string, 8)
	for i := range ids {
		ids[i] = fmt.Sprintf("%q", s.titled(list, fmt.Sprintf("task %d", i)))
	}
	body := fmt.Sprintf(`{"ids":[%s]}`, strings.Join(ids, ","))

	hammer(s, 20, 25, func(worker, i int) {
		if worker%2 == 0 {
			s.hit(http.MethodPost, "/lists/"+list+"/tasks", fmt.Sprintf(`{"title":"t%d-%d"}`, worker, i))
			return
		}
		s.hit(http.MethodPost, "/tasks/bulk/complete", body)
	})
	if s.raceReport("store.Store.CompleteAll") {
		t.Fatal("CompleteAll locks a copy of the store")
	}
}

func TestT3DeadlockDelete(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	s.titled(list, "one")

	s.hit(http.MethodDelete, "/tasks/task_9999", "")

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(s.url + "/lists")
	if err != nil {
		t.Fatalf("GET /lists after deleting an unknown task: %v", err)
	}
	resp.Body.Close()
}

func TestDecoyCompletionPercentFloor(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	id := s.titled(list, "one")
	s.titled(list, "two")
	s.titled(list, "three")
	s.call(http.MethodPost, "/tasks/"+id+"/complete", "")

	_, body := s.call(http.MethodGet, "/stats", "")
	if percent := number(t, body, "completion_percent"); percent != 33 {
		t.Fatalf("completion_percent is %d", percent)
	}
}

func TestDecoyListNameValidation(t *testing.T) {
	s := start(t)

	status, body := s.call(http.MethodPost, "/lists", `{"name":"  "}`)
	if status != http.StatusBadRequest {
		t.Fatalf("empty list name: %d %v", status, body)
	}
	status, body = s.call(http.MethodPost, "/lists", `{"name":"Home"}`)
	if status != http.StatusCreated {
		t.Fatalf("valid list name: %d %v", status, body)
	}
}

package tests

import (
	"fmt"
	"net/http"
	"slices"
	"testing"
)

func TestT1RangeCopy(t *testing.T) {
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
}

func TestT1AtoiIgnored(t *testing.T) {
	s := start(t)
	s.list("Home")

	status, body := s.call(http.MethodGet, "/tasks?per_page=abc", "")
	if status != http.StatusBadRequest {
		t.Fatalf("per_page=abc: %d %v", status, body)
	}
}

func TestT1NilMap(t *testing.T) {
	s := start(t)
	list := s.list("Home")
	s.task(list, `{"title":"one","tags":["home","errand"]}`)

	status, body := s.call(http.MethodGet, "/stats", "")
	if status != http.StatusOK {
		t.Fatalf("stats: %d %v", status, body)
	}
	counts, ok := body["by_tag"].([]any)
	if !ok || len(counts) != 2 {
		t.Fatalf("expected two tag counts, got %v", body["by_tag"])
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
}

func TestT2ErrShadow(t *testing.T) {
	s := start(t)
	s.list("Home")

	status, body := s.call(http.MethodDelete, "/lists/list_9999", "")
	if status < 400 {
		t.Fatalf("deleting an unknown list reported success: %d %v", status, body)
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

	s.call(http.MethodGet, "/tasks?sort=title", "")

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

package tests

import (
	"encoding/csv"
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
	} {
		status, response := s.call(http.MethodPost, "/lists/"+list+"/tasks", body)
		if status != http.StatusBadRequest {
			t.Errorf("%s: %d %v", body, status, response)
		}
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
	if status, _ := s.call(http.MethodDelete, "/tasks/"+id+"/tags/home", ""); status != http.StatusNotFound {
		t.Fatalf("removing a missing tag: %d", status)
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

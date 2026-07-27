package tests

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var binary string

func TestMain(m *testing.M) {
	dir := os.Getenv("SOLVE_THIS_DIR")
	if dir == "" {
		fmt.Fprintln(os.Stderr, "SOLVE_THIS_DIR is not set")
		os.Exit(1)
	}
	work, err := os.MkdirTemp("", "solve-this-bin")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	binary = filepath.Join(work, "api")
	build := exec.Command("go", "build", "-o", binary, "./cmd/api")
	build.Dir = dir
	build.Stdout = os.Stderr
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n", err)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(work)
	os.Exit(code)
}

type server struct {
	t   *testing.T
	url string
}

func start(t *testing.T) *server {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	cmd := exec.Command(binary, "-addr", fmt.Sprintf("127.0.0.1:%d", port))
	var logs bytes.Buffer
	cmd.Stdout = &logs
	cmd.Stderr = &logs
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cmd.Process.Kill()
		cmd.Wait()
		if t.Failed() {
			t.Logf("server log:\n%s", logs.String())
		}
	})

	s := &server{t: t, url: fmt.Sprintf("http://127.0.0.1:%d", port)}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return s
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("server did not start")
	return nil
}

func (s *server) raw(method, path, body string) (int, []byte) {
	s.t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, s.url+path, reader)
	if err != nil {
		s.t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.t.Fatal(err)
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		s.t.Fatal(err)
	}
	return resp.StatusCode, payload
}

func (s *server) call(method, path, body string) (int, map[string]any) {
	s.t.Helper()
	status, payload := s.raw(method, path, body)
	decoded := map[string]any{}
	if len(bytes.TrimSpace(payload)) > 0 {
		if err := json.Unmarshal(payload, &decoded); err != nil {
			s.t.Fatalf("%s %s: %v: %s", method, path, err, payload)
		}
	}
	return status, decoded
}

func (s *server) list(name string) string {
	s.t.Helper()
	status, body := s.call(http.MethodPost, "/lists", fmt.Sprintf(`{"name":%q}`, name))
	if status < 200 || status > 299 {
		s.t.Fatalf("create list: %d %v", status, body)
	}
	return text(s.t, body, "id")
}

func (s *server) task(list, body string) map[string]any {
	s.t.Helper()
	status, task := s.call(http.MethodPost, "/lists/"+list+"/tasks", body)
	if status < 200 || status > 299 {
		s.t.Fatalf("create task: %d %v", status, task)
	}
	return task
}

func (s *server) titled(list, title string) string {
	s.t.Helper()
	return text(s.t, s.task(list, fmt.Sprintf(`{"title":%q}`, title)), "id")
}

func text(t *testing.T, body map[string]any, key string) string {
	t.Helper()
	value, ok := body[key].(string)
	if !ok {
		t.Fatalf("%q missing or not a string in %v", key, body)
	}
	return value
}

func number(t *testing.T, body map[string]any, key string) int {
	t.Helper()
	value, ok := body[key].(float64)
	if !ok {
		t.Fatalf("%q missing or not a number in %v", key, body)
	}
	return int(value)
}

func boolean(t *testing.T, body map[string]any, key string) bool {
	t.Helper()
	value, ok := body[key].(bool)
	if !ok {
		t.Fatalf("%q missing or not a bool in %v", key, body)
	}
	return value
}

func items(t *testing.T, body map[string]any) []map[string]any {
	t.Helper()
	raw, ok := body["items"].([]any)
	if !ok {
		t.Fatalf("items missing or not an array in %v", body)
	}
	out := make([]map[string]any, 0, len(raw))
	for _, entry := range raw {
		row, ok := entry.(map[string]any)
		if !ok {
			t.Fatalf("item is not an object: %v", entry)
		}
		out = append(out, row)
	}
	return out
}

func stringList(t *testing.T, body map[string]any, key string) []string {
	t.Helper()
	raw, ok := body[key].([]any)
	if !ok {
		t.Fatalf("%q missing or not an array in %v", key, body)
	}
	out := make([]string, 0, len(raw))
	for _, entry := range raw {
		value, ok := entry.(string)
		if !ok {
			t.Fatalf("%q holds a non string: %v", key, entry)
		}
		out = append(out, value)
	}
	return out
}

func exported(t *testing.T, s *server) []string {
	t.Helper()
	status, payload := s.raw(http.MethodGet, "/export.csv", "")
	if status != http.StatusOK {
		t.Fatalf("export: %d", status)
	}
	rows, err := csv.NewReader(bytes.NewReader(payload)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, row := range rows[1:] {
		ids = append(ids, row[0])
	}
	return ids
}

func filled(t *testing.T, body map[string]any) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, row := range items(t, body) {
		if text(t, row, "id") != "" {
			out = append(out, row)
		}
	}
	return out
}

func field(t *testing.T, rows []map[string]any, key string) []string {
	t.Helper()
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, text(t, row, key))
	}
	return out
}

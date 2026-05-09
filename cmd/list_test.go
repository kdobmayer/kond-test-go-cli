package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kdobmayer/kond-test-go-cli/internal"
)

func captureStdout(f func()) string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func newTestStore(t *testing.T) *internal.Store {
	t.Helper()
	s, err := internal.NewStoreWithPath(filepath.Join(t.TempDir(), "tasks.json"))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestPrintTableNoTasks(t *testing.T) {
	store = newTestStore(t)

	out := captureStdout(func() {
		printTable(store.List())
	})
	if out != "" {
		t.Errorf("expected empty output, got %q", out)
	}
}

func TestPrintTableWithTasks(t *testing.T) {
	store = newTestStore(t)
	store.Add("pending task")
	task2 := store.Add("done task")
	store.MarkDone(task2.ID)

	out := captureStdout(func() {
		printTable(store.List())
	})

	if !strings.Contains(out, "[ ] pending task") {
		t.Errorf("expected pending task line, got %q", out)
	}
	if !strings.Contains(out, "[x] done task") {
		t.Errorf("expected done task line, got %q", out)
	}
}

func TestPrintCSV(t *testing.T) {
	store = newTestStore(t)
	store.Add("write tests")
	store.MarkDone(1)

	out := captureStdout(func() {
		if err := printCSV(store.List()); err != nil {
			t.Errorf("printCSV returned error: %v", err)
		}
	})

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (header + 1 row), got %d: %q", len(lines), out)
	}
	if lines[0] != "id,done,title" {
		t.Errorf("expected header 'id,done,title', got %q", lines[0])
	}
	if lines[1] != "1,true,write tests" {
		t.Errorf("expected data row '1,true,write tests', got %q", lines[1])
	}
}

func TestListInvalidFormat(t *testing.T) {
	store = newTestStore(t)
	store.Add("a task")

	listFormat = "json"
	err := listCmd.RunE(listCmd, nil)
	listFormat = "table"

	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "must be table or csv") {
		t.Errorf("expected 'must be table or csv' in error, got %q", err.Error())
	}
}

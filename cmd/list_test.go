package cmd

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/kdobmayer/kond-test-go-cli/internal"
)

func setupListStore(t *testing.T) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tasks.json")
	s, err := internal.NewStoreWithPath(path)
	if err != nil {
		t.Fatal(err)
	}
	store = s
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestListJSONEmpty(t *testing.T) {
	setupListStore(t)
	listJSON = true
	defer func() { listJSON = false }()

	out := captureStdout(t, func() {
		if err := listCmd.RunE(listCmd, nil); err != nil {
			t.Fatal(err)
		}
	})

	var tasks []internal.Task
	if err := json.Unmarshal([]byte(out), &tasks); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(tasks) != 0 {
		t.Errorf("expected empty array, got %d tasks", len(tasks))
	}
}

func TestListJSONWithTasks(t *testing.T) {
	setupListStore(t)
	store.Add("Buy groceries")
	store.Add("Write tests")
	listJSON = true
	defer func() { listJSON = false }()

	out := captureStdout(t, func() {
		if err := listCmd.RunE(listCmd, nil); err != nil {
			t.Fatal(err)
		}
	})

	var tasks []internal.Task
	if err := json.Unmarshal([]byte(out), &tasks); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0].Title != "Buy groceries" {
		t.Errorf("expected 'Buy groceries', got %q", tasks[0].Title)
	}
	if tasks[1].Title != "Write tests" {
		t.Errorf("expected 'Write tests', got %q", tasks[1].Title)
	}
}

func TestListJSONDoneField(t *testing.T) {
	setupListStore(t)
	store.Add("Task one")
	store.MarkDone(1)
	listJSON = true
	defer func() { listJSON = false }()

	out := captureStdout(t, func() {
		if err := listCmd.RunE(listCmd, nil); err != nil {
			t.Fatal(err)
		}
	})

	var tasks []internal.Task
	if err := json.Unmarshal([]byte(out), &tasks); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if !tasks[0].Done {
		t.Error("expected done=true in JSON output")
	}
}

package cmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/kdobmayer/kond-test-go-cli/internal"
)

func TestListJSONOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	s, err := internal.NewStoreWithPath(path)
	if err != nil {
		t.Fatal(err)
	}
	s.Add("First task")
	s.Add("Second task")
	if err := s.MarkDone(1); err != nil {
		t.Fatal(err)
	}

	origStore := store
	t.Cleanup(func() { store = origStore })
	store = s

	origJSON := jsonOutput
	t.Cleanup(func() { jsonOutput = origJSON })
	jsonOutput = true

	buf := &bytes.Buffer{}
	listCmd.SetOut(buf)
	t.Cleanup(func() { listCmd.SetOut(nil) })

	if err := listCmd.RunE(listCmd, nil); err != nil {
		t.Fatal(err)
	}

	var tasks []internal.Task
	if err := json.Unmarshal(buf.Bytes(), &tasks); err != nil {
		t.Fatalf("invalid JSON output: %v\noutput: %s", err, buf)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0].Title != "First task" {
		t.Errorf("expected 'First task', got %q", tasks[0].Title)
	}
	if !tasks[0].Done {
		t.Error("first task should be done")
	}
	if tasks[1].Done {
		t.Error("second task should not be done")
	}
}

func TestListJSONEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	s, err := internal.NewStoreWithPath(path)
	if err != nil {
		t.Fatal(err)
	}

	origStore := store
	t.Cleanup(func() { store = origStore })
	store = s

	origJSON := jsonOutput
	t.Cleanup(func() { jsonOutput = origJSON })
	jsonOutput = true

	buf := &bytes.Buffer{}
	listCmd.SetOut(buf)
	t.Cleanup(func() { listCmd.SetOut(nil) })

	if err := listCmd.RunE(listCmd, nil); err != nil {
		t.Fatal(err)
	}

	var tasks []internal.Task
	if err := json.Unmarshal(buf.Bytes(), &tasks); err != nil {
		t.Fatalf("invalid JSON output: %v\noutput: %s", err, buf)
	}
	if len(tasks) != 0 {
		t.Errorf("expected empty array, got %d tasks", len(tasks))
	}
}

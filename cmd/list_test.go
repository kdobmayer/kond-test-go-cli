package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/kdobmayer/kond-test-go-cli/internal"
)

func captureStdout(fn func() error) (string, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	old := os.Stdout
	os.Stdout = w
	runErr := fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String(), runErr
}

func TestListJSONFlag(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	s, err := internal.NewStoreWithPath(path)
	if err != nil {
		t.Fatal(err)
	}
	s.Add("Task A")
	s.Add("Task B")
	store = s
	jsonOutput = true
	defer func() { jsonOutput = false }()

	output, err := captureStdout(func() error {
		return listCmd.RunE(listCmd, nil)
	})
	if err != nil {
		t.Fatal(err)
	}

	var tasks []internal.Task
	if err := json.Unmarshal([]byte(output), &tasks); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, output)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0].Title != "Task A" || tasks[1].Title != "Task B" {
		t.Errorf("unexpected task titles: %v", tasks)
	}
}

func TestListJSONFlagEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	s, err := internal.NewStoreWithPath(path)
	if err != nil {
		t.Fatal(err)
	}
	store = s
	jsonOutput = true
	defer func() { jsonOutput = false }()

	output, err := captureStdout(func() error {
		return listCmd.RunE(listCmd, nil)
	})
	if err != nil {
		t.Fatal(err)
	}

	var tasks []internal.Task
	if err := json.Unmarshal([]byte(output), &tasks); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, output)
	}
	if len(tasks) != 0 {
		t.Errorf("expected empty JSON array, got %d tasks", len(tasks))
	}
}

func TestListJSONFlagDoneField(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	s, err := internal.NewStoreWithPath(path)
	if err != nil {
		t.Fatal(err)
	}
	s.Add("Task A")
	if err := s.MarkDone(1); err != nil {
		t.Fatal(err)
	}
	store = s
	jsonOutput = true
	defer func() { jsonOutput = false }()

	output, err := captureStdout(func() error {
		return listCmd.RunE(listCmd, nil)
	})
	if err != nil {
		t.Fatal(err)
	}

	var tasks []internal.Task
	if err := json.Unmarshal([]byte(output), &tasks); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, output)
	}
	if !tasks[0].Done {
		t.Error("expected done=true in JSON output")
	}
}

package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAdd(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	s, err := NewStoreWithPath(path)
	if err != nil {
		t.Fatal(err)
	}

	task := s.Add("Buy groceries")
	if task.ID != 1 {
		t.Errorf("expected ID 1, got %d", task.ID)
	}
	if task.Title != "Buy groceries" {
		t.Errorf("expected title 'Buy groceries', got %q", task.Title)
	}
	if task.Done {
		t.Error("new task should not be done")
	}
}

func TestAddIncrementsID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	s, err := NewStoreWithPath(path)
	if err != nil {
		t.Fatal(err)
	}

	s.Add("First")
	second := s.Add("Second")
	if second.ID != 2 {
		t.Errorf("expected ID 2, got %d", second.ID)
	}
}

func TestMarkDone(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	s, err := NewStoreWithPath(path)
	if err != nil {
		t.Fatal(err)
	}

	s.Add("Task")
	if err := s.MarkDone(1); err != nil {
		t.Fatal(err)
	}
	tasks := s.List()
	if !tasks[0].Done {
		t.Error("task should be done")
	}
}

func TestMarkDoneNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	s, err := NewStoreWithPath(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.MarkDone(99); err == nil {
		t.Error("expected error for non-existent task")
	}
}

func TestDelete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	s, err := NewStoreWithPath(path)
	if err != nil {
		t.Fatal(err)
	}

	s.Add("Task")
	if err := s.Delete(1); err != nil {
		t.Fatal(err)
	}
	if len(s.List()) != 0 {
		t.Error("expected empty list after delete")
	}
}

func TestDeleteNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	s, err := NewStoreWithPath(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Delete(99); err == nil {
		t.Error("expected error for non-existent task")
	}
}

func TestSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	s, err := NewStoreWithPath(path)
	if err != nil {
		t.Fatal(err)
	}

	s.Add("Persist me")
	s.MarkDone(1)
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	s2, err := NewStoreWithPath(path)
	if err != nil {
		t.Fatal(err)
	}
	tasks := s2.List()
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Title != "Persist me" {
		t.Errorf("expected 'Persist me', got %q", tasks[0].Title)
	}
	if !tasks[0].Done {
		t.Error("task should be done after reload")
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")
	s, err := NewStoreWithPath(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.List()) != 0 {
		t.Error("expected empty list for non-existent file")
	}
}

func TestLoadCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	os.WriteFile(path, []byte("not json"), 0644)
	_, err := NewStoreWithPath(path)
	if err == nil {
		t.Error("expected error for corrupt file")
	}
}

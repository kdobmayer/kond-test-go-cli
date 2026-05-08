// Package internal implements the task storage layer.
package internal

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// Task represents a single TODO item.
type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// Store manages task persistence to a JSON file.
type Store struct {
	path  string
	tasks []Task
}

// NewStore creates a Store backed by ~/.tasks.json.
func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("determining home directory: %w", err)
	}
	path := filepath.Join(home, ".tasks.json")
	s := &Store{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

// NewStoreWithPath creates a Store backed by the given file path.
// Useful for testing.
func NewStoreWithPath(path string) (*Store, error) {
	s := &Store{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		slog.Debug("no existing tasks file", "path", s.path)
		s.tasks = []Task{}
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading tasks file: %w", err)
	}
	if err := json.Unmarshal(data, &s.tasks); err != nil {
		return fmt.Errorf("parsing tasks file: %w", err)
	}
	slog.Debug("loaded tasks", "count", len(s.tasks), "path", s.path)
	return nil
}

// Save persists the current task list to disk.
func (s *Store) Save() error {
	data, err := json.MarshalIndent(s.tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling tasks: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return fmt.Errorf("writing tasks file: %w", err)
	}
	return nil
}

// Add creates a new task and returns it.
func (s *Store) Add(title string) Task {
	id := 1
	for _, t := range s.tasks {
		if t.ID >= id {
			id = t.ID + 1
		}
	}
	task := Task{ID: id, Title: title, Done: false}
	s.tasks = append(s.tasks, task)
	return task
}

// List returns all tasks.
func (s *Store) List() []Task {
	return s.tasks
}

// MarkDone marks the task with the given ID as done.
func (s *Store) MarkDone(id int) error {
	for i := range s.tasks {
		if s.tasks[i].ID == id {
			s.tasks[i].Done = true
			return nil
		}
	}
	return fmt.Errorf("task %d not found", id)
}

// Delete removes the task with the given ID.
func (s *Store) Delete(id int) error {
	for i, t := range s.tasks {
		if t.ID == id {
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("task %d not found", id)
}

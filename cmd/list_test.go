package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/kdobmayer/kond-test-go-cli/internal"
)

func TestListJSONSuccess(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	s, err := internal.NewStoreWithPath(path)
	if err != nil {
		t.Fatal(err)
	}
	s.Add("Buy milk")
	s.Add("Walk dog")

	var buf bytes.Buffer
	if err := writeListJSON(&buf, s.List(), nil); err != nil {
		t.Fatal(err)
	}

	var resp jsonListResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("expected null error, got %q", *resp.Error)
	}
	if len(resp.Result) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(resp.Result))
	}
	if resp.Result[0].Title != "Buy milk" {
		t.Errorf("expected first task 'Buy milk', got %q", resp.Result[0].Title)
	}
}

func TestListJSONError(t *testing.T) {
	var buf bytes.Buffer
	if err := writeListJSON(&buf, nil, errors.New("store unavailable")); err != nil {
		t.Fatal(err)
	}

	var resp jsonListResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.Result != nil {
		t.Errorf("expected null result, got %v", resp.Result)
	}
	if resp.Error == nil {
		t.Fatal("expected non-null error field")
	}
	if *resp.Error != "store unavailable" {
		t.Errorf("expected 'store unavailable', got %q", *resp.Error)
	}
}

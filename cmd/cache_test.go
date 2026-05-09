package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/kdobmayer/kond-test-go-cli/internal"
)

// captureStdout runs fn and returns everything written to os.Stdout.
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = old
	return buf.String()
}

// setupCacheTest wires a temp store + FakeCache into the package vars and returns
// a cleanup func that restores the originals.
func setupCacheTest(t *testing.T, initialTasks []internal.Task) (*internal.FakeCache, func()) {
	t.Helper()

	f, err := os.CreateTemp("", "tasks-test-*.json")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Close()

	if len(initialTasks) > 0 {
		data, _ := json.Marshal(initialTasks)
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
	} else {
		os.Remove(path)
	}

	s, err := internal.NewStoreWithPath(path)
	if err != nil {
		t.Fatal(err)
	}

	fc := internal.NewFakeCache()

	origStore, origCache, origUseCache := store, cache, useCache
	store = s
	cache = fc
	useCache = true

	return fc, func() {
		os.Remove(path)
		store, cache, useCache = origStore, origCache, origUseCache
	}
}

func TestListCacheMiss(t *testing.T) {
	tasks := []internal.Task{{ID: 1, Title: "buy milk", Done: false}}
	fc, cleanup := setupCacheTest(t, tasks)
	defer cleanup()

	out := captureStdout(func() {
		if err := listCmd.RunE(listCmd, nil); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "buy milk") {
		t.Errorf("expected output to contain 'buy milk', got: %q", out)
	}

	// Cache should now be populated after the miss.
	data, err := fc.Get(context.Background(), cacheKey)
	if err != nil {
		t.Fatalf("expected cache to be populated after miss: %v", err)
	}
	var cached []internal.Task
	if err := json.Unmarshal(data, &cached); err != nil {
		t.Fatal(err)
	}
	if len(cached) != 1 || cached[0].Title != "buy milk" {
		t.Errorf("unexpected cached tasks: %+v", cached)
	}
}

func TestListCacheHit(t *testing.T) {
	diskTasks := []internal.Task{{ID: 1, Title: "disk task", Done: false}}
	fc, cleanup := setupCacheTest(t, diskTasks)
	defer cleanup()

	// Pre-populate cache with different data than what's on disk.
	cachedTasks := []internal.Task{{ID: 99, Title: "cached task", Done: false}}
	data, _ := json.Marshal(cachedTasks)
	fc.Set(context.Background(), cacheKey, data, cacheTTL)

	out := captureStdout(func() {
		if err := listCmd.RunE(listCmd, nil); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "cached task") {
		t.Errorf("expected output to contain 'cached task', got: %q", out)
	}
	if strings.Contains(out, "disk task") {
		t.Errorf("expected output NOT to contain 'disk task', got: %q", out)
	}
}

func TestInvalidationOnAdd(t *testing.T) {
	tasks := []internal.Task{{ID: 1, Title: "existing", Done: false}}
	fc, cleanup := setupCacheTest(t, tasks)
	defer cleanup()
	useCache = false // add does not use --cache

	data, _ := json.Marshal(tasks)
	fc.Set(context.Background(), cacheKey, data, cacheTTL)

	captureStdout(func() {
		if err := addCmd.RunE(addCmd, []string{"new task"}); err != nil {
			t.Fatal(err)
		}
	})

	_, err := fc.Get(context.Background(), cacheKey)
	if err != internal.ErrCacheMiss {
		t.Errorf("expected cache miss after add, got: %v", err)
	}
}

func TestInvalidationOnDone(t *testing.T) {
	tasks := []internal.Task{{ID: 1, Title: "existing", Done: false}}
	fc, cleanup := setupCacheTest(t, tasks)
	defer cleanup()
	useCache = false

	data, _ := json.Marshal(tasks)
	fc.Set(context.Background(), cacheKey, data, cacheTTL)

	captureStdout(func() {
		if err := doneCmd.RunE(doneCmd, []string{"1"}); err != nil {
			t.Fatal(err)
		}
	})

	_, err := fc.Get(context.Background(), cacheKey)
	if err != internal.ErrCacheMiss {
		t.Errorf("expected cache miss after done, got: %v", err)
	}
}

func TestInvalidationOnDelete(t *testing.T) {
	tasks := []internal.Task{{ID: 1, Title: "existing", Done: false}}
	fc, cleanup := setupCacheTest(t, tasks)
	defer cleanup()
	useCache = false

	data, _ := json.Marshal(tasks)
	fc.Set(context.Background(), cacheKey, data, cacheTTL)

	captureStdout(func() {
		if err := deleteCmd.RunE(deleteCmd, []string{"1"}); err != nil {
			t.Fatal(err)
		}
	})

	_, err := fc.Get(context.Background(), cacheKey)
	if err != internal.ErrCacheMiss {
		t.Errorf("expected cache miss after delete, got: %v", err)
	}
}

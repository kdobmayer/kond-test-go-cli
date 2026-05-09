package internal

import (
	"context"
	"sync"
	"time"
)

// FakeCache is an in-memory Cache implementation for testing.
type FakeCache struct {
	mu      sync.Mutex
	entries map[string]fakeEntry
}

type fakeEntry struct {
	value  []byte
	expiry time.Time
}

// NewFakeCache creates an empty FakeCache.
func NewFakeCache() *FakeCache {
	return &FakeCache{entries: make(map[string]fakeEntry)}
}

func (f *FakeCache) Get(_ context.Context, key string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	e, ok := f.entries[key]
	if !ok || time.Now().After(e.expiry) {
		return nil, ErrCacheMiss
	}
	return e.value, nil
}

func (f *FakeCache) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.entries[key] = fakeEntry{value: value, expiry: time.Now().Add(ttl)}
	return nil
}

func (f *FakeCache) Delete(_ context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.entries, key)
	return nil
}

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "info")
	}
	if cfg.Output != "table" {
		t.Errorf("Output = %q, want %q", cfg.Output, "table")
	}
	if cfg.PipelineDir == "" {
		t.Error("PipelineDir should not be empty")
	}
	if cfg.RunDir == "" {
		t.Error("RunDir should not be empty")
	}
}

func TestLoadFrom_NonExistent(t *testing.T) {
	cfg, err := LoadFrom("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}
	// Should return defaults
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "info")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := DefaultConfig()
	cfg.LogLevel = "debug"
	cfg.Defaults["timeout"] = "60s"

	if err := cfg.SaveTo(path); err != nil {
		t.Fatalf("SaveTo() error = %v", err)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}

	if loaded.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", loaded.LogLevel, "debug")
	}
	if loaded.Defaults["timeout"] != "60s" {
		t.Errorf("Defaults[timeout] = %q, want %q", loaded.Defaults["timeout"], "60s")
	}
}

func TestGet(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Defaults["foo"] = "bar"
	tests := []struct{ key, want string }{
		{"log_level", "info"},
		{"output", "table"},
		{"defaults.foo", "bar"},
	}
	for _, tt := range tests {
		val, err := cfg.Get(tt.key)
		if err != nil {
			t.Errorf("Get(%q) error = %v", tt.key, err)
			continue
		}
		if val != tt.want {
			t.Errorf("Get(%q) = %q, want %q", tt.key, val, tt.want)
		}
	}
}

func TestGet_Unknown(t *testing.T) {
	cfg := DefaultConfig()
	_, err := cfg.Get("nonexistent")
	if err == nil {
		t.Error("expected error for unknown key")
	}
}

func TestGet_DefaultsNotFound(t *testing.T) {
	cfg := DefaultConfig()
	_, err := cfg.Get("defaults.nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent defaults key")
	}
}

func TestSet(t *testing.T) {
	cfg := DefaultConfig()
	tests := []struct{ key, value string }{
		{"pipeline_dir", "/tmp/pipelines"},
		{"run_dir", "/tmp/runs"},
		{"log_level", "debug"},
		{"output", "json"},
		{"defaults.timeout", "30s"},
	}
	for _, tt := range tests {
		if err := cfg.Set(tt.key, tt.value); err != nil {
			t.Errorf("Set(%q, %q) error = %v", tt.key, tt.value, err)
			continue
		}
		val, _ := cfg.Get(tt.key)
		if val != tt.value {
			t.Errorf("after Set, Get(%q) = %q, want %q", tt.key, val, tt.value)
		}
	}
}

func TestSet_InvalidLogLevel(t *testing.T) {
	cfg := DefaultConfig()
	err := cfg.Set("log_level", "invalid")
	if err == nil {
		t.Error("expected error for invalid log level")
	}
}

func TestSet_InvalidOutput(t *testing.T) {
	cfg := DefaultConfig()
	err := cfg.Set("output", "invalid")
	if err == nil {
		t.Error("expected error for invalid output format")
	}
}

func TestSet_UnknownKey(t *testing.T) {
	cfg := DefaultConfig()
	err := cfg.Set("nonexistent", "value")
	if err == nil {
		t.Error("expected error for unknown key")
	}
}

func TestListKeys(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Defaults["a"] = "1"
	cfg.Defaults["b"] = "2"

	keys := cfg.ListKeys()
	if len(keys) < 4 {
		t.Errorf("expected at least 4 keys, got %d", len(keys))
	}

	// Should contain base keys
	found := map[string]bool{}
	for _, k := range keys {
		found[k] = true
	}
	for _, expected := range []string{"pipeline_dir", "run_dir", "log_level", "output"} {
		if !found[expected] {
			t.Errorf("missing key %q", expected)
		}
	}
}

func TestLoadFrom_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	os.WriteFile(path, []byte("not: [valid: yaml: {{"), 0644)

	_, err := LoadFrom(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestConfigPath(t *testing.T) {
	path := ConfigPath()
	if path == "" {
		t.Error("ConfigPath() should not be empty")
	}
}

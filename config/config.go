package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	PipelineDir string            `yaml:"pipeline_dir"`
	RunDir      string            `yaml:"run_dir"`
	LogLevel    string            `yaml:"log_level"`
	Output      string            `yaml:"output"`
	Defaults    map[string]string `yaml:"defaults,omitempty"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		PipelineDir: filepath.Join(home, ".pipeline", "pipelines"),
		RunDir:      filepath.Join(home, ".pipeline", "runs"),
		LogLevel:    "info",
		Output:      "table",
		Defaults:    make(map[string]string),
	}
}

// ConfigPath returns the default config file path
func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".pipeline", "config.yaml")
}

// Load reads the config from disk, or returns defaults if not found
func Load() (*Config, error) {
	return LoadFrom(ConfigPath())
}

// LoadFrom reads config from a specific path
func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}

// Save writes the config to disk
func (c *Config) Save() error {
	return c.SaveTo(ConfigPath())
}

// SaveTo writes the config to a specific path
func (c *Config) SaveTo(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// Get retrieves a config value by key (dot-notation for nested)
func (c *Config) Get(key string) (string, error) {
	switch key {
	case "pipeline_dir":
		return c.PipelineDir, nil
	case "run_dir":
		return c.RunDir, nil
	case "log_level":
		return c.LogLevel, nil
	case "output":
		return c.Output, nil
	default:
		if strings.HasPrefix(key, "defaults.") {
			subKey := strings.TrimPrefix(key, "defaults.")
			if val, ok := c.Defaults[subKey]; ok {
				return val, nil
			}
			return "", fmt.Errorf("key not found: %s", key)
		}
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

// Set updates a config value by key
func (c *Config) Set(key, value string) error {
	switch key {
	case "pipeline_dir":
		c.PipelineDir = value
	case "run_dir":
		c.RunDir = value
	case "log_level":
		if !isValidLogLevel(value) {
			return fmt.Errorf("invalid log level: %s (valid: debug, info, warn, error)", value)
		}
		c.LogLevel = value
	case "output":
		if !isValidOutput(value) {
			return fmt.Errorf("invalid output format: %s (valid: table, json, yaml)", value)
		}
		c.Output = value
	default:
		if strings.HasPrefix(key, "defaults.") {
			subKey := strings.TrimPrefix(key, "defaults.")
			if c.Defaults == nil {
				c.Defaults = make(map[string]string)
			}
			c.Defaults[subKey] = value
			return nil
		}
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

// ListKeys returns all available config keys
func (c *Config) ListKeys() []string {
	keys := []string{"pipeline_dir", "run_dir", "log_level", "output"}
	for k := range c.Defaults {
		keys = append(keys, "defaults."+k)
	}
	return keys
}

func isValidLogLevel(level string) bool {
	switch level {
	case "debug", "info", "warn", "error":
		return true
	}
	return false
}

func isValidOutput(output string) bool {
	switch output {
	case "table", "json", "yaml":
		return true
	}
	return false
}

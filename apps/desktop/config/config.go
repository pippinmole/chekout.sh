package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const maxRecentOpens = 5

// RepoEntry holds per-repo configuration.
type RepoEntry struct {
	Path string `json:"path"`
	IDE  string `json:"ide,omitempty"`
}

// Config is the top-level config structure persisted to disk.
type Config struct {
	mu sync.RWMutex `json:"-"`

	DefaultIDE  string               `json:"default_ide"`
	Registered  bool                 `json:"registered"`
	Repos       map[string]RepoEntry `json:"repos"`
	RecentOpens []string             `json:"recent_opens"`
}

// configPath returns the path to the config file.
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "chekout", "config.json"), nil
}

// Load reads the config from disk, creating defaults if the file does not exist.
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Return a sensible default config.
		return &Config{
			DefaultIDE:  "vscode",
			Repos:       make(map[string]RepoEntry),
			RecentOpens: []string{},
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.Repos == nil {
		cfg.Repos = make(map[string]RepoEntry)
	}
	if cfg.DefaultIDE == "" {
		cfg.DefaultIDE = "vscode"
	}
	return &cfg, nil
}

// Save writes the config to disk atomically.
func Save(cfg *Config) error {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	// Write atomically via a temp file.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return os.Rename(tmp, path)
}

// Path returns the absolute path to the config file (even if it doesn't exist yet).
func Path() string {
	p, _ := configPath()
	return p
}

// AddRecentOpen prepends repoKey to RecentOpens, deduplicating and capping at maxRecentOpens.
func (cfg *Config) AddRecentOpen(repoKey string) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	// Remove existing occurrence.
	filtered := cfg.RecentOpens[:0]
	for _, r := range cfg.RecentOpens {
		if r != repoKey {
			filtered = append(filtered, r)
		}
	}

	// Prepend and cap.
	cfg.RecentOpens = append([]string{repoKey}, filtered...)
	if len(cfg.RecentOpens) > maxRecentOpens {
		cfg.RecentOpens = cfg.RecentOpens[:maxRecentOpens]
	}
}

// GetRepo returns the RepoEntry for the given key, with a boolean indicating presence.
func (cfg *Config) GetRepo(key string) (RepoEntry, bool) {
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()
	e, ok := cfg.Repos[key]
	return e, ok
}

// SetRepo sets the RepoEntry for the given key.
func (cfg *Config) SetRepo(key string, entry RepoEntry) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	cfg.Repos[key] = entry
}

package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	ScanRoots    []string `toml:"scan_roots"`
	AllowedRepos []string `toml:"allowed_repos,omitempty"` // Whitelist (if non-empty, only these are exposed)
	HiddenRepos  []string `toml:"hidden_repos,omitempty"`  // Blacklist (these are always hidden)
	IgnoredDirs  []string `toml:"ignored_dirs"`
	Port         string   `toml:"port"`
	DefaultRepo  string   `toml:"default_repo,omitempty"`
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		ScanRoots: []string{
			filepath.Join(home, "projects"),
			filepath.Join(home, ".config", "nix-config"),
		},
		AllowedRepos: []string{},
		HiddenRepos:  []string{},
		IgnoredDirs: []string{
			".git",
			"node_modules",
			"target",
			"vendor",
			".cache",
			"tmp",
			"dist",
			"build",
			".idea",
			".tokensave",
		},
		Port: "8425",
	}
}

func LoadConfig(customPath string) (*Config, error) {
	home, _ := os.UserHomeDir()
	configFile := customPath

	if configFile == "" {
		configDir := filepath.Join(home, ".config", "beads-everywhere")
		configFile = filepath.Join(configDir, "config.toml")

		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			// Fallback check for legacy beads-fleet config
			legacyFile := filepath.Join(home, ".config", "beads-fleet", "config.toml")
			if _, lErr := os.Stat(legacyFile); lErr == nil {
				configFile = legacyFile
			} else {
				cfg := DefaultConfig()
				_ = os.MkdirAll(configDir, 0755)
				data, _ := toml.Marshal(cfg)
				_ = os.WriteFile(configFile, data, 0644)
				return cfg, nil
			}
		}
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return DefaultConfig(), nil
	}

	cfg := DefaultConfig()
	if err := toml.Unmarshal(data, cfg); err != nil {
		return DefaultConfig(), nil
	}

	// Expand tildes in scan roots
	for i, root := range cfg.ScanRoots {
		if strings.HasPrefix(root, "~/") {
			cfg.ScanRoots[i] = filepath.Join(home, root[2:])
		}
	}

	// Expand tildes in allowed / hidden repos
	for i, r := range cfg.AllowedRepos {
		if strings.HasPrefix(r, "~/") {
			cfg.AllowedRepos[i] = filepath.Join(home, r[2:])
		}
	}
	for i, r := range cfg.HiddenRepos {
		if strings.HasPrefix(r, "~/") {
			cfg.HiddenRepos[i] = filepath.Join(home, r[2:])
		}
	}

	return cfg, nil
}

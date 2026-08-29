package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	ScanRoots   []string `toml:"scan_roots"`
	IgnoredDirs []string `toml:"ignored_dirs"`
	Port        string   `toml:"port"`
	DefaultRepo string   `toml:"default_repo,omitempty"`
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		ScanRoots: []string{
			filepath.Join(home, "projects"),
			filepath.Join(home, ".config", "nix-config"),
		},
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
		Port: "8420",
	}
}

func LoadConfig() (*Config, error) {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".config", "beads-fleet")
	configFile := filepath.Join(configDir, "config.toml")

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		cfg := DefaultConfig()
		_ = os.MkdirAll(configDir, 0755)
		data, _ := toml.Marshal(cfg)
		_ = os.WriteFile(configFile, data, 0644)
		return cfg, nil
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

	return cfg, nil
}

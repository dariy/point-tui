package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const defaultBaseURL = "https://darii.net"

type Config struct {
	BaseURL string `toml:"base_url"`
	Login   bool   `toml:"login"`
}

type tomlFile struct {
	BaseURL string `toml:"base_url"`
}

// Load resolves config with precedence: flag > env > config file > default.
// flagURL and flagLogin are the values passed on the command line (zero
// values mean "not set by the user").
func Load(flagURL string, flagLogin bool) (*Config, error) {
	cfg := &Config{
		BaseURL: defaultBaseURL,
		Login:   false,
	}

	if err := loadTOML(cfg); err != nil {
		return nil, err
	}

	if v := os.Getenv("POINT_TUI_BASE_URL"); v != "" {
		cfg.BaseURL = normalizeURL(v)
	}

	if flagURL != "" {
		cfg.BaseURL = normalizeURL(flagURL)
	}
	if flagLogin {
		cfg.Login = true
	}

	return cfg, nil
}

func normalizeURL(s string) string {
	if s == "" {
		return ""
	}
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		return "https://" + s
	}
	return s
}

func loadTOML(cfg *Config) error {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil // skip if we can't determine config dir
	}
	path := filepath.Join(dir, "point-tui", "config.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading config file: %w", err)
	}
	var f tomlFile
	if _, err := toml.Decode(string(data), &f); err != nil {
		return fmt.Errorf("parsing config file %s: %w", path, err)
	}
	if f.BaseURL != "" {
		cfg.BaseURL = normalizeURL(f.BaseURL)
	}
	return nil
}

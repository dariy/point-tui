package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrecedence(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		cfg, err := Load("", false)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.BaseURL != defaultBaseURL {
			t.Errorf("got %q, want %q", cfg.BaseURL, defaultBaseURL)
		}
	})

	t.Run("env overrides default", func(t *testing.T) {
		t.Setenv("POINT_TUI_BASE_URL", "https://env.example.com")
		cfg, err := Load("", false)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.BaseURL != "https://env.example.com" {
			t.Errorf("got %q", cfg.BaseURL)
		}
	})

	t.Run("flag overrides env", func(t *testing.T) {
		t.Setenv("POINT_TUI_BASE_URL", "https://env.example.com")
		cfg, err := Load("https://flag.example.com", false)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.BaseURL != "https://flag.example.com" {
			t.Errorf("got %q", cfg.BaseURL)
		}
	})

	t.Run("toml file overrides default", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, "point-tui"), 0755); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, "point-tui", "config.toml")
		if err := os.WriteFile(path, []byte(`base_url = "https://toml.example.com"`), 0644); err != nil {
			t.Fatal(err)
		}
		t.Setenv("XDG_CONFIG_HOME", dir)
		// UserConfigDir on Linux uses $XDG_CONFIG_HOME when set.
		cfg, err := Load("", false)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.BaseURL != "https://toml.example.com" {
			t.Errorf("got %q", cfg.BaseURL)
		}
	})

	t.Run("login flag", func(t *testing.T) {
		cfg, err := Load("", true)
		if err != nil {
			t.Fatal(err)
		}
		if !cfg.Login {
			t.Error("login should be true")
		}
	})

	t.Run("url normalization", func(t *testing.T) {
		cfg, err := Load("darii.net", false)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.BaseURL != "https://darii.net" {
			t.Errorf("got %q, want https://darii.net", cfg.BaseURL)
		}

		cfg, err = Load("http://darii.net", false)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.BaseURL != "http://darii.net" {
			t.Errorf("got %q", cfg.BaseURL)
		}
	})
}

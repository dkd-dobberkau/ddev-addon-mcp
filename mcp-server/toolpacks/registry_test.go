package toolpacks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "mcp-config.yaml")
	err := os.WriteFile(configPath, []byte("toolpacks:\n  redis: true\n  solr: false\n  mailhog: true\nautostart: true\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.ToolPacks["redis"] {
		t.Error("expected redis to be enabled")
	}
	if cfg.ToolPacks["solr"] {
		t.Error("expected solr to be disabled")
	}
	if !cfg.ToolPacks["mailhog"] {
		t.Error("expected mailhog to be enabled")
	}
	if !cfg.AutoStart {
		t.Error("expected autostart to be true")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/path/mcp-config.yaml")
	if err != nil {
		t.Fatalf("should not error on missing file: %v", err)
	}
	if len(cfg.ToolPacks) != 0 {
		t.Error("expected empty toolpacks for missing config")
	}
}

func TestEnabledPacks(t *testing.T) {
	cfg := &Config{ToolPacks: map[string]bool{"redis": true, "solr": false, "mailhog": true}}
	enabled := cfg.EnabledPacks()
	if len(enabled) != 2 {
		t.Errorf("expected 2 enabled packs, got %d", len(enabled))
	}
}

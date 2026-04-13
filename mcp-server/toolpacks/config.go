package toolpacks

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ToolPacks map[string]bool `yaml:"toolpacks"`
	AutoStart bool            `yaml:"autostart"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{ToolPacks: make(map[string]bool)}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.ToolPacks == nil {
		cfg.ToolPacks = make(map[string]bool)
	}
	return &cfg, nil
}

func (c *Config) EnabledPacks() []string {
	var packs []string
	for name, enabled := range c.ToolPacks {
		if enabled {
			packs = append(packs, name)
		}
	}
	return packs
}

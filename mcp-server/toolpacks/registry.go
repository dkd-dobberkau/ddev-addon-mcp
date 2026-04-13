package toolpacks

import (
	"context"
	"log"
	"time"

	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/discovery"
	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/executor"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ToolPack interface {
	Name() string
	ServiceRequired() string
	Register(server *mcp.Server)
}

func LoadEnabled(server *mcp.Server, configPath string) {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		log.Printf("Warning: could not load MCP config: %v", err)
		return
	}
	enabledPacks := cfg.EnabledPacks()
	if len(enabledPacks) == 0 {
		return
	}
	result, err := executor.ExecuteDdev(context.Background(), []string{"describe", "-j"}, 30*time.Second)
	if err != nil {
		log.Printf("Warning: could not discover services: %v", err)
		return
	}
	services, err := discovery.ParseServices([]byte(result.Stdout))
	if err != nil {
		log.Printf("Warning: could not parse service info: %v", err)
		return
	}
	allPacks := []ToolPack{
		&RedisPack{},
		&SolrPack{},
		&MailhogPack{},
	}
	for _, pack := range allPacks {
		if !cfg.ToolPacks[pack.Name()] {
			continue
		}
		if !services.Has(pack.ServiceRequired()) {
			log.Printf("Tool-pack %q enabled but service %q not running, skipping", pack.Name(), pack.ServiceRequired())
			continue
		}
		pack.Register(server)
		log.Printf("Loaded tool-pack: %s", pack.Name())
	}
}

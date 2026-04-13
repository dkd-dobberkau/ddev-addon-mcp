package toolpacks

import (
	"context"
	"strings"
	"time"

	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/executor"
	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type RedisPack struct{}

func (p *RedisPack) Name() string            { return "redis" }
func (p *RedisPack) ServiceRequired() string { return "redis" }

func (p *RedisPack) Register(server *mcp.Server) {
	type RedisCliInput struct {
		Command string `json:"command" jsonschema:"redis-cli command (e.g. 'GET key', 'KEYS *')"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "ddev_redis_cli", Description: "Execute a redis-cli command in the DDEV Redis container."}, func(ctx context.Context, req *mcp.CallToolRequest, input RedisCliInput) (*mcp.CallToolResult, any, error) {
		args := append([]string{"exec", "--service", "redis", "redis-cli"}, strings.Fields(input.Command)...)
		result, err := executor.ExecuteDdev(ctx, args, 30*time.Second)
		if err != nil {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
		}
		if result.ExitCode != 0 {
			return tools.ToolError(result), nil, nil
		}
		return tools.ToolSuccess(result), nil, nil
	})

	type RedisInfoInput struct{}
	mcp.AddTool(server, &mcp.Tool{Name: "ddev_redis_info", Description: "Get Redis server info."}, func(ctx context.Context, req *mcp.CallToolRequest, input RedisInfoInput) (*mcp.CallToolResult, any, error) {
		result, err := executor.ExecuteDdev(ctx, []string{"exec", "--service", "redis", "redis-cli", "INFO"}, 30*time.Second)
		if err != nil {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
		}
		if result.ExitCode != 0 {
			return tools.ToolError(result), nil, nil
		}
		return tools.ToolSuccess(result), nil, nil
	})

	type RedisFlushInput struct {
		Confirm bool `json:"confirm" jsonschema:"Must be true to execute. This flushes all Redis data."`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "ddev_redis_flush", Description: "Flush all Redis data. Requires confirm: true."}, func(ctx context.Context, req *mcp.CallToolRequest, input RedisFlushInput) (*mcp.CallToolResult, any, error) {
		if !input.Confirm {
			return tools.ConfirmationRequired("ddev_redis_flush", "This will delete all data in Redis."), nil, nil
		}
		result, err := executor.ExecuteDdev(ctx, []string{"exec", "--service", "redis", "redis-cli", "FLUSHALL"}, 30*time.Second)
		if err != nil {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
		}
		if result.ExitCode != 0 {
			return tools.ToolError(result), nil, nil
		}
		return tools.ToolSuccess(result), nil, nil
	})
}

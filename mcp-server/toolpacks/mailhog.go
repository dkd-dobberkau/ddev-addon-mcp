package toolpacks

import (
	"context"
	"time"

	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/executor"
	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type MailhogPack struct{}

func (p *MailhogPack) Name() string            { return "mailhog" }
func (p *MailhogPack) ServiceRequired() string { return "mailhog" }

func (p *MailhogPack) Register(server *mcp.Server) {
	type MailhogListInput struct{}
	mcp.AddTool(server, &mcp.Tool{Name: "ddev_mailhog_list", Description: "List captured emails in Mailhog."}, func(ctx context.Context, req *mcp.CallToolRequest, input MailhogListInput) (*mcp.CallToolResult, any, error) {
		result, err := executor.ExecuteDdev(ctx, []string{"exec", "curl", "-s", "http://mailhog:8025/api/v2/messages"}, 30*time.Second)
		if err != nil {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
		}
		if result.ExitCode != 0 {
			return tools.ToolError(result), nil, nil
		}
		return tools.ToolSuccess(result), nil, nil
	})

	type MailhogSearchInput struct {
		Kind  string `json:"kind" jsonschema:"Search kind: 'from', 'to', 'containing'"`
		Query string `json:"query" jsonschema:"Search query string"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "ddev_mailhog_search", Description: "Search captured emails by from, to, or containing text."}, func(ctx context.Context, req *mcp.CallToolRequest, input MailhogSearchInput) (*mcp.CallToolResult, any, error) {
		url := "http://mailhog:8025/api/v2/search?kind=" + input.Kind + "&query=" + input.Query
		result, err := executor.ExecuteDdev(ctx, []string{"exec", "curl", "-s", url}, 30*time.Second)
		if err != nil {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
		}
		if result.ExitCode != 0 {
			return tools.ToolError(result), nil, nil
		}
		return tools.ToolSuccess(result), nil, nil
	})

	type MailhogDeleteInput struct {
		Confirm bool `json:"confirm" jsonschema:"Must be true. Deletes all captured emails."`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "ddev_mailhog_delete", Description: "Delete all captured emails in Mailhog. Requires confirm: true."}, func(ctx context.Context, req *mcp.CallToolRequest, input MailhogDeleteInput) (*mcp.CallToolResult, any, error) {
		if !input.Confirm {
			return tools.ConfirmationRequired("ddev_mailhog_delete", "This will delete all captured emails."), nil, nil
		}
		result, err := executor.ExecuteDdev(ctx, []string{"exec", "curl", "-s", "-X", "DELETE", "http://mailhog:8025/api/v1/messages"}, 30*time.Second)
		if err != nil {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
		}
		if result.ExitCode != 0 {
			return tools.ToolError(result), nil, nil
		}
		return tools.ToolSuccess(result), nil, nil
	})
}

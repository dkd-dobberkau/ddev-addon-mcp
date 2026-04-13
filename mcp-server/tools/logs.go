package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/executor"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type LogsInput struct {
	Project string `json:"project,omitempty" jsonschema:"Project name (optional)"`
	Service string `json:"service,omitempty" jsonschema:"Service name (e.g. 'web', 'db'). Default: all services."`
	Tail    *int   `json:"tail,omitempty" jsonschema:"Number of lines to show from the end of the log"`
}

func Logs(ctx context.Context, req *mcp.CallToolRequest, input LogsInput) (*mcp.CallToolResult, any, error) {
	args := []string{"logs"}
	if input.Project != "" {
		args = append(args, "--project", input.Project)
	}
	if input.Service != "" {
		args = append(args, "--service", input.Service)
	}
	if input.Tail != nil {
		args = append(args, "--tail", fmt.Sprintf("%d", *input.Tail))
	}
	result, err := executor.ExecuteDdev(ctx, args, 30*time.Second)
	if err != nil {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
	}
	if result.ExitCode != 0 {
		return ToolError(result), nil, nil
	}
	return ToolSuccess(result), nil, nil
}

func RegisterLogs(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{Name: "ddev_logs", Description: "Get logs from a DDEV service (web, db, or other containers)."}, Logs)
}

package tools

import (
	"context"
	"strings"
	"time"

	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/executor"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ComposerInput struct {
	Command string `json:"command" jsonschema:"Composer command and arguments (e.g. 'require typo3/cms-core', 'install', 'update')"`
	Project string `json:"project,omitempty" jsonschema:"Project name (optional)"`
}

func Composer(ctx context.Context, req *mcp.CallToolRequest, input ComposerInput) (*mcp.CallToolResult, any, error) {
	args := []string{"composer"}
	if input.Project != "" {
		args = append(args, "--project", input.Project)
	}
	args = append(args, strings.Fields(input.Command)...)
	result, err := executor.ExecuteDdev(ctx, args, 30*time.Second)
	if err != nil {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
	}
	if result.ExitCode != 0 {
		return ToolError(result), nil, nil
	}
	return ToolSuccess(result), nil, nil
}

func RegisterComposer(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{Name: "ddev_composer", Description: "Run a composer command inside the DDEV container (e.g. require, update, install)."}, Composer)
}

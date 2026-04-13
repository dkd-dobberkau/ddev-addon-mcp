package tools

import (
	"context"
	"time"

	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/executor"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const lifecycleTimeout = 120 * time.Second

type StartInput struct {
	Project string `json:"project,omitempty" jsonschema:"Project name (optional, uses current directory if omitted)"`
}

func Start(ctx context.Context, req *mcp.CallToolRequest, input StartInput) (*mcp.CallToolResult, any, error) {
	args := []string{"start"}
	if input.Project != "" {
		args = append(args, "--project", input.Project)
	}
	result, err := executor.ExecuteDdev(ctx, args, lifecycleTimeout)
	if err != nil {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
	}
	if result.ExitCode != 0 {
		return ToolError(result), nil, nil
	}
	return ToolSuccess(result), nil, nil
}

type StopInput struct {
	Project string `json:"project,omitempty" jsonschema:"Project name (optional)"`
}

func Stop(ctx context.Context, req *mcp.CallToolRequest, input StopInput) (*mcp.CallToolResult, any, error) {
	args := []string{"stop"}
	if input.Project != "" {
		args = append(args, "--project", input.Project)
	}
	result, err := executor.ExecuteDdev(ctx, args, lifecycleTimeout)
	if err != nil {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
	}
	if result.ExitCode != 0 {
		return ToolError(result), nil, nil
	}
	return ToolSuccess(result), nil, nil
}

type RestartInput struct {
	Project string `json:"project,omitempty" jsonschema:"Project name (optional)"`
}

func Restart(ctx context.Context, req *mcp.CallToolRequest, input RestartInput) (*mcp.CallToolResult, any, error) {
	args := []string{"restart"}
	if input.Project != "" {
		args = append(args, "--project", input.Project)
	}
	result, err := executor.ExecuteDdev(ctx, args, lifecycleTimeout)
	if err != nil {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
	}
	if result.ExitCode != 0 {
		return ToolError(result), nil, nil
	}
	return ToolSuccess(result), nil, nil
}

type PoweroffInput struct {
	Confirm bool `json:"confirm" jsonschema:"Must be true to execute. This stops ALL projects, not just one."`
}

func Poweroff(ctx context.Context, req *mcp.CallToolRequest, input PoweroffInput) (*mcp.CallToolResult, any, error) {
	if !input.Confirm {
		return ConfirmationRequired("ddev_poweroff", "This will stop ALL running DDEV projects and containers."), nil, nil
	}
	result, err := executor.ExecuteDdev(ctx, []string{"poweroff"}, lifecycleTimeout)
	if err != nil {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
	}
	if result.ExitCode != 0 {
		return ToolError(result), nil, nil
	}
	return ToolSuccess(result), nil, nil
}

func RegisterLifecycle(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{Name: "ddev_start", Description: "Start a DDEV project. Downloads Docker images on first run (may take minutes)."}, Start)
	mcp.AddTool(server, &mcp.Tool{Name: "ddev_stop", Description: "Stop a DDEV project."}, Stop)
	mcp.AddTool(server, &mcp.Tool{Name: "ddev_restart", Description: "Restart a DDEV project."}, Restart)
	mcp.AddTool(server, &mcp.Tool{Name: "ddev_poweroff", Description: "Stop ALL running DDEV projects and containers. Requires confirm: true."}, Poweroff)
}

package tools

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/executor"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const infoTimeout = 30 * time.Second

type ListInput struct{}

func List(ctx context.Context, req *mcp.CallToolRequest, input ListInput) (*mcp.CallToolResult, any, error) {
	result, err := executor.ExecuteDdev(ctx, []string{"list", "-j"}, infoTimeout)
	if err != nil {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
	}
	if result.ExitCode != 0 {
		return ToolError(result), nil, nil
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(result.Stdout), &parsed); err == nil {
		if raw, ok := parsed["raw"]; ok {
			jsonData, _ := json.MarshalIndent(map[string]any{"projects": raw}, "", "  ")
			return ToolSuccessJSON(result, string(jsonData)), nil, nil
		}
	}
	return ToolSuccess(result), nil, nil
}

type DescribeInput struct {
	Project string `json:"project,omitempty" jsonschema:"Project name (optional)"`
}

func Describe(ctx context.Context, req *mcp.CallToolRequest, input DescribeInput) (*mcp.CallToolResult, any, error) {
	args := []string{"describe", "-j"}
	if input.Project != "" {
		args = append(args, "--project", input.Project)
	}
	result, err := executor.ExecuteDdev(ctx, args, infoTimeout)
	if err != nil {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
	}
	if result.ExitCode != 0 {
		return ToolError(result), nil, nil
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(result.Stdout), &parsed); err == nil {
		if raw, ok := parsed["raw"]; ok {
			jsonData, _ := json.MarshalIndent(map[string]any{"project": raw}, "", "  ")
			return ToolSuccessJSON(result, string(jsonData)), nil, nil
		}
	}
	return ToolSuccess(result), nil, nil
}

func RegisterInfo(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{Name: "ddev_list", Description: "List all DDEV projects with their status, type, and URLs."}, List)
	mcp.AddTool(server, &mcp.Tool{Name: "ddev_describe", Description: "Get detailed info about a DDEV project: URLs, database credentials, PHP version, services."}, Describe)
}

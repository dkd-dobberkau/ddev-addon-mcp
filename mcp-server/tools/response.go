package tools

import (
	"fmt"
	"strings"

	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/executor"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func ToolSuccess(result *executor.Result) *mcp.CallToolResult {
	parts := []string{fmt.Sprintf("> Executed: `%s`", result.Command)}
	if strings.TrimSpace(result.Stdout) != "" {
		parts = append(parts, strings.TrimSpace(result.Stdout))
	}
	if strings.TrimSpace(result.Stderr) != "" {
		parts = append(parts, "stderr: "+strings.TrimSpace(result.Stderr))
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: strings.Join(parts, "\n\n")},
		},
	}
}

func ToolSuccessJSON(result *executor.Result, jsonData string) *mcp.CallToolResult {
	parts := []string{
		fmt.Sprintf("> Executed: `%s`", result.Command),
		jsonData,
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: strings.Join(parts, "\n\n")},
		},
	}
}

func ToolError(result *executor.Result) *mcp.CallToolResult {
	parts := []string{
		fmt.Sprintf("> Executed: `%s`", result.Command),
		fmt.Sprintf("Error (exit code %d):", result.ExitCode),
	}
	if strings.TrimSpace(result.Stderr) != "" {
		parts = append(parts, strings.TrimSpace(result.Stderr))
	}
	if strings.TrimSpace(result.Stdout) != "" {
		parts = append(parts, strings.TrimSpace(result.Stdout))
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: strings.Join(parts, "\n\n")},
		},
		IsError: true,
	}
}

func ConfirmationRequired(toolName, warning string) *mcp.CallToolResult {
	text := fmt.Sprintf("⚠ %s requires confirmation. %s Pass `confirm: true` to execute.", toolName, warning)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
		IsError: true,
	}
}

package tools

import (
	"context"
	"strings"
	"time"

	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/executor"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ExecInput struct {
	Command string `json:"command" jsonschema:"Command to execute (e.g. 'php -v', 'vendor/bin/typo3 cache:flush')"`
	Project string `json:"project,omitempty" jsonschema:"Project name (optional)"`
	Service string `json:"service,omitempty" jsonschema:"Container service (default: web). Use 'db' for database container."`
	Raw     bool   `json:"raw,omitempty" jsonschema:"Use --raw for unprocessed output"`
}

func Exec(ctx context.Context, req *mcp.CallToolRequest, input ExecInput) (*mcp.CallToolResult, any, error) {
	args := []string{"exec"}
	if input.Project != "" {
		args = append(args, "--project", input.Project)
	}
	if input.Service != "" {
		args = append(args, "--service", input.Service)
	}
	if input.Raw {
		args = append(args, "--raw")
	}
	args = append(args, splitCommand(input.Command)...)
	result, err := executor.ExecuteDdev(ctx, args, 30*time.Second)
	if err != nil {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
	}
	if result.ExitCode != 0 {
		return ToolError(result), nil, nil
	}
	return ToolSuccess(result), nil, nil
}

func splitCommand(cmd string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false
	quoteChar := byte(0)
	for i := 0; i < len(cmd); i++ {
		c := cmd[i]
		if inQuote {
			if c == quoteChar {
				inQuote = false
			} else {
				current.WriteByte(c)
			}
		} else if c == '\'' || c == '"' {
			inQuote = true
			quoteChar = c
		} else if c == ' ' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(c)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

func RegisterExec(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{Name: "ddev_exec", Description: "Run a command inside the DDEV web container. Use for framework CLI tools (e.g. vendor/bin/typo3, php artisan, wp)."}, Exec)
}

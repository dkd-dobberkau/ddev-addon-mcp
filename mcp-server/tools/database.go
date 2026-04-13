package tools

import (
	"context"
	"time"

	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/executor"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ImportDBInput struct {
	File    string `json:"file" jsonschema:"Path to the SQL dump file"`
	Project string `json:"project,omitempty" jsonschema:"Project name (optional)"`
	Confirm bool   `json:"confirm" jsonschema:"Must be true to execute. This overwrites the current database."`
}

func ImportDB(ctx context.Context, req *mcp.CallToolRequest, input ImportDBInput) (*mcp.CallToolResult, any, error) {
	if !input.Confirm {
		return ConfirmationRequired("ddev_import_db", "This will overwrite the current project database."), nil, nil
	}
	args := []string{"import-db", "--file", input.File}
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

type ExportDBInput struct {
	File    string `json:"file,omitempty" jsonschema:"Output file path (optional, outputs to stdout if omitted)"`
	Project string `json:"project,omitempty" jsonschema:"Project name (optional)"`
}

func ExportDB(ctx context.Context, req *mcp.CallToolRequest, input ExportDBInput) (*mcp.CallToolResult, any, error) {
	args := []string{"export-db"}
	if input.File != "" {
		args = append(args, "--file", input.File)
	}
	if input.Project != "" {
		args = append(args, "--project", input.Project)
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

type SnapshotInput struct {
	Project string  `json:"project,omitempty" jsonschema:"Project name (optional)"`
	Name    string  `json:"name,omitempty" jsonschema:"Snapshot name"`
	Action  *string `json:"action,omitempty" jsonschema:"Action: 'create' (default) or 'restore'"`
}

func Snapshot(ctx context.Context, req *mcp.CallToolRequest, input SnapshotInput) (*mcp.CallToolResult, any, error) {
	args := []string{"snapshot"}
	if input.Action != nil && *input.Action == "restore" {
		args = append(args, "restore")
		if input.Name != "" {
			args = append(args, input.Name)
		} else {
			args = append(args, "--latest")
		}
	} else {
		if input.Name != "" {
			args = append(args, "--name", input.Name)
		}
	}
	if input.Project != "" {
		args = append(args, "--project", input.Project)
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

func RegisterDatabase(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{Name: "ddev_import_db", Description: "Import a SQL dump file into the project database. Overwrites existing data. Requires confirm: true."}, ImportDB)
	mcp.AddTool(server, &mcp.Tool{Name: "ddev_export_db", Description: "Export the project database to a file or stdout."}, ExportDB)
	mcp.AddTool(server, &mcp.Tool{Name: "ddev_snapshot", Description: "Create or restore a database snapshot."}, Snapshot)
}

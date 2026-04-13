package tools

import (
	"testing"

	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/executor"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestToolSuccess(t *testing.T) {
	result := &executor.Result{
		Command:  "ddev list",
		Stdout:   "project1 running",
		Stderr:   "",
		ExitCode: 0,
	}
	resp := ToolSuccess(result)
	if resp.IsError {
		t.Error("expected IsError to be false")
	}
	if len(resp.Content) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(resp.Content))
	}
	text := resp.Content[0].(*mcp.TextContent).Text
	if text == "" {
		t.Error("expected non-empty text")
	}
}

func TestToolSuccessJSON(t *testing.T) {
	result := &executor.Result{
		Command:  "ddev list -j",
		Stdout:   `{"raw":[]}`,
		Stderr:   "",
		ExitCode: 0,
	}
	resp := ToolSuccessJSON(result, `{"projects":[]}`)
	text := resp.Content[0].(*mcp.TextContent).Text
	if text == "" {
		t.Error("expected non-empty text")
	}
}

func TestToolError(t *testing.T) {
	result := &executor.Result{
		Command:  "ddev start",
		Stdout:   "",
		Stderr:   "Docker not running",
		ExitCode: 1,
	}
	resp := ToolError(result)
	if !resp.IsError {
		t.Error("expected IsError to be true")
	}
	text := resp.Content[0].(*mcp.TextContent).Text
	if text == "" {
		t.Error("expected non-empty text")
	}
}

func TestConfirmationRequired(t *testing.T) {
	resp := ConfirmationRequired("ddev_poweroff", "This will stop ALL projects.")
	if !resp.IsError {
		t.Error("expected IsError to be true")
	}
	text := resp.Content[0].(*mcp.TextContent).Text
	if text == "" {
		t.Error("expected non-empty text")
	}
}

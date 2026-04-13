package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func setupFakeDdev(t *testing.T, script string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "ddev")
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
}

func TestStart_Success(t *testing.T) {
	setupFakeDdev(t, "#!/bin/sh\necho \"Successfully started\"\nexit 0\n")
	result, _, err := Start(context.Background(), nil, StartInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success, got error")
	}
}

func TestStart_WithProject(t *testing.T) {
	setupFakeDdev(t, "#!/bin/sh\necho \"Started $@\"\nexit 0\n")
	result, _, err := Start(context.Background(), nil, StartInput{Project: "my-site"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}

func TestPoweroff_RequiresConfirm(t *testing.T) {
	result, _, err := Poweroff(context.Background(), nil, PoweroffInput{Confirm: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error when confirm is false")
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if text == "" {
		t.Error("expected confirmation message")
	}
}

func TestPoweroff_Confirmed(t *testing.T) {
	setupFakeDdev(t, "#!/bin/sh\necho \"All projects stopped\"\nexit 0\n")
	result, _, err := Poweroff(context.Background(), nil, PoweroffInput{Confirm: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success when confirmed")
	}
}

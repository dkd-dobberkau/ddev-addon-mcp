# ddev-addon-mcp Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a DDEV add-on that provides a Go-based MCP server, enabling AI agents to manage DDEV projects via stdio with optional tool-packs for additional services.

**Architecture:** Go binary communicates over stdio using the official Go MCP SDK. Core tools wrap `ddev` CLI commands. Tool-packs extend functionality for Redis, Solr, Mailhog based on config and service discovery. DDEV add-on installs the binary and a custom `ddev mcp` command.

**Tech Stack:** Go 1.22+, `github.com/modelcontextprotocol/go-sdk`, `os/exec`, `gopkg.in/yaml.v3`

---

### Task 1: Project Scaffolding

**Files:**
- Create: `mcp-server/go.mod`
- Create: `mcp-server/main.go` (stub)
- Create: `.gitignore`
- Create: `.goreleaser.yml`

- [ ] **Step 1: Initialize Go module**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp
mkdir -p mcp-server
cd mcp-server
go mod init github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server
```

- [ ] **Step 2: Add dependencies**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go get github.com/modelcontextprotocol/go-sdk/mcp
go get gopkg.in/yaml.v3
```

- [ ] **Step 3: Create stub main.go**

Create `mcp-server/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("ddev-mcp-server " + version)
		os.Exit(0)
	}

	server := mcp.NewServer(
		&mcp.Implementation{Name: "ddev-mcp", Version: version},
		nil,
	)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
```

- [ ] **Step 4: Verify it compiles**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go build -o ddev-mcp-server .
./ddev-mcp-server --version
```

Expected: `ddev-mcp-server dev`

- [ ] **Step 5: Create .gitignore**

Create `/Users/olivier/Versioncontrol/github/ddev-addon-mcp/.gitignore`:

```
mcp-server/ddev-mcp-server
dist/
.ddev/
```

- [ ] **Step 6: Create .goreleaser.yml**

Create `/Users/olivier/Versioncontrol/github/ddev-addon-mcp/.goreleaser.yml`:

```yaml
version: 2

builds:
  - id: ddev-mcp-server
    dir: mcp-server
    binary: ddev-mcp-server
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}}

archives:
  - format: binary
    name_template: "ddev-mcp-server_{{ .Os }}_{{ .Arch }}"
```

- [ ] **Step 7: Init git and commit**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp
git init
git add .gitignore .goreleaser.yml mcp-server/go.mod mcp-server/go.sum mcp-server/main.go
git commit -m "chore: scaffold Go MCP server project"
```

---

### Task 2: Executor

**Files:**
- Create: `mcp-server/executor/executor.go`
- Create: `mcp-server/executor/executor_test.go`

- [ ] **Step 1: Write executor test**

Create `mcp-server/executor/executor_test.go`:

```go
package executor

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestExecuteDdev_Success tests successful command execution using a fake ddev script.
func TestExecuteDdev_Success(t *testing.T) {
	fakeDdev := createFakeDdev(t, `#!/bin/sh
echo "project1 running"
exit 0
`)
	t.Setenv("PATH", filepath.Dir(fakeDdev)+":"+os.Getenv("PATH"))

	result, err := ExecuteDdev(context.Background(), []string{"list"}, 30*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
	if result.Command != "ddev list" {
		t.Errorf("expected command 'ddev list', got %q", result.Command)
	}
	if result.Stdout != "project1 running\n" {
		t.Errorf("unexpected stdout: %q", result.Stdout)
	}
}

// TestExecuteDdev_Failure tests command failure with non-zero exit code.
func TestExecuteDdev_Failure(t *testing.T) {
	fakeDdev := createFakeDdev(t, `#!/bin/sh
echo "Error: project not found" >&2
exit 1
`)
	t.Setenv("PATH", filepath.Dir(fakeDdev)+":"+os.Getenv("PATH"))

	result, err := ExecuteDdev(context.Background(), []string{"describe", "--project", "nope"}, 30*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", result.ExitCode)
	}
	if result.Command != "ddev describe --project nope" {
		t.Errorf("unexpected command: %q", result.Command)
	}
}

// TestExecuteDdev_Timeout tests that execution times out.
func TestExecuteDdev_Timeout(t *testing.T) {
	fakeDdev := createFakeDdev(t, `#!/bin/sh
sleep 10
`)
	t.Setenv("PATH", filepath.Dir(fakeDdev)+":"+os.Getenv("PATH"))

	_, err := ExecuteDdev(context.Background(), []string{"start"}, 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

// createFakeDdev writes a temporary script named "ddev" and returns its path.
func createFakeDdev(t *testing.T, script string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "ddev")
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	// Verify it's executable
	if _, err := exec.LookPath(path); err != nil {
		t.Fatal(err)
	}
	return path
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go test ./executor/ -v
```

Expected: FAIL — package/function not found

- [ ] **Step 3: Implement executor**

Create `mcp-server/executor/executor.go`:

```go
package executor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Result holds the output of a ddev command execution.
type Result struct {
	Command  string
	Stdout   string
	Stderr   string
	ExitCode int
}

// ExecuteDdev runs a ddev command with the given arguments and timeout.
// It uses exec.CommandContext for timeout support and never invokes a shell.
func ExecuteDdev(ctx context.Context, args []string, timeout time.Duration) (*Result, error) {
	command := "ddev " + strings.Join(args, " ")

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ddev", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("command timed out after %v: %s", timeout, command)
	}

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute %s: %w", command, err)
		}
	}

	return &Result{
		Command:  command,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}
```

- [ ] **Step 4: Run tests**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go test ./executor/ -v
```

Expected: All 3 tests PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp
git add mcp-server/executor/
git commit -m "feat: add ddev command executor with tests"
```

---

### Task 3: Response Helpers

**Files:**
- Create: `mcp-server/tools/response.go`
- Create: `mcp-server/tools/response_test.go`

- [ ] **Step 1: Write response helper test**

Create `mcp-server/tools/response_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go test ./tools/ -v
```

Expected: FAIL — functions not found

- [ ] **Step 3: Implement response helpers**

Create `mcp-server/tools/response.go`:

```go
package tools

import (
	"fmt"
	"strings"

	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/executor"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolSuccess formats a successful command result.
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

// ToolSuccessJSON formats a successful result with custom JSON data.
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

// ToolError formats a failed command result.
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

// ConfirmationRequired returns an error result asking the agent to confirm.
func ConfirmationRequired(toolName, warning string) *mcp.CallToolResult {
	text := fmt.Sprintf("⚠ %s requires confirmation. %s Pass `confirm: true` to execute.", toolName, warning)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
		IsError: true,
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go test ./tools/ -v
```

Expected: All 4 tests PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp
git add mcp-server/tools/response.go mcp-server/tools/response_test.go
git commit -m "feat: add response helpers for tool results"
```

---

### Task 4: Lifecycle Tools

**Files:**
- Create: `mcp-server/tools/lifecycle.go`
- Create: `mcp-server/tools/lifecycle_test.go`

- [ ] **Step 1: Write lifecycle tools test**

Create `mcp-server/tools/lifecycle_test.go`:

```go
package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

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
	setupFakeDdev(t, `#!/bin/sh
echo "Successfully started"
exit 0
`)

	result, _, err := Start(context.Background(), nil, StartInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success, got error")
	}
}

func TestStart_WithProject(t *testing.T) {
	setupFakeDdev(t, `#!/bin/sh
echo "Started $@"
exit 0
`)

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
	setupFakeDdev(t, `#!/bin/sh
echo "All projects stopped"
exit 0
`)

	result, _, err := Poweroff(context.Background(), nil, PoweroffInput{Confirm: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success when confirmed")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go test ./tools/ -run TestStart -v
```

Expected: FAIL — functions not defined

- [ ] **Step 3: Implement lifecycle tools**

Create `mcp-server/tools/lifecycle.go`:

```go
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
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			IsError: true,
		}, nil, nil
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
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			IsError: true,
		}, nil, nil
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
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			IsError: true,
		}, nil, nil
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
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			IsError: true,
		}, nil, nil
	}
	if result.ExitCode != 0 {
		return ToolError(result), nil, nil
	}
	return ToolSuccess(result), nil, nil
}

// RegisterLifecycle registers all lifecycle tools on the server.
func RegisterLifecycle(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ddev_start",
		Description: "Start a DDEV project. Downloads Docker images on first run (may take minutes).",
	}, Start)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ddev_stop",
		Description: "Stop a DDEV project.",
	}, Stop)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ddev_restart",
		Description: "Restart a DDEV project.",
	}, Restart)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ddev_poweroff",
		Description: "Stop ALL running DDEV projects and containers. Requires confirm: true.",
	}, Poweroff)
}
```

- [ ] **Step 4: Run tests**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go test ./tools/ -v
```

Expected: All tests PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp
git add mcp-server/tools/lifecycle.go mcp-server/tools/lifecycle_test.go
git commit -m "feat: add lifecycle tools (start, stop, restart, poweroff)"
```

---

### Task 5: Info Tools

**Files:**
- Create: `mcp-server/tools/info.go`
- Create: `mcp-server/tools/info_test.go`

- [ ] **Step 1: Write info tools test**

Create `mcp-server/tools/info_test.go`:

```go
package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func TestList_ParsesJSON(t *testing.T) {
	ddevOutput := `{"raw":[{"name":"my-project","status":"running"}]}`
	setupFakeDdev(t, `#!/bin/sh
echo '`+ddevOutput+`'
exit 0
`)

	result, _, err := List(context.Background(), nil, ListInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}

func TestList_FallbackOnInvalidJSON(t *testing.T) {
	setupFakeDdev(t, `#!/bin/sh
echo 'not valid json'
exit 0
`)

	result, _, err := List(context.Background(), nil, ListInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success even with invalid JSON")
	}
}

func TestDescribe_ParsesJSON(t *testing.T) {
	raw := map[string]any{
		"raw": map[string]any{
			"name":   "my-project",
			"status": "running",
		},
	}
	data, _ := json.Marshal(raw)
	setupFakeDdev(t, `#!/bin/sh
echo '`+string(data)+`'
exit 0
`)

	result, _, err := Describe(context.Background(), nil, DescribeInput{Project: "my-project"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go test ./tools/ -run TestList -v
```

Expected: FAIL

- [ ] **Step 3: Implement info tools**

Create `mcp-server/tools/info.go`:

```go
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
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			IsError: true,
		}, nil, nil
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
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			IsError: true,
		}, nil, nil
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

// RegisterInfo registers all info tools on the server.
func RegisterInfo(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ddev_list",
		Description: "List all DDEV projects with their status, type, and URLs.",
	}, List)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ddev_describe",
		Description: "Get detailed info about a DDEV project: URLs, database credentials, PHP version, services.",
	}, Describe)
}
```

- [ ] **Step 4: Run tests**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go test ./tools/ -v
```

Expected: All tests PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp
git add mcp-server/tools/info.go mcp-server/tools/info_test.go
git commit -m "feat: add info tools (list, describe) with JSON parsing"
```

---

### Task 6: Exec, Composer, Logs Tools

**Files:**
- Create: `mcp-server/tools/exec.go`
- Create: `mcp-server/tools/composer.go`
- Create: `mcp-server/tools/logs.go`
- Create: `mcp-server/tools/exec_test.go`

- [ ] **Step 1: Write exec test**

Create `mcp-server/tools/exec_test.go`:

```go
package tools

import (
	"context"
	"testing"
)

func TestExec_BasicCommand(t *testing.T) {
	setupFakeDdev(t, `#!/bin/sh
echo "PHP 8.2.0"
exit 0
`)

	result, _, err := Exec(context.Background(), nil, ExecInput{Command: "php -v"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}

func TestExec_WithService(t *testing.T) {
	setupFakeDdev(t, `#!/bin/sh
echo "$@"
exit 0
`)

	result, _, err := Exec(context.Background(), nil, ExecInput{
		Command: "mysql -e 'SHOW TABLES'",
		Service: "db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}

func TestComposer_BasicCommand(t *testing.T) {
	setupFakeDdev(t, `#!/bin/sh
echo "Installing dependencies"
exit 0
`)

	result, _, err := Composer(context.Background(), nil, ComposerInput{Command: "install"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}

func TestLogs_Defaults(t *testing.T) {
	setupFakeDdev(t, `#!/bin/sh
echo "web log line 1"
exit 0
`)

	result, _, err := Logs(context.Background(), nil, LogsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}

func TestLogs_WithOptions(t *testing.T) {
	setupFakeDdev(t, `#!/bin/sh
echo "db logs"
exit 0
`)

	tail := 50
	result, _, err := Logs(context.Background(), nil, LogsInput{
		Service: "db",
		Tail:    &tail,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go test ./tools/ -run TestExec -v
```

Expected: FAIL

- [ ] **Step 3: Implement exec tool**

Create `mcp-server/tools/exec.go`:

```go
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
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			IsError: true,
		}, nil, nil
	}
	if result.ExitCode != 0 {
		return ToolError(result), nil, nil
	}
	return ToolSuccess(result), nil, nil
}

// splitCommand splits a command string respecting quoted substrings.
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

// RegisterExec registers the exec tool on the server.
func RegisterExec(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ddev_exec",
		Description: "Run a command inside the DDEV web container. Use for framework CLI tools (e.g. vendor/bin/typo3, php artisan, wp).",
	}, Exec)
}
```

- [ ] **Step 4: Implement composer tool**

Create `mcp-server/tools/composer.go`:

```go
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
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			IsError: true,
		}, nil, nil
	}
	if result.ExitCode != 0 {
		return ToolError(result), nil, nil
	}
	return ToolSuccess(result), nil, nil
}

// RegisterComposer registers the composer tool on the server.
func RegisterComposer(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ddev_composer",
		Description: "Run a composer command inside the DDEV container (e.g. require, update, install).",
	}, Composer)
}
```

- [ ] **Step 5: Implement logs tool**

Create `mcp-server/tools/logs.go`:

```go
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
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			IsError: true,
		}, nil, nil
	}
	if result.ExitCode != 0 {
		return ToolError(result), nil, nil
	}
	return ToolSuccess(result), nil, nil
}

// RegisterLogs registers the logs tool on the server.
func RegisterLogs(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ddev_logs",
		Description: "Get logs from a DDEV service (web, db, or other containers).",
	}, Logs)
}
```

- [ ] **Step 6: Run all tests**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go test ./... -v
```

Expected: All tests PASS

- [ ] **Step 7: Commit**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp
git add mcp-server/tools/exec.go mcp-server/tools/composer.go mcp-server/tools/logs.go mcp-server/tools/exec_test.go
git commit -m "feat: add exec, composer, and logs tools"
```

---

### Task 7: Database Tools

**Files:**
- Create: `mcp-server/tools/database.go`
- Create: `mcp-server/tools/database_test.go`

- [ ] **Step 1: Write database tools test**

Create `mcp-server/tools/database_test.go`:

```go
package tools

import (
	"context"
	"testing"
)

func TestImportDB_RequiresConfirm(t *testing.T) {
	result, _, err := ImportDB(context.Background(), nil, ImportDBInput{
		File:    "/tmp/dump.sql",
		Confirm: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error when confirm is false")
	}
}

func TestImportDB_Confirmed(t *testing.T) {
	setupFakeDdev(t, `#!/bin/sh
echo "Successfully imported"
exit 0
`)

	result, _, err := ImportDB(context.Background(), nil, ImportDBInput{
		File:    "/tmp/dump.sql",
		Confirm: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success when confirmed")
	}
}

func TestExportDB_Success(t *testing.T) {
	setupFakeDdev(t, `#!/bin/sh
echo "Exported"
exit 0
`)

	result, _, err := ExportDB(context.Background(), nil, ExportDBInput{
		File: "/tmp/export.sql.gz",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}

func TestSnapshot_Create(t *testing.T) {
	setupFakeDdev(t, `#!/bin/sh
echo "Snapshot created"
exit 0
`)

	result, _, err := Snapshot(context.Background(), nil, SnapshotInput{
		Name: "before-migration",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}

func TestSnapshot_Restore(t *testing.T) {
	setupFakeDdev(t, `#!/bin/sh
echo "Restored"
exit 0
`)

	action := "restore"
	result, _, err := Snapshot(context.Background(), nil, SnapshotInput{
		Action: &action,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go test ./tools/ -run TestImport -v
```

Expected: FAIL

- [ ] **Step 3: Implement database tools**

Create `mcp-server/tools/database.go`:

```go
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
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			IsError: true,
		}, nil, nil
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
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			IsError: true,
		}, nil, nil
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
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			IsError: true,
		}, nil, nil
	}
	if result.ExitCode != 0 {
		return ToolError(result), nil, nil
	}
	return ToolSuccess(result), nil, nil
}

// RegisterDatabase registers all database tools on the server.
func RegisterDatabase(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ddev_import_db",
		Description: "Import a SQL dump file into the project database. Overwrites existing data. Requires confirm: true.",
	}, ImportDB)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ddev_export_db",
		Description: "Export the project database to a file or stdout.",
	}, ExportDB)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ddev_snapshot",
		Description: "Create or restore a database snapshot.",
	}, Snapshot)
}
```

- [ ] **Step 4: Run all tests**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go test ./... -v
```

Expected: All tests PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp
git add mcp-server/tools/database.go mcp-server/tools/database_test.go
git commit -m "feat: add database tools (import, export, snapshot)"
```

---

### Task 8: Wire Up Server

**Files:**
- Modify: `mcp-server/main.go`

- [ ] **Step 1: Update main.go to register all tools**

Replace `mcp-server/main.go` with:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("ddev-mcp-server " + version)
		os.Exit(0)
	}

	server := mcp.NewServer(
		&mcp.Implementation{Name: "ddev-mcp", Version: version},
		nil,
	)

	tools.RegisterLifecycle(server)
	tools.RegisterInfo(server)
	tools.RegisterExec(server)
	tools.RegisterDatabase(server)
	tools.RegisterComposer(server)
	tools.RegisterLogs(server)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
```

- [ ] **Step 2: Build and verify**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go build -o ddev-mcp-server .
./ddev-mcp-server --version
```

Expected: `ddev-mcp-server dev`

- [ ] **Step 3: Run all tests**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go test ./... -v
```

Expected: All tests PASS

- [ ] **Step 4: Commit**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp
git add mcp-server/main.go
git commit -m "feat: wire up all core tools in server entry point"
```

---

### Task 9: Service Discovery

**Files:**
- Create: `mcp-server/discovery/discovery.go`
- Create: `mcp-server/discovery/discovery_test.go`

- [ ] **Step 1: Write discovery test**

Create `mcp-server/discovery/discovery_test.go`:

```go
package discovery

import (
	"testing"
)

func TestParseServices(t *testing.T) {
	jsonData := `{
		"raw": {
			"name": "my-project",
			"services": {
				"web": {"status": "running"},
				"db": {"status": "running"},
				"redis": {"status": "running"},
				"solr": {"status": "running"}
			}
		}
	}`

	services, err := ParseServices([]byte(jsonData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !services.Has("redis") {
		t.Error("expected redis to be detected")
	}
	if !services.Has("solr") {
		t.Error("expected solr to be detected")
	}
	if services.Has("mailhog") {
		t.Error("expected mailhog to not be detected")
	}
}

func TestParseServices_NoExtra(t *testing.T) {
	jsonData := `{
		"raw": {
			"name": "my-project",
			"services": {
				"web": {"status": "running"},
				"db": {"status": "running"}
			}
		}
	}`

	services, err := ParseServices([]byte(jsonData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if services.Has("redis") {
		t.Error("expected redis to not be detected")
	}
}

func TestParseServices_InvalidJSON(t *testing.T) {
	_, err := ParseServices([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go test ./discovery/ -v
```

Expected: FAIL

- [ ] **Step 3: Implement discovery**

Create `mcp-server/discovery/discovery.go`:

```go
package discovery

import (
	"encoding/json"
	"fmt"
)

// ProjectServices tracks which extra services are running in a DDEV project.
type ProjectServices struct {
	services map[string]bool
}

// Has returns true if the named service is running.
func (ps *ProjectServices) Has(name string) bool {
	return ps.services[name]
}

// ParseServices parses ddev describe -j output and detects running services.
func ParseServices(data []byte) (*ProjectServices, error) {
	var parsed struct {
		Raw struct {
			Services map[string]struct {
				Status string `json:"status"`
			} `json:"services"`
		} `json:"raw"`
	}

	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse ddev describe output: %w", err)
	}

	services := make(map[string]bool)
	for name, svc := range parsed.Raw.Services {
		if svc.Status == "running" {
			services[name] = true
		}
	}

	return &ProjectServices{services: services}, nil
}
```

- [ ] **Step 4: Run tests**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go test ./discovery/ -v
```

Expected: All 3 tests PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp
git add mcp-server/discovery/
git commit -m "feat: add service discovery from ddev describe output"
```

---

### Task 10: Tool-Pack Registry and Config

**Files:**
- Create: `mcp-server/toolpacks/config.go`
- Create: `mcp-server/toolpacks/registry.go`
- Create: `mcp-server/toolpacks/registry_test.go`
- Create: `mcp-config.yaml.example`

- [ ] **Step 1: Write registry test**

Create `mcp-server/toolpacks/registry_test.go`:

```go
package toolpacks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "mcp-config.yaml")
	err := os.WriteFile(configPath, []byte(`
toolpacks:
  redis: true
  solr: false
  mailhog: true
autostart: true
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.ToolPacks["redis"] {
		t.Error("expected redis to be enabled")
	}
	if cfg.ToolPacks["solr"] {
		t.Error("expected solr to be disabled")
	}
	if !cfg.ToolPacks["mailhog"] {
		t.Error("expected mailhog to be enabled")
	}
	if !cfg.AutoStart {
		t.Error("expected autostart to be true")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/path/mcp-config.yaml")
	if err != nil {
		t.Fatalf("should not error on missing file: %v", err)
	}
	if len(cfg.ToolPacks) != 0 {
		t.Error("expected empty toolpacks for missing config")
	}
}

func TestEnabledPacks(t *testing.T) {
	cfg := &Config{
		ToolPacks: map[string]bool{
			"redis":   true,
			"solr":    false,
			"mailhog": true,
		},
	}

	enabled := cfg.EnabledPacks()
	if len(enabled) != 2 {
		t.Errorf("expected 2 enabled packs, got %d", len(enabled))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go test ./toolpacks/ -v
```

Expected: FAIL

- [ ] **Step 3: Implement config**

Create `mcp-server/toolpacks/config.go`:

```go
package toolpacks

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the MCP add-on configuration.
type Config struct {
	ToolPacks map[string]bool `yaml:"toolpacks"`
	AutoStart bool            `yaml:"autostart"`
}

// LoadConfig reads the mcp-config.yaml file. Returns empty config if file doesn't exist.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{ToolPacks: make(map[string]bool)}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.ToolPacks == nil {
		cfg.ToolPacks = make(map[string]bool)
	}
	return &cfg, nil
}

// EnabledPacks returns the names of all enabled tool-packs.
func (c *Config) EnabledPacks() []string {
	var packs []string
	for name, enabled := range c.ToolPacks {
		if enabled {
			packs = append(packs, name)
		}
	}
	return packs
}
```

- [ ] **Step 4: Implement registry**

Create `mcp-server/toolpacks/registry.go`:

```go
package toolpacks

import (
	"context"
	"log"
	"time"

	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/discovery"
	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/executor"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolPack defines an optional set of tools for a specific service.
type ToolPack interface {
	Name() string
	ServiceRequired() string
	Register(server *mcp.Server)
}

// LoadEnabled loads tool-packs that are enabled in config and whose service is running.
func LoadEnabled(server *mcp.Server, configPath string) {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		log.Printf("Warning: could not load MCP config: %v", err)
		return
	}

	enabledPacks := cfg.EnabledPacks()
	if len(enabledPacks) == 0 {
		return
	}

	// Discover running services
	result, err := executor.ExecuteDdev(context.Background(), []string{"describe", "-j"}, 30*time.Second)
	if err != nil {
		log.Printf("Warning: could not discover services: %v", err)
		return
	}

	services, err := discovery.ParseServices([]byte(result.Stdout))
	if err != nil {
		log.Printf("Warning: could not parse service info: %v", err)
		return
	}

	// Register packs for all known pack types
	allPacks := []ToolPack{
		&RedisPack{},
		&SolrPack{},
		&MailhogPack{},
	}

	for _, pack := range allPacks {
		if !cfg.ToolPacks[pack.Name()] {
			continue
		}
		if !services.Has(pack.ServiceRequired()) {
			log.Printf("Tool-pack %q enabled but service %q not running, skipping", pack.Name(), pack.ServiceRequired())
			continue
		}
		pack.Register(server)
		log.Printf("Loaded tool-pack: %s", pack.Name())
	}
}
```

- [ ] **Step 5: Create example config**

Create `/Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-config.yaml.example`:

```yaml
# ddev-addon-mcp configuration
# Enable/disable tool-packs for additional services.
# Packs are only loaded if the service is actually running.
toolpacks:
  redis: false
  solr: false
  mailhog: false

# Auto-start MCP server with ddev start
autostart: true
```

- [ ] **Step 6: Run tests**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go test ./toolpacks/ -v
```

Expected: All 3 tests PASS

- [ ] **Step 7: Commit**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp
git add mcp-server/toolpacks/config.go mcp-server/toolpacks/registry.go mcp-server/toolpacks/registry_test.go mcp-config.yaml.example
git commit -m "feat: add tool-pack registry with config loading and service discovery"
```

---

### Task 11: Tool-Packs (Redis, Solr, Mailhog)

**Files:**
- Create: `mcp-server/toolpacks/redis.go`
- Create: `mcp-server/toolpacks/solr.go`
- Create: `mcp-server/toolpacks/mailhog.go`

- [ ] **Step 1: Implement Redis pack**

Create `mcp-server/toolpacks/redis.go`:

```go
package toolpacks

import (
	"context"
	"strings"
	"time"

	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/executor"
	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type RedisPack struct{}

func (p *RedisPack) Name() string            { return "redis" }
func (p *RedisPack) ServiceRequired() string  { return "redis" }

func (p *RedisPack) Register(server *mcp.Server) {
	type RedisCliInput struct {
		Command string `json:"command" jsonschema:"redis-cli command (e.g. 'GET key', 'KEYS *')"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ddev_redis_cli",
		Description: "Execute a redis-cli command in the DDEV Redis container.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input RedisCliInput) (*mcp.CallToolResult, any, error) {
		args := append([]string{"exec", "--service", "redis", "redis-cli"}, strings.Fields(input.Command)...)
		result, err := executor.ExecuteDdev(ctx, args, 30*time.Second)
		if err != nil {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
		}
		if result.ExitCode != 0 {
			return tools.ToolError(result), nil, nil
		}
		return tools.ToolSuccess(result), nil, nil
	})

	type RedisInfoInput struct{}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ddev_redis_info",
		Description: "Get Redis server info.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input RedisInfoInput) (*mcp.CallToolResult, any, error) {
		result, err := executor.ExecuteDdev(ctx, []string{"exec", "--service", "redis", "redis-cli", "INFO"}, 30*time.Second)
		if err != nil {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
		}
		if result.ExitCode != 0 {
			return tools.ToolError(result), nil, nil
		}
		return tools.ToolSuccess(result), nil, nil
	})

	type RedisFlushInput struct {
		Confirm bool `json:"confirm" jsonschema:"Must be true to execute. This flushes all Redis data."`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ddev_redis_flush",
		Description: "Flush all Redis data. Requires confirm: true.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input RedisFlushInput) (*mcp.CallToolResult, any, error) {
		if !input.Confirm {
			return tools.ConfirmationRequired("ddev_redis_flush", "This will delete all data in Redis."), nil, nil
		}
		result, err := executor.ExecuteDdev(ctx, []string{"exec", "--service", "redis", "redis-cli", "FLUSHALL"}, 30*time.Second)
		if err != nil {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
		}
		if result.ExitCode != 0 {
			return tools.ToolError(result), nil, nil
		}
		return tools.ToolSuccess(result), nil, nil
	})
}
```

- [ ] **Step 2: Implement Solr pack**

Create `mcp-server/toolpacks/solr.go`:

```go
package toolpacks

import (
	"context"
	"time"

	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/executor"
	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SolrPack struct{}

func (p *SolrPack) Name() string            { return "solr" }
func (p *SolrPack) ServiceRequired() string  { return "solr" }

func (p *SolrPack) Register(server *mcp.Server) {
	type SolrStatusInput struct{}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ddev_solr_status",
		Description: "Get Solr core status.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SolrStatusInput) (*mcp.CallToolResult, any, error) {
		result, err := executor.ExecuteDdev(ctx, []string{"exec", "--service", "solr", "curl", "-s", "http://localhost:8983/solr/admin/cores?action=STATUS"}, 30*time.Second)
		if err != nil {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
		}
		if result.ExitCode != 0 {
			return tools.ToolError(result), nil, nil
		}
		return tools.ToolSuccess(result), nil, nil
	})

	type SolrReloadInput struct {
		Core string `json:"core" jsonschema:"Name of the Solr core to reload"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ddev_solr_reload",
		Description: "Reload a Solr core.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SolrReloadInput) (*mcp.CallToolResult, any, error) {
		url := "http://localhost:8983/solr/admin/cores?action=RELOAD&core=" + input.Core
		result, err := executor.ExecuteDdev(ctx, []string{"exec", "--service", "solr", "curl", "-s", url}, 30*time.Second)
		if err != nil {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
		}
		if result.ExitCode != 0 {
			return tools.ToolError(result), nil, nil
		}
		return tools.ToolSuccess(result), nil, nil
	})

	type SolrQueryInput struct {
		Core  string `json:"core" jsonschema:"Solr core name"`
		Query string `json:"query" jsonschema:"Solr query string (e.g. '*:*')"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ddev_solr_query",
		Description: "Execute a Solr query.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SolrQueryInput) (*mcp.CallToolResult, any, error) {
		url := "http://localhost:8983/solr/" + input.Core + "/select?q=" + input.Query + "&wt=json"
		result, err := executor.ExecuteDdev(ctx, []string{"exec", "--service", "solr", "curl", "-s", url}, 30*time.Second)
		if err != nil {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
		}
		if result.ExitCode != 0 {
			return tools.ToolError(result), nil, nil
		}
		return tools.ToolSuccess(result), nil, nil
	})
}
```

- [ ] **Step 3: Implement Mailhog pack**

Create `mcp-server/toolpacks/mailhog.go`:

```go
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
func (p *MailhogPack) ServiceRequired() string  { return "mailhog" }

func (p *MailhogPack) Register(server *mcp.Server) {
	type MailhogListInput struct{}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ddev_mailhog_list",
		Description: "List captured emails in Mailhog.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input MailhogListInput) (*mcp.CallToolResult, any, error) {
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

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ddev_mailhog_search",
		Description: "Search captured emails by from, to, or containing text.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input MailhogSearchInput) (*mcp.CallToolResult, any, error) {
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

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ddev_mailhog_delete",
		Description: "Delete all captured emails in Mailhog. Requires confirm: true.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input MailhogDeleteInput) (*mcp.CallToolResult, any, error) {
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
```

- [ ] **Step 4: Verify build**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go build ./...
```

Expected: No errors

- [ ] **Step 5: Run all tests**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go test ./... -v
```

Expected: All tests PASS

- [ ] **Step 6: Commit**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp
git add mcp-server/toolpacks/redis.go mcp-server/toolpacks/solr.go mcp-server/toolpacks/mailhog.go
git commit -m "feat: add tool-packs for Redis, Solr, and Mailhog"
```

---

### Task 12: DDEV Add-on Files

**Files:**
- Create: `install.yaml`
- Create: `commands/host/mcp`

- [ ] **Step 1: Create install.yaml**

Create `/Users/olivier/Versioncontrol/github/ddev-addon-mcp/install.yaml`:

```yaml
name: mcp

pre_install_commands:
  - |
    mkdir -p .ddev/bin
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    case "$ARCH" in
      x86_64) ARCH="amd64" ;;
      aarch64|arm64) ARCH="arm64" ;;
    esac
    VERSION="latest"
    URL="https://github.com/dkd-dobberkau/ddev-addon-mcp/releases/${VERSION}/download/ddev-mcp-server_${OS}_${ARCH}"
    echo "Downloading ddev-mcp-server for ${OS}/${ARCH}..."
    curl -sSL "$URL" -o .ddev/bin/ddev-mcp-server
    chmod +x .ddev/bin/ddev-mcp-server

project_files:
  - commands/host/mcp
  - mcp-config.yaml.example

post_install_commands:
  - |
    if [ ! -f .ddev/mcp-config.yaml ]; then
      cp .ddev/mcp-config.yaml.example .ddev/mcp-config.yaml
    fi
    echo ""
    echo "✓ MCP add-on installed."
    echo ""
    echo "Configure your AI agent with:"
    echo '  "command": ".ddev/bin/ddev-mcp-server"'
    echo ""
    echo "Or use: ddev mcp serve"

removal_commands:
  - rm -f .ddev/bin/ddev-mcp-server
  - rm -f .ddev/commands/host/mcp
  - rm -f .ddev/mcp-config.yaml.example

ddev_version_constraint: ">= v1.24.0"
```

- [ ] **Step 2: Create custom command**

```bash
mkdir -p /Users/olivier/Versioncontrol/github/ddev-addon-mcp/commands/host
```

Create `/Users/olivier/Versioncontrol/github/ddev-addon-mcp/commands/host/mcp`:

```bash
#!/bin/bash

## Description: Manage the MCP server for AI agent integration
## Usage: mcp [command]
## Example: ddev mcp serve\nddev mcp status\nddev mcp config

case "${1:-status}" in
  serve)
    exec .ddev/bin/ddev-mcp-server
    ;;
  status)
    if [ -f .ddev/bin/ddev-mcp-server ]; then
      echo "MCP server: installed"
      .ddev/bin/ddev-mcp-server --version 2>/dev/null || true
      echo ""
      echo "Config: .ddev/mcp-config.yaml"
      if [ -f .ddev/mcp-config.yaml ]; then
        cat .ddev/mcp-config.yaml
      else
        echo "(no config found)"
      fi
    else
      echo "MCP server: not installed"
      echo "Run: ddev addon get ddev-addon-mcp"
    fi
    ;;
  config)
    ${EDITOR:-vi} .ddev/mcp-config.yaml
    ;;
  *)
    echo "Usage: ddev mcp [serve|status|config]"
    echo ""
    echo "Commands:"
    echo "  serve   Start MCP server (stdio mode)"
    echo "  status  Show MCP installation and config status"
    echo "  config  Edit MCP configuration"
    ;;
esac
```

- [ ] **Step 3: Make command executable**

```bash
chmod +x /Users/olivier/Versioncontrol/github/ddev-addon-mcp/commands/host/mcp
```

- [ ] **Step 4: Commit**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp
git add install.yaml commands/host/mcp
git commit -m "feat: add DDEV add-on install.yaml and ddev mcp command"
```

---

### Task 13: README

**Files:**
- Create: `README.md`

- [ ] **Step 1: Create README**

Create `/Users/olivier/Versioncontrol/github/ddev-addon-mcp/README.md`:

````markdown
# ddev-addon-mcp

DDEV add-on that provides an MCP (Model Context Protocol) server for AI agent integration. Enables AI agents to manage your DDEV projects via stdio.

## Installation

```bash
ddev addon get ddev-addon-mcp
```

## Usage

### Claude Code

Add to your project's `.mcp.json`:

```json
{
  "mcpServers": {
    "ddev": {
      "command": ".ddev/bin/ddev-mcp-server"
    }
  }
}
```

### Claude Desktop

Add to your Claude Desktop config:

```json
{
  "mcpServers": {
    "ddev": {
      "command": "/path/to/project/.ddev/bin/ddev-mcp-server"
    }
  }
}
```

### Manual

```bash
ddev mcp serve    # Start MCP server (stdio)
ddev mcp status   # Show installation status
ddev mcp config   # Edit configuration
```

## Core Tools

| Tool | Description |
|------|-------------|
| `ddev_start` | Start a DDEV project |
| `ddev_stop` | Stop a DDEV project |
| `ddev_restart` | Restart a DDEV project |
| `ddev_poweroff` | Stop ALL projects (requires `confirm: true`) |
| `ddev_list` | List all projects with status |
| `ddev_describe` | Detailed project info |
| `ddev_exec` | Run commands in containers |
| `ddev_composer` | Run composer commands |
| `ddev_import_db` | Import SQL dump (requires `confirm: true`) |
| `ddev_export_db` | Export database |
| `ddev_snapshot` | Create/restore snapshots |
| `ddev_logs` | View service logs |

## Tool-Packs

Optional tools for additional services. Edit `.ddev/mcp-config.yaml`:

```yaml
toolpacks:
  redis: true    # redis-cli, info, flush
  solr: true     # status, reload, query
  mailhog: true  # list, search, delete emails
```

Tool-packs only load if the service is actually running.

## Requirements

- DDEV >= v1.24.0

## License

MIT
````

- [ ] **Step 2: Commit**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp
git add README.md
git commit -m "docs: add README"
```

---

### Task 14: Final Verification

- [ ] **Step 1: Run all Go tests**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go test ./... -v
```

Expected: All tests PASS

- [ ] **Step 2: Build binary**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go build -o ddev-mcp-server .
./ddev-mcp-server --version
```

Expected: `ddev-mcp-server dev`

- [ ] **Step 3: Verify go vet passes**

```bash
cd /Users/olivier/Versioncontrol/github/ddev-addon-mcp/mcp-server
go vet ./...
```

Expected: No issues

- [ ] **Step 4: Verify file structure**

```bash
find /Users/olivier/Versioncontrol/github/ddev-addon-mcp -not -path '*/.git/*' -type f | sort
```

Expected files:
```
.gitignore
.goreleaser.yml
README.md
commands/host/mcp
install.yaml
mcp-config.yaml.example
mcp-server/discovery/discovery.go
mcp-server/discovery/discovery_test.go
mcp-server/executor/executor.go
mcp-server/executor/executor_test.go
mcp-server/go.mod
mcp-server/go.sum
mcp-server/main.go
mcp-server/toolpacks/config.go
mcp-server/toolpacks/mailhog.go
mcp-server/toolpacks/redis.go
mcp-server/toolpacks/registry.go
mcp-server/toolpacks/registry_test.go
mcp-server/toolpacks/solr.go
mcp-server/tools/composer.go
mcp-server/tools/database.go
mcp-server/tools/database_test.go
mcp-server/tools/exec.go
mcp-server/tools/exec_test.go
mcp-server/tools/info.go
mcp-server/tools/info_test.go
mcp-server/tools/lifecycle.go
mcp-server/tools/lifecycle_test.go
mcp-server/tools/logs.go
mcp-server/tools/response.go
mcp-server/tools/response_test.go
```

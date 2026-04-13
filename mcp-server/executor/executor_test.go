package executor

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestExecuteDdev_Success(t *testing.T) {
	fakeDdev := createFakeDdev(t, "#!/bin/sh\necho \"project1 running\"\nexit 0\n")
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

func TestExecuteDdev_Failure(t *testing.T) {
	fakeDdev := createFakeDdev(t, "#!/bin/sh\necho \"Error: project not found\" >&2\nexit 1\n")
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

func TestExecuteDdev_Timeout(t *testing.T) {
	fakeDdev := createFakeDdev(t, "#!/bin/sh\nsleep 10\n")
	t.Setenv("PATH", filepath.Dir(fakeDdev)+":"+os.Getenv("PATH"))

	_, err := ExecuteDdev(context.Background(), []string{"start"}, 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func createFakeDdev(t *testing.T, script string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "ddev")
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := exec.LookPath(path); err != nil {
		t.Fatal(err)
	}
	return path
}

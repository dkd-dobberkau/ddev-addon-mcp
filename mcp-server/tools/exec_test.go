package tools

import (
	"context"
	"testing"
)

func TestExec_BasicCommand(t *testing.T) {
	setupFakeDdev(t, "#!/bin/sh\necho \"PHP 8.2.0\"\nexit 0\n")
	result, _, err := Exec(context.Background(), nil, ExecInput{Command: "php -v"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}

func TestExec_WithService(t *testing.T) {
	setupFakeDdev(t, "#!/bin/sh\necho \"$@\"\nexit 0\n")
	result, _, err := Exec(context.Background(), nil, ExecInput{Command: "mysql -e 'SHOW TABLES'", Service: "db"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}

func TestComposer_BasicCommand(t *testing.T) {
	setupFakeDdev(t, "#!/bin/sh\necho \"Installing dependencies\"\nexit 0\n")
	result, _, err := Composer(context.Background(), nil, ComposerInput{Command: "install"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}

func TestLogs_Defaults(t *testing.T) {
	setupFakeDdev(t, "#!/bin/sh\necho \"web log line 1\"\nexit 0\n")
	result, _, err := Logs(context.Background(), nil, LogsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}

func TestLogs_WithOptions(t *testing.T) {
	setupFakeDdev(t, "#!/bin/sh\necho \"db logs\"\nexit 0\n")
	tail := 50
	result, _, err := Logs(context.Background(), nil, LogsInput{Service: "db", Tail: &tail})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}

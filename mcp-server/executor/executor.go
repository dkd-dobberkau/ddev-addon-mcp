package executor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Result struct {
	Command  string
	Stdout   string
	Stderr   string
	ExitCode int
}

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

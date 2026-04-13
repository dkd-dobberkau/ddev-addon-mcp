package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func TestList_ParsesJSON(t *testing.T) {
	ddevOutput := `{"raw":[{"name":"my-project","status":"running"}]}`
	setupFakeDdev(t, "#!/bin/sh\necho '"+ddevOutput+"'\nexit 0\n")
	result, _, err := List(context.Background(), nil, ListInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}

func TestList_FallbackOnInvalidJSON(t *testing.T) {
	setupFakeDdev(t, "#!/bin/sh\necho 'not valid json'\nexit 0\n")
	result, _, err := List(context.Background(), nil, ListInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success even with invalid JSON")
	}
}

func TestDescribe_ParsesJSON(t *testing.T) {
	raw := map[string]any{"raw": map[string]any{"name": "my-project", "status": "running"}}
	data, _ := json.Marshal(raw)
	setupFakeDdev(t, "#!/bin/sh\necho '"+string(data)+"'\nexit 0\n")
	result, _, err := Describe(context.Background(), nil, DescribeInput{Project: "my-project"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}

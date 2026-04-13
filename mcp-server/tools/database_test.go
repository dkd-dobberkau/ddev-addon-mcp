package tools

import (
	"context"
	"testing"
)

func TestImportDB_RequiresConfirm(t *testing.T) {
	result, _, err := ImportDB(context.Background(), nil, ImportDBInput{File: "/tmp/dump.sql", Confirm: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error when confirm is false")
	}
}

func TestImportDB_Confirmed(t *testing.T) {
	setupFakeDdev(t, "#!/bin/sh\necho \"Successfully imported\"\nexit 0\n")
	result, _, err := ImportDB(context.Background(), nil, ImportDBInput{File: "/tmp/dump.sql", Confirm: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success when confirmed")
	}
}

func TestExportDB_Success(t *testing.T) {
	setupFakeDdev(t, "#!/bin/sh\necho \"Exported\"\nexit 0\n")
	result, _, err := ExportDB(context.Background(), nil, ExportDBInput{File: "/tmp/export.sql.gz"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}

func TestSnapshot_Create(t *testing.T) {
	setupFakeDdev(t, "#!/bin/sh\necho \"Snapshot created\"\nexit 0\n")
	result, _, err := Snapshot(context.Background(), nil, SnapshotInput{Name: "before-migration"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}

func TestSnapshot_Restore(t *testing.T) {
	setupFakeDdev(t, "#!/bin/sh\necho \"Restored\"\nexit 0\n")
	action := "restore"
	result, _, err := Snapshot(context.Background(), nil, SnapshotInput{Action: &action})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}

package api

import (
	"os"
	"path/filepath"
	"testing"

	"warden-mcp/internal/domain"
	"warden-mcp/internal/mcp/contracts"
)

func TestUpdateDedupesWarnings(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, ".agent", "PLAN.md")
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	content := `---
plan_id: sample-plan
title: Sample Plan
version: 1.0
status: active
current_phase: PH01
can_finish: false
completed_tasks: 0
---

# Sample Plan

## Phase 1 — Setup
- [ ] PH01-T01 create repo
`
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan failed: %v", err)
	}
	app := New(root, nil)
	envelope := app.Update(filepath.Join(".agent", "PLAN.md"), contracts.UpdateTaskRequest{TaskID: "PH01-T01", Status: domain.TaskInProgress, ActorType: domain.ActorAgent})
	if !envelope.OK {
		t.Fatalf("expected successful update, got %+v", envelope)
	}
	if len(envelope.Warnings) != 1 {
		t.Fatalf("expected one deduped warning, got %+v", envelope.Warnings)
	}
}

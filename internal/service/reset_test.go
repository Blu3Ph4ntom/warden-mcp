package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"warden-mcp/internal/domain"
	"warden-mcp/internal/mcp/contracts"
)

func TestResetTaskReopensDoneTask(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, "PLAN.md")
	content := `---
plan_id: reset-plan
title: Reset Plan
version: 1.0.0
status: active
current_phase: PH01
can_finish: false
completed_tasks: 1
---

# Reset Plan

## Phase 1 — Design
- [x] PH01-T01 done work
- [ ] PH01-T02 pending work

## Phase 2 — Build
- [ ] PH02-T01 implement
- [ ] PH02-T02 verify
`
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan failed: %v", err)
	}
	data, _, err := ResetTask(planPath, contracts.ResetTaskRequest{PlanID: "reset-plan", TaskID: "PH01-T01", Status: domain.TaskInProgress, Reason: "reopen after failed verification"})
	if err != nil {
		t.Fatalf("expected reset success, got %v", err)
	}
	if data.Task.Status != domain.TaskInProgress {
		t.Fatalf("expected reopened task status, got %+v", data)
	}
	updated, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("read updated plan failed: %v", err)
	}
	if !strings.Contains(string(updated), "reset_reason: reopen after failed verification") {
		t.Fatalf("expected reset reason to be persisted, got %s", string(updated))
	}
}

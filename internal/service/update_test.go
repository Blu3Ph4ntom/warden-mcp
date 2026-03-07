package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Blu3Ph4ntom/warden-mcp/internal/domain"
	"github.com/Blu3Ph4ntom/warden-mcp/internal/mcp/contracts"
)

func TestUpdateTaskPersistsNotesAndEvidence(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, "PLAN.md")
	content := `---
plan_id: update-plan
title: Update Plan
version: 1.0.0
status: active
current_phase: PH01
can_finish: false
completed_tasks: 0
---

# Update Plan

## Phase 1 — Design
- [ ] PH01-T01 define scope
	- [ ] PH01-T02 review scope

	## Phase 2 — Build
	- [ ] PH02-T01 implement server
	- [ ] PH02-T02 verify server
`
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan failed: %v", err)
	}
	data, _, err := UpdateTask(planPath, contracts.UpdateTaskRequest{PlanID: "update-plan", TaskID: "PH01-T01", Status: domain.TaskInProgress, ActorType: domain.ActorAgent, Note: "picked up task", Reason: "top priority", Evidence: []string{"ref|run-log|captured CLI output"}})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if data.Task.Status != domain.TaskInProgress || len(data.Task.Notes) != 2 || len(data.Task.Evidence) != 1 {
		t.Fatalf("expected persisted task metadata, got %+v", data.Task)
	}
	updated, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("read updated plan failed: %v", err)
	}
	for _, fragment := range []string{"picked up task", "reason: top priority", "run-log"} {
		if !strings.Contains(string(updated), fragment) {
			t.Fatalf("expected %s in updated markdown, got %s", fragment, string(updated))
		}
	}
}

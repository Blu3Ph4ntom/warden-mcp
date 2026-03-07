package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Blu3Ph4ntom/warden-mcp/internal/mcp/contracts"
)

func TestEditPlanSupportsStructuralAndFieldOperations(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, "PLAN.md")
	content := `---
plan_id: edit-plan
title: Edit Plan
version: 1.0.0
status: active
current_phase: PH01
can_finish: false
completed_tasks: 0
---

# Edit Plan

## Phase 1 — Design
- [ ] PH01-T01 first task
- [ ] PH01-T02 second task

## Phase 2 — Build
- [ ] PH02-T01 implement
- [ ] PH02-T02 verify
`
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan failed: %v", err)
	}
	payload, _ := json.Marshal(map[string]any{"priority": "P0"})
	if _, _, err := EditPlan(planPath, contracts.EditPlanRequest{PlanID: "edit-plan", Operation: contracts.EditReprioritizeTask, TargetID: "PH01-T01", Payload: payload}); err != nil {
		t.Fatalf("reprioritize failed: %v", err)
	}
	payload, _ = json.Marshal(map[string]any{"phase_id": "PH02", "title": "integration check", "after_task_id": "PH02-T01", "required": false})
	if _, _, err := EditPlan(planPath, contracts.EditPlanRequest{PlanID: "edit-plan", Operation: contracts.EditAddTask, Payload: payload}); err != nil {
		t.Fatalf("add task failed: %v", err)
	}
	payload, _ = json.Marshal(map[string]any{"depends_on": "PH01-T01"})
	if _, _, err := EditPlan(planPath, contracts.EditPlanRequest{PlanID: "edit-plan", Operation: contracts.EditAddDependency, TargetID: "PH02-T01", Payload: payload}); err != nil {
		t.Fatalf("add dependency failed: %v", err)
	}
	updated, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("read updated plan failed: %v", err)
	}
	text := string(updated)
	for _, fragment := range []string{"PH01-T01 first task | priority:P0", "PH02-T03 integration check | required:false", "PH02-T01 implement | depends_on:PH01-T01"} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("expected %s in updated plan, got %s", fragment, text)
		}
	}
}

func TestEditPlanRejectsInvalidSelfDependency(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, "PLAN.md")
	content := `---
plan_id: edit-plan
title: Edit Plan
version: 1.0.0
status: active
current_phase: PH01
can_finish: false
completed_tasks: 0
---

# Edit Plan

## Phase 1 — Design
- [ ] PH01-T01 first task
- [ ] PH01-T02 second task

## Phase 2 — Build
- [ ] PH02-T01 implement
- [ ] PH02-T02 verify
`
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan failed: %v", err)
	}
	payload, _ := json.Marshal(map[string]any{"depends_on": "PH01-T01"})
	if _, _, err := EditPlan(planPath, contracts.EditPlanRequest{PlanID: "edit-plan", Operation: contracts.EditAddDependency, TargetID: "PH01-T01", Payload: payload}); err == nil {
		t.Fatal("expected self-dependency edit to fail")
	}
}

package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"warden-mcp/internal/mcp/contracts"
)

func TestReconcilePlanDryRunAndApply(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, "PLAN.md")
	active := `---
plan_id: reconcile-plan
title: Reconcile Plan
version: 1.0.0
status: active
current_phase: PH01
can_finish: false
completed_tasks: 0
---

# Reconcile Plan

## Phase 1 — Design
- [ ] PH01-T01 first task
- [ ] PH01-T02 second task

## Phase 2 — Build
- [ ] PH02-T01 implement
- [ ] PH02-T02 verify
`
	if err := os.WriteFile(planPath, []byte(active), 0o644); err != nil {
		t.Fatalf("write active plan failed: %v", err)
	}
	candidate := strings.ReplaceAll(active, "PH01-T01 first task", "PH01-T01 first task | priority:P0")
	dryRun, _, err := ReconcilePlan(planPath, contracts.ReconcilePlanRequest{PlanID: "reconcile-plan", MarkdownContent: candidate, Mode: contracts.ReconcileDryRun})
	if err != nil {
		t.Fatalf("dry run failed: %v", err)
	}
	if !dryRun.Reconciled || len(dryRun.ChangedIDs) == 0 || dryRun.ChangedIDs[0] != "PH01-T01" {
		t.Fatalf("unexpected dry-run response %+v", dryRun)
	}
	applied, _, err := ReconcilePlan(planPath, contracts.ReconcilePlanRequest{PlanID: "reconcile-plan", MarkdownContent: candidate, Mode: contracts.ReconcileApply})
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	updated, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("read updated plan failed: %v", err)
	}
	if !applied.Reconciled || !strings.Contains(string(updated), "priority:P0") {
		t.Fatalf("expected applied reconcile change, data=%+v content=%s", applied, string(updated))
	}
}

func TestReconcilePlanFlagsValidationErrorsWithoutWriting(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, "PLAN.md")
	active := `---
plan_id: reconcile-plan
title: Reconcile Plan
version: 1.0.0
status: active
current_phase: PH01
can_finish: false
completed_tasks: 0
---

# Reconcile Plan

## Phase 1 — Design
- [ ] PH01-T01 first task
- [ ] PH01-T02 second task

## Phase 2 — Build
- [ ] PH02-T01 implement
- [ ] PH02-T02 verify
`
	if err := os.WriteFile(planPath, []byte(active), 0o644); err != nil {
		t.Fatalf("write active plan failed: %v", err)
	}
	candidate := strings.Replace(active, "PH01-T01 first task", "PH01-T01    ", 1)
	data, _, err := ReconcilePlan(planPath, contracts.ReconcilePlanRequest{PlanID: "reconcile-plan", MarkdownContent: candidate, Mode: contracts.ReconcileApply})
	if err != nil {
		t.Fatalf("expected validation conflict response, got error %v", err)
	}
	if data.Reconciled || len(data.Conflicts) == 0 {
		t.Fatalf("expected reconcile conflicts for invalid candidate, got %+v", data)
	}
}

func TestReconcilePlanRejectsTopLevelStateDrift(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, "PLAN.md")
	active := `---
plan_id: reconcile-plan
title: Reconcile Plan
version: 1.0.0
status: active
current_phase: PH01
can_finish: false
completed_tasks: 0
---

# Reconcile Plan

## Phase 1 — Design
- [ ] PH01-T01 first task
- [ ] PH01-T02 second task

## Phase 2 — Build
- [ ] PH02-T01 implement
- [ ] PH02-T02 verify
`
	if err := os.WriteFile(planPath, []byte(active), 0o644); err != nil {
		t.Fatalf("write active plan failed: %v", err)
	}
	candidate := strings.Replace(active, "status: active", "status: completed", 1)
	data, _, err := ReconcilePlan(planPath, contracts.ReconcilePlanRequest{PlanID: "reconcile-plan", MarkdownContent: candidate, Mode: contracts.ReconcileApply})
	if err != nil {
		t.Fatalf("expected drift conflict response, got error %v", err)
	}
	if data.Reconciled || len(data.Conflicts) == 0 || data.Conflicts[0].Code != "PLAN_STATUS_DRIFT" {
		t.Fatalf("expected top-level state drift conflict, got %+v", data)
	}
}

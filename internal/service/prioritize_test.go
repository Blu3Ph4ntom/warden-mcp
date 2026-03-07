package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"warden-mcp/internal/domain"
	"warden-mcp/internal/mcp/contracts"
)

func TestPrioritizeTasksUpdatesMarkdownMetadata(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, "PLAN.md")
	content := `---
plan_id: prioritize-plan
title: Prioritize Plan
version: 1.0.0
status: active
current_phase: PH01
can_finish: false
completed_tasks: 0
---

# Prioritize Plan

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
	data, _, err := PrioritizeTasks(planPath, contracts.PrioritizeTasksRequest{PlanID: "prioritize-plan", Updates: []contracts.PriorityUpdate{{TaskID: "PH01-T01", Priority: domain.PriorityP0}, {TaskID: "PH02-T01", Priority: domain.PriorityP1}}})
	if err != nil {
		t.Fatalf("prioritize failed: %v", err)
	}
	if len(data.UpdatedTaskIDs) != 2 {
		t.Fatalf("expected 2 updated IDs, got %+v", data)
	}
	updated, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("read updated plan failed: %v", err)
	}
	text := string(updated)
	if !strings.Contains(text, "PH01-T01 first task | priority:P0") || !strings.Contains(text, "PH02-T01 implement | priority:P1") {
		t.Fatalf("expected priority metadata in markdown, got %s", text)
	}
}

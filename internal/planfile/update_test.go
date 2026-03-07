package planfile

import (
	"strings"
	"testing"

	"github.com/Blu3Ph4ntom/warden-mcp/internal/domain"
)

func TestUpdateTaskStatusContentRewritesCheckboxAndFrontmatter(t *testing.T) {
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
- [ ] PH01-T02 add tests
`
	updated, err := UpdateTaskStatusContent(content, "PH01-T01", domain.TaskInProgress)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	for _, fragment := range []string{"- [/] PH01-T01 create repo", "current_phase: PH01", "completed_tasks: 0", "can_finish: false"} {
		if !strings.Contains(updated, fragment) {
			t.Fatalf("expected updated content to contain %s, got %s", fragment, updated)
		}
	}
	updated, err = UpdateTaskStatusContent(updated, "PH01-T01", domain.TaskDone)
	if err != nil {
		t.Fatalf("second update failed: %v", err)
	}
	for _, fragment := range []string{"- [x] PH01-T01 create repo", "completed_tasks: 1"} {
		if !strings.Contains(updated, fragment) {
			t.Fatalf("expected done update to contain %s, got %s", fragment, updated)
		}
	}
}

func TestUpdateTaskStatusContentRejectsIllegalTransition(t *testing.T) {
	content := `---
plan_id: sample-plan
title: Sample Plan
version: 1.0
status: active
current_phase: PH01
can_finish: false
completed_tasks: 1
---

# Sample Plan

## Phase 1 — Setup
- [x] PH01-T01 create repo
`
	if _, err := UpdateTaskStatusContent(content, "PH01-T01", domain.TaskInProgress); err == nil {
		t.Fatal("expected illegal transition to fail")
	}
}

func TestUpdateTaskContentPersistsReasonNoteAndEvidence(t *testing.T) {
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
	updated, err := UpdateTaskContent(content, TaskUpdateMutation{TaskID: "PH01-T01", TargetStatus: domain.TaskInProgress, ActorType: domain.ActorAgent, Note: "started work", Reason: "picked from queue", Evidence: []domain.EvidenceItem{{Kind: "ref", Ref: "run-log", Summary: "initial start"}}})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	for _, fragment := range []string{"- [/] PH01-T01 create repo", "notes:", "evidence:", "picked from queue", "run-log"} {
		if !strings.Contains(updated, fragment) {
			t.Fatalf("expected updated content to contain %s, got %s", fragment, updated)
		}
	}
}

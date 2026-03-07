package planfile

import (
	"testing"
	"time"

	"warden-mcp/internal/domain"
)

func TestParseBuildsPlanFromMarkdownProjection(t *testing.T) {
	content := `---
plan_id: sample-plan
title: Sample Plan
version: 1.0
status: active
current_phase: PH02
---

# Sample Plan

## Phase 1 — Setup
- [x] PH01-T01 create repo
- [x] PH01-T02 add tests

## Phase 2 — Build
- [/] PH02-T01 start implementation
- [ ] PH02-T02 finish implementation
`
	plan, issues, err := Parse(content, time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if plan.PlanID != "sample-plan" || plan.Version != "1.0.0" {
		t.Fatalf("unexpected plan header: %+v", plan)
	}
	if len(plan.Phases) != 2 || plan.Phases[0].Status != domain.PhaseCompleted || plan.Phases[1].Status != domain.PhaseInProgress {
		t.Fatalf("unexpected phase rollups: %+v", plan.Phases)
	}
	if len(issues) == 0 || issues[0].Code != "PLAN_VERSION_NORMALIZED" {
		t.Fatalf("expected version normalization warning, got %+v", issues)
	}
	if plan.CanFinish {
		t.Fatal("expected plan with open tasks to deny finish")
	}
	if plan.Phases[1].Tasks[0].Status != domain.TaskInProgress {
		t.Fatalf("expected checkbox / to map to in_progress, got %+v", plan.Phases[1].Tasks[0])
	}
}

func TestParseRejectsMissingFrontmatter(t *testing.T) {
	if _, _, err := Parse("# missing", time.Now().UTC()); err == nil {
		t.Fatal("expected parse failure for missing frontmatter")
	}
}

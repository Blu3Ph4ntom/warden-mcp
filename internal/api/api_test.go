package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Blu3Ph4ntom/warden-mcp/internal/domain"
	"github.com/Blu3Ph4ntom/warden-mcp/internal/mcp/contracts"
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
	- [ ] PH01-T02 review repo

	## Phase 2 — Build
	- [ ] PH02-T01 implement server
	- [ ] PH02-T02 verify server
`
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan failed: %v", err)
	}
	app := New(root, nil)
	envelope := app.Update(filepath.Join(".agent", "PLAN.md"), contracts.UpdateTaskRequest{PlanID: "sample-plan", TaskID: "PH01-T01", Status: domain.TaskInProgress, ActorType: domain.ActorAgent})
	if !envelope.OK {
		t.Fatalf("expected successful update, got %+v", envelope)
	}
	if len(envelope.Warnings) != 1 {
		t.Fatalf("expected one deduped warning, got %+v", envelope.Warnings)
	}
}

func TestInitCreatesPlanAndListFindsIt(t *testing.T) {
	root := t.TempDir()
	app := New(root, nil)
	result := app.Init(contracts.InitPlanRequest{
		Title:   "Agent Delivery Plan",
		Version: "1.0",
		Phases: []contracts.InitPlanPhaseInput{
			{Title: "Design", Tasks: []contracts.InitPlanTaskInput{{Title: "define contracts"}, {Title: "review plan"}}},
			{Title: "Build", Tasks: []contracts.InitPlanTaskInput{{Title: "implement server"}, {Title: "run tests"}}},
		},
	})
	if !result.OK {
		t.Fatalf("expected init success, got %+v", result)
	}
	if _, err := os.Stat(filepath.Join(root, ".agent", "PLAN.md")); err != nil {
		t.Fatalf("expected plan file to exist: %v", err)
	}
	list := app.List(contracts.ListPlansRequest{})
	if !list.OK || len(list.Data.Plans) != 1 || list.Data.Plans[0].PlanID != "agent-delivery-plan" {
		t.Fatalf("unexpected list response %+v", list)
	}
}

func TestImportJSONCreatesNormalizedMarkdown(t *testing.T) {
	root := t.TempDir()
	app := New(root, nil)
	contentBytes, err := json.Marshal(domain.Plan{
		PlanID:  "json-import-plan",
		Title:   "JSON Import Plan",
		Version: "1.0.0",
		Phases: []domain.Phase{
			{Title: "Plan", Tasks: []domain.Task{{Title: "draft work", Required: true}, {Title: "review work", Required: true}}},
			{Title: "Ship", Tasks: []domain.Task{{Title: "release work", Required: true}, {Title: "verify work", Required: true}}},
		},
	})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	result := app.Import(contracts.ImportPlanRequest{Format: contracts.ImportJSON, Content: string(contentBytes), Mode: contracts.ImportCreate})
	if !result.OK {
		t.Fatalf("expected import success, got %+v", result)
	}
	status := app.Status(filepath.Join(".agent", "PLAN.md"), contracts.GetStatusRequest{PlanID: "json-import-plan", IncludeTasks: true})
	if !status.OK || status.Data.Plan.PlanID != "json-import-plan" || len(status.Data.Tasks) != 4 {
		t.Fatalf("unexpected status after import %+v", status)
	}
}

func TestImportRejectsCrossWorkspaceJSONMetadata(t *testing.T) {
	root := t.TempDir()
	app := New(root, nil)
	contentBytes, err := json.Marshal(domain.Plan{
		PlanID:        "foreign-plan",
		Title:         "Foreign Plan",
		Version:       "1.0.0",
		WorkspaceRoot: filepath.Join(t.TempDir(), "other-workspace"),
		PlanPath:      filepath.Join(t.TempDir(), "other-workspace", ".agent", "PLAN.md"),
		Phases: []domain.Phase{
			{Title: "Plan", Tasks: []domain.Task{{Title: "draft work", Required: true}}},
		},
	})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	result := app.Import(contracts.ImportPlanRequest{Format: contracts.ImportJSON, Content: string(contentBytes), Mode: contracts.ImportCreate})
	if result.OK || result.Error == nil || result.Error.Code != contracts.ErrWorkspacePlanMismatch {
		t.Fatalf("expected workspace mismatch rejection, got %+v", result)
	}
}

func TestStatusRejectsPlanIDMismatch(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, ".agent", "PLAN.md")
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	content := `---
plan_id: sample-plan
title: Sample Plan
version: 1.0.0
status: active
current_phase: PH01
---

# Sample Plan

## Phase 1 — Setup
- [ ] PH01-T01 create repo
`
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan failed: %v", err)
	}
	app := New(root, nil)
	envelope := app.Status(filepath.Join(".agent", "PLAN.md"), contracts.GetStatusRequest{PlanID: "other-plan"})
	if envelope.OK || envelope.Error == nil || envelope.Error.Code != contracts.ErrSyncConflict {
		t.Fatalf("expected sync conflict, got %+v", envelope)
	}
}

func TestArchiveMovesFinishedPlanIntoArchive(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, ".agent", "PLAN.md")
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	content := `---
plan_id: archive-plan
title: Archive Plan
version: 1.0.0
status: completed
current_phase: PH02
can_finish: true
completed_tasks: 4
---

# Archive Plan

## Phase 1 — Design
- [x] PH01-T01 define work
- [x] PH01-T02 review work

## Phase 2 — Build
- [x] PH02-T01 implement
- [x] PH02-T02 verify
`
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan failed: %v", err)
	}
	app := New(root, nil)
	result := app.Archive(contracts.ArchivePlanRequest{PlanID: "archive-plan", CreateFinalExport: true})
	if !result.OK || !result.Data.Archived {
		t.Fatalf("expected archive success, got %+v", result)
	}
	if _, err := os.Stat(planPath); !os.IsNotExist(err) {
		t.Fatalf("expected active plan to be removed, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, result.Data.ArchivePath)); err != nil {
		t.Fatalf("expected archived markdown to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, strings.TrimSuffix(result.Data.ArchivePath, ".md")+".json")); err != nil {
		t.Fatalf("expected archived json export to exist: %v", err)
	}
}

func TestStatusIgnoresUnsafeAbsoluteDefaultPlanPath(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, ".agent", "PLAN.md")
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	content := `---
plan_id: fallback-plan
title: Fallback Plan
version: 1.0.0
status: active
current_phase: PH01
---

# Fallback Plan

## Phase 1 — Setup
- [ ] PH01-T01 create repo
`
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan failed: %v", err)
	}
	base := t.TempDir()
	windir := filepath.Join(base, "Windows")
	t.Setenv("WINDIR", windir)
	app := New(root, nil)
	envelope := app.Status(filepath.Join(windir, "System32", ".agent", "PLAN.md"), contracts.GetStatusRequest{})
	if !envelope.OK || envelope.Data.Plan.PlanID != "fallback-plan" {
		t.Fatalf("expected status to use workspace default plan, got %+v", envelope)
	}
}

func TestHealthIgnoresUnsafeAbsoluteDefaultPlanPath(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, ".agent", "PLAN.md")
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	content := `---
plan_id: fallback-plan
title: Fallback Plan
version: 1.0.0
status: active
current_phase: PH01
---

# Fallback Plan

## Phase 1 — Setup
- [ ] PH01-T01 create repo
`
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan failed: %v", err)
	}
	base := t.TempDir()
	windir := filepath.Join(base, "Windows")
	t.Setenv("WINDIR", windir)
	app := New(root, nil)
	envelope := app.Health(filepath.Join(windir, "System32", ".agent", "PLAN.md"))
	if !envelope.OK || envelope.Data.PlanID != "fallback-plan" {
		t.Fatalf("expected health to use workspace default plan, got %+v", envelope)
	}
}

package service

import (
	"testing"
	"time"

	"github.com/Blu3Ph4ntom/warden-mcp/internal/domain"
	"github.com/Blu3Ph4ntom/warden-mcp/internal/mcp/contracts"
)

func TestGetNextTaskReturnsFirstOpenTaskInPhaseOrder(t *testing.T) {
	plan := samplePlan()
	data := GetNextTask(plan, contracts.GetNextTaskRequest{PlanID: plan.PlanID, RespectPhaseOrder: true, RespectDependencies: true})
	if data.NextTask == nil || data.NextTask.TaskID != "PH02-T01" {
		t.Fatalf("expected PH02-T01 as next task, got %+v", data)
	}
}

func TestRequestFinishDeniesWhenRequiredWorkRemains(t *testing.T) {
	plan := samplePlan()
	data := RequestFinish(plan, contracts.RequestFinishRequest{PlanID: plan.PlanID, ActorType: domain.ActorAgent})
	if data.CanFinish {
		t.Fatalf("expected finish denial, got %+v", data)
	}
	if data.RecommendedNextTaskID != "PH02-T01" {
		t.Fatalf("expected recommended next task PH02-T01, got %+v", data)
	}
	if len(data.IncompleteTaskIDs) != 2 {
		t.Fatalf("expected incomplete task IDs, got %+v", data.IncompleteTaskIDs)
	}
	assertHasBlockingCode(t, data.BlockingReasons, "PHASE_INCOMPLETE")
	assertHasBlockingCode(t, data.BlockingReasons, contracts.ErrFinishDenied)
}

func TestRequestFinishAllowsCompletedPlan(t *testing.T) {
	plan := samplePlan()
	plan.Phases[1].Tasks[0].Status = domain.TaskDone
	plan.Phases[1].Tasks[1].Status = domain.TaskDone
	plan.Phases[1].Status = domain.PhaseCompleted
	data := RequestFinish(plan, contracts.RequestFinishRequest{PlanID: plan.PlanID, ActorType: domain.ActorAgent})
	if !data.CanFinish {
		t.Fatalf("expected finish approval, got %+v", data)
	}
	if len(data.BlockingReasons) != 0 {
		t.Fatalf("expected no blocking reasons, got %+v", data.BlockingReasons)
	}
}

func TestOptionalTasksDoNotBlockNextTaskOrFinish(t *testing.T) {
	plan := samplePlan()
	plan.Phases[1].Tasks[0].Status = domain.TaskDone
	plan.Phases[1].Tasks[1].Required = false
	plan.Phases[1].Tasks[1].Status = domain.TaskNotStarted
	plan.Phases[1].Status = domain.PhaseCompleted
	next := GetNextTask(plan, contracts.GetNextTaskRequest{PlanID: plan.PlanID, RespectPhaseOrder: true, RespectDependencies: true})
	if next.NextTask != nil || next.Blocked {
		t.Fatalf("expected no blocking next task, got %+v", next)
	}
	finish := RequestFinish(plan, contracts.RequestFinishRequest{PlanID: plan.PlanID, ActorType: domain.ActorAgent})
	if !finish.CanFinish || len(finish.IncompleteTaskIDs) != 0 {
		t.Fatalf("expected finish approval with optional work remaining, got %+v", finish)
	}
}

func samplePlan() domain.Plan {
	now := time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC)
	return domain.Plan{
		PlanID:         "warden-plan",
		Title:          "Warden",
		Status:         domain.PlanActive,
		Version:        "1.0.0",
		CurrentPhaseID: "PH02",
		UpdatedAt:      now,
		Phases: []domain.Phase{
			{PhaseID: "PH01", Title: "Foundations", Status: domain.PhaseCompleted, Tasks: []domain.Task{{TaskID: "PH01-T01", PhaseID: "PH01", Title: "define scope", Status: domain.TaskDone, Priority: domain.PriorityP1, Required: true, UpdatedAt: now}, {TaskID: "PH01-T02", PhaseID: "PH01", Title: "define rules", Status: domain.TaskDone, Priority: domain.PriorityP1, Required: true, UpdatedAt: now}}},
			{PhaseID: "PH02", Title: "Build", Status: domain.PhaseInProgress, Tasks: []domain.Task{{TaskID: "PH02-T01", PhaseID: "PH02", Title: "implement service", Status: domain.TaskInProgress, Priority: domain.PriorityP1, Required: true, UpdatedAt: now}, {TaskID: "PH02-T02", PhaseID: "PH02", Title: "finish tests", Status: domain.TaskNotStarted, Priority: domain.PriorityP2, Required: true, DependsOn: []string{"PH02-T01"}, UpdatedAt: now}}},
		},
	}
}

func assertHasBlockingCode(t *testing.T, blocking []contracts.BlockingReason, code string) {
	t.Helper()
	for _, reason := range blocking {
		if reason.Code == code {
			return
		}
	}
	t.Fatalf("expected blocking code %s in %+v", code, blocking)
}

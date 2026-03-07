package domain

import (
	"testing"
	"time"
)

func TestPlanRequiresMultiplePhases(t *testing.T) {
	taskA := Task{TaskID: "PH01-T01", PhaseID: "PH01", Title: "One", Status: TaskNotStarted, Priority: PriorityP2}
	taskB := Task{TaskID: "PH01-T02", PhaseID: "PH01", Title: "Two", Status: TaskNotStarted, Priority: PriorityP2}
	plan := Plan{
		PlanID:  "warden-plan",
		Title:   "Test",
		Status:  PlanDraft,
		Version: "0.1.0",
		Phases:  []Phase{{PhaseID: "PH01", Title: "Phase 1", Status: PhaseNotStarted, Tasks: []Task{taskA, taskB}}},
	}

	issues := plan.Validate()
	assertHasIssueCode(t, issues, "PLAN_TOO_SHALLOW")
}

func TestPhaseRequiresMultipleTasks(t *testing.T) {
	plan := validPlan()
	plan.Phases[0].Tasks = plan.Phases[0].Tasks[:1]

	issues := plan.Validate()
	assertHasIssueCode(t, issues, "PHASE_TOO_SHALLOW")
}

func TestTerminalStatesMatchGovernedTaskStates(t *testing.T) {
	if !(Task{Status: TaskDone}).IsTerminal() {
		t.Fatal("done must be terminal")
	}
	if !(Task{Status: TaskWaived}).IsTerminal() {
		t.Fatal("waived must be terminal")
	}
	if (Task{Status: TaskBlocked}).IsTerminal() {
		t.Fatal("blocked must not be terminal")
	}
}

func TestTransitionRulesAreStrict(t *testing.T) {
	if !CanTransitionTask(TaskNotStarted, TaskInProgress) {
		t.Fatal("expected not_started -> in_progress to be allowed")
	}
	if CanTransitionTask(TaskDone, TaskInProgress) {
		t.Fatal("expected done -> in_progress to be denied")
	}
	if !CanTransitionPhase(PhaseInProgress, PhaseCompleted) {
		t.Fatal("expected in_progress -> completed to be allowed")
	}
	if CanTransitionPhase(PhaseCompleted, PhaseInProgress) {
		t.Fatal("expected completed -> in_progress to be denied")
	}
	if !CanTransitionPlan(PlanActive, PlanCompleted) {
		t.Fatal("expected active -> completed to be allowed")
	}
	if CanTransitionPlan(PlanArchived, PlanActive) {
		t.Fatal("expected archived -> active to be denied")
	}
}

func TestDependencyValidationCatchesMissingAndCyclicReferences(t *testing.T) {
	plan := validPlan()
	plan.Phases[0].Tasks[0].DependsOn = []string{"PH02-T01"}
	plan.Phases[1].Tasks[0].DependsOn = []string{"PH01-T01"}
	plan.Phases[0].Tasks[1].DependsOn = []string{"PH99-T99"}

	issues := plan.Validate()
	assertHasIssueCode(t, issues, "DEPENDENCY_CYCLE")
	assertHasIssueCode(t, issues, "DEPENDENCY_NOT_FOUND")
}

func TestIDsAndFinishRequestAreValidated(t *testing.T) {
	plan := validPlan()
	plan.PlanID = "BAD"
	plan.Version = "latest"
	plan.Phases[0].PhaseID = "phase-1"
	plan.Phases[0].Tasks[0].TaskID = "task-1"

	issues := plan.Validate()
	assertHasIssueCode(t, issues, "PLAN_ID_INVALID")
	assertHasIssueCode(t, issues, "PLAN_VERSION_INVALID")
	assertHasIssueCode(t, issues, "PHASE_ID_INVALID")
	assertHasIssueCode(t, issues, "TASK_ID_INVALID")

	request := FinishRequest{PlanID: "BAD", ActorType: "robot", RequestedAt: time.Time{}}
	requestIssues := request.Validate()
	assertHasIssueCode(t, requestIssues, "PLAN_ID_INVALID")
	assertHasIssueCode(t, requestIssues, "ACTOR_TYPE_INVALID")
	assertHasIssueCode(t, requestIssues, "REQUESTED_AT_MISSING")
}

func TestGovernanceRequestsAndVersioningAreValidated(t *testing.T) {
	now := time.Now().UTC()

	snapshot := VersionSnapshot{PlanID: "warden-plan", PlanVersion: "0.2.0", Revision: 2, SchemaVersion: "1", RecordedAt: now}
	if issues := snapshot.Validate(); len(issues) != 0 {
		t.Fatalf("expected valid version snapshot, got %+v", issues)
	}

	archive := ArchiveRecord{PlanID: "warden-plan", PlanVersion: "0.2.0", ArchivedAt: now}
	if issues := archive.Validate(); len(issues) != 0 {
		t.Fatalf("expected valid archive record, got %+v", issues)
	}

	reset := ResetTaskRequest{PlanID: "warden-plan", TaskID: "PH02-T01", TargetStatus: TaskInProgress, Reason: "reopen after failed validation", RequestedAt: now}
	if issues := reset.Validate(); len(issues) != 0 {
		t.Fatalf("expected valid reset request, got %+v", issues)
	}
	if !CanResetTask(TaskDone, TaskNotStarted) {
		t.Fatal("expected reset from done to not_started to be allowed through governed reset flow")
	}

	closure := TaskClosureRequest{PlanID: "warden-plan", TaskID: "PH02-T02", TargetStatus: TaskWaived, Reason: "dependency replaced", ActorType: ActorHuman, RequestedAt: now}
	if issues := closure.Validate(); len(issues) != 0 {
		t.Fatalf("expected valid closure request, got %+v", issues)
	}
	if !RequiresClosureReason(TaskCancelled) || !RequiresClosureReason(TaskWaived) {
		t.Fatal("cancelled and waived tasks must require explicit rationale")
	}
	if RequiresClosureReason(TaskDone) {
		t.Fatal("done must not be treated as a governance closure state requiring rationale")
	}
}

func validPlan() Plan {
	return Plan{
		PlanID:  "warden-plan",
		Title:   "Warden",
		Status:  PlanActive,
		Version: "0.1.0",
		Phases: []Phase{
			{
				PhaseID: "PH01",
				Title:   "Phase 1",
				Status:  PhaseInProgress,
				Tasks: []Task{
					{TaskID: "PH01-T01", PhaseID: "PH01", Title: "Define product", Status: TaskDone, Priority: PriorityP1},
					{TaskID: "PH01-T02", PhaseID: "PH01", Title: "Define scope", Status: TaskInProgress, Priority: PriorityP2},
				},
			},
			{
				PhaseID: "PH02",
				Title:   "Phase 2",
				Status:  PhaseNotStarted,
				Tasks: []Task{
					{TaskID: "PH02-T01", PhaseID: "PH02", Title: "Define entities", Status: TaskNotStarted, Priority: PriorityP1},
					{TaskID: "PH02-T02", PhaseID: "PH02", Title: "Define IDs", Status: TaskNotStarted, Priority: PriorityP2},
				},
			},
		},
	}
}

func assertHasIssueCode(t *testing.T, issues []ValidationIssue, code string) {
	t.Helper()
	for _, issue := range issues {
		if issue.Code == code {
			return
		}
	}
	t.Fatalf("expected issue code %s, got %+v", code, issues)
}

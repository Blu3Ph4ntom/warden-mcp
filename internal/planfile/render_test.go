package planfile

import (
	"strings"
	"testing"
	"time"

	"warden-mcp/internal/domain"
)

func TestRenderProducesParsableMarkdownProjection(t *testing.T) {
	plan := domain.Plan{
		PlanID:         "render-plan",
		Title:          "Render Plan",
		Status:         domain.PlanActive,
		Version:        "1.0.0",
		CurrentPhaseID: "PH01",
		CanFinish:      false,
		UpdatedAt:      time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC),
		Phases:         []domain.Phase{{PhaseID: "PH01", Title: "Setup", Status: domain.PhaseInProgress, Tasks: []domain.Task{{TaskID: "PH01-T01", PhaseID: "PH01", Title: "create repo", Status: domain.TaskDone, Priority: domain.PriorityP1, Required: true, UpdatedAt: time.Now().UTC()}, {TaskID: "PH01-T02", PhaseID: "PH01", Title: "run tests", Status: domain.TaskInProgress, Priority: domain.PriorityP2, DependsOn: []string{"PH01-T01"}, Required: false, UpdatedAt: time.Now().UTC()}}}, {PhaseID: "PH02", Title: "Ship", Status: domain.PhaseNotStarted, Tasks: []domain.Task{{TaskID: "PH02-T01", PhaseID: "PH02", Title: "release", Status: domain.TaskNotStarted, Priority: domain.PriorityP2, Required: true, UpdatedAt: time.Now().UTC()}}}},
	}
	content := Render(plan)
	for _, fragment := range []string{"plan_id: render-plan", "- [x] PH01-T01 create repo | priority:P1", "- [/] PH01-T02 run tests | depends_on:PH01-T01 | required:false", "- [ ] PH02-T01 release"} {
		if !strings.Contains(content, fragment) {
			t.Fatalf("expected rendered content to contain %s, got %s", fragment, content)
		}
	}
	parsed, _, err := Parse(content, time.Now().UTC())
	if err != nil {
		t.Fatalf("parse rendered content failed: %v", err)
	}
	if parsed.PlanID != "render-plan" || len(parsed.Phases) != 2 || parsed.TotalTasks() != 3 {
		t.Fatalf("unexpected parsed plan %+v", parsed)
	}
	if parsed.Phases[0].Tasks[0].Priority != domain.PriorityP1 || parsed.Phases[0].Tasks[1].Required || len(parsed.Phases[0].Tasks[1].DependsOn) != 1 {
		t.Fatalf("expected metadata round-trip, got %+v", parsed.Phases[0].Tasks)
	}
}

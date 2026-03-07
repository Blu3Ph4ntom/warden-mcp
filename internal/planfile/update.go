package planfile

import (
	"fmt"
	"os"
	"strings"
	"time"

	"warden-mcp/internal/domain"
)

type TaskUpdateMutation struct {
	TaskID       string
	TargetStatus domain.TaskStatus
	ActorType    domain.ActorType
	Note         string
	Reason       string
	Evidence     []domain.EvidenceItem
	Timestamp    time.Time
}

func UpdateTaskStatusFile(path, taskID string, target domain.TaskStatus) (domain.Plan, []domain.ValidationIssue, error) {
	return UpdateTaskFile(path, TaskUpdateMutation{TaskID: taskID, TargetStatus: target})
}

func UpdateTaskFile(path string, mutation TaskUpdateMutation) (domain.Plan, []domain.ValidationIssue, error) {
	info, err := os.Stat(path)
	if err != nil {
		return domain.Plan{}, nil, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Plan{}, nil, err
	}
	updated, err := UpdateTaskContent(string(content), mutation)
	if err != nil {
		return domain.Plan{}, nil, err
	}
	if err := os.WriteFile(path, []byte(updated), info.Mode()); err != nil {
		return domain.Plan{}, nil, err
	}
	return Parse(updated, time.Now().UTC())
}

func UpdateTaskStatusContent(content, taskID string, target domain.TaskStatus) (string, error) {
	return UpdateTaskContent(content, TaskUpdateMutation{TaskID: taskID, TargetStatus: target})
}

func UpdateTaskContent(content string, mutation TaskUpdateMutation) (string, error) {
	plan, _, err := Parse(content, mutationTime(mutation.Timestamp))
	if err != nil {
		return "", err
	}
	if err := applyTaskUpdate(&plan, mutation); err != nil {
		return "", err
	}
	return Render(plan), nil
}

func applyTaskUpdate(plan *domain.Plan, mutation TaskUpdateMutation) error {
	task, err := findTaskForMutation(plan, mutation.TaskID)
	if err != nil {
		return err
	}
	if mutation.TargetStatus == "" {
		return fmt.Errorf("target status is required")
	}
	if task.Status != mutation.TargetStatus && !domain.CanTransitionTask(task.Status, mutation.TargetStatus) {
		return fmt.Errorf("invalid task transition: %s -> %s", task.Status, mutation.TargetStatus)
	}
	now := mutationTime(mutation.Timestamp)
	task.Status = mutation.TargetStatus
	appendNote(task, normalizeActor(mutation.ActorType), mutation.Note, now)
	appendReasonNote(task, normalizeActor(mutation.ActorType), mutation.Reason, now, "reason")
	if len(mutation.Evidence) > 0 {
		task.Evidence = append(task.Evidence, mutation.Evidence...)
	}
	task.UpdatedAt = now
	normalizePlanAfterTaskMutation(plan, now)
	return nil
}

func nextCurrentPhaseID(plan domain.Plan) string {
	for _, phase := range plan.Phases {
		if phase.Status != domain.PhaseCompleted {
			return phase.PhaseID
		}
	}
	if len(plan.Phases) == 0 {
		return ""
	}
	return plan.Phases[len(plan.Phases)-1].PhaseID
}

func findTaskForMutation(plan *domain.Plan, taskID string) (*domain.Task, error) {
	var found *domain.Task
	for phaseIndex := range plan.Phases {
		for taskIndex := range plan.Phases[phaseIndex].Tasks {
			if plan.Phases[phaseIndex].Tasks[taskIndex].TaskID != taskID {
				continue
			}
			if found != nil {
				return nil, fmt.Errorf("task appears multiple times: %s", taskID)
			}
			found = &plan.Phases[phaseIndex].Tasks[taskIndex]
		}
	}
	if found == nil {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	return found, nil
}

func normalizePlanAfterTaskMutation(plan *domain.Plan, now time.Time) {
	for index := range plan.Phases {
		plan.Phases[index].Status = rollupPhaseStatus(plan.Phases[index])
	}
	plan.CanFinish = canFinish(*plan)
	plan.Status = rollupPlanStatus(*plan)
	plan.CurrentPhaseID = nextCurrentPhaseID(*plan)
	plan.UpdatedAt = now
}

func mutationTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value.UTC()
}

func normalizeActor(actor domain.ActorType) domain.ActorType {
	if actor.Valid() {
		return actor
	}
	return domain.ActorSystem
}

func appendNote(task *domain.Task, actor domain.ActorType, text string, createdAt time.Time) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	task.Notes = append(task.Notes, domain.Note{ActorType: actor, Text: text, CreatedAt: createdAt})
}

func appendReasonNote(task *domain.Task, actor domain.ActorType, reason string, createdAt time.Time, prefix string) {
	if reason == "" {
		return
	}
	appendNote(task, actor, prefix+": "+reason, createdAt)
}

func statusMarker(status domain.TaskStatus) (string, error) {
	switch status {
	case domain.TaskDone:
		return "x", nil
	case domain.TaskInProgress:
		return "/", nil
	case domain.TaskNotStarted:
		return " ", nil
	case domain.TaskBlocked:
		return "-", nil
	case domain.TaskCancelled:
		return "-", nil
	case domain.TaskWaived:
		return "-", nil
	default:
		return "", fmt.Errorf("unsupported task status: %s", status)
	}
}

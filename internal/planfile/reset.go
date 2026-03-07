package planfile

import (
	"fmt"
	"time"

	"warden-mcp/internal/domain"
)

type TaskResetMutation struct {
	TaskID       string
	TargetStatus domain.TaskStatus
	Reason       string
	ActorType    domain.ActorType
	Timestamp    time.Time
}

func ResetTaskStatusFile(path, taskID string, target domain.TaskStatus) (domain.Plan, []domain.ValidationIssue, error) {
	return ResetTaskFile(path, TaskResetMutation{TaskID: taskID, TargetStatus: target})
}

func ResetTaskFile(path string, mutation TaskResetMutation) (domain.Plan, []domain.ValidationIssue, error) {
	plan, warnings, err := Load(path)
	if err != nil {
		return domain.Plan{}, nil, err
	}
	updated, err := ResetTaskContent(Render(plan), mutation)
	if err != nil {
		return domain.Plan{}, warnings, err
	}
	parsed, writeWarnings, err := Parse(updated, time.Now().UTC())
	if err != nil {
		return domain.Plan{}, append(warnings, writeWarnings...), err
	}
	parsed, writeWarnings, err = WritePlanFile(path, parsed)
	return parsed, append(warnings, writeWarnings...), err
}

func ResetTaskStatusContent(content, taskID string, target domain.TaskStatus) (string, error) {
	return ResetTaskContent(content, TaskResetMutation{TaskID: taskID, TargetStatus: target})
}

func ResetTaskContent(content string, mutation TaskResetMutation) (string, error) {
	plan, _, err := Parse(content, mutationTime(mutation.Timestamp))
	if err != nil {
		return "", err
	}
	task, err := findTaskForMutation(&plan, mutation.TaskID)
	if err != nil {
		return "", err
	}
	if task.Status != domain.TaskDone && task.Status != domain.TaskCancelled && task.Status != domain.TaskWaived {
		return "", fmt.Errorf("task is not terminal and cannot be reset: %s", mutation.TaskID)
	}
	if !domain.CanResetTask(task.Status, mutation.TargetStatus) {
		return "", fmt.Errorf("invalid reset target: %s -> %s", task.Status, mutation.TargetStatus)
	}
	now := mutationTime(mutation.Timestamp)
	task.Status = mutation.TargetStatus
	appendReasonNote(task, normalizeActor(mutation.ActorType), mutation.Reason, now, "reset_reason")
	task.UpdatedAt = now
	normalizePlanAfterTaskMutation(&plan, now)
	return Render(plan), nil
}

package service

import (
	"fmt"

	"warden-mcp/internal/domain"
	"warden-mcp/internal/mcp/contracts"
	"warden-mcp/internal/planfile"
)

func UpdateTask(planPath string, req contracts.UpdateTaskRequest) (contracts.UpdateTaskData, []domain.ValidationIssue, error) {
	plan, warnings, err := planfile.UpdateTaskStatusFile(planPath, req.TaskID, req.Status)
	if err != nil {
		return contracts.UpdateTaskData{}, nil, err
	}
	task, phase, ok := findTaskAndPhase(plan, req.TaskID)
	if !ok {
		return contracts.UpdateTaskData{}, warnings, fmt.Errorf("task not found after update: %s", req.TaskID)
	}
	return contracts.UpdateTaskData{Task: summarizeTask(task), Phase: summarizePhase(phase), Plan: summarizePlan(plan, finishReady(plan)), TransitionAccepted: true}, warnings, nil
}

func findTaskAndPhase(plan domain.Plan, taskID string) (domain.Task, domain.Phase, bool) {
	for _, phase := range plan.Phases {
		for _, task := range phase.Tasks {
			if task.TaskID == taskID {
				return task, phase, true
			}
		}
	}
	return domain.Task{}, domain.Phase{}, false
}

func summarizePhase(phase domain.Phase) contracts.PhaseSummary {
	return contracts.PhaseSummary{PhaseID: phase.PhaseID, Title: phase.Title, Status: phase.Status, TaskCount: len(phase.Tasks), CompletedTaskCount: phase.CompletedTaskCount(), BlockedTaskCount: phase.BlockedTaskCount()}
}

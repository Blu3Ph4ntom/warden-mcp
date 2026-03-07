package service

import (
	"fmt"
	"time"

	"github.com/Blu3Ph4ntom/warden-mcp/internal/domain"
	"github.com/Blu3Ph4ntom/warden-mcp/internal/mcp/contracts"
	"github.com/Blu3Ph4ntom/warden-mcp/internal/planfile"
)

func ResetTask(planPath string, req contracts.ResetTaskRequest) (contracts.ResetTaskData, []domain.ValidationIssue, error) {
	target := req.Status
	if target == "" {
		target = domain.TaskNotStarted
	}
	resetReq := domain.ResetTaskRequest{PlanID: req.PlanID, TaskID: req.TaskID, TargetStatus: target, Reason: req.Reason, RequestedAt: time.Now().UTC()}
	if issues := resetReq.Validate(); len(issues) > 0 {
		return contracts.ResetTaskData{}, issues, fmt.Errorf("%s", issues[0].Message)
	}
	plan, warnings, err := planfile.ResetTaskFile(planPath, planfile.TaskResetMutation{TaskID: req.TaskID, TargetStatus: target, Reason: req.Reason, ActorType: domain.ActorSystem, Timestamp: time.Now().UTC()})
	if err != nil {
		return contracts.ResetTaskData{}, warnings, err
	}
	task, _, ok := findTaskAndPhase(plan, req.TaskID)
	if !ok {
		return contracts.ResetTaskData{}, warnings, fmt.Errorf("task not found after reset: %s", req.TaskID)
	}
	return contracts.ResetTaskData{Task: summarizeTask(task), Plan: summarizePlan(plan, finishReady(plan))}, warnings, nil
}

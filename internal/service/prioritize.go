package service

import (
	"fmt"
	"time"

	"github.com/Blu3Ph4ntom/warden-mcp/internal/domain"
	"github.com/Blu3Ph4ntom/warden-mcp/internal/mcp/contracts"
	"github.com/Blu3Ph4ntom/warden-mcp/internal/planfile"
)

func PrioritizeTasks(planPath string, req contracts.PrioritizeTasksRequest) (contracts.PrioritizeTasksData, []domain.ValidationIssue, error) {
	plan, warnings, err := planfile.Load(planPath)
	if err != nil {
		return contracts.PrioritizeTasksData{}, warnings, err
	}
	if len(req.Updates) == 0 {
		return contracts.PrioritizeTasksData{}, warnings, fmt.Errorf("prioritize requires at least one update")
	}
	updatedIDs := make([]string, 0, len(req.Updates))
	for _, update := range req.Updates {
		if !update.Priority.Valid() {
			return contracts.PrioritizeTasksData{}, warnings, fmt.Errorf("invalid priority for %s", update.TaskID)
		}
		found := false
		for phaseIndex := range plan.Phases {
			for taskIndex := range plan.Phases[phaseIndex].Tasks {
				if plan.Phases[phaseIndex].Tasks[taskIndex].TaskID != update.TaskID {
					continue
				}
				plan.Phases[phaseIndex].Tasks[taskIndex].Priority = update.Priority
				plan.Phases[phaseIndex].Tasks[taskIndex].UpdatedAt = time.Now().UTC()
				updatedIDs = append(updatedIDs, update.TaskID)
				found = true
			}
		}
		if !found {
			return contracts.PrioritizeTasksData{}, warnings, fmt.Errorf("task not found: %s", update.TaskID)
		}
	}
	plan.UpdatedAt = time.Now().UTC()
	plan, warnings, err = planfile.WritePlanFile(planPath, plan)
	if err != nil {
		return contracts.PrioritizeTasksData{}, warnings, err
	}
	return contracts.PrioritizeTasksData{UpdatedTaskIDs: updatedIDs, Plan: summarizePlan(plan, finishReady(plan))}, warnings, nil
}

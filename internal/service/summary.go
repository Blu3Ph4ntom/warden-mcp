package service

import (
	"warden-mcp/internal/domain"
	"warden-mcp/internal/mcp/contracts"
)

func PlanSummary(plan domain.Plan, canFinish bool) contracts.PlanSummary {
	return summarizePlan(plan, canFinish)
}

func PhaseSummaries(phases []domain.Phase) []contracts.PhaseSummary {
	return summarizePhases(phases)
}

func TaskSummaries(phases []domain.Phase, includeCompleted bool) []contracts.TaskSummary {
	return summarizeTasks(phases, includeCompleted)
}

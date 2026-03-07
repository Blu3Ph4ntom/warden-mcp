package service

import (
	"github.com/Blu3Ph4ntom/warden-mcp/internal/domain"
	"github.com/Blu3Ph4ntom/warden-mcp/internal/mcp/contracts"
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

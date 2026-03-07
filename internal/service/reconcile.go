package service

import (
	"reflect"
	"sort"
	"time"

	"github.com/Blu3Ph4ntom/warden-mcp/internal/domain"
	"github.com/Blu3Ph4ntom/warden-mcp/internal/mcp/contracts"
	"github.com/Blu3Ph4ntom/warden-mcp/internal/planfile"
)

func ReconcilePlan(planPath string, req contracts.ReconcilePlanRequest) (contracts.ReconcilePlanData, []domain.ValidationIssue, error) {
	active, warnings, err := planfile.Load(planPath)
	if err != nil {
		return contracts.ReconcilePlanData{}, warnings, err
	}
	candidate, parseWarnings, err := planfile.Parse(req.MarkdownContent, time.Now().UTC())
	warnings = append(warnings, parseWarnings...)
	if err != nil {
		return contracts.ReconcilePlanData{}, warnings, err
	}
	normalized := normalizeReconciledPlan(candidate)
	conflicts, changedIDs := reconcileDiff(active, candidate, req.PlanID)
	conflicts = append(conflicts, topLevelStateConflicts(candidate, normalized)...)
	conflicts = append(conflicts, validationConflicts(normalized)...)
	data := contracts.ReconcilePlanData{Reconciled: len(conflicts) == 0, Conflicts: conflicts, ChangedIDs: changedIDs, Plan: summarizePlan(active, finishReady(active))}
	if len(conflicts) > 0 || req.Mode == contracts.ReconcileDryRun || req.Mode == "" {
		return data, warnings, nil
	}
	updated, writeWarnings, err := planfile.WritePlanFile(planPath, normalized)
	warnings = append(warnings, writeWarnings...)
	if err != nil {
		return contracts.ReconcilePlanData{}, warnings, err
	}
	data.Plan = summarizePlan(updated, finishReady(updated))
	return data, warnings, nil
}

func reconcileDiff(active, candidate domain.Plan, requestedPlanID string) ([]contracts.Conflict, []string) {
	conflicts := make([]contracts.Conflict, 0)
	changedIDs := make([]string, 0)
	if requestedPlanID != "" && requestedPlanID != active.PlanID {
		conflicts = append(conflicts, contracts.Conflict{Code: "PLAN_ID_MISMATCH", Message: "requested plan_id does not match active plan", TargetID: requestedPlanID})
	}
	if candidate.PlanID != active.PlanID {
		conflicts = append(conflicts, contracts.Conflict{Code: "PLAN_ID_MISMATCH", Message: "candidate markdown plan_id does not match active plan", TargetID: candidate.PlanID})
	}
	activePhaseMap := make(map[string]domain.Phase, len(active.Phases))
	candidatePhaseMap := make(map[string]domain.Phase, len(candidate.Phases))
	for _, phase := range active.Phases {
		activePhaseMap[phase.PhaseID] = phase
	}
	for _, phase := range candidate.Phases {
		candidatePhaseMap[phase.PhaseID] = phase
	}
	for phaseID := range activePhaseMap {
		if _, ok := candidatePhaseMap[phaseID]; !ok {
			conflicts = append(conflicts, contracts.Conflict{Code: "PHASE_REMOVED", Message: "candidate markdown removed an existing phase", TargetID: phaseID})
		}
	}
	for phaseID := range candidatePhaseMap {
		if _, ok := activePhaseMap[phaseID]; !ok {
			conflicts = append(conflicts, contracts.Conflict{Code: "PHASE_ADDED", Message: "candidate markdown added a new phase", TargetID: phaseID})
		}
	}
	if len(conflicts) > 0 {
		return conflicts, changedIDs
	}
	for _, activePhase := range active.Phases {
		candidatePhase := candidatePhaseMap[activePhase.PhaseID]
		if candidatePhase.Title != activePhase.Title {
			changedIDs = append(changedIDs, activePhase.PhaseID)
		}
		activeTasks := tasksByID(activePhase.Tasks)
		candidateTasks := tasksByID(candidatePhase.Tasks)
		for taskID := range activeTasks {
			if _, ok := candidateTasks[taskID]; !ok {
				conflicts = append(conflicts, contracts.Conflict{Code: "TASK_REMOVED", Message: "candidate markdown removed an existing task", TargetID: taskID})
			}
		}
		for taskID := range candidateTasks {
			if _, ok := activeTasks[taskID]; !ok {
				conflicts = append(conflicts, contracts.Conflict{Code: "TASK_ADDED", Message: "candidate markdown added a new task", TargetID: taskID})
			}
		}
		if len(conflicts) > 0 {
			return conflicts, changedIDs
		}
		for taskID, activeTask := range activeTasks {
			candidateTask := candidateTasks[taskID]
			if taskMetadataDrifted(activeTask, candidateTask) {
				conflicts = append(conflicts, contracts.Conflict{Code: "TASK_METADATA_DRIFT", Message: "candidate markdown changed audit metadata that must be preserved", TargetID: taskID})
				continue
			}
			if taskChanged(activeTask, candidateTask) {
				changedIDs = append(changedIDs, taskID)
			}
		}
	}
	sort.Strings(changedIDs)
	return conflicts, dedupeStrings(changedIDs)
}

func tasksByID(tasks []domain.Task) map[string]domain.Task {
	result := make(map[string]domain.Task, len(tasks))
	for _, task := range tasks {
		result[task.TaskID] = task
	}
	return result
}

func taskChanged(active, candidate domain.Task) bool {
	return active.Title != candidate.Title || active.Status != candidate.Status || active.Priority != candidate.Priority || active.Required != candidate.Required || !reflect.DeepEqual(sortedStrings(active.DependsOn), sortedStrings(candidate.DependsOn))
}

func taskMetadataDrifted(active, candidate domain.Task) bool {
	return !reflect.DeepEqual(active.Notes, candidate.Notes) || !reflect.DeepEqual(active.Evidence, candidate.Evidence)
}

func sortedStrings(values []string) []string {
	clone := append([]string(nil), values...)
	sort.Strings(clone)
	return clone
}

func dedupeStrings(values []string) []string {
	if len(values) < 2 {
		return values
	}
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func validationConflicts(plan domain.Plan) []contracts.Conflict {
	conflicts := make([]contracts.Conflict, 0)
	for _, issue := range plan.Validate() {
		if issue.Severity != "error" {
			continue
		}
		conflicts = append(conflicts, contracts.Conflict{Code: issue.Code, Message: issue.Message, TargetID: issue.Path})
	}
	return conflicts
}

func topLevelStateConflicts(raw, normalized domain.Plan) []contracts.Conflict {
	conflicts := make([]contracts.Conflict, 0)
	if raw.Status != normalized.Status {
		conflicts = append(conflicts, contracts.Conflict{Code: "PLAN_STATUS_DRIFT", Message: "candidate markdown plan status does not match derived plan status", TargetID: string(raw.Status)})
	}
	if raw.CurrentPhaseID != normalized.CurrentPhaseID {
		conflicts = append(conflicts, contracts.Conflict{Code: "CURRENT_PHASE_DRIFT", Message: "candidate markdown current_phase does not match derived current phase", TargetID: raw.CurrentPhaseID})
	}
	return conflicts
}

func normalizeReconciledPlan(plan domain.Plan) domain.Plan {
	for index := range plan.Phases {
		plan.Phases[index].Status = planfile.RollupPhaseStatus(plan.Phases[index])
	}
	plan.Status = planfile.RollupPlanStatus(plan)
	plan.CanFinish = planfile.CanFinishPlan(plan)
	plan.CurrentPhaseID = nextReconcileCurrentPhaseID(plan)
	return plan
}

func nextReconcileCurrentPhaseID(plan domain.Plan) string {
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

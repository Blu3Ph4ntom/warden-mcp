package service

import (
	"fmt"
	"time"

	"warden-mcp/internal/domain"
	"warden-mcp/internal/mcp/contracts"
	"warden-mcp/internal/planfile"
)

func GetStatus(plan domain.Plan, includeTasks bool) contracts.GetStatusData {
	next := GetNextTask(plan, contracts.GetNextTaskRequest{PlanID: plan.PlanID, RespectPhaseOrder: true, RespectDependencies: true})
	data := contracts.GetStatusData{
		Plan:            summarizePlan(plan, finishReady(plan)),
		Phases:          summarizePhases(plan.Phases),
		BlockingReasons: statusBlockingReasons(plan),
		Stalled:         next.Blocked && next.NextTask == nil,
	}
	if next.NextTask != nil {
		data.NextTaskID = next.NextTask.TaskID
	}
	if includeTasks {
		data.Tasks = summarizeTasks(plan.Phases, true)
	}
	return data
}

func GetNextTask(plan domain.Plan, req contracts.GetNextTaskRequest) contracts.GetNextTaskData {
	for _, phase := range plan.Phases {
		candidates := make([]domain.Task, 0)
		blockers := make([]contracts.BlockingReason, 0)
		for _, task := range phase.Tasks {
			if !task.Required {
				continue
			}
			if task.IsTerminal() {
				continue
			}
			if task.Status == domain.TaskBlocked {
				blockers = append(blockers, contracts.BlockingReason{Code: "TASK_BLOCKED", Message: "task is blocked", PhaseID: phase.PhaseID, TaskID: task.TaskID})
				continue
			}
			if req.RespectDependencies {
				missing := unmetDependencies(plan, task)
				if len(missing) > 0 {
					blockers = append(blockers, contracts.BlockingReason{Code: contracts.ErrDependencyViolation, Message: fmt.Sprintf("task is waiting on %v", missing), PhaseID: phase.PhaseID, TaskID: task.TaskID})
					continue
				}
			}
			candidates = append(candidates, task)
		}
		if chosen, ok := chooseTask(candidates, req.PriorityBias); ok {
			summary := summarizeTask(chosen)
			return contracts.GetNextTaskData{NextTask: &summary, Reason: "selected next required task from phase order", Blocked: false}
		}
		if req.RespectPhaseOrder && (len(candidates) > 0 || len(blockers) > 0) {
			return contracts.GetNextTaskData{Reason: "current phase has blocked work", Blocked: true, BlockingReasons: blockers}
		}
	}
	return contracts.GetNextTaskData{Reason: "no remaining required non-terminal tasks", Blocked: false}
}

func RequestFinish(plan domain.Plan, req contracts.RequestFinishRequest) contracts.RequestFinishData {
	blocking := make([]contracts.BlockingReason, 0)
	incomplete := make([]string, 0)
	for _, issue := range plan.Validate() {
		if issue.Severity == "error" {
			blocking = append(blocking, contracts.BlockingReason{Code: contracts.ErrPlanInvalid, Message: issue.Message})
		}
	}
	for _, phase := range plan.Phases {
		if planfile.RollupPhaseStatus(phase) != domain.PhaseCompleted {
			blocking = append(blocking, contracts.BlockingReason{Code: "PHASE_INCOMPLETE", Message: "phase is not complete", PhaseID: phase.PhaseID})
		}
		for _, task := range phase.Tasks {
			if task.Required && !task.IsTerminal() {
				incomplete = append(incomplete, task.TaskID)
			}
			if task.Status == domain.TaskBlocked && task.Required {
				blocking = append(blocking, contracts.BlockingReason{Code: "TASK_BLOCKED", Message: "required task is blocked", PhaseID: phase.PhaseID, TaskID: task.TaskID})
			}
		}
	}
	next := GetNextTask(plan, contracts.GetNextTaskRequest{PlanID: req.PlanID, RespectPhaseOrder: true, RespectDependencies: true})
	actions := make([]string, 0)
	if next.NextTask != nil {
		actions = append(actions, "Continue "+next.NextTask.TaskID)
	}
	if len(blocking) > 0 && len(actions) == 0 {
		actions = append(actions, "Resolve blocking reasons")
	}
	canFinish := len(blocking) == 0 && len(incomplete) == 0
	if !canFinish {
		blocking = append([]contracts.BlockingReason{{Code: contracts.ErrFinishDenied, Message: "finish denied until required work is complete"}}, blocking...)
	}
	data := contracts.RequestFinishData{
		CanFinish:           canFinish,
		Plan:                summarizePlan(plan, canFinish),
		BlockingReasons:     blocking,
		IncompleteTaskIDs:   incomplete,
		NextRequiredActions: actions,
	}
	if next.NextTask != nil {
		data.RecommendedNextTaskID = next.NextTask.TaskID
	}
	return data
}

func finishReady(plan domain.Plan) bool {
	return RequestFinish(plan, contracts.RequestFinishRequest{PlanID: plan.PlanID}).CanFinish
}

func statusBlockingReasons(plan domain.Plan) []contracts.BlockingReason {
	blocking := make([]contracts.BlockingReason, 0)
	for _, phase := range plan.Phases {
		for _, task := range phase.Tasks {
			if task.Status == domain.TaskBlocked {
				blocking = append(blocking, contracts.BlockingReason{Code: "TASK_BLOCKED", Message: "task is blocked", PhaseID: phase.PhaseID, TaskID: task.TaskID})
			}
		}
	}
	return blocking
}

func chooseTask(tasks []domain.Task, bias domain.Priority) (domain.Task, bool) {
	for _, task := range tasks {
		if bias != "" && task.Priority == bias {
			return task, true
		}
	}
	if len(tasks) == 0 {
		return domain.Task{}, false
	}
	return tasks[0], true
}

func unmetDependencies(plan domain.Plan, task domain.Task) []string {
	missing := make([]string, 0)
	for _, dependencyID := range task.DependsOn {
		dependency, ok := findTask(plan, dependencyID)
		if !ok || !dependency.IsTerminal() {
			missing = append(missing, dependencyID)
		}
	}
	return missing
}

func findTask(plan domain.Plan, taskID string) (domain.Task, bool) {
	for _, phase := range plan.Phases {
		for _, task := range phase.Tasks {
			if task.TaskID == taskID {
				return task, true
			}
		}
	}
	return domain.Task{}, false
}

func summarizePlan(plan domain.Plan, canFinish bool) contracts.PlanSummary {
	return contracts.PlanSummary{PlanID: plan.PlanID, Title: plan.Title, Status: plan.Status, Version: plan.Version, CurrentPhaseID: plan.CurrentPhaseID, TotalTasks: plan.TotalTasks(), CompletedTasks: plan.CompletedTaskCount(), CanFinish: canFinish, UpdatedAt: formatTime(plan.UpdatedAt)}
}

func summarizePhases(phases []domain.Phase) []contracts.PhaseSummary {
	out := make([]contracts.PhaseSummary, 0, len(phases))
	for _, phase := range phases {
		out = append(out, contracts.PhaseSummary{PhaseID: phase.PhaseID, Title: phase.Title, Status: phase.Status, TaskCount: len(phase.Tasks), CompletedTaskCount: phase.CompletedTaskCount(), BlockedTaskCount: phase.BlockedTaskCount()})
	}
	return out
}

func summarizeTasks(phases []domain.Phase, includeCompleted bool) []contracts.TaskSummary {
	out := make([]contracts.TaskSummary, 0)
	for _, phase := range phases {
		for _, task := range phase.Tasks {
			if !includeCompleted && task.IsTerminal() {
				continue
			}
			out = append(out, summarizeTask(task))
		}
	}
	return out
}

func summarizeTask(task domain.Task) contracts.TaskSummary {
	return contracts.TaskSummary{TaskID: task.TaskID, PhaseID: task.PhaseID, Title: task.Title, Status: task.Status, Priority: task.Priority, DependsOn: append([]string(nil), task.DependsOn...), Required: task.Required, Evidence: append([]domain.EvidenceItem(nil), task.Evidence...), Notes: append([]domain.Note(nil), task.Notes...), UpdatedAt: formatTime(task.UpdatedAt)}
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

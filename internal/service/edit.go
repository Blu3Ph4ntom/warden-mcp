package service

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"warden-mcp/internal/domain"
	"warden-mcp/internal/mcp/contracts"
	"warden-mcp/internal/planfile"
)

func EditPlan(planPath string, req contracts.EditPlanRequest) (contracts.EditPlanData, []domain.ValidationIssue, error) {
	plan, warnings, err := planfile.Load(planPath)
	if err != nil {
		return contracts.EditPlanData{}, warnings, err
	}
	changedIDs, diffSummary, err := applyEdit(&plan, req)
	if err != nil {
		return contracts.EditPlanData{}, warnings, err
	}
	plan = normalizeEditedPlan(plan)
	if issues := plan.Validate(); hasErrorValidationIssues(issues) {
		return contracts.EditPlanData{}, issues, fmt.Errorf("%s", firstValidationErrorIssue(issues).Message)
	}
	plan, warnings, err = planfile.WritePlanFile(planPath, plan)
	if err != nil {
		return contracts.EditPlanData{}, warnings, err
	}
	return contracts.EditPlanData{Plan: summarizePlan(plan, finishReady(plan)), ChangedIDs: changedIDs, DiffSummary: diffSummary}, warnings, nil
}

func applyEdit(plan *domain.Plan, req contracts.EditPlanRequest) ([]string, string, error) {
	switch req.Operation {
	case contracts.EditAddPhase:
		var payload addPhasePayload
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			return nil, "", err
		}
		return addPhase(plan, payload)
	case contracts.EditAddTask:
		var payload addTaskPayload
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			return nil, "", err
		}
		return addTask(plan, req.TargetID, payload)
	case contracts.EditUpdateTaskFields:
		var payload updateTaskFieldsPayload
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			return nil, "", err
		}
		return updateTaskFields(plan, req.TargetID, payload)
	case contracts.EditMoveTask:
		var payload moveTaskPayload
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			return nil, "", err
		}
		return moveTask(plan, req.TargetID, payload)
	case contracts.EditSplitTask:
		var payload splitTaskPayload
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			return nil, "", err
		}
		return splitTask(plan, req.TargetID, payload.Tasks)
	case contracts.EditReprioritizeTask:
		var payload reprioritizePayload
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			return nil, "", err
		}
		return reprioritizeTask(plan, req.TargetID, payload.Priority)
	case contracts.EditAddDependency:
		var payload dependencyPayload
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			return nil, "", err
		}
		return addDependency(plan, req.TargetID, payload.DependsOn)
	case contracts.EditRemoveDependency:
		var payload dependencyPayload
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			return nil, "", err
		}
		return removeDependency(plan, req.TargetID, payload.DependsOn)
	case contracts.EditWaiveTask:
		return closeTask(plan, req.TargetID, domain.TaskWaived, req.Reason)
	case contracts.EditCancelTask:
		return closeTask(plan, req.TargetID, domain.TaskCancelled, req.Reason)
	default:
		return nil, "", fmt.Errorf("unsupported edit operation: %s", req.Operation)
	}
}

type initTaskPayload struct {
	Title     string          `json:"title"`
	Priority  domain.Priority `json:"priority,omitempty"`
	DependsOn []string        `json:"depends_on,omitempty"`
	Required  *bool           `json:"required,omitempty"`
}
type addPhasePayload struct {
	Title        string            `json:"title"`
	AfterPhaseID string            `json:"after_phase_id,omitempty"`
	Tasks        []initTaskPayload `json:"tasks,omitempty"`
}
type addTaskPayload struct {
	PhaseID     string          `json:"phase_id,omitempty"`
	Title       string          `json:"title"`
	Priority    domain.Priority `json:"priority,omitempty"`
	DependsOn   []string        `json:"depends_on,omitempty"`
	Required    *bool           `json:"required,omitempty"`
	AfterTaskID string          `json:"after_task_id,omitempty"`
}
type updateTaskFieldsPayload struct {
	Title    *string         `json:"title,omitempty"`
	Priority domain.Priority `json:"priority,omitempty"`
	Required *bool           `json:"required,omitempty"`
}
type moveTaskPayload struct {
	PhaseID     string `json:"phase_id,omitempty"`
	AfterTaskID string `json:"after_task_id,omitempty"`
}
type splitTaskPayload struct {
	Tasks []initTaskPayload `json:"tasks"`
}
type reprioritizePayload struct {
	Priority domain.Priority `json:"priority"`
}
type dependencyPayload struct {
	DependsOn string `json:"depends_on"`
}

func addPhase(plan *domain.Plan, payload addPhasePayload) ([]string, string, error) {
	if strings.TrimSpace(payload.Title) == "" {
		return nil, "", fmt.Errorf("phase title is required")
	}
	phaseID := nextPhaseID(*plan)
	phase := domain.Phase{PhaseID: phaseID, Title: strings.TrimSpace(payload.Title), Tasks: make([]domain.Task, 0, len(payload.Tasks))}
	for _, item := range payload.Tasks {
		phase.Tasks = append(phase.Tasks, buildTask(phaseID, nextTaskIDForPhase(*plan, phaseID, phase.Tasks), item))
	}
	index := len(plan.Phases)
	if payload.AfterPhaseID != "" {
		phaseIndex, ok := findPhaseIndex(*plan, payload.AfterPhaseID)
		if !ok {
			return nil, "", fmt.Errorf("phase not found: %s", payload.AfterPhaseID)
		}
		index = phaseIndex + 1
	}
	plan.Phases = slices.Insert(plan.Phases, index, phase)
	return []string{phaseID}, "added phase " + phaseID, nil
}

func addTask(plan *domain.Plan, targetID string, payload addTaskPayload) ([]string, string, error) {
	phaseID := payload.PhaseID
	if phaseID == "" {
		phaseID = targetID
	}
	if strings.TrimSpace(payload.Title) == "" {
		return nil, "", fmt.Errorf("task title is required")
	}
	phaseIndex, ok := findPhaseIndex(*plan, phaseID)
	if !ok {
		return nil, "", fmt.Errorf("phase not found: %s", phaseID)
	}
	taskID := nextTaskIDForPhase(*plan, phaseID, nil)
	task := buildTask(phaseID, taskID, initTaskPayload{Title: payload.Title, Priority: payload.Priority, DependsOn: payload.DependsOn, Required: payload.Required})
	insertIndex := len(plan.Phases[phaseIndex].Tasks)
	if payload.AfterTaskID != "" {
		_, taskIndex, ok := findTaskLocation(*plan, payload.AfterTaskID)
		if !ok {
			return nil, "", fmt.Errorf("task not found: %s", payload.AfterTaskID)
		}
		insertIndex = taskIndex + 1
	}
	plan.Phases[phaseIndex].Tasks = slices.Insert(plan.Phases[phaseIndex].Tasks, insertIndex, task)
	return []string{taskID}, "added task " + taskID, nil
}

func updateTaskFields(plan *domain.Plan, taskID string, payload updateTaskFieldsPayload) ([]string, string, error) {
	phaseIndex, taskIndex, ok := findTaskLocation(*plan, taskID)
	if !ok {
		return nil, "", fmt.Errorf("task not found: %s", taskID)
	}
	task := &plan.Phases[phaseIndex].Tasks[taskIndex]
	if payload.Title != nil {
		task.Title = strings.TrimSpace(*payload.Title)
		if task.Title == "" {
			return nil, "", fmt.Errorf("task title is required")
		}
	}
	if payload.Priority != "" {
		if !payload.Priority.Valid() {
			return nil, "", fmt.Errorf("invalid priority: %s", payload.Priority)
		}
		task.Priority = payload.Priority
	}
	if payload.Required != nil {
		task.Required = *payload.Required
	}
	task.UpdatedAt = time.Now().UTC()
	return []string{taskID}, "updated task fields for " + taskID, nil
}

func moveTask(plan *domain.Plan, taskID string, payload moveTaskPayload) ([]string, string, error) {
	fromPhaseIndex, fromTaskIndex, ok := findTaskLocation(*plan, taskID)
	if !ok {
		return nil, "", fmt.Errorf("task not found: %s", taskID)
	}
	task := plan.Phases[fromPhaseIndex].Tasks[fromTaskIndex]
	plan.Phases[fromPhaseIndex].Tasks = append(plan.Phases[fromPhaseIndex].Tasks[:fromTaskIndex], plan.Phases[fromPhaseIndex].Tasks[fromTaskIndex+1:]...)
	toPhaseID := payload.PhaseID
	if toPhaseID == "" {
		toPhaseID = task.PhaseID
	}
	toPhaseIndex, ok := findPhaseIndex(*plan, toPhaseID)
	if !ok {
		return nil, "", fmt.Errorf("phase not found: %s", toPhaseID)
	}
	changedIDs := []string{taskID}
	if toPhaseID != task.PhaseID {
		oldID := task.TaskID
		task.PhaseID = toPhaseID
		task.TaskID = nextTaskIDForPhase(*plan, toPhaseID, nil)
		updateDependencyReferences(plan, oldID, task.TaskID)
		changedIDs = append(changedIDs, task.TaskID)
	}
	insertIndex := len(plan.Phases[toPhaseIndex].Tasks)
	if payload.AfterTaskID != "" {
		afterPhaseIndex, afterTaskIndex, ok := findTaskLocation(*plan, payload.AfterTaskID)
		if !ok {
			return nil, "", fmt.Errorf("task not found: %s", payload.AfterTaskID)
		}
		if afterPhaseIndex != toPhaseIndex {
			return nil, "", fmt.Errorf("after_task_id must be in destination phase")
		}
		insertIndex = afterTaskIndex + 1
	}
	task.UpdatedAt = time.Now().UTC()
	plan.Phases[toPhaseIndex].Tasks = slices.Insert(plan.Phases[toPhaseIndex].Tasks, insertIndex, task)
	return changedIDs, "moved task " + taskID, nil
}

func splitTask(plan *domain.Plan, taskID string, payload []initTaskPayload) ([]string, string, error) {
	if len(payload) < 2 {
		return nil, "", fmt.Errorf("split_task requires at least two tasks")
	}
	phaseIndex, taskIndex, ok := findTaskLocation(*plan, taskID)
	if !ok {
		return nil, "", fmt.Errorf("task not found: %s", taskID)
	}
	phaseID := plan.Phases[phaseIndex].PhaseID
	changedIDs := []string{taskID}
	plan.Phases[phaseIndex].Tasks[taskIndex].Title = strings.TrimSpace(payload[0].Title)
	if payload[0].Priority != "" {
		plan.Phases[phaseIndex].Tasks[taskIndex].Priority = payload[0].Priority
	}
	if payload[0].Required != nil {
		plan.Phases[phaseIndex].Tasks[taskIndex].Required = *payload[0].Required
	}
	if payload[0].DependsOn != nil {
		plan.Phases[phaseIndex].Tasks[taskIndex].DependsOn = append([]string(nil), payload[0].DependsOn...)
	}
	insertIndex := taskIndex + 1
	for _, item := range payload[1:] {
		newTaskID := nextTaskIDForPhase(*plan, phaseID, nil)
		newTask := buildTask(phaseID, newTaskID, item)
		plan.Phases[phaseIndex].Tasks = slices.Insert(plan.Phases[phaseIndex].Tasks, insertIndex, newTask)
		insertIndex++
		changedIDs = append(changedIDs, newTaskID)
	}
	return changedIDs, "split task " + taskID, nil
}

func reprioritizeTask(plan *domain.Plan, taskID string, priority domain.Priority) ([]string, string, error) {
	return updateTaskFields(plan, taskID, updateTaskFieldsPayload{Priority: priority})
}

func addDependency(plan *domain.Plan, taskID, dependsOn string) ([]string, string, error) {
	phaseIndex, taskIndex, ok := findTaskLocation(*plan, taskID)
	if !ok {
		return nil, "", fmt.Errorf("task not found: %s", taskID)
	}
	if _, _, ok := findTaskLocation(*plan, dependsOn); !ok {
		return nil, "", fmt.Errorf("dependency task not found: %s", dependsOn)
	}
	if dependsOn == taskID {
		return nil, "", fmt.Errorf("task cannot depend on itself: %s", taskID)
	}
	task := &plan.Phases[phaseIndex].Tasks[taskIndex]
	if !slices.Contains(task.DependsOn, dependsOn) {
		task.DependsOn = append(task.DependsOn, dependsOn)
	}
	task.UpdatedAt = time.Now().UTC()
	return []string{taskID}, "added dependency to " + taskID, nil
}

func removeDependency(plan *domain.Plan, taskID, dependsOn string) ([]string, string, error) {
	phaseIndex, taskIndex, ok := findTaskLocation(*plan, taskID)
	if !ok {
		return nil, "", fmt.Errorf("task not found: %s", taskID)
	}
	task := &plan.Phases[phaseIndex].Tasks[taskIndex]
	filtered := make([]string, 0, len(task.DependsOn))
	for _, value := range task.DependsOn {
		if value != dependsOn {
			filtered = append(filtered, value)
		}
	}
	task.DependsOn = filtered
	task.UpdatedAt = time.Now().UTC()
	return []string{taskID}, "removed dependency from " + taskID, nil
}

func closeTask(plan *domain.Plan, taskID string, status domain.TaskStatus, reason string) ([]string, string, error) {
	if strings.TrimSpace(reason) == "" {
		return nil, "", fmt.Errorf("closure reason is required")
	}
	phaseIndex, taskIndex, ok := findTaskLocation(*plan, taskID)
	if !ok {
		return nil, "", fmt.Errorf("task not found: %s", taskID)
	}
	task := &plan.Phases[phaseIndex].Tasks[taskIndex]
	if !domain.RequiresClosureReason(status) {
		return nil, "", fmt.Errorf("invalid closure status: %s", status)
	}
	if !domain.CanTransitionTask(task.Status, status) {
		return nil, "", fmt.Errorf("invalid closure transition: %s -> %s", task.Status, status)
	}
	task.Status = status
	now := time.Now().UTC()
	task.Notes = append(task.Notes, domain.Note{ActorType: domain.ActorSystem, Text: "closure_reason: " + strings.TrimSpace(reason), CreatedAt: now})
	task.UpdatedAt = now
	return []string{taskID}, string(status) + " task " + taskID, nil
}

func buildTask(phaseID, taskID string, payload initTaskPayload) domain.Task {
	required := true
	if payload.Required != nil {
		required = *payload.Required
	}
	priority := payload.Priority
	if priority == "" {
		priority = domain.PriorityP2
	}
	return domain.Task{TaskID: taskID, PhaseID: phaseID, Title: strings.TrimSpace(payload.Title), Status: domain.TaskNotStarted, Priority: priority, DependsOn: append([]string(nil), payload.DependsOn...), Required: required, UpdatedAt: time.Now().UTC()}
}

func nextPhaseID(plan domain.Plan) string {
	max := 0
	for _, phase := range plan.Phases {
		if n := parseNumericSuffix(phase.PhaseID); n > max {
			max = n
		}
	}
	return fmt.Sprintf("PH%02d", max+1)
}

func nextTaskIDForPhase(plan domain.Plan, phaseID string, extra []domain.Task) string {
	max := 0
	for _, phase := range plan.Phases {
		if phase.PhaseID != phaseID {
			continue
		}
		for _, task := range phase.Tasks {
			if n := parseTaskIndex(task.TaskID); n > max {
				max = n
			}
		}
	}
	for _, task := range extra {
		if n := parseTaskIndex(task.TaskID); n > max {
			max = n
		}
	}
	return fmt.Sprintf("%s-T%02d", phaseID, max+1)
}

func parseNumericSuffix(phaseID string) int {
	var n int
	_, _ = fmt.Sscanf(phaseID, "PH%d", &n)
	return n
}
func parseTaskIndex(taskID string) int {
	var phase, task int
	_, _ = fmt.Sscanf(taskID, "PH%d-T%d", &phase, &task)
	return task
}

func findPhaseIndex(plan domain.Plan, phaseID string) (int, bool) {
	for i, phase := range plan.Phases {
		if phase.PhaseID == phaseID {
			return i, true
		}
	}
	return 0, false
}

func findTaskLocation(plan domain.Plan, taskID string) (int, int, bool) {
	for phaseIndex, phase := range plan.Phases {
		for taskIndex, task := range phase.Tasks {
			if task.TaskID == taskID {
				return phaseIndex, taskIndex, true
			}
		}
	}
	return 0, 0, false
}

func updateDependencyReferences(plan *domain.Plan, oldID, newID string) {
	for phaseIndex := range plan.Phases {
		for taskIndex := range plan.Phases[phaseIndex].Tasks {
			for depIndex, dep := range plan.Phases[phaseIndex].Tasks[taskIndex].DependsOn {
				if dep == oldID {
					plan.Phases[phaseIndex].Tasks[taskIndex].DependsOn[depIndex] = newID
				}
			}
		}
	}
}

func normalizeEditedPlan(plan domain.Plan) domain.Plan {
	for phaseIndex := range plan.Phases {
		phase := &plan.Phases[phaseIndex]
		requiredTasks := make([]domain.Task, 0, len(phase.Tasks))
		for taskIndex := range phase.Tasks {
			task := &phase.Tasks[taskIndex]
			if task.Priority == "" {
				task.Priority = domain.PriorityP2
			}
			if task.UpdatedAt.IsZero() {
				task.UpdatedAt = time.Now().UTC()
			}
			if task.Required {
				requiredTasks = append(requiredTasks, *task)
			}
		}
		phase.Status = planfile.RollupPhaseStatus(domain.Phase{PhaseID: phase.PhaseID, Title: phase.Title, Tasks: requiredTasks})
	}
	plan.CanFinish = finishReady(plan)
	plan.Status = planfile.RollupPlanStatus(plan)
	plan.CurrentPhaseID = ""
	for _, phase := range plan.Phases {
		if phase.Status != domain.PhaseCompleted {
			plan.CurrentPhaseID = phase.PhaseID
			break
		}
	}
	if plan.CurrentPhaseID == "" && len(plan.Phases) > 0 {
		plan.CurrentPhaseID = plan.Phases[len(plan.Phases)-1].PhaseID
	}
	plan.UpdatedAt = time.Now().UTC()
	return plan
}

func hasErrorValidationIssues(issues []domain.ValidationIssue) bool {
	for _, issue := range issues {
		if issue.Severity == "error" {
			return true
		}
	}
	return false
}

func firstValidationErrorIssue(issues []domain.ValidationIssue) domain.ValidationIssue {
	for _, issue := range issues {
		if issue.Severity == "error" {
			return issue
		}
	}
	return domain.ValidationIssue{}
}

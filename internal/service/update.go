package service

import (
	"fmt"
	"strings"
	"time"

	"warden-mcp/internal/domain"
	"warden-mcp/internal/mcp/contracts"
	"warden-mcp/internal/planfile"
)

func UpdateTask(planPath string, req contracts.UpdateTaskRequest) (contracts.UpdateTaskData, []domain.ValidationIssue, error) {
	evidence := make([]domain.EvidenceItem, 0, len(req.Evidence))
	for _, raw := range req.Evidence {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		evidence = append(evidence, parseEvidenceInput(raw))
	}
	if req.ActorType == "" {
		req.ActorType = domain.ActorAgent
	}
	plan, warnings, err := planfile.UpdateTaskFile(planPath, planfile.TaskUpdateMutation{TaskID: req.TaskID, TargetStatus: req.Status, ActorType: req.ActorType, Note: req.Note, Reason: req.Reason, Evidence: evidence, Timestamp: time.Now().UTC()})
	if err != nil {
		return contracts.UpdateTaskData{}, nil, err
	}
	task, phase, ok := findTaskAndPhase(plan, req.TaskID)
	if !ok {
		return contracts.UpdateTaskData{}, warnings, fmt.Errorf("task not found after update: %s", req.TaskID)
	}
	return contracts.UpdateTaskData{Task: summarizeTask(task), Phase: summarizePhase(phase), Plan: summarizePlan(plan, finishReady(plan)), TransitionAccepted: true}, warnings, nil
}

func parseEvidenceInput(raw string) domain.EvidenceItem {
	parts := strings.Split(raw, "|")
	for index := range parts {
		parts[index] = strings.TrimSpace(parts[index])
	}
	switch len(parts) {
	case 0:
		return domain.EvidenceItem{}
	case 1:
		return domain.EvidenceItem{Kind: "ref", Ref: parts[0]}
	case 2:
		return domain.EvidenceItem{Kind: parts[0], Ref: parts[1]}
	default:
		return domain.EvidenceItem{Kind: parts[0], Ref: parts[1], Summary: strings.Join(parts[2:], " | ")}
	}
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

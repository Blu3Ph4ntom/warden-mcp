package planfile

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"warden-mcp/internal/domain"
)

func UpdateTaskStatusFile(path, taskID string, target domain.TaskStatus) (domain.Plan, []domain.ValidationIssue, error) {
	info, err := os.Stat(path)
	if err != nil {
		return domain.Plan{}, nil, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Plan{}, nil, err
	}
	updated, err := UpdateTaskStatusContent(string(content), taskID, target)
	if err != nil {
		return domain.Plan{}, nil, err
	}
	if err := os.WriteFile(path, []byte(updated), info.Mode()); err != nil {
		return domain.Plan{}, nil, err
	}
	return Parse(updated, time.Now().UTC())
}

func UpdateTaskStatusContent(content, taskID string, target domain.TaskStatus) (string, error) {
	lines := splitNormalizedLines(content)
	matches := 0
	for i, line := range lines {
		match := taskLinePattern.FindStringSubmatch(line)
		if match == nil || match[2] != taskID {
			continue
		}
		current := checkboxStatus(match[1])
		if current != target && !domain.CanTransitionTask(current, target) {
			return "", fmt.Errorf("invalid task transition: %s -> %s", current, target)
		}
		marker, err := statusMarker(target)
		if err != nil {
			return "", err
		}
		lines[i] = replaceTaskMarker(line, match[1], marker)
		matches++
	}
	return finalizeUpdatedContent(lines, matches, taskID)
}

func splitNormalizedLines(content string) []string {
	return strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
}

func replaceTaskMarker(line, currentMarker, targetMarker string) string {
	return strings.Replace(line, "["+currentMarker+"]", "["+targetMarker+"]", 1)
}

func finalizeUpdatedContent(lines []string, matches int, taskID string) (string, error) {
	if matches == 0 {
		return "", fmt.Errorf("task not found: %s", taskID)
	}
	if matches > 1 {
		return "", fmt.Errorf("task appears multiple times: %s", taskID)
	}
	updated := strings.Join(lines, "\n")
	plan, _, err := Parse(updated, time.Now().UTC())
	if err != nil {
		return "", err
	}
	return rewriteFrontmatter(updated, plan), nil
}

func rewriteFrontmatter(content string, plan domain.Plan) string {
	lines := strings.Split(content, "\n")
	_, bodyStart, err := parseFrontmatter(lines)
	if err != nil {
		return content
	}
	replacements := map[string]string{
		"current_phase":   nextCurrentPhaseID(plan),
		"completed_tasks": strconv.Itoa(plan.CompletedTaskCount()),
		"can_finish":      strconv.FormatBool(canFinish(plan)),
	}
	for i := 1; i < bodyStart-1; i++ {
		match := frontmatterLinePattern.FindStringSubmatch(lines[i])
		if match == nil {
			continue
		}
		value, ok := replacements[match[1]]
		if !ok {
			continue
		}
		lines[i] = match[1] + ": " + value
	}
	return strings.Join(lines, "\n")
}

func nextCurrentPhaseID(plan domain.Plan) string {
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

func statusMarker(status domain.TaskStatus) (string, error) {
	switch status {
	case domain.TaskDone:
		return "x", nil
	case domain.TaskInProgress:
		return "/", nil
	case domain.TaskNotStarted:
		return " ", nil
	case domain.TaskBlocked:
		return "-", nil
	case domain.TaskCancelled:
		return "-", nil
	case domain.TaskWaived:
		return "-", nil
	default:
		return "", fmt.Errorf("unsupported task status: %s", status)
	}
}

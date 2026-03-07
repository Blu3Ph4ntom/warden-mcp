package planfile

import (
	"fmt"
	"os"
	"time"

	"warden-mcp/internal/domain"
)

func ResetTaskStatusFile(path, taskID string, target domain.TaskStatus) (domain.Plan, []domain.ValidationIssue, error) {
	info, err := os.Stat(path)
	if err != nil {
		return domain.Plan{}, nil, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Plan{}, nil, err
	}
	updated, err := ResetTaskStatusContent(string(content), taskID, target)
	if err != nil {
		return domain.Plan{}, nil, err
	}
	if err := os.WriteFile(path, []byte(updated), info.Mode()); err != nil {
		return domain.Plan{}, nil, err
	}
	return Parse(updated, time.Now().UTC())
}

func ResetTaskStatusContent(content, taskID string, target domain.TaskStatus) (string, error) {
	lines := splitNormalizedLines(content)
	matches := 0
	for i, line := range lines {
		match := taskLinePattern.FindStringSubmatch(line)
		if match == nil || match[2] != taskID {
			continue
		}
		current := checkboxStatus(match[1])
		if current != domain.TaskDone && current != domain.TaskCancelled && current != domain.TaskWaived {
			return "", fmt.Errorf("task is not terminal and cannot be reset: %s", taskID)
		}
		if !domain.CanResetTask(current, target) {
			return "", fmt.Errorf("invalid reset target: %s -> %s", current, target)
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

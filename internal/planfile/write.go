package planfile

import (
	"fmt"
	"os"
	"time"

	"github.com/Blu3Ph4ntom/warden-mcp/internal/domain"
	"github.com/Blu3Ph4ntom/warden-mcp/internal/fsutil"
)

func WritePlanFile(path string, plan domain.Plan) (domain.Plan, []domain.ValidationIssue, error) {
	info, err := os.Stat(path)
	mode := os.FileMode(0o644)
	if err == nil {
		mode = info.Mode()
	} else if !os.IsNotExist(err) {
		return domain.Plan{}, nil, err
	}
	content := Render(plan)
	issues := plan.Validate()
	if hasErrorIssues(issues) {
		return domain.Plan{}, issues, fmt.Errorf("plan validation failed: %s", firstErrorIssue(issues).Message)
	}
	parsed, warnings, err := Parse(content, time.Now().UTC())
	if err != nil {
		return domain.Plan{}, warnings, err
	}
	parsedIssues := parsed.Validate()
	warnings = append(warnings, parsedIssues...)
	if hasErrorIssues(parsedIssues) {
		return domain.Plan{}, warnings, fmt.Errorf("plan validation failed: %s", firstErrorIssue(parsedIssues).Message)
	}
	if err := fsutil.WriteFileAtomic(path, []byte(content), mode); err != nil {
		return domain.Plan{}, warnings, err
	}
	return parsed, warnings, nil
}

func hasErrorIssues(issues []domain.ValidationIssue) bool {
	for _, issue := range issues {
		if issue.Severity == "error" {
			return true
		}
	}
	return false
}

func firstErrorIssue(issues []domain.ValidationIssue) domain.ValidationIssue {
	for _, issue := range issues {
		if issue.Severity == "error" {
			return issue
		}
	}
	return domain.ValidationIssue{}
}

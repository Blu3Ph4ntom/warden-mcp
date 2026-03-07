package planfile

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"warden-mcp/internal/domain"
)

var (
	frontmatterLinePattern = regexp.MustCompile(`^([a-z_]+):\s*(.*)$`)
	taskLinePattern        = regexp.MustCompile(`^- \[([ xX/\-])\] (PH[0-9]{2}-T[0-9]{2}) (.+)$`)
	phaseHeadingPattern    = regexp.MustCompile(`^## Phase [0-9]+ [—-] (.+)$`)
	twoPartVersionPattern  = regexp.MustCompile(`^[0-9]+\.[0-9]+$`)
)

func Load(path string) (domain.Plan, []domain.ValidationIssue, error) {
	info, err := os.Stat(path)
	if err != nil {
		return domain.Plan{}, nil, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Plan{}, nil, err
	}
	return Parse(string(content), info.ModTime().UTC())
}

func Parse(content string, updatedAt time.Time) (domain.Plan, []domain.ValidationIssue, error) {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	frontmatter, bodyStart, err := parseFrontmatter(lines)
	if err != nil {
		return domain.Plan{}, nil, err
	}
	issues := make([]domain.ValidationIssue, 0)
	version := strings.TrimSpace(frontmatter["version"])
	if twoPartVersionPattern.MatchString(version) {
		version += ".0"
		issues = append(issues, domain.ValidationIssue{Severity: "warning", Code: "PLAN_VERSION_NORMALIZED", Message: "two-part version normalized to semver", Path: "version"})
	}
	plan := domain.Plan{
		PlanID:         strings.TrimSpace(frontmatter["plan_id"]),
		Title:          strings.TrimSpace(frontmatter["title"]),
		Status:         domain.PlanStatus(strings.TrimSpace(frontmatter["status"])),
		Version:        version,
		CurrentPhaseID: strings.TrimSpace(frontmatter["current_phase"]),
		UpdatedAt:      updatedAt,
	}
	phases, parseIssues := parsePhases(lines[bodyStart:], updatedAt)
	plan.Phases = phases
	issues = append(issues, parseIssues...)
	for i := range plan.Phases {
		plan.Phases[i].Status = rollupPhaseStatus(plan.Phases[i])
	}
	plan.CanFinish = canFinish(plan)
	if plan.Status == "" {
		plan.Status = rollupPlanStatus(plan)
	}
	return plan, issues, nil
}

func parseFrontmatter(lines []string) (map[string]string, int, error) {
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil, 0, fmt.Errorf("plan frontmatter opening delimiter is required")
	}
	values := map[string]string{}
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) == "---" {
			return values, i + 1, nil
		}
		match := frontmatterLinePattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		values[match[1]] = match[2]
	}
	return nil, 0, fmt.Errorf("plan frontmatter closing delimiter is required")
}

func parsePhases(lines []string, updatedAt time.Time) ([]domain.Phase, []domain.ValidationIssue) {
	phases := make([]domain.Phase, 0)
	issues := make([]domain.ValidationIssue, 0)
	var current *domain.Phase
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if match := phaseHeadingPattern.FindStringSubmatch(line); match != nil {
			if current != nil {
				phases = append(phases, *current)
			}
			current = &domain.Phase{Title: strings.TrimSpace(match[1])}
			continue
		}
		match := taskLinePattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		if current == nil {
			current = &domain.Phase{Title: match[2][:4]}
		}
		phaseID := match[2][:4]
		if current.PhaseID == "" {
			current.PhaseID = phaseID
		}
		if current.PhaseID != phaseID {
			issues = append(issues, domain.ValidationIssue{Severity: "error", Code: "TASK_PHASE_SECTION_MISMATCH", Message: "task ID does not match current phase section", Path: match[2]})
			continue
		}
		current.Tasks = append(current.Tasks, domain.Task{TaskID: match[2], PhaseID: phaseID, Title: strings.TrimSpace(match[3]), Status: checkboxStatus(match[1]), Priority: domain.PriorityP2, Required: true, UpdatedAt: updatedAt})
	}
	if current != nil {
		phases = append(phases, *current)
	}
	return phases, issues
}

func checkboxStatus(mark string) domain.TaskStatus {
	switch mark {
	case "x", "X":
		return domain.TaskDone
	case "/":
		return domain.TaskInProgress
	case "-":
		return domain.TaskBlocked
	default:
		return domain.TaskNotStarted
	}
}

func rollupPhaseStatus(phase domain.Phase) domain.PhaseStatus {
	if len(phase.Tasks) == 0 {
		return domain.PhaseNotStarted
	}
	allDone := true
	anyStarted := false
	anyBlocked := false
	for _, task := range phase.Tasks {
		if task.Status == domain.TaskBlocked {
			anyBlocked = true
		}
		if task.Status != domain.TaskNotStarted {
			anyStarted = true
		}
		if task.Status != domain.TaskDone && task.Status != domain.TaskCancelled && task.Status != domain.TaskWaived {
			allDone = false
		}
	}
	if allDone {
		return domain.PhaseCompleted
	}
	if anyBlocked && !anyStarted {
		return domain.PhaseBlocked
	}
	if anyStarted {
		return domain.PhaseInProgress
	}
	return domain.PhaseNotStarted
}

func rollupPlanStatus(plan domain.Plan) domain.PlanStatus {
	allCompleted := len(plan.Phases) > 0
	for _, phase := range plan.Phases {
		if phase.Status != domain.PhaseCompleted {
			allCompleted = false
			break
		}
	}
	if allCompleted {
		return domain.PlanCompleted
	}
	return domain.PlanActive
}

func canFinish(plan domain.Plan) bool {
	for _, phase := range plan.Phases {
		for _, task := range phase.Tasks {
			if task.Required && !task.IsTerminal() {
				return false
			}
		}
	}
	return true
}

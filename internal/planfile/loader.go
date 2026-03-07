package planfile

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"warden-mcp/internal/domain"
)

var (
	ErrPlanTooLarge        = errors.New("plan file exceeds maximum size")
	ErrPlanTooManyLines    = errors.New("plan file exceeds maximum line count")
	frontmatterLinePattern = regexp.MustCompile(`^([a-z_]+):\s*(.*)$`)
	taskLinePattern        = regexp.MustCompile(`^- \[([ xX/\-])\] (PH[0-9]{2}-T[0-9]{2}) (.+)$`)
	phaseHeadingPattern    = regexp.MustCompile(`^## Phase [0-9]+ [—-] (.+)$`)
	twoPartVersionPattern  = regexp.MustCompile(`^[0-9]+\.[0-9]+$`)
)

const (
	DefaultMaxPlanBytes = 1 << 20
	DefaultMaxPlanLines = 10000
)

func Load(path string) (domain.Plan, []domain.ValidationIssue, error) {
	info, err := os.Stat(path)
	if err != nil {
		return domain.Plan{}, nil, err
	}
	if info.Size() > DefaultMaxPlanBytes {
		return domain.Plan{}, nil, fmt.Errorf("%w: %d bytes", ErrPlanTooLarge, info.Size())
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Plan{}, nil, err
	}
	return Parse(string(content), info.ModTime().UTC())
}

func Parse(content string, updatedAt time.Time) (domain.Plan, []domain.ValidationIssue, error) {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	if len(lines) > DefaultMaxPlanLines {
		return domain.Plan{}, nil, fmt.Errorf("%w: %d lines", ErrPlanTooManyLines, len(lines))
	}
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
		plan.Phases[i].Status = RollupPhaseStatus(plan.Phases[i])
	}
	plan.CanFinish = CanFinishPlan(plan)
	if plan.Status == "" {
		plan.Status = RollupPlanStatus(plan)
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
	for index := 0; index < len(lines); index++ {
		raw := lines[index]
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
		title, priority, dependsOn, required, status, metaIssues := parseTaskMetadata(match[3], match[2], checkboxStatus(match[1]))
		issues = append(issues, metaIssues...)
		notes, evidence, nextIndex, blockIssues := parseTaskBlocks(lines, index+1, match[2])
		issues = append(issues, blockIssues...)
		current.Tasks = append(current.Tasks, domain.Task{TaskID: match[2], PhaseID: phaseID, Title: title, Status: status, Priority: priority, DependsOn: dependsOn, Required: required, Evidence: evidence, Notes: notes, UpdatedAt: updatedAt})
		index = nextIndex - 1
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

func RollupPhaseStatus(phase domain.Phase) domain.PhaseStatus {
	if len(phase.Tasks) == 0 {
		return domain.PhaseNotStarted
	}
	requiredTasks := requiredTasks(phase.Tasks)
	if len(requiredTasks) == 0 {
		return domain.PhaseCompleted
	}
	allDone := true
	anyStarted := false
	anyBlocked := false
	for _, task := range requiredTasks {
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

func RollupPlanStatus(plan domain.Plan) domain.PlanStatus {
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

func CanFinishPlan(plan domain.Plan) bool {
	for _, phase := range plan.Phases {
		for _, task := range phase.Tasks {
			if task.Required && !task.IsTerminal() {
				return false
			}
		}
	}
	return true
}

func requiredTasks(tasks []domain.Task) []domain.Task {
	filtered := make([]domain.Task, 0, len(tasks))
	for _, task := range tasks {
		if task.Required {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

func parseTaskMetadata(raw, taskID string, defaultStatus domain.TaskStatus) (string, domain.Priority, []string, bool, domain.TaskStatus, []domain.ValidationIssue) {
	parts := strings.Split(raw, " | ")
	title := strings.TrimSpace(parts[0])
	priority := domain.PriorityP2
	dependsOn := make([]string, 0)
	required := true
	status := defaultStatus
	issues := make([]domain.ValidationIssue, 0)
	for _, segment := range parts[1:] {
		key, value, ok := strings.Cut(strings.TrimSpace(segment), ":")
		if !ok {
			issues = append(issues, domain.ValidationIssue{Severity: "warning", Code: "TASK_METADATA_IGNORED", Message: "task metadata segment could not be parsed", Path: taskID})
			continue
		}
		key = strings.TrimSpace(strings.ToLower(key))
		value = strings.TrimSpace(value)
		switch key {
		case "priority":
			candidate := domain.Priority(strings.ToUpper(value))
			if !candidate.Valid() {
				issues = append(issues, domain.ValidationIssue{Severity: "warning", Code: "TASK_PRIORITY_INVALID", Message: "task priority metadata is invalid", Path: taskID})
				continue
			}
			priority = candidate
		case "depends_on":
			dependsOn = parseCSVList(value)
		case "required":
			switch strings.ToLower(value) {
			case "true":
				required = true
			case "false":
				required = false
			default:
				issues = append(issues, domain.ValidationIssue{Severity: "warning", Code: "TASK_REQUIRED_INVALID", Message: "task required metadata must be true or false", Path: taskID})
			}
		case "status":
			candidate := domain.TaskStatus(strings.ToLower(value))
			if !candidate.Valid() {
				issues = append(issues, domain.ValidationIssue{Severity: "warning", Code: "TASK_STATUS_INVALID", Message: "task status metadata is invalid", Path: taskID})
				continue
			}
			status = candidate
		default:
			issues = append(issues, domain.ValidationIssue{Severity: "warning", Code: "TASK_METADATA_IGNORED", Message: "unknown task metadata key ignored", Path: taskID})
		}
	}
	return title, priority, dependsOn, required, status, issues
}

func parseTaskBlocks(lines []string, start int, taskID string) ([]domain.Note, []domain.EvidenceItem, int, []domain.ValidationIssue) {
	notes := make([]domain.Note, 0)
	evidence := make([]domain.EvidenceItem, 0)
	issues := make([]domain.ValidationIssue, 0)
	section := ""
	index := start
	for ; index < len(lines); index++ {
		raw := lines[index]
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(raw, "  ") {
			break
		}
		switch trimmed {
		case "notes:":
			section = "notes"
		case "evidence:":
			section = "evidence"
		default:
			if !strings.HasPrefix(raw, "    - ") {
				issues = append(issues, domain.ValidationIssue{Severity: "warning", Code: "TASK_BLOCK_IGNORED", Message: "task metadata block line ignored", Path: taskID})
				continue
			}
			payload := strings.TrimSpace(strings.TrimPrefix(raw, "    - "))
			switch section {
			case "notes":
				var note domain.Note
				if err := json.Unmarshal([]byte(payload), &note); err != nil {
					issues = append(issues, domain.ValidationIssue{Severity: "warning", Code: "TASK_NOTE_INVALID", Message: "task note block could not be parsed", Path: taskID})
					continue
				}
				notes = append(notes, note)
			case "evidence":
				var item domain.EvidenceItem
				if err := json.Unmarshal([]byte(payload), &item); err != nil {
					issues = append(issues, domain.ValidationIssue{Severity: "warning", Code: "TASK_EVIDENCE_INVALID", Message: "task evidence block could not be parsed", Path: taskID})
					continue
				}
				evidence = append(evidence, item)
			default:
				issues = append(issues, domain.ValidationIssue{Severity: "warning", Code: "TASK_BLOCK_IGNORED", Message: "task block item appeared before a supported section header", Path: taskID})
			}
		}
	}
	return notes, evidence, index, issues
}

func parseCSVList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	sort.Strings(result)
	return result
}

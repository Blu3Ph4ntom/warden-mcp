package domain

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

type PlanStatus string
type PhaseStatus string
type TaskStatus string
type Priority string
type ActorType string

const (
	PlanDraft     PlanStatus = "draft"
	PlanActive    PlanStatus = "active"
	PlanBlocked   PlanStatus = "blocked"
	PlanCompleted PlanStatus = "completed"
	PlanArchived  PlanStatus = "archived"

	PhaseNotStarted PhaseStatus = "not_started"
	PhaseInProgress PhaseStatus = "in_progress"
	PhaseBlocked    PhaseStatus = "blocked"
	PhaseCompleted  PhaseStatus = "completed"

	TaskNotStarted TaskStatus = "not_started"
	TaskInProgress TaskStatus = "in_progress"
	TaskDone       TaskStatus = "done"
	TaskBlocked    TaskStatus = "blocked"
	TaskCancelled  TaskStatus = "cancelled"
	TaskWaived     TaskStatus = "waived"

	PriorityP0 Priority = "P0"
	PriorityP1 Priority = "P1"
	PriorityP2 Priority = "P2"
	PriorityP3 Priority = "P3"

	ActorAgent  ActorType = "agent"
	ActorHuman  ActorType = "human"
	ActorSystem ActorType = "system"
)

var (
	planIDPattern  = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{2,63}$`)
	phaseIDPattern = regexp.MustCompile(`^PH[0-9]{2}$`)
	taskIDPattern  = regexp.MustCompile(`^PH[0-9]{2}-T[0-9]{2}$`)
)

type ValidationIssue struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	Path     string `json:"path"`
}

type EvidenceItem struct {
	Kind    string `json:"kind"`
	Ref     string `json:"ref"`
	Summary string `json:"summary"`
}

type Note struct {
	ActorType ActorType `json:"actor_type"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type Task struct {
	TaskID    string         `json:"task_id"`
	PhaseID   string         `json:"phase_id"`
	Title     string         `json:"title"`
	Status    TaskStatus     `json:"status"`
	Priority  Priority       `json:"priority"`
	DependsOn []string       `json:"depends_on,omitempty"`
	Required  bool           `json:"required"`
	Evidence  []EvidenceItem `json:"evidence,omitempty"`
	Notes     []Note         `json:"notes,omitempty"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type Phase struct {
	PhaseID string      `json:"phase_id"`
	Title   string      `json:"title"`
	Status  PhaseStatus `json:"status"`
	Tasks   []Task      `json:"tasks"`
}

type Plan struct {
	PlanID         string     `json:"plan_id"`
	Title          string     `json:"title"`
	Status         PlanStatus `json:"status"`
	Version        string     `json:"version"`
	WorkspaceRoot  string     `json:"workspace_root,omitempty"`
	PlanPath       string     `json:"plan_path,omitempty"`
	CurrentPhaseID string     `json:"current_phase_id,omitempty"`
	Phases         []Phase    `json:"phases"`
	CanFinish      bool       `json:"can_finish"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type AuditEvent struct {
	EventID   string    `json:"event_id"`
	PlanID    string    `json:"plan_id"`
	ActorType ActorType `json:"actor_type"`
	EventType string    `json:"event_type"`
	Accepted  bool      `json:"accepted"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

type FinishRequest struct {
	PlanID      string    `json:"plan_id"`
	ActorType   ActorType `json:"actor_type"`
	Summary     string    `json:"summary,omitempty"`
	RequestedAt time.Time `json:"requested_at"`
}

func (s PlanStatus) Valid() bool {
	switch s {
	case PlanDraft, PlanActive, PlanBlocked, PlanCompleted, PlanArchived:
		return true
	default:
		return false
	}
}

func (s PhaseStatus) Valid() bool {
	switch s {
	case PhaseNotStarted, PhaseInProgress, PhaseBlocked, PhaseCompleted:
		return true
	default:
		return false
	}
}

func (s TaskStatus) Valid() bool {
	switch s {
	case TaskNotStarted, TaskInProgress, TaskDone, TaskBlocked, TaskCancelled, TaskWaived:
		return true
	default:
		return false
	}
}

func (p Priority) Valid() bool {
	switch p {
	case PriorityP0, PriorityP1, PriorityP2, PriorityP3:
		return true
	default:
		return false
	}
}

func (a ActorType) Valid() bool {
	switch a {
	case ActorAgent, ActorHuman, ActorSystem:
		return true
	default:
		return false
	}
}

func (t Task) IsTerminal() bool {
	return t.Status == TaskDone || t.Status == TaskCancelled || t.Status == TaskWaived
}

func (p Phase) CompletedTaskCount() int {
	count := 0
	for _, task := range p.Tasks {
		if task.Status == TaskDone {
			count++
		}
	}
	return count
}

func (p Phase) BlockedTaskCount() int {
	count := 0
	for _, task := range p.Tasks {
		if task.Status == TaskBlocked {
			count++
		}
	}
	return count
}

func (p Plan) TotalTasks() int {
	total := 0
	for _, phase := range p.Phases {
		total += len(phase.Tasks)
	}
	return total
}

func (p Plan) CompletedTaskCount() int {
	count := 0
	for _, phase := range p.Phases {
		count += phase.CompletedTaskCount()
	}
	return count
}

func ValidatePlanID(id string) bool  { return planIDPattern.MatchString(id) }
func ValidatePhaseID(id string) bool { return phaseIDPattern.MatchString(id) }
func ValidateTaskID(id string) bool  { return taskIDPattern.MatchString(id) }

func (r FinishRequest) Validate() []ValidationIssue {
	issues := make([]ValidationIssue, 0)
	if !ValidatePlanID(r.PlanID) {
		issues = append(issues, issue("error", "PLAN_ID_INVALID", "finish request plan ID is invalid", "plan_id"))
	}
	if !r.ActorType.Valid() {
		issues = append(issues, issue("error", "ACTOR_TYPE_INVALID", "actor type is invalid", "actor_type"))
	}
	if r.RequestedAt.IsZero() {
		issues = append(issues, issue("error", "REQUESTED_AT_MISSING", "requested_at is required", "requested_at"))
	}
	return issues
}

func (e AuditEvent) Validate() []ValidationIssue {
	issues := make([]ValidationIssue, 0)
	if strings.TrimSpace(e.EventID) == "" {
		issues = append(issues, issue("error", "EVENT_ID_MISSING", "event ID is required", "event_id"))
	}
	if !ValidatePlanID(e.PlanID) {
		issues = append(issues, issue("error", "PLAN_ID_INVALID", "audit event plan ID is invalid", "plan_id"))
	}
	if !e.ActorType.Valid() {
		issues = append(issues, issue("error", "ACTOR_TYPE_INVALID", "actor type is invalid", "actor_type"))
	}
	if strings.TrimSpace(e.EventType) == "" {
		issues = append(issues, issue("error", "EVENT_TYPE_MISSING", "event type is required", "event_type"))
	}
	if e.Timestamp.IsZero() {
		issues = append(issues, issue("error", "TIMESTAMP_MISSING", "timestamp is required", "timestamp"))
	}
	return issues
}

func (p Plan) Validate() []ValidationIssue {
	issues := make([]ValidationIssue, 0)
	if !ValidatePlanID(p.PlanID) {
		issues = append(issues, issue("error", "PLAN_ID_INVALID", "plan ID must be lowercase slug-like text", "plan_id"))
	}
	if strings.TrimSpace(p.Title) == "" {
		issues = append(issues, issue("error", "PLAN_TITLE_MISSING", "plan title is required", "title"))
	}
	if !ValidatePlanVersion(p.Version) {
		issues = append(issues, issue("error", "PLAN_VERSION_INVALID", "plan version must look like semver", "version"))
	}
	if !p.Status.Valid() {
		issues = append(issues, issue("error", "PLAN_STATUS_INVALID", "plan status is invalid", "status"))
	}
	if len(p.Phases) < 2 {
		issues = append(issues, issue("error", "PLAN_TOO_SHALLOW", "each plan needs at least two phases", "phases"))
	}

	seenPhaseIDs := make(map[string]struct{}, len(p.Phases))
	taskToPhase := map[string]string{}
	adjacency := map[string][]string{}

	for phaseIndex, phase := range p.Phases {
		phasePath := fmt.Sprintf("phases[%d]", phaseIndex)
		issues = append(issues, phase.validate(phasePath)...)
		if _, exists := seenPhaseIDs[phase.PhaseID]; exists {
			issues = append(issues, issue("error", "DUPLICATE_PHASE_ID", "phase IDs must be unique", phasePath+".phase_id"))
		}
		seenPhaseIDs[phase.PhaseID] = struct{}{}

		for taskIndex, task := range phase.Tasks {
			taskPath := fmt.Sprintf("%s.tasks[%d]", phasePath, taskIndex)
			if priorPhaseID, exists := taskToPhase[task.TaskID]; exists {
				issues = append(issues, issue("error", "DUPLICATE_TASK_ID", fmt.Sprintf("task ID already used in phase %s", priorPhaseID), taskPath+".task_id"))
			} else {
				taskToPhase[task.TaskID] = phase.PhaseID
			}
			adjacency[task.TaskID] = append([]string(nil), task.DependsOn...)
		}
	}

	for phaseIndex, phase := range p.Phases {
		for taskIndex, task := range phase.Tasks {
			taskPath := fmt.Sprintf("phases[%d].tasks[%d]", phaseIndex, taskIndex)
			for depIndex, dependencyID := range task.DependsOn {
				if _, exists := taskToPhase[dependencyID]; !exists {
					issues = append(issues, issue("error", "DEPENDENCY_NOT_FOUND", "dependency must reference an existing task ID", fmt.Sprintf("%s.depends_on[%d]", taskPath, depIndex)))
				}
			}
		}
	}

	for _, cycle := range dependencyCycles(adjacency) {
		issues = append(issues, issue("error", "DEPENDENCY_CYCLE", fmt.Sprintf("dependency cycle detected: %s", cycle), "phases"))
	}
	if p.CurrentPhaseID != "" {
		if _, exists := seenPhaseIDs[p.CurrentPhaseID]; !exists {
			issues = append(issues, issue("error", "CURRENT_PHASE_INVALID", "current_phase_id must reference an existing phase", "current_phase_id"))
		}
	}

	return issues
}

func (p Phase) validate(path string) []ValidationIssue {
	issues := make([]ValidationIssue, 0)
	if !ValidatePhaseID(p.PhaseID) {
		issues = append(issues, issue("error", "PHASE_ID_INVALID", "phase ID must match PH##", path+".phase_id"))
	}
	if strings.TrimSpace(p.Title) == "" {
		issues = append(issues, issue("error", "PHASE_TITLE_MISSING", "phase title is required", path+".title"))
	}
	if !p.Status.Valid() {
		issues = append(issues, issue("error", "PHASE_STATUS_INVALID", "phase status is invalid", path+".status"))
	}
	if len(p.Tasks) < 2 {
		issues = append(issues, issue("error", "PHASE_TOO_SHALLOW", "each phase needs at least two tasks", path+".tasks"))
	}
	for taskIndex, task := range p.Tasks {
		issues = append(issues, task.validate(fmt.Sprintf("%s.tasks[%d]", path, taskIndex), p.PhaseID)...)
	}
	return issues
}

func (t Task) validate(path, expectedPhaseID string) []ValidationIssue {
	issues := make([]ValidationIssue, 0)
	if !ValidateTaskID(t.TaskID) {
		issues = append(issues, issue("error", "TASK_ID_INVALID", "task ID must match PH##-T##", path+".task_id"))
	}
	if t.PhaseID != expectedPhaseID {
		issues = append(issues, issue("error", "TASK_PHASE_MISMATCH", "task phase_id must match containing phase", path+".phase_id"))
	}
	if strings.TrimSpace(t.Title) == "" {
		issues = append(issues, issue("error", "TASK_TITLE_MISSING", "task title is required", path+".title"))
	}
	if !t.Status.Valid() {
		issues = append(issues, issue("error", "TASK_STATUS_INVALID", "task status is invalid", path+".status"))
	}
	if !t.Priority.Valid() {
		issues = append(issues, issue("error", "TASK_PRIORITY_INVALID", "task priority is invalid", path+".priority"))
	}
	seenDeps := map[string]struct{}{}
	for index, dependencyID := range t.DependsOn {
		if dependencyID == t.TaskID {
			issues = append(issues, issue("error", "SELF_DEPENDENCY", "task cannot depend on itself", fmt.Sprintf("%s.depends_on[%d]", path, index)))
		}
		if _, exists := seenDeps[dependencyID]; exists {
			issues = append(issues, issue("error", "DUPLICATE_DEPENDENCY", "duplicate task dependency", fmt.Sprintf("%s.depends_on[%d]", path, index)))
		}
		seenDeps[dependencyID] = struct{}{}
	}
	for index, evidence := range t.Evidence {
		if strings.TrimSpace(evidence.Kind) == "" || strings.TrimSpace(evidence.Ref) == "" {
			issues = append(issues, issue("error", "EVIDENCE_INVALID", "evidence requires kind and ref", fmt.Sprintf("%s.evidence[%d]", path, index)))
		}
	}
	for index, note := range t.Notes {
		if !note.ActorType.Valid() || strings.TrimSpace(note.Text) == "" {
			issues = append(issues, issue("error", "NOTE_INVALID", "note requires actor_type and text", fmt.Sprintf("%s.notes[%d]", path, index)))
		}
	}
	return issues
}

func dependencyCycles(adjacency map[string][]string) []string {
	state := map[string]int{}
	cycles := map[string]struct{}{}
	var visit func(string)
	visit = func(node string) {
		state[node] = 1
		for _, next := range adjacency[node] {
			switch state[next] {
			case 0:
				visit(next)
			case 1:
				cycles[node+"->"+next] = struct{}{}
			}
		}
		state[node] = 2
	}
	for node := range adjacency {
		if state[node] == 0 {
			visit(node)
		}
	}
	out := make([]string, 0, len(cycles))
	for cycle := range cycles {
		out = append(out, cycle)
	}
	sort.Strings(out)
	return out
}

func issue(severity, code, message, path string) ValidationIssue {
	return ValidationIssue{Severity: severity, Code: code, Message: message, Path: path}
}

package contracts

import (
	"encoding/json"

	"warden-mcp/internal/domain"
)

type ExportFormat string
type ImportFormat string
type EnforcementProfile string
type SourceType string
type ValidateMode string
type EditOperation string
type ImportMode string
type ReconcileMode string

const (
	ExportMarkdown ExportFormat = "markdown"
	ExportJSON     ExportFormat = "json"
	ExportCSV      ExportFormat = "csv"

	ImportMarkdown ImportFormat = "markdown"
	ImportJSON     ImportFormat = "json"

	EnforcementStrict   EnforcementProfile = "strict"
	EnforcementBalanced EnforcementProfile = "balanced"
	EnforcementAdvisory EnforcementProfile = "advisory"

	SourceMCP      SourceType = "mcp"
	SourceMarkdown SourceType = "markdown"
	SourceJSON     SourceType = "json"
	SourceHuman    SourceType = "human"

	ValidateStrict ValidateMode = "strict"
	ValidateImport ValidateMode = "import"
	ValidateLint   ValidateMode = "lint"

	EditAddPhase         EditOperation = "add_phase"
	EditAddTask          EditOperation = "add_task"
	EditUpdateTaskFields EditOperation = "update_task_fields"
	EditMoveTask         EditOperation = "move_task"
	EditSplitTask        EditOperation = "split_task"
	EditReprioritizeTask EditOperation = "reprioritize_task"
	EditAddDependency    EditOperation = "add_dependency"
	EditRemoveDependency EditOperation = "remove_dependency"
	EditWaiveTask        EditOperation = "waive_task"
	EditCancelTask       EditOperation = "cancel_task"

	ImportCreate  ImportMode = "create"
	ImportMerge   ImportMode = "merge"
	ImportReplace ImportMode = "replace"

	ReconcileDryRun ReconcileMode = "dry_run"
	ReconcileApply  ReconcileMode = "apply"
)

const (
	ErrPlanNotFound          = "PLAN_NOT_FOUND"
	ErrPlanInvalid           = "PLAN_INVALID"
	ErrPhaseInvalid          = "PHASE_INVALID"
	ErrTaskNotFound          = "TASK_NOT_FOUND"
	ErrTaskTransitionInvalid = "TASK_TRANSITION_INVALID"
	ErrDependencyViolation   = "DEPENDENCY_VIOLATION"
	ErrFinishDenied          = "FINISH_DENIED"
	ErrSyncConflict          = "SYNC_CONFLICT"
	ErrImportInvalid         = "IMPORT_INVALID"
	ErrExportFailed          = "EXPORT_FAILED"
	ErrArchiveDenied         = "ARCHIVE_DENIED"
	ErrInternal              = "INTERNAL_ERROR"
)

type ErrorObject struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Retryable bool           `json:"retryable"`
	Details   map[string]any `json:"details"`
}

type BlockingReason struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	PhaseID string `json:"phase_id,omitempty"`
	TaskID  string `json:"task_id,omitempty"`
}

type TaskSummary struct {
	TaskID    string                `json:"task_id"`
	PhaseID   string                `json:"phase_id"`
	Title     string                `json:"title"`
	Status    domain.TaskStatus     `json:"status"`
	Priority  domain.Priority       `json:"priority"`
	DependsOn []string              `json:"depends_on"`
	Required  bool                  `json:"required"`
	Evidence  []domain.EvidenceItem `json:"evidence,omitempty"`
	Notes     []domain.Note         `json:"notes,omitempty"`
	UpdatedAt string                `json:"updated_at"`
}

type PhaseSummary struct {
	PhaseID            string             `json:"phase_id"`
	Title              string             `json:"title"`
	Status             domain.PhaseStatus `json:"status"`
	TaskCount          int                `json:"task_count"`
	CompletedTaskCount int                `json:"completed_task_count"`
	BlockedTaskCount   int                `json:"blocked_task_count"`
}

type PlanSummary struct {
	PlanID         string            `json:"plan_id"`
	Title          string            `json:"title"`
	Status         domain.PlanStatus `json:"status"`
	Version        string            `json:"version"`
	CurrentPhaseID string            `json:"current_phase_id,omitempty"`
	TotalTasks     int               `json:"total_tasks"`
	CompletedTasks int               `json:"completed_tasks"`
	CanFinish      bool              `json:"can_finish"`
	UpdatedAt      string            `json:"updated_at"`
}

type ToolResponseEnvelope[T any] struct {
	OK        bool                     `json:"ok"`
	Tool      string                   `json:"tool"`
	Timestamp string                   `json:"timestamp"`
	PlanID    string                   `json:"plan_id,omitempty"`
	Warnings  []domain.ValidationIssue `json:"warnings"`
	Error     *ErrorObject             `json:"error,omitempty"`
	Data      T                        `json:"data"`
}

type InitPlanTaskInput struct {
	Title     string          `json:"title"`
	DependsOn []string        `json:"depends_on,omitempty"`
	Priority  domain.Priority `json:"priority,omitempty"`
	Required  *bool           `json:"required,omitempty"`
}

type InitPlanPhaseInput struct {
	Title string              `json:"title"`
	Tasks []InitPlanTaskInput `json:"tasks"`
}

type InitPlanRequest struct {
	Title                    string               `json:"title"`
	Goal                     string               `json:"goal,omitempty"`
	SourceText               string               `json:"source_text,omitempty"`
	Phases                   []InitPlanPhaseInput `json:"phases,omitempty"`
	PlanID                   string               `json:"plan_id,omitempty"`
	Version                  string               `json:"version,omitempty"`
	EnforcementProfile       EnforcementProfile   `json:"enforcement_profile,omitempty"`
	SourceType               SourceType           `json:"source_type,omitempty"`
	CreateMarkdownProjection bool                 `json:"create_markdown_projection,omitempty"`
}

type InitPlanData struct {
	Plan             PlanSummary              `json:"plan"`
	Phases           []PhaseSummary           `json:"phases"`
	Tasks            []TaskSummary            `json:"tasks"`
	ValidationIssues []domain.ValidationIssue `json:"validation_issues"`
	Normalized       bool                     `json:"normalized"`
}

type ValidatePlanRequest struct {
	Plan json.RawMessage `json:"plan"`
	Mode ValidateMode    `json:"mode,omitempty"`
}

type NormalizedCounts struct {
	PhaseCount int `json:"phase_count"`
	TaskCount  int `json:"task_count"`
}

type ValidatePlanData struct {
	Valid            bool                     `json:"valid"`
	Issues           []domain.ValidationIssue `json:"issues"`
	NormalizedCounts NormalizedCounts         `json:"normalized_counts"`
}

type GetStatusRequest struct {
	PlanID                string `json:"plan_id"`
	IncludeTasks          bool   `json:"include_tasks,omitempty"`
	IncludeCompletedTasks bool   `json:"include_completed_tasks,omitempty"`
	IncludeAuditSummary   bool   `json:"include_audit_summary,omitempty"`
}

type GetStatusData struct {
	Plan            PlanSummary      `json:"plan"`
	Phases          []PhaseSummary   `json:"phases"`
	Tasks           []TaskSummary    `json:"tasks,omitempty"`
	BlockingReasons []BlockingReason `json:"blocking_reasons"`
	NextTaskID      string           `json:"next_task_id,omitempty"`
	Stalled         bool             `json:"stalled"`
	StalledSince    string           `json:"stalled_since,omitempty"`
}

type HealthCheckRequest struct {
	PlanPath string `json:"plan_path,omitempty"`
}

type HealthCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type HealthCheckData struct {
	Status    string        `json:"status"`
	PlanID    string        `json:"plan_id,omitempty"`
	Checks    []HealthCheck `json:"checks"`
	CheckedAt string        `json:"checked_at"`
}

type UpdateTaskRequest struct {
	PlanID    string            `json:"plan_id"`
	TaskID    string            `json:"task_id"`
	Status    domain.TaskStatus `json:"status"`
	Note      string            `json:"note,omitempty"`
	Evidence  []string          `json:"evidence,omitempty"`
	Reason    string            `json:"reason,omitempty"`
	ActorType domain.ActorType  `json:"actor_type,omitempty"`
}

type UpdateTaskData struct {
	Task               TaskSummary  `json:"task"`
	Phase              PhaseSummary `json:"phase"`
	Plan               PlanSummary  `json:"plan"`
	TransitionAccepted bool         `json:"transition_accepted"`
}

type GetNextTaskRequest struct {
	PlanID              string          `json:"plan_id"`
	RespectPhaseOrder   bool            `json:"respect_phase_order,omitempty"`
	RespectDependencies bool            `json:"respect_dependencies,omitempty"`
	PriorityBias        domain.Priority `json:"priority_bias,omitempty"`
}

type GetNextTaskData struct {
	NextTask        *TaskSummary     `json:"next_task,omitempty"`
	Reason          string           `json:"reason"`
	Blocked         bool             `json:"blocked"`
	BlockingReasons []BlockingReason `json:"blocking_reasons"`
}

type RequestFinishRequest struct {
	PlanID    string           `json:"plan_id"`
	ActorType domain.ActorType `json:"actor_type,omitempty"`
	Summary   string           `json:"summary,omitempty"`
}

type RequestFinishData struct {
	CanFinish             bool             `json:"can_finish"`
	Plan                  PlanSummary      `json:"plan"`
	BlockingReasons       []BlockingReason `json:"blocking_reasons"`
	IncompleteTaskIDs     []string         `json:"incomplete_task_ids"`
	NextRequiredActions   []string         `json:"next_required_actions"`
	RecommendedNextTaskID string           `json:"recommended_next_task_id,omitempty"`
}

type EditPlanRequest struct {
	PlanID    string          `json:"plan_id"`
	Operation EditOperation   `json:"operation"`
	TargetID  string          `json:"target_id,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Reason    string          `json:"reason,omitempty"`
}

type EditPlanData struct {
	Plan        PlanSummary `json:"plan"`
	ChangedIDs  []string    `json:"changed_ids"`
	DiffSummary string      `json:"diff_summary"`
}

type ResetTaskRequest struct {
	PlanID string            `json:"plan_id"`
	TaskID string            `json:"task_id"`
	Status domain.TaskStatus `json:"status,omitempty"`
	Reason string            `json:"reason"`
}

type ResetTaskData struct {
	Task TaskSummary `json:"task"`
	Plan PlanSummary `json:"plan"`
}

type PriorityUpdate struct {
	TaskID   string          `json:"task_id"`
	Priority domain.Priority `json:"priority"`
}

type PrioritizeTasksRequest struct {
	PlanID  string           `json:"plan_id"`
	Updates []PriorityUpdate `json:"updates"`
}

type PrioritizeTasksData struct {
	UpdatedTaskIDs []string    `json:"updated_task_ids"`
	Plan           PlanSummary `json:"plan"`
}

type ListPlansRequest struct {
	Status          domain.PlanStatus `json:"status,omitempty"`
	IncludeArchived bool              `json:"include_archived,omitempty"`
}

type ListPlansData struct {
	Plans []PlanSummary `json:"plans"`
}

type ArchivePlanRequest struct {
	PlanID            string `json:"plan_id"`
	Reason            string `json:"reason,omitempty"`
	CreateFinalExport bool   `json:"create_final_export,omitempty"`
}

type ArchivePlanData struct {
	Archived    bool        `json:"archived"`
	Plan        PlanSummary `json:"plan"`
	ArchivePath string      `json:"archive_path,omitempty"`
}

type ImportPlanRequest struct {
	Format  ImportFormat `json:"format"`
	Content string       `json:"content"`
	PlanID  string       `json:"plan_id,omitempty"`
	Mode    ImportMode   `json:"mode,omitempty"`
}

type ImportPlanData struct {
	Plan              PlanSummary              `json:"plan"`
	Issues            []domain.ValidationIssue `json:"issues"`
	ConflictsDetected bool                     `json:"conflicts_detected"`
}

type ExportPlanRequest struct {
	PlanID              string       `json:"plan_id"`
	Format              ExportFormat `json:"format"`
	IncludeAuditSummary bool         `json:"include_audit_summary,omitempty"`
}

type ExportPlanData struct {
	Format      ExportFormat `json:"format"`
	Content     string       `json:"content"`
	ContentPath string       `json:"content_path,omitempty"`
}

type ReconcilePlanRequest struct {
	PlanID          string        `json:"plan_id"`
	MarkdownContent string        `json:"markdown_content"`
	Mode            ReconcileMode `json:"mode,omitempty"`
}

type Conflict struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	TargetID string `json:"target_id,omitempty"`
}

type ReconcilePlanData struct {
	Reconciled bool        `json:"reconciled"`
	Conflicts  []Conflict  `json:"conflicts"`
	ChangedIDs []string    `json:"changed_ids"`
	Plan       PlanSummary `json:"plan"`
}

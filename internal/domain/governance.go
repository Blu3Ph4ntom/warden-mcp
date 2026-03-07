package domain

import (
	"regexp"
	"strings"
	"time"
)

var versionPattern = regexp.MustCompile(`^v?[0-9]+\.[0-9]+\.[0-9]+$`)

type VersionSnapshot struct {
	PlanID        string
	PlanVersion   string
	Revision      int
	SchemaVersion string
	RecordedAt    time.Time
}

type ArchiveRecord struct {
	PlanID          string
	PlanVersion     string
	ArchivedAt      time.Time
	Reason          string
	FinalExportPath string
}

type ResetTaskRequest struct {
	PlanID       string
	TaskID       string
	TargetStatus TaskStatus
	Reason       string
	RequestedAt  time.Time
}

type TaskClosureRequest struct {
	PlanID       string
	TaskID       string
	TargetStatus TaskStatus
	Reason       string
	ActorType    ActorType
	RequestedAt  time.Time
}

func ValidatePlanVersion(version string) bool {
	return versionPattern.MatchString(version)
}

func CanResetTask(current, target TaskStatus) bool {
	if current == target {
		return false
	}
	return target == TaskNotStarted || target == TaskInProgress
}

func RequiresClosureReason(target TaskStatus) bool {
	return target == TaskCancelled || target == TaskWaived
}

func (v VersionSnapshot) Validate() []ValidationIssue {
	issues := make([]ValidationIssue, 0)
	if !ValidatePlanID(v.PlanID) {
		issues = append(issues, issue("error", "PLAN_ID_INVALID", "version snapshot plan ID is invalid", "plan_id"))
	}
	if !ValidatePlanVersion(v.PlanVersion) {
		issues = append(issues, issue("error", "PLAN_VERSION_INVALID", "plan version must look like semver", "plan_version"))
	}
	if v.Revision < 1 {
		issues = append(issues, issue("error", "REVISION_INVALID", "revision must be at least 1", "revision"))
	}
	if strings.TrimSpace(v.SchemaVersion) == "" {
		issues = append(issues, issue("error", "SCHEMA_VERSION_MISSING", "schema version is required", "schema_version"))
	}
	if v.RecordedAt.IsZero() {
		issues = append(issues, issue("error", "RECORDED_AT_MISSING", "recorded_at is required", "recorded_at"))
	}
	return issues
}

func (a ArchiveRecord) Validate() []ValidationIssue {
	issues := make([]ValidationIssue, 0)
	if !ValidatePlanID(a.PlanID) {
		issues = append(issues, issue("error", "PLAN_ID_INVALID", "archive record plan ID is invalid", "plan_id"))
	}
	if !ValidatePlanVersion(a.PlanVersion) {
		issues = append(issues, issue("error", "PLAN_VERSION_INVALID", "plan version must look like semver", "plan_version"))
	}
	if a.ArchivedAt.IsZero() {
		issues = append(issues, issue("error", "ARCHIVED_AT_MISSING", "archived_at is required", "archived_at"))
	}
	return issues
}

func (r ResetTaskRequest) Validate() []ValidationIssue {
	issues := make([]ValidationIssue, 0)
	if !ValidatePlanID(r.PlanID) {
		issues = append(issues, issue("error", "PLAN_ID_INVALID", "reset request plan ID is invalid", "plan_id"))
	}
	if !ValidateTaskID(r.TaskID) {
		issues = append(issues, issue("error", "TASK_ID_INVALID", "reset request task ID is invalid", "task_id"))
	}
	if !CanResetTask(TaskDone, r.TargetStatus) {
		issues = append(issues, issue("error", "RESET_TARGET_INVALID", "reset target must be not_started or in_progress", "target_status"))
	}
	if strings.TrimSpace(r.Reason) == "" {
		issues = append(issues, issue("error", "RESET_REASON_MISSING", "reset reason is required", "reason"))
	}
	if r.RequestedAt.IsZero() {
		issues = append(issues, issue("error", "REQUESTED_AT_MISSING", "requested_at is required", "requested_at"))
	}
	return issues
}

func (r TaskClosureRequest) Validate() []ValidationIssue {
	issues := make([]ValidationIssue, 0)
	if !ValidatePlanID(r.PlanID) {
		issues = append(issues, issue("error", "PLAN_ID_INVALID", "closure request plan ID is invalid", "plan_id"))
	}
	if !ValidateTaskID(r.TaskID) {
		issues = append(issues, issue("error", "TASK_ID_INVALID", "closure request task ID is invalid", "task_id"))
	}
	if !RequiresClosureReason(r.TargetStatus) {
		issues = append(issues, issue("error", "CLOSURE_TARGET_INVALID", "closure target must be cancelled or waived", "target_status"))
	}
	if strings.TrimSpace(r.Reason) == "" {
		issues = append(issues, issue("error", "CLOSURE_REASON_MISSING", "closure reason is required", "reason"))
	}
	if !r.ActorType.Valid() {
		issues = append(issues, issue("error", "ACTOR_TYPE_INVALID", "actor type is invalid", "actor_type"))
	}
	if r.RequestedAt.IsZero() {
		issues = append(issues, issue("error", "REQUESTED_AT_MISSING", "requested_at is required", "requested_at"))
	}
	return issues
}

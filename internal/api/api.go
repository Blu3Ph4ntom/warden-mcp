package api

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"warden-mcp/internal/domain"
	"warden-mcp/internal/mcp/contracts"
	"warden-mcp/internal/observe"
	"warden-mcp/internal/planfile"
	"warden-mcp/internal/security"
	"warden-mcp/internal/service"
)

type API struct {
	WorkspaceRoot string
	Recorder      observe.Recorder
	Now           func() time.Time
}

func New(workspaceRoot string, recorder observe.Recorder) API {
	return API{WorkspaceRoot: workspaceRoot, Recorder: recorder, Now: func() time.Time { return time.Now().UTC() }}
}

func (a API) Status(planPath string, includeTasks bool) contracts.ToolResponseEnvelope[contracts.GetStatusData] {
	start := a.now()
	resolved, plan, warnings, errObj := a.loadPlan(planPath)
	if errObj != nil {
		a.record(observe.Event{Kind: "command", Command: "status", Accepted: observe.Accepted(false), DurationMS: observe.Since(start), Message: errObj.Message, ErrorCode: errObj.Code})
		return contracts.ToolResponseEnvelope[contracts.GetStatusData]{OK: false, Tool: "get_status", Timestamp: timestamp(a.now()), Warnings: warnings, Error: errObj}
	}
	data := service.GetStatus(plan, includeTasks)
	a.record(observe.Event{Kind: "command", Command: "status", PlanID: plan.PlanID, Accepted: observe.Accepted(true), DurationMS: observe.Since(start), Message: "plan status loaded", Fields: map[string]any{"plan_path": resolved}})
	return contracts.ToolResponseEnvelope[contracts.GetStatusData]{OK: true, Tool: "get_status", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Data: data}
}

func (a API) Health(planPath string) contracts.ToolResponseEnvelope[contracts.HealthCheckData] {
	start := a.now()
	checks := make([]contracts.HealthCheck, 0, 3)
	status := "ok"
	checks = append(checks, contracts.HealthCheck{Name: "workspace_root", Status: "ok", Message: "workspace root available"})
	resolved, err := security.ResolveWorkspacePath(a.WorkspaceRoot, planPath, ".md")
	if err != nil {
		status = combineHealthStatus(status, "failing")
		checks = append(checks, contracts.HealthCheck{Name: "plan_path", Status: "failing", Message: security.RedactSecretLikeText(err.Error())})
		data := contracts.HealthCheckData{Status: status, Checks: checks, CheckedAt: timestamp(a.now())}
		a.record(observe.Event{Kind: "command", Command: "health", Accepted: observe.Accepted(false), DurationMS: observe.Since(start), Message: "health check failed", ErrorCode: contracts.ErrPlanInvalid})
		return contracts.ToolResponseEnvelope[contracts.HealthCheckData]{OK: true, Tool: "health_check", Timestamp: timestamp(a.now()), Data: data}
	}
	checks = append(checks, contracts.HealthCheck{Name: "plan_path", Status: "ok", Message: "plan path resolved inside workspace"})
	plan, warnings, err := planfile.Load(resolved)
	if err != nil {
		checkStatus := "degraded"
		if errors.Is(err, planfile.ErrPlanTooLarge) || errors.Is(err, planfile.ErrPlanTooManyLines) {
			checkStatus = "failing"
		}
		status = combineHealthStatus(status, checkStatus)
		checks = append(checks, contracts.HealthCheck{Name: "plan_load", Status: checkStatus, Message: security.RedactSecretLikeText(err.Error())})
		data := contracts.HealthCheckData{Status: status, Checks: checks, CheckedAt: timestamp(a.now())}
		a.record(observe.Event{Kind: "command", Command: "health", Accepted: observe.Accepted(checkStatus == "ok"), DurationMS: observe.Since(start), Message: "health check completed", ErrorCode: contracts.ErrPlanInvalid})
		return contracts.ToolResponseEnvelope[contracts.HealthCheckData]{OK: true, Tool: "health_check", Timestamp: timestamp(a.now()), Warnings: warnings, Data: data}
	}
	checks = append(checks, contracts.HealthCheck{Name: "plan_load", Status: "ok", Message: "plan parsed successfully"})
	data := contracts.HealthCheckData{Status: status, PlanID: plan.PlanID, Checks: checks, CheckedAt: timestamp(a.now())}
	a.record(observe.Event{Kind: "command", Command: "health", PlanID: plan.PlanID, Accepted: observe.Accepted(true), DurationMS: observe.Since(start), Message: "health check completed"})
	return contracts.ToolResponseEnvelope[contracts.HealthCheckData]{OK: true, Tool: "health_check", Timestamp: timestamp(a.now()), Warnings: warnings, Data: data}
}

func (a API) Export(planPath string, req contracts.ExportPlanRequest, outputPath string) contracts.ToolResponseEnvelope[contracts.ExportPlanData] {
	start := a.now()
	resolved, plan, warnings, errObj := a.loadPlan(planPath)
	if errObj != nil {
		a.record(observe.Event{Kind: "command", Command: "export", Accepted: observe.Accepted(false), DurationMS: observe.Since(start), Message: errObj.Message, ErrorCode: errObj.Code})
		return contracts.ToolResponseEnvelope[contracts.ExportPlanData]{OK: false, Tool: "export_plan", Timestamp: timestamp(a.now()), Warnings: warnings, Error: errObj}
	}
	format := req.Format
	if format == "" {
		format = contracts.ExportMarkdown
	}
	content, errObj := buildExportContent(resolved, plan, format)
	if errObj != nil {
		a.record(observe.Event{Kind: "command", Command: "export", PlanID: plan.PlanID, Accepted: observe.Accepted(false), DurationMS: observe.Since(start), Message: errObj.Message, ErrorCode: errObj.Code})
		return contracts.ToolResponseEnvelope[contracts.ExportPlanData]{OK: false, Tool: "export_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	contentPath := ""
	if outputPath != "" {
		contentPath, errObj = a.writeExport(content, format, outputPath)
		if errObj != nil {
			a.record(observe.Event{Kind: "command", Command: "export", PlanID: plan.PlanID, Accepted: observe.Accepted(false), DurationMS: observe.Since(start), Message: errObj.Message, ErrorCode: errObj.Code})
			return contracts.ToolResponseEnvelope[contracts.ExportPlanData]{OK: false, Tool: "export_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
		}
	}
	data := contracts.ExportPlanData{Format: format, Content: content, ContentPath: contentPath}
	a.record(observe.Event{Kind: "command", Command: "export", PlanID: plan.PlanID, Accepted: observe.Accepted(true), DurationMS: observe.Since(start), Message: "plan exported", Fields: map[string]any{"format": format, "content_path": contentPath}})
	return contracts.ToolResponseEnvelope[contracts.ExportPlanData]{OK: true, Tool: "export_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Data: data}
}

func (a API) Next(planPath string, req contracts.GetNextTaskRequest) contracts.ToolResponseEnvelope[contracts.GetNextTaskData] {
	start := a.now()
	resolved, plan, warnings, errObj := a.loadPlan(planPath)
	if errObj != nil {
		a.record(observe.Event{Kind: "command", Command: "next", Accepted: observe.Accepted(false), DurationMS: observe.Since(start), Message: errObj.Message, ErrorCode: errObj.Code})
		return contracts.ToolResponseEnvelope[contracts.GetNextTaskData]{OK: false, Tool: "get_next_task", Timestamp: timestamp(a.now()), Warnings: warnings, Error: errObj}
	}
	if req.PlanID == "" {
		req.PlanID = plan.PlanID
	}
	data := service.GetNextTask(plan, req)
	a.record(observe.Event{Kind: "command", Command: "next", PlanID: plan.PlanID, TaskID: valueOrEmpty(data.NextTask), Accepted: observe.Accepted(!data.Blocked), DurationMS: observe.Since(start), Message: "next task evaluated", Fields: map[string]any{"plan_path": resolved}})
	return contracts.ToolResponseEnvelope[contracts.GetNextTaskData]{OK: true, Tool: "get_next_task", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Data: data}
}

func (a API) Finish(planPath string, req contracts.RequestFinishRequest) contracts.ToolResponseEnvelope[contracts.RequestFinishData] {
	start := a.now()
	resolved, plan, warnings, errObj := a.loadPlan(planPath)
	if errObj != nil {
		a.record(observe.Event{Kind: "command", Command: "finish", ActorType: string(req.ActorType), Accepted: observe.Accepted(false), DurationMS: observe.Since(start), Message: errObj.Message, ErrorCode: errObj.Code})
		return contracts.ToolResponseEnvelope[contracts.RequestFinishData]{OK: false, Tool: "request_finish", Timestamp: timestamp(a.now()), Warnings: warnings, Error: errObj}
	}
	if req.PlanID == "" {
		req.PlanID = plan.PlanID
	}
	data := service.RequestFinish(plan, req)
	a.record(observe.Event{Kind: "command", Command: "finish", PlanID: plan.PlanID, ActorType: string(req.ActorType), TaskID: data.RecommendedNextTaskID, Accepted: observe.Accepted(data.CanFinish), DurationMS: observe.Since(start), Message: "finish evaluated", ErrorCode: firstBlockingCode(data.BlockingReasons), Fields: map[string]any{"plan_path": resolved}})
	return contracts.ToolResponseEnvelope[contracts.RequestFinishData]{OK: true, Tool: "request_finish", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Data: data}
}

func (a API) Update(planPath string, req contracts.UpdateTaskRequest) contracts.ToolResponseEnvelope[contracts.UpdateTaskData] {
	start := a.now()
	resolved, plan, warnings, errObj := a.loadPlan(planPath)
	if errObj != nil {
		a.record(observe.Event{Kind: "command", Command: "update", TaskID: req.TaskID, ActorType: string(req.ActorType), Accepted: observe.Accepted(false), DurationMS: observe.Since(start), Message: errObj.Message, ErrorCode: errObj.Code})
		return contracts.ToolResponseEnvelope[contracts.UpdateTaskData]{OK: false, Tool: "update_task", Timestamp: timestamp(a.now()), Warnings: warnings, Error: errObj}
	}
	if req.TaskID == "" || req.Status == "" {
		errObj = errorObject(contracts.ErrTaskNotFound, "update requires task_id and status")
		a.record(observe.Event{Kind: "command", Command: "update", PlanID: plan.PlanID, TaskID: req.TaskID, ActorType: string(req.ActorType), Accepted: observe.Accepted(false), DurationMS: observe.Since(start), Message: errObj.Message, ErrorCode: errObj.Code, Fields: map[string]any{"plan_path": resolved}})
		return contracts.ToolResponseEnvelope[contracts.UpdateTaskData]{OK: false, Tool: "update_task", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	if req.PlanID == "" {
		req.PlanID = plan.PlanID
	}
	data, updateWarnings, err := service.UpdateTask(resolved, req)
	if err != nil {
		errObj = errorObject(contracts.ErrTaskTransitionInvalid, err.Error())
		a.record(observe.Event{Kind: "command", Command: "update", PlanID: plan.PlanID, TaskID: req.TaskID, ActorType: string(req.ActorType), Accepted: observe.Accepted(false), DurationMS: observe.Since(start), Message: errObj.Message, ErrorCode: errObj.Code, Fields: map[string]any{"plan_path": resolved}})
		return contracts.ToolResponseEnvelope[contracts.UpdateTaskData]{OK: false, Tool: "update_task", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	warnings = dedupeWarnings(append(warnings, updateWarnings...))
	a.record(observe.Event{Kind: "command", Command: "update", PlanID: plan.PlanID, PhaseID: data.Phase.PhaseID, TaskID: req.TaskID, ActorType: string(req.ActorType), Accepted: observe.Accepted(true), DurationMS: observe.Since(start), Message: "task update accepted", Fields: map[string]any{"plan_path": resolved, "status": req.Status}})
	return contracts.ToolResponseEnvelope[contracts.UpdateTaskData]{OK: true, Tool: "update_task", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Data: data}
}

func (a API) loadPlan(planPath string) (string, domain.Plan, []domain.ValidationIssue, *contracts.ErrorObject) {
	resolved, err := security.ResolveWorkspacePath(a.WorkspaceRoot, planPath, ".md")
	if err != nil {
		return "", domain.Plan{}, nil, errorObject(contracts.ErrPlanInvalid, err.Error())
	}
	plan, warnings, err := planfile.Load(resolved)
	if err != nil {
		if errors.Is(err, planfile.ErrPlanTooLarge) || errors.Is(err, planfile.ErrPlanTooManyLines) {
			return resolved, domain.Plan{}, warnings, errorObject(contracts.ErrPlanInvalid, err.Error())
		}
		return resolved, domain.Plan{}, warnings, errorObject(contracts.ErrPlanNotFound, err.Error())
	}
	return resolved, plan, warnings, nil
}

func (a API) record(event observe.Event) {
	if a.Recorder != nil {
		a.Recorder.Record(event)
	}
}

func (a API) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now().UTC()
}

func timestamp(now time.Time) string {
	return now.UTC().Format(time.RFC3339)
}

func errorObject(code, message string) *contracts.ErrorObject {
	return &contracts.ErrorObject{Code: code, Message: security.RedactSecretLikeText(message), Retryable: false, Details: map[string]any{}}
}

func firstBlockingCode(blocking []contracts.BlockingReason) string {
	if len(blocking) == 0 {
		return ""
	}
	return blocking[0].Code
}

func valueOrEmpty(task *contracts.TaskSummary) string {
	if task == nil {
		return ""
	}
	return task.TaskID
}

func combineHealthStatus(current, next string) string {
	rank := map[string]int{"ok": 0, "degraded": 1, "failing": 2}
	if rank[next] > rank[current] {
		return next
	}
	return current
}

func buildExportContent(resolvedPlanPath string, plan domain.Plan, format contracts.ExportFormat) (string, *contracts.ErrorObject) {
	switch format {
	case contracts.ExportMarkdown:
		content, err := os.ReadFile(resolvedPlanPath)
		if err != nil {
			return "", errorObject(contracts.ErrExportFailed, err.Error())
		}
		return string(content), nil
	case contracts.ExportJSON:
		content, err := json.MarshalIndent(plan, "", "  ")
		if err != nil {
			return "", errorObject(contracts.ErrExportFailed, err.Error())
		}
		return string(content), nil
	default:
		return "", errorObject(contracts.ErrExportFailed, "unsupported export format")
	}
}

func (a API) writeExport(content string, format contracts.ExportFormat, outputPath string) (string, *contracts.ErrorObject) {
	exts := []string{".md", ".json", ".txt"}
	if format == contracts.ExportJSON {
		exts = []string{".json", ".txt"}
	}
	resolved, err := security.ResolveWorkspacePath(a.WorkspaceRoot, outputPath, exts...)
	if err != nil {
		return "", errorObject(contracts.ErrExportFailed, err.Error())
	}
	if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
		return "", errorObject(contracts.ErrExportFailed, err.Error())
	}
	if err := os.WriteFile(resolved, []byte(content), 0o644); err != nil {
		return "", errorObject(contracts.ErrExportFailed, err.Error())
	}
	return resolved, nil
}

func dedupeWarnings(warnings []domain.ValidationIssue) []domain.ValidationIssue {
	if len(warnings) < 2 {
		return warnings
	}
	seen := make(map[string]struct{}, len(warnings))
	result := make([]domain.ValidationIssue, 0, len(warnings))
	for _, warning := range warnings {
		key := warning.Severity + "|" + warning.Code + "|" + warning.Message + "|" + warning.Path
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, warning)
	}
	return result
}

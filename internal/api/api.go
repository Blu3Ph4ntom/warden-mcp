package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

func (a API) Init(req contracts.InitPlanRequest) contracts.ToolResponseEnvelope[contracts.InitPlanData] {
	start := a.now()
	plan, warnings := a.planFromInit(req)
	issues := dedupeWarnings(append(append([]domain.ValidationIssue{}, warnings...), plan.Validate()...))
	data := contracts.InitPlanData{Plan: service.PlanSummary(plan, plan.CanFinish), Phases: service.PhaseSummaries(plan.Phases), Tasks: service.TaskSummaries(plan.Phases, true), ValidationIssues: issues, Normalized: len(warnings) > 0}
	if hasErrorIssues(issues) {
		a.record(observe.Event{Kind: "command", Command: "init", PlanID: plan.PlanID, Accepted: observe.Accepted(false), DurationMS: observe.Since(start), Message: "plan init rejected", ErrorCode: contracts.ErrPlanInvalid})
		return contracts.ToolResponseEnvelope[contracts.InitPlanData]{OK: true, Tool: "init_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Data: data}
	}
	resolved, err := security.ResolveWorkspacePath(a.WorkspaceRoot, defaultPlanPath(), ".md")
	if err != nil {
		errObj := errorObject(contracts.ErrPlanInvalid, err.Error())
		return contracts.ToolResponseEnvelope[contracts.InitPlanData]{OK: false, Tool: "init_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	if _, err := os.Stat(resolved); err == nil {
		errObj := errorObject(contracts.ErrSyncConflict, "active plan already exists")
		return contracts.ToolResponseEnvelope[contracts.InitPlanData]{OK: false, Tool: "init_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
		errObj := errorObject(contracts.ErrInternal, err.Error())
		return contracts.ToolResponseEnvelope[contracts.InitPlanData]{OK: false, Tool: "init_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	if err := os.WriteFile(resolved, []byte(planfile.Render(plan)), 0o644); err != nil {
		errObj := errorObject(contracts.ErrInternal, err.Error())
		return contracts.ToolResponseEnvelope[contracts.InitPlanData]{OK: false, Tool: "init_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	a.record(observe.Event{Kind: "command", Command: "init", PlanID: plan.PlanID, Accepted: observe.Accepted(true), DurationMS: observe.Since(start), Message: "plan initialized", Fields: map[string]any{"plan_path": resolved}})
	return contracts.ToolResponseEnvelope[contracts.InitPlanData]{OK: true, Tool: "init_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Data: data}
}

func (a API) List(req contracts.ListPlansRequest) contracts.ToolResponseEnvelope[contracts.ListPlansData] {
	start := a.now()
	root := filepath.Join(a.WorkspaceRoot, ".agent")
	plans := make([]contracts.PlanSummary, 0)
	warnings := make([]domain.ValidationIssue, 0)
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			warnings = append(warnings, domain.ValidationIssue{Severity: "warning", Code: "PLAN_SCAN_SKIPPED", Message: security.RedactSecretLikeText(err.Error()), Path: path})
			return nil
		}
		if entry.IsDir() {
			if !req.IncludeArchived && strings.Contains(path, string(filepath.Separator)+"archive") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		if !strings.EqualFold(entry.Name(), "PLAN.md") && !strings.Contains(path, string(filepath.Separator)+"archive"+string(filepath.Separator)) {
			return nil
		}
		plan, planWarnings, loadErr := planfile.Load(path)
		warnings = append(warnings, planWarnings...)
		if loadErr != nil {
			warnings = append(warnings, domain.ValidationIssue{Severity: "warning", Code: "PLAN_SCAN_SKIPPED", Message: security.RedactSecretLikeText(loadErr.Error()), Path: path})
			return nil
		}
		if strings.Contains(path, string(filepath.Separator)+"archive"+string(filepath.Separator)) {
			plan.Status = domain.PlanArchived
		}
		if req.Status != "" && plan.Status != req.Status {
			return nil
		}
		plans = append(plans, service.PlanSummary(plan, plan.CanFinish))
		return nil
	})
	sort.Slice(plans, func(i, j int) bool { return plans[i].PlanID < plans[j].PlanID })
	warnings = dedupeWarnings(warnings)
	a.record(observe.Event{Kind: "command", Command: "list", Accepted: observe.Accepted(true), DurationMS: observe.Since(start), Message: "plans listed", Fields: map[string]any{"count": len(plans)}})
	return contracts.ToolResponseEnvelope[contracts.ListPlansData]{OK: true, Tool: "list_plans", Timestamp: timestamp(a.now()), Warnings: warnings, Data: contracts.ListPlansData{Plans: plans}}
}

func (a API) Import(req contracts.ImportPlanRequest) contracts.ToolResponseEnvelope[contracts.ImportPlanData] {
	start := a.now()
	plan, warnings, errObj := a.planFromImport(req)
	if errObj != nil {
		a.record(observe.Event{Kind: "command", Command: "import", Accepted: observe.Accepted(false), DurationMS: observe.Since(start), Message: errObj.Message, ErrorCode: errObj.Code})
		return contracts.ToolResponseEnvelope[contracts.ImportPlanData]{OK: false, Tool: "import_plan", Timestamp: timestamp(a.now()), Warnings: warnings, Error: errObj}
	}
	issues := dedupeWarnings(append(append([]domain.ValidationIssue{}, warnings...), plan.Validate()...))
	data := contracts.ImportPlanData{Plan: service.PlanSummary(plan, plan.CanFinish), Issues: issues, ConflictsDetected: false}
	if hasErrorIssues(issues) {
		a.record(observe.Event{Kind: "command", Command: "import", PlanID: plan.PlanID, Accepted: observe.Accepted(false), DurationMS: observe.Since(start), Message: "plan import rejected", ErrorCode: contracts.ErrImportInvalid})
		return contracts.ToolResponseEnvelope[contracts.ImportPlanData]{OK: true, Tool: "import_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Data: data}
	}
	mode := req.Mode
	if mode == "" {
		mode = contracts.ImportReplace
	}
	resolved, err := security.ResolveWorkspacePath(a.WorkspaceRoot, defaultPlanPath(), ".md")
	if err != nil {
		errObj = errorObject(contracts.ErrImportInvalid, err.Error())
		return contracts.ToolResponseEnvelope[contracts.ImportPlanData]{OK: false, Tool: "import_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	_, statErr := os.Stat(resolved)
	if mode == contracts.ImportCreate && statErr == nil {
		errObj = errorObject(contracts.ErrImportInvalid, "active plan already exists")
		return contracts.ToolResponseEnvelope[contracts.ImportPlanData]{OK: false, Tool: "import_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	if mode == contracts.ImportMerge && statErr == nil {
		errObj = errorObject(contracts.ErrImportInvalid, "merge mode is not implemented yet")
		return contracts.ToolResponseEnvelope[contracts.ImportPlanData]{OK: false, Tool: "import_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
		errObj = errorObject(contracts.ErrInternal, err.Error())
		return contracts.ToolResponseEnvelope[contracts.ImportPlanData]{OK: false, Tool: "import_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	if err := os.WriteFile(resolved, []byte(planfile.Render(plan)), 0o644); err != nil {
		errObj = errorObject(contracts.ErrInternal, err.Error())
		return contracts.ToolResponseEnvelope[contracts.ImportPlanData]{OK: false, Tool: "import_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	a.record(observe.Event{Kind: "command", Command: "import", PlanID: plan.PlanID, Accepted: observe.Accepted(true), DurationMS: observe.Since(start), Message: "plan imported", Fields: map[string]any{"mode": mode}})
	return contracts.ToolResponseEnvelope[contracts.ImportPlanData]{OK: true, Tool: "import_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Data: data}
}

func (a API) Archive(req contracts.ArchivePlanRequest) contracts.ToolResponseEnvelope[contracts.ArchivePlanData] {
	start := a.now()
	resolved, plan, warnings, errObj := a.loadPlan(defaultPlanPath())
	if errObj != nil {
		a.record(observe.Event{Kind: "command", Command: "archive", Accepted: observe.Accepted(false), DurationMS: observe.Since(start), Message: errObj.Message, ErrorCode: errObj.Code})
		return contracts.ToolResponseEnvelope[contracts.ArchivePlanData]{OK: false, Tool: "archive_plan", Timestamp: timestamp(a.now()), Warnings: warnings, Error: errObj}
	}
	if req.PlanID != "" && req.PlanID != plan.PlanID {
		errObj = errorObject(contracts.ErrPlanNotFound, "requested plan does not match active plan")
		return contracts.ToolResponseEnvelope[contracts.ArchivePlanData]{OK: false, Tool: "archive_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	if !service.RequestFinish(plan, contracts.RequestFinishRequest{PlanID: plan.PlanID}).CanFinish {
		errObj = errorObject(contracts.ErrArchiveDenied, "plan cannot be archived until finish checks pass")
		return contracts.ToolResponseEnvelope[contracts.ArchivePlanData]{OK: false, Tool: "archive_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	archived := plan
	archived.Status = domain.PlanArchived
	archived.CanFinish = true
	archiveRel := filepath.Join(".agent", "archive", fmt.Sprintf("%s-%s.md", archived.PlanID, a.now().Format("20060102-150405")))
	archivePath, err := security.ResolveWorkspacePath(a.WorkspaceRoot, archiveRel, ".md")
	if err != nil {
		errObj = errorObject(contracts.ErrArchiveDenied, err.Error())
		return contracts.ToolResponseEnvelope[contracts.ArchivePlanData]{OK: false, Tool: "archive_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	if err := os.MkdirAll(filepath.Dir(archivePath), 0o755); err != nil {
		errObj = errorObject(contracts.ErrInternal, err.Error())
		return contracts.ToolResponseEnvelope[contracts.ArchivePlanData]{OK: false, Tool: "archive_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	if err := os.WriteFile(archivePath, []byte(planfile.Render(archived)), 0o644); err != nil {
		errObj = errorObject(contracts.ErrInternal, err.Error())
		return contracts.ToolResponseEnvelope[contracts.ArchivePlanData]{OK: false, Tool: "archive_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	if req.CreateFinalExport {
		jsonContent, exportErr := buildExportContent(resolved, archived, contracts.ExportJSON)
		if exportErr == nil {
			_ = os.WriteFile(strings.TrimSuffix(archivePath, ".md")+".json", []byte(jsonContent), 0o644)
		}
	}
	if err := os.Remove(resolved); err != nil {
		errObj = errorObject(contracts.ErrInternal, err.Error())
		return contracts.ToolResponseEnvelope[contracts.ArchivePlanData]{OK: false, Tool: "archive_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	data := contracts.ArchivePlanData{Archived: true, Plan: service.PlanSummary(archived, true), ArchivePath: archiveRel}
	a.record(observe.Event{Kind: "command", Command: "archive", PlanID: plan.PlanID, Accepted: observe.Accepted(true), DurationMS: observe.Since(start), Message: "plan archived", Fields: map[string]any{"archive_path": archiveRel}})
	return contracts.ToolResponseEnvelope[contracts.ArchivePlanData]{OK: true, Tool: "archive_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Data: data}
}

func (a API) Status(planPath string, req contracts.GetStatusRequest) contracts.ToolResponseEnvelope[contracts.GetStatusData] {
	start := a.now()
	resolved, plan, warnings, errObj := a.loadPlan(planPath)
	if errObj != nil {
		a.record(observe.Event{Kind: "command", Command: "status", Accepted: observe.Accepted(false), DurationMS: observe.Since(start), Message: errObj.Message, ErrorCode: errObj.Code})
		return contracts.ToolResponseEnvelope[contracts.GetStatusData]{OK: false, Tool: "get_status", Timestamp: timestamp(a.now()), Warnings: warnings, Error: errObj}
	}
	if errObj = requirePlanIdentity(plan, req.PlanID); errObj != nil {
		return contracts.ToolResponseEnvelope[contracts.GetStatusData]{OK: false, Tool: "get_status", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	data := service.GetStatus(plan, req.IncludeTasks)
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

func (a API) Validate(planPath string, req contracts.ValidatePlanRequest) contracts.ToolResponseEnvelope[contracts.ValidatePlanData] {
	start := a.now()
	resolved, plan, warnings, errObj := a.loadPlan(planPath)
	if errObj != nil {
		a.record(observe.Event{Kind: "command", Command: "validate", Accepted: observe.Accepted(false), DurationMS: observe.Since(start), Message: errObj.Message, ErrorCode: errObj.Code})
		return contracts.ToolResponseEnvelope[contracts.ValidatePlanData]{OK: false, Tool: "validate_plan", Timestamp: timestamp(a.now()), Warnings: warnings, Error: errObj}
	}
	issues := append([]domain.ValidationIssue{}, warnings...)
	issues = append(issues, plan.Validate()...)
	data := contracts.ValidatePlanData{Valid: !hasErrorIssues(issues), Issues: issues, NormalizedCounts: contracts.NormalizedCounts{PhaseCount: len(plan.Phases), TaskCount: plan.TotalTasks()}}
	a.record(observe.Event{Kind: "command", Command: "validate", PlanID: plan.PlanID, Accepted: observe.Accepted(data.Valid), DurationMS: observe.Since(start), Message: "plan validated", Fields: map[string]any{"plan_path": resolved, "mode": req.Mode}})
	return contracts.ToolResponseEnvelope[contracts.ValidatePlanData]{OK: true, Tool: "validate_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Data: data}
}

func (a API) Next(planPath string, req contracts.GetNextTaskRequest) contracts.ToolResponseEnvelope[contracts.GetNextTaskData] {
	start := a.now()
	resolved, plan, warnings, errObj := a.loadPlan(planPath)
	if errObj != nil {
		a.record(observe.Event{Kind: "command", Command: "next", Accepted: observe.Accepted(false), DurationMS: observe.Since(start), Message: errObj.Message, ErrorCode: errObj.Code})
		return contracts.ToolResponseEnvelope[contracts.GetNextTaskData]{OK: false, Tool: "get_next_task", Timestamp: timestamp(a.now()), Warnings: warnings, Error: errObj}
	}
	if errObj = requirePlanIdentity(plan, req.PlanID); errObj != nil {
		return contracts.ToolResponseEnvelope[contracts.GetNextTaskData]{OK: false, Tool: "get_next_task", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
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
	if errObj = requirePlanIdentity(plan, req.PlanID); errObj != nil {
		return contracts.ToolResponseEnvelope[contracts.RequestFinishData]{OK: false, Tool: "request_finish", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
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
	if errObj = requirePlanIdentity(plan, req.PlanID); errObj != nil {
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

func (a API) Reset(planPath string, req contracts.ResetTaskRequest) contracts.ToolResponseEnvelope[contracts.ResetTaskData] {
	start := a.now()
	resolved, plan, warnings, errObj := a.loadPlan(planPath)
	if errObj != nil {
		a.record(observe.Event{Kind: "command", Command: "reset", TaskID: req.TaskID, Accepted: observe.Accepted(false), DurationMS: observe.Since(start), Message: errObj.Message, ErrorCode: errObj.Code})
		return contracts.ToolResponseEnvelope[contracts.ResetTaskData]{OK: false, Tool: "reset_task", Timestamp: timestamp(a.now()), Warnings: warnings, Error: errObj}
	}
	if req.TaskID == "" {
		errObj = errorObject(contracts.ErrTaskNotFound, "reset requires task_id")
		return contracts.ToolResponseEnvelope[contracts.ResetTaskData]{OK: false, Tool: "reset_task", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	if errObj = requirePlanIdentity(plan, req.PlanID); errObj != nil {
		return contracts.ToolResponseEnvelope[contracts.ResetTaskData]{OK: false, Tool: "reset_task", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	if req.PlanID == "" {
		req.PlanID = plan.PlanID
	}
	data, resetWarnings, err := service.ResetTask(resolved, req)
	if err != nil {
		errObj = errorObject(contracts.ErrTaskTransitionInvalid, err.Error())
		return contracts.ToolResponseEnvelope[contracts.ResetTaskData]{OK: false, Tool: "reset_task", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: dedupeWarnings(append(warnings, resetWarnings...)), Error: errObj}
	}
	warnings = dedupeWarnings(append(warnings, resetWarnings...))
	a.record(observe.Event{Kind: "command", Command: "reset", PlanID: plan.PlanID, TaskID: req.TaskID, Accepted: observe.Accepted(true), DurationMS: observe.Since(start), Message: "task reset accepted", Fields: map[string]any{"plan_path": resolved, "status": req.Status}})
	return contracts.ToolResponseEnvelope[contracts.ResetTaskData]{OK: true, Tool: "reset_task", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Data: data}
}

func (a API) Prioritize(planPath string, req contracts.PrioritizeTasksRequest) contracts.ToolResponseEnvelope[contracts.PrioritizeTasksData] {
	start := a.now()
	resolved, plan, warnings, errObj := a.loadPlan(planPath)
	if errObj != nil {
		a.record(observe.Event{Kind: "command", Command: "prioritize", Accepted: observe.Accepted(false), DurationMS: observe.Since(start), Message: errObj.Message, ErrorCode: errObj.Code})
		return contracts.ToolResponseEnvelope[contracts.PrioritizeTasksData]{OK: false, Tool: "prioritize_tasks", Timestamp: timestamp(a.now()), Warnings: warnings, Error: errObj}
	}
	if errObj = requirePlanIdentity(plan, req.PlanID); errObj != nil {
		return contracts.ToolResponseEnvelope[contracts.PrioritizeTasksData]{OK: false, Tool: "prioritize_tasks", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	if req.PlanID == "" {
		req.PlanID = plan.PlanID
	}
	data, updateWarnings, err := service.PrioritizeTasks(resolved, req)
	if err != nil {
		errObj = errorObject(contracts.ErrTaskTransitionInvalid, err.Error())
		return contracts.ToolResponseEnvelope[contracts.PrioritizeTasksData]{OK: false, Tool: "prioritize_tasks", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: dedupeWarnings(append(warnings, updateWarnings...)), Error: errObj}
	}
	warnings = dedupeWarnings(append(warnings, updateWarnings...))
	a.record(observe.Event{Kind: "command", Command: "prioritize", PlanID: plan.PlanID, Accepted: observe.Accepted(true), DurationMS: observe.Since(start), Message: "task priorities updated", Fields: map[string]any{"plan_path": resolved, "updated_count": len(data.UpdatedTaskIDs)}})
	return contracts.ToolResponseEnvelope[contracts.PrioritizeTasksData]{OK: true, Tool: "prioritize_tasks", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Data: data}
}

func (a API) Reconcile(planPath string, req contracts.ReconcilePlanRequest) contracts.ToolResponseEnvelope[contracts.ReconcilePlanData] {
	start := a.now()
	resolved, plan, warnings, errObj := a.loadPlan(planPath)
	if errObj != nil {
		a.record(observe.Event{Kind: "command", Command: "reconcile", Accepted: observe.Accepted(false), DurationMS: observe.Since(start), Message: errObj.Message, ErrorCode: errObj.Code})
		return contracts.ToolResponseEnvelope[contracts.ReconcilePlanData]{OK: false, Tool: "reconcile_plan", Timestamp: timestamp(a.now()), Warnings: warnings, Error: errObj}
	}
	if errObj = requirePlanIdentity(plan, req.PlanID); errObj != nil {
		return contracts.ToolResponseEnvelope[contracts.ReconcilePlanData]{OK: false, Tool: "reconcile_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	if req.PlanID == "" {
		req.PlanID = plan.PlanID
	}
	data, reconcileWarnings, err := service.ReconcilePlan(resolved, req)
	if err != nil {
		errObj = errorObject(contracts.ErrSyncConflict, err.Error())
		return contracts.ToolResponseEnvelope[contracts.ReconcilePlanData]{OK: false, Tool: "reconcile_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: dedupeWarnings(append(warnings, reconcileWarnings...)), Error: errObj}
	}
	warnings = dedupeWarnings(append(warnings, reconcileWarnings...))
	a.record(observe.Event{Kind: "command", Command: "reconcile", PlanID: plan.PlanID, Accepted: observe.Accepted(len(data.Conflicts) == 0), DurationMS: observe.Since(start), Message: "plan reconciliation evaluated", Fields: map[string]any{"plan_path": resolved, "mode": req.Mode, "changed_count": len(data.ChangedIDs), "conflict_count": len(data.Conflicts)}})
	return contracts.ToolResponseEnvelope[contracts.ReconcilePlanData]{OK: true, Tool: "reconcile_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Data: data}
}

func (a API) Edit(planPath string, req contracts.EditPlanRequest) contracts.ToolResponseEnvelope[contracts.EditPlanData] {
	start := a.now()
	resolved, plan, warnings, errObj := a.loadPlan(planPath)
	if errObj != nil {
		a.record(observe.Event{Kind: "command", Command: "edit", Accepted: observe.Accepted(false), DurationMS: observe.Since(start), Message: errObj.Message, ErrorCode: errObj.Code})
		return contracts.ToolResponseEnvelope[contracts.EditPlanData]{OK: false, Tool: "edit_plan", Timestamp: timestamp(a.now()), Warnings: warnings, Error: errObj}
	}
	if errObj = requirePlanIdentity(plan, req.PlanID); errObj != nil {
		return contracts.ToolResponseEnvelope[contracts.EditPlanData]{OK: false, Tool: "edit_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Error: errObj}
	}
	if req.PlanID == "" {
		req.PlanID = plan.PlanID
	}
	data, editWarnings, err := service.EditPlan(resolved, req)
	if err != nil {
		errObj = errorObject(contracts.ErrPlanInvalid, err.Error())
		return contracts.ToolResponseEnvelope[contracts.EditPlanData]{OK: false, Tool: "edit_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: dedupeWarnings(append(warnings, editWarnings...)), Error: errObj}
	}
	warnings = dedupeWarnings(append(warnings, editWarnings...))
	a.record(observe.Event{Kind: "command", Command: "edit", PlanID: plan.PlanID, Accepted: observe.Accepted(true), DurationMS: observe.Since(start), Message: "plan edit applied", Fields: map[string]any{"plan_path": resolved, "operation": req.Operation, "changed_count": len(data.ChangedIDs)}})
	return contracts.ToolResponseEnvelope[contracts.EditPlanData]{OK: true, Tool: "edit_plan", Timestamp: timestamp(a.now()), PlanID: plan.PlanID, Warnings: warnings, Data: data}
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

func requirePlanIdentity(plan domain.Plan, requested string) *contracts.ErrorObject {
	if requested == "" || requested == plan.PlanID {
		return nil
	}
	return errorObject(contracts.ErrSyncConflict, "requested plan_id does not match active plan")
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

func hasErrorIssues(issues []domain.ValidationIssue) bool {
	for _, issue := range issues {
		if issue.Severity == "error" {
			return true
		}
	}
	return false
}

func defaultPlanPath() string {
	return filepath.Join(".agent", "PLAN.md")
}

func (a API) planFromInit(req contracts.InitPlanRequest) (domain.Plan, []domain.ValidationIssue) {
	planID := strings.TrimSpace(req.PlanID)
	if planID == "" {
		planID = slugify(req.Title)
	}
	version, warnings := normalizeVersion(req.Version)
	if version == "" {
		version = "1.0.0"
	}
	plan := domain.Plan{PlanID: planID, Title: strings.TrimSpace(req.Title), Status: domain.PlanActive, Version: version, UpdatedAt: a.now(), Phases: make([]domain.Phase, 0, len(req.Phases))}
	for phaseIndex, phaseInput := range req.Phases {
		phaseID := fmt.Sprintf("PH%02d", phaseIndex+1)
		phase := domain.Phase{PhaseID: phaseID, Title: strings.TrimSpace(phaseInput.Title), Tasks: make([]domain.Task, 0, len(phaseInput.Tasks))}
		for taskIndex, taskInput := range phaseInput.Tasks {
			required := true
			if taskInput.Required != nil {
				required = *taskInput.Required
			}
			priority := taskInput.Priority
			if priority == "" {
				priority = domain.PriorityP2
			}
			phase.Tasks = append(phase.Tasks, domain.Task{TaskID: fmt.Sprintf("%s-T%02d", phaseID, taskIndex+1), PhaseID: phaseID, Title: strings.TrimSpace(taskInput.Title), Status: domain.TaskNotStarted, Priority: priority, DependsOn: append([]string(nil), taskInput.DependsOn...), Required: required, UpdatedAt: a.now()})
		}
		phase.Status = derivePhaseStatus(phase)
		plan.Phases = append(plan.Phases, phase)
	}
	return normalizePlan(plan), warnings
}

func (a API) planFromImport(req contracts.ImportPlanRequest) (domain.Plan, []domain.ValidationIssue, *contracts.ErrorObject) {
	if strings.TrimSpace(req.Content) == "" {
		return domain.Plan{}, nil, errorObject(contracts.ErrImportInvalid, "import content is required")
	}
	var (
		plan     domain.Plan
		warnings []domain.ValidationIssue
		err      error
	)
	switch req.Format {
	case contracts.ImportMarkdown:
		plan, warnings, err = planfile.Parse(req.Content, a.now())
	case contracts.ImportJSON:
		err = json.Unmarshal([]byte(req.Content), &plan)
	default:
		return domain.Plan{}, nil, errorObject(contracts.ErrImportInvalid, "unsupported import format")
	}
	if err != nil {
		return domain.Plan{}, warnings, errorObject(contracts.ErrImportInvalid, err.Error())
	}
	if req.PlanID != "" {
		if plan.PlanID != "" && plan.PlanID != req.PlanID {
			return domain.Plan{}, warnings, errorObject(contracts.ErrImportInvalid, "plan_id does not match imported content")
		}
		plan.PlanID = req.PlanID
	}
	return normalizePlan(plan), warnings, nil
}

func normalizePlan(plan domain.Plan) domain.Plan {
	if plan.Status == "" {
		plan.Status = domain.PlanActive
	}
	if plan.Version == "" {
		plan.Version = "1.0.0"
	}
	for phaseIndex := range plan.Phases {
		if plan.Phases[phaseIndex].PhaseID == "" {
			plan.Phases[phaseIndex].PhaseID = fmt.Sprintf("PH%02d", phaseIndex+1)
		}
		if strings.TrimSpace(plan.Phases[phaseIndex].Title) == "" {
			plan.Phases[phaseIndex].Title = plan.Phases[phaseIndex].PhaseID
		}
		for taskIndex := range plan.Phases[phaseIndex].Tasks {
			task := &plan.Phases[phaseIndex].Tasks[taskIndex]
			if task.TaskID == "" {
				task.TaskID = fmt.Sprintf("%s-T%02d", plan.Phases[phaseIndex].PhaseID, taskIndex+1)
			}
			if task.PhaseID == "" {
				task.PhaseID = plan.Phases[phaseIndex].PhaseID
			}
			if task.Status == "" {
				task.Status = domain.TaskNotStarted
			}
			if task.Priority == "" {
				task.Priority = domain.PriorityP2
			}
			if task.UpdatedAt.IsZero() {
				task.UpdatedAt = time.Now().UTC()
			}
		}
		plan.Phases[phaseIndex].Status = derivePhaseStatus(plan.Phases[phaseIndex])
	}
	if plan.UpdatedAt.IsZero() {
		plan.UpdatedAt = time.Now().UTC()
	}
	if plan.Status != domain.PlanArchived {
		plan.Status = derivePlanStatus(plan)
	}
	plan.CurrentPhaseID = nextCurrentPhaseID(plan)
	plan.CanFinish = calculateCanFinish(plan)
	return plan
}

func derivePlanStatus(plan domain.Plan) domain.PlanStatus {
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

func derivePhaseStatus(phase domain.Phase) domain.PhaseStatus {
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

func calculateCanFinish(plan domain.Plan) bool {
	for _, phase := range plan.Phases {
		for _, task := range phase.Tasks {
			if task.Required && !task.IsTerminal() {
				return false
			}
		}
	}
	return true
}

func nextCurrentPhaseID(plan domain.Plan) string {
	for _, phase := range plan.Phases {
		if phase.Status != domain.PhaseCompleted {
			return phase.PhaseID
		}
	}
	if len(plan.Phases) > 0 {
		return plan.Phases[len(plan.Phases)-1].PhaseID
	}
	return ""
}

func normalizeVersion(version string) (string, []domain.ValidationIssue) {
	version = strings.TrimSpace(version)
	if version == "" {
		return "", nil
	}
	if strings.Count(version, ".") == 1 {
		return version + ".0", []domain.ValidationIssue{{Severity: "warning", Code: "PLAN_VERSION_NORMALIZED", Message: "two-part version normalized to semver", Path: "version"}}
	}
	return version, nil
}

func slugify(title string) string {
	value := strings.ToLower(strings.TrimSpace(title))
	var builder strings.Builder
	lastDash := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastDash = false
		case !lastDash:
			builder.WriteRune('-')
			lastDash = true
		}
	}
	result := strings.Trim(builder.String(), "-")
	if len(result) < 3 {
		return "new-plan"
	}
	if len(result) > 64 {
		return result[:64]
	}
	return result
}

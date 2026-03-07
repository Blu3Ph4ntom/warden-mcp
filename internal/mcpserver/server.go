package mcpserver

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Blu3Ph4ntom/warden-mcp/internal/api"
	"github.com/Blu3Ph4ntom/warden-mcp/internal/domain"
	"github.com/Blu3Ph4ntom/warden-mcp/internal/mcp/contracts"
	"github.com/Blu3Ph4ntom/warden-mcp/internal/observe"
)

const ProtocolVersion = "2025-11-25"
const AgentGuideVersion = "2026-03-07"
const ServerVersion = "0.1.5"

var supportedProtocolVersions = []string{
	"2024-11-05",
	"2025-03-26",
	"2025-06-18",
	ProtocolVersion,
}

type Server struct {
	API         api.API
	Recorder    observe.Recorder
	initialized bool
	negotiated  bool
	protocol    string
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *responseError  `json:"error,omitempty"`
}

type responseError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type transportMode int

const (
	transportModeUnknown transportMode = iota
	transportModeFramed
	transportModeLineDelimited
)

type initializeParams struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities,omitempty"`
	ClientInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"clientInfo,omitempty"`
}

type toolsCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

func (s *Server) Serve(reader io.Reader, writer io.Writer) error {
	buffered := bufio.NewReader(reader)
	mode := transportModeUnknown
	for {
		payload, detectedMode, err := readFrame(buffered)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if mode == transportModeUnknown {
			mode = detectedMode
		}
		var req request
		if err := json.Unmarshal(payload, &req); err != nil {
			if err := writeFrame(writer, response{JSONRPC: "2.0", Error: &responseError{Code: -32700, Message: "parse error"}}, mode); err != nil {
				return err
			}
			continue
		}
		resp := s.handle(req)
		if len(req.ID) == 0 || resp == nil {
			continue
		}
		if err := writeFrame(writer, resp, mode); err != nil {
			return err
		}
	}
}

func (s *Server) handle(req request) *response {
	s.record(observe.Event{Kind: "mcp", Method: req.Method, Accepted: observe.Accepted(true), Message: "request received"})
	switch req.Method {
	case "initialize":
		var params initializeParams
		_ = json.Unmarshal(req.Params, &params)
		s.protocol = negotiateProtocolVersion(params.ProtocolVersion)
		s.negotiated = true
		return &response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"protocolVersion": s.protocol, "capabilities": serverCapabilities(), "serverInfo": map[string]any{"name": "warden-mcp", "version": ServerVersion}, "instructions": "Use Warden tools to inspect and update the active plan safely."}}
	case "notifications/initialized":
		s.initialized = true
		s.record(observe.Event{Kind: "mcp", Method: req.Method, Accepted: observe.Accepted(true), Message: "client initialized"})
		return nil
	case "ping":
		return &response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{}}
	case "tools/list":
		if errResp := s.requireNegotiated(req); errResp != nil {
			return errResp
		}
		return &response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"tools": toolDefinitions()}}
	case "resources/list":
		if errResp := s.requireNegotiated(req); errResp != nil {
			return errResp
		}
		return &response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"resources": []map[string]any{}}}
	case "resources/templates/list":
		if errResp := s.requireNegotiated(req); errResp != nil {
			return errResp
		}
		return &response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"resourceTemplates": []map[string]any{}}}
	case "prompts/list":
		if errResp := s.requireNegotiated(req); errResp != nil {
			return errResp
		}
		return &response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"prompts": []map[string]any{}}}
	case "logging/setLevel":
		if errResp := s.requireNegotiated(req); errResp != nil {
			return errResp
		}
		return &response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{}}
	case "completion/complete":
		if errResp := s.requireNegotiated(req); errResp != nil {
			return errResp
		}
		return &response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"completion": map[string]any{"values": []string{}}}}
	case "tools/call":
		if errResp := s.requireInitialized(req); errResp != nil {
			return errResp
		}
		return s.handleToolCall(req)
	default:
		return &response{JSONRPC: "2.0", ID: req.ID, Error: &responseError{Code: -32601, Message: "method not found"}}
	}
}

func (s *Server) requireNegotiated(req request) *response {
	if !s.negotiated {
		return &response{JSONRPC: "2.0", ID: req.ID, Error: &responseError{Code: -32001, Message: "server not initialized"}}
	}
	return nil
}

func (s *Server) requireInitialized(req request) *response {
	if errResp := s.requireNegotiated(req); errResp != nil {
		return errResp
	}
	if !s.initialized {
		return &response{JSONRPC: "2.0", ID: req.ID, Error: &responseError{Code: -32002, Message: "client must send notifications/initialized before tool requests"}}
	}
	return nil
}

func negotiateProtocolVersion(requested string) string {
	if slices.Contains(supportedProtocolVersions, requested) {
		return requested
	}
	return ProtocolVersion
}

func serverCapabilities() map[string]any {
	return map[string]any{
		"tools":       map[string]any{},
		"prompts":     map[string]any{},
		"resources":   map[string]any{},
		"logging":     map[string]any{},
		"completions": map[string]any{},
	}
}

func (s *Server) handleToolCall(req request) *response {
	var params toolsCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &response{JSONRPC: "2.0", ID: req.ID, Error: &responseError{Code: -32602, Message: "invalid tool call params"}}
	}
	result, ok := s.callTool(params)
	if !ok {
		return &response{JSONRPC: "2.0", ID: req.ID, Error: &responseError{Code: -32602, Message: "unknown tool"}}
	}
	return &response{JSONRPC: "2.0", ID: req.ID, Result: result}
}

func (s *Server) callTool(params toolsCallParams) (map[string]any, bool) {
	planPath := ".agent/PLAN.md"
	switch params.Name {
	case "get_status":
		var args struct {
			PlanPath     string `json:"plan_path,omitempty"`
			PlanID       string `json:"plan_id,omitempty"`
			IncludeTasks bool   `json:"include_tasks,omitempty"`
		}
		_ = json.Unmarshal(params.Arguments, &args)
		if args.PlanPath != "" {
			planPath = args.PlanPath
		}
		env := s.API.Status(planPath, contracts.GetStatusRequest{PlanID: args.PlanID, IncludeTasks: args.IncludeTasks})
		return toolResult(env), true
	case "init_plan":
		var args contracts.InitPlanRequest
		_ = json.Unmarshal(params.Arguments, &args)
		env := s.API.Init(args)
		return toolResult(env), true
	case "health_check":
		var args struct {
			PlanPath string `json:"plan_path,omitempty"`
		}
		_ = json.Unmarshal(params.Arguments, &args)
		if args.PlanPath != "" {
			planPath = args.PlanPath
		}
		env := s.API.Health(planPath)
		return toolResult(env), true
	case "get_agent_guide":
		var args contracts.GetAgentGuideRequest
		_ = json.Unmarshal(params.Arguments, &args)
		if args.PlanPath != "" {
			planPath = args.PlanPath
		}
		env := s.agentGuide(planPath, args.PlanID, args.DetailLevel)
		return toolResult(env), true
	case "export_plan":
		var args struct {
			PlanPath string                 `json:"plan_path,omitempty"`
			PlanID   string                 `json:"plan_id,omitempty"`
			Format   contracts.ExportFormat `json:"format,omitempty"`
		}
		_ = json.Unmarshal(params.Arguments, &args)
		if args.PlanPath != "" {
			planPath = args.PlanPath
		}
		env := s.API.Export(planPath, contracts.ExportPlanRequest{PlanID: args.PlanID, Format: args.Format}, "")
		return toolResult(env), true
	case "import_plan":
		var args contracts.ImportPlanRequest
		_ = json.Unmarshal(params.Arguments, &args)
		env := s.API.Import(args)
		return toolResult(env), true
	case "list_plans":
		var args contracts.ListPlansRequest
		_ = json.Unmarshal(params.Arguments, &args)
		env := s.API.List(args)
		return toolResult(env), true
	case "archive_plan":
		var args contracts.ArchivePlanRequest
		_ = json.Unmarshal(params.Arguments, &args)
		env := s.API.Archive(args)
		return toolResult(env), true
	case "validate_plan":
		var args struct {
			PlanPath string                 `json:"plan_path,omitempty"`
			Mode     contracts.ValidateMode `json:"mode,omitempty"`
		}
		_ = json.Unmarshal(params.Arguments, &args)
		if args.PlanPath != "" {
			planPath = args.PlanPath
		}
		env := s.API.Validate(planPath, contracts.ValidatePlanRequest{Mode: args.Mode})
		return toolResult(env), true
	case "get_next_task":
		var args struct {
			PlanPath            string          `json:"plan_path,omitempty"`
			PlanID              string          `json:"plan_id,omitempty"`
			RespectPhaseOrder   bool            `json:"respect_phase_order,omitempty"`
			RespectDependencies bool            `json:"respect_dependencies,omitempty"`
			PriorityBias        domain.Priority `json:"priority_bias,omitempty"`
		}
		_ = json.Unmarshal(params.Arguments, &args)
		if args.PlanPath != "" {
			planPath = args.PlanPath
		}
		env := s.API.Next(planPath, contracts.GetNextTaskRequest{PlanID: args.PlanID, RespectPhaseOrder: args.RespectPhaseOrder || !fieldPresent(params.Arguments, "respect_phase_order"), RespectDependencies: args.RespectDependencies || !fieldPresent(params.Arguments, "respect_dependencies"), PriorityBias: args.PriorityBias})
		return toolResult(env), true
	case "request_finish":
		var args struct {
			PlanPath  string           `json:"plan_path,omitempty"`
			PlanID    string           `json:"plan_id,omitempty"`
			ActorType domain.ActorType `json:"actor_type,omitempty"`
			Summary   string           `json:"summary,omitempty"`
		}
		_ = json.Unmarshal(params.Arguments, &args)
		if args.PlanPath != "" {
			planPath = args.PlanPath
		}
		if args.ActorType == "" {
			args.ActorType = domain.ActorAgent
		}
		env := s.API.Finish(planPath, contracts.RequestFinishRequest{PlanID: args.PlanID, ActorType: args.ActorType, Summary: args.Summary})
		return toolResult(env), true
	case "update_task":
		var args struct {
			PlanPath  string            `json:"plan_path,omitempty"`
			PlanID    string            `json:"plan_id,omitempty"`
			TaskID    string            `json:"task_id"`
			Status    domain.TaskStatus `json:"status"`
			ActorType domain.ActorType  `json:"actor_type,omitempty"`
			Note      string            `json:"note,omitempty"`
			Evidence  []string          `json:"evidence,omitempty"`
			Reason    string            `json:"reason,omitempty"`
		}
		_ = json.Unmarshal(params.Arguments, &args)
		if args.PlanPath != "" {
			planPath = args.PlanPath
		}
		if args.ActorType == "" {
			args.ActorType = domain.ActorAgent
		}
		env := s.API.Update(planPath, contracts.UpdateTaskRequest{PlanID: args.PlanID, TaskID: args.TaskID, Status: args.Status, ActorType: args.ActorType, Note: args.Note, Evidence: args.Evidence, Reason: args.Reason})
		return toolResult(env), true
	case "reset_task":
		var args struct {
			PlanPath string            `json:"plan_path,omitempty"`
			PlanID   string            `json:"plan_id,omitempty"`
			TaskID   string            `json:"task_id"`
			Status   domain.TaskStatus `json:"status,omitempty"`
			Reason   string            `json:"reason"`
		}
		_ = json.Unmarshal(params.Arguments, &args)
		if args.PlanPath != "" {
			planPath = args.PlanPath
		}
		env := s.API.Reset(planPath, contracts.ResetTaskRequest{PlanID: args.PlanID, TaskID: args.TaskID, Status: args.Status, Reason: args.Reason})
		return toolResult(env), true
	case "prioritize_tasks":
		var args struct {
			PlanPath string                     `json:"plan_path,omitempty"`
			PlanID   string                     `json:"plan_id,omitempty"`
			Updates  []contracts.PriorityUpdate `json:"updates"`
		}
		_ = json.Unmarshal(params.Arguments, &args)
		if args.PlanPath != "" {
			planPath = args.PlanPath
		}
		env := s.API.Prioritize(planPath, contracts.PrioritizeTasksRequest{PlanID: args.PlanID, Updates: args.Updates})
		return toolResult(env), true
	case "reconcile_plan":
		var args contracts.ReconcilePlanRequest
		_ = json.Unmarshal(params.Arguments, &args)
		env := s.API.Reconcile(planPath, args)
		return toolResult(env), true
	case "edit_plan":
		var args contracts.EditPlanRequest
		_ = json.Unmarshal(params.Arguments, &args)
		env := s.API.Edit(planPath, args)
		return toolResult(env), true
	default:
		return nil, false
	}
}

func (s *Server) record(event observe.Event) {
	if s.Recorder != nil {
		s.Recorder.Record(event)
	}
}

func toolDefinitions() []map[string]any {
	return []map[string]any{
		{"name": "init_plan", "description": "Create a new governed plan from phased task input.", "inputSchema": objectSchema([]string{"title"}, map[string]any{"title": stringSchema("Plan title."), "plan_id": stringSchema("Optional stable plan ID."), "version": stringSchema("Optional semver version."), "goal": stringSchema("Optional goal statement."), "source_text": stringSchema("Optional source text."), "create_markdown_projection": boolSchema("Whether to create the markdown plan projection."), "phases": map[string]any{"type": "array", "items": map[string]any{"type": "object"}}})},
		{"name": "health_check", "description": "Run basic workspace and plan readiness checks.", "inputSchema": objectSchema([]string{}, map[string]any{"plan_path": stringSchema("Workspace-relative path to the plan markdown.")})},
		{"name": "get_agent_guide", "description": "Return the recommended end-to-end playbook for agents using Warden MCP.", "inputSchema": objectSchema([]string{}, map[string]any{"plan_path": stringSchema("Workspace-relative path to the plan markdown."), "plan_id": stringSchema("Optional identity guard for live context."), "detail_level": stringSchema("Guide detail level: brief or full.")})},
		{"name": "list_plans", "description": "List known active and optionally archived plans in the workspace.", "inputSchema": objectSchema([]string{}, map[string]any{"status": stringSchema("Optional plan status filter."), "include_archived": boolSchema("Include archived plans.")})},
		{"name": "import_plan", "description": "Import markdown or JSON plan content into the workspace.", "inputSchema": objectSchema([]string{"format", "content"}, map[string]any{"format": stringSchema("Import format: markdown or json."), "content": stringSchema("Plan content to import."), "plan_id": stringSchema("Optional plan ID override."), "mode": stringSchema("Import mode: create, merge, or replace.")})},
		{"name": "archive_plan", "description": "Archive the active plan after finish-gate checks pass.", "inputSchema": objectSchema([]string{}, map[string]any{"plan_id": stringSchema("Optional plan ID hint."), "reason": stringSchema("Optional archive reason."), "create_final_export": boolSchema("Write a final JSON export beside the archived plan.")})},
		{"name": "export_plan", "description": "Export the active plan as markdown or JSON.", "inputSchema": objectSchema([]string{}, map[string]any{"plan_path": stringSchema("Workspace-relative path to the plan markdown."), "plan_id": stringSchema("Optional plan ID hint."), "format": stringSchema("Optional export format: markdown or json.")})},
		{"name": "validate_plan", "description": "Validate the active plan and return issues plus normalized counts.", "inputSchema": objectSchema([]string{}, map[string]any{"plan_path": stringSchema("Workspace-relative path to the plan markdown."), "mode": stringSchema("Optional validation mode.")})},
		{"name": "get_status", "description": "Return the active plan status and optional task list.", "inputSchema": objectSchema([]string{}, map[string]any{"plan_path": stringSchema("Workspace-relative path to the plan markdown."), "plan_id": stringSchema("Optional identity guard for the active plan."), "include_tasks": boolSchema("Include summarized tasks in the response.")})},
		{"name": "get_next_task", "description": "Return the next recommended task under current phase-order rules.", "inputSchema": objectSchema([]string{}, map[string]any{"plan_path": stringSchema("Workspace-relative path to the plan markdown."), "plan_id": stringSchema("Optional plan ID hint."), "respect_phase_order": boolSchema("Prefer current phase ordering."), "respect_dependencies": boolSchema("Respect task dependency edges."), "priority_bias": stringSchema("Optional priority bias such as p1.")})},
		{"name": "request_finish", "description": "Evaluate whether the active plan can finish and explain blocking work.", "inputSchema": objectSchema([]string{}, map[string]any{"plan_path": stringSchema("Workspace-relative path to the plan markdown."), "plan_id": stringSchema("Optional plan ID hint."), "actor_type": stringSchema("Actor type requesting finish."), "summary": stringSchema("Optional operator summary.")})},
		{"name": "update_task", "description": "Update one task status in the active markdown plan projection.", "inputSchema": objectSchema([]string{"task_id", "status"}, map[string]any{"plan_path": stringSchema("Workspace-relative path to the plan markdown."), "plan_id": stringSchema("Optional plan ID hint."), "task_id": stringSchema("Task ID like PH09-T01."), "status": stringSchema("Target task status."), "actor_type": stringSchema("Actor type applying the update."), "note": stringSchema("Optional note."), "reason": stringSchema("Optional reason."), "evidence": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}})},
		{"name": "reset_task", "description": "Reset a terminal task back to not_started or in_progress with a required reason.", "inputSchema": objectSchema([]string{"task_id", "reason"}, map[string]any{"plan_path": stringSchema("Workspace-relative path to the plan markdown."), "plan_id": stringSchema("Optional plan ID hint."), "task_id": stringSchema("Task ID like PH09-T01."), "status": stringSchema("Reset target status: not_started or in_progress."), "reason": stringSchema("Why the task is being reopened.")})},
		{"name": "prioritize_tasks", "description": "Update task priorities in the active plan projection.", "inputSchema": objectSchema([]string{"updates"}, map[string]any{"plan_path": stringSchema("Workspace-relative path to the plan markdown."), "plan_id": stringSchema("Optional plan ID hint."), "updates": map[string]any{"type": "array", "items": map[string]any{"type": "object"}}})},
		{"name": "reconcile_plan", "description": "Compare candidate markdown to the active plan and optionally apply safe changes.", "inputSchema": objectSchema([]string{"markdown_content"}, map[string]any{"plan_id": stringSchema("Optional plan ID hint."), "markdown_content": stringSchema("Candidate markdown plan content."), "mode": stringSchema("dry_run or apply.")})},
		{"name": "edit_plan", "description": "Apply a bounded structured plan edit operation to the active plan.", "inputSchema": objectSchema([]string{"operation"}, map[string]any{"plan_id": stringSchema("Optional plan ID hint."), "target_id": stringSchema("Optional phase/task target."), "operation": stringSchema("Structured edit operation."), "reason": stringSchema("Required for waive/cancel operations."), "payload": map[string]any{"type": "object"}})},
	}
}

func objectSchema(required []string, props map[string]any) map[string]any {
	return map[string]any{"type": "object", "properties": props, "required": required, "additionalProperties": false}
}

func stringSchema(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}

func boolSchema(description string) map[string]any {
	return map[string]any{"type": "boolean", "description": description}
}

func (s *Server) agentGuide(planPath, requestedPlanID, detailLevel string) contracts.ToolResponseEnvelope[contracts.AgentGuideData] {
	if detailLevel == "" {
		detailLevel = "full"
	}
	if detailLevel != "brief" && detailLevel != "full" {
		detailLevel = "full"
	}
	warnings := []domain.ValidationIssue{}
	guide := baseAgentGuide(detailLevel, planPath)
	guide.LiveContext = s.buildAgentGuideLiveContext(planPath, requestedPlanID, &warnings)
	return contracts.ToolResponseEnvelope[contracts.AgentGuideData]{
		OK:        true,
		Tool:      "get_agent_guide",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		PlanID:    agentGuidePlanID(guide),
		Warnings:  dedupeValidationIssues(warnings),
		Data:      guide,
	}
}

func baseAgentGuide(detailLevel, planPath string) contracts.AgentGuideData {
	guide := contracts.AgentGuideData{
		GuideVersion: AgentGuideVersion,
		Summary:      "Use Warden first, preserve plan identity, record evidence as you go, and ask Warden to evaluate finish before claiming the work is complete.",
		CoreRules: []string{
			"Start with get_status before making assumptions about the plan.",
			"Use get_next_task when choosing work instead of inferring the next step from markdown alone.",
			"Carry plan_id forward once known and treat it as an identity guard on future calls.",
			"Record meaningful progress with update_task including note, evidence, or reason where applicable.",
			"Run validate_plan and request_finish before declaring completion.",
		},
		RecommendedSequence: []contracts.AgentGuideStep{
			{Order: 1, Call: "health_check", Why: "Confirm the workspace and plan path are readable."},
			{Order: 2, Call: "get_status", Why: "Learn the active plan_id, current phase, blocking reasons, and next task hint."},
			{Order: 3, Call: "get_next_task", Why: "Choose the next recommended task under phase-order and dependency rules."},
			{Order: 4, Call: "update_task", Why: "Record status changes with notes, evidence, or reasons as work progresses."},
			{Order: 5, Call: "validate_plan", Why: "Check that the active plan remains internally valid after edits or task changes."},
			{Order: 6, Call: "request_finish", Why: "Ask Warden whether the plan is actually finishable before claiming done."},
		},
		ExampleCalls: []contracts.AgentGuideExampleCall{
			{Name: "get_status", Arguments: map[string]any{"plan_path": planPath}},
			{Name: "get_next_task", Arguments: map[string]any{"plan_path": planPath, "respect_phase_order": true, "respect_dependencies": true}},
			{Name: "update_task", Arguments: map[string]any{"plan_path": planPath, "plan_id": "<from_get_status>", "task_id": "PHXX-TYY", "status": "in_progress", "note": "Started implementation."}},
			{Name: "request_finish", Arguments: map[string]any{"plan_path": planPath, "plan_id": "<from_get_status>", "actor_type": string(domain.ActorAgent)}},
		},
	}
	if detailLevel == "brief" {
		guide.ExampleCalls = guide.ExampleCalls[:2]
		return guide
	}
	guide.ToolPlaybook = []contracts.AgentGuideToolEntry{
		{Name: "get_status", UseWhen: "Whenever you need canonical plan state, identity, blocking reasons, or current phase context.", Notes: []string{"Prefer this before reading markdown directly.", "Carry forward the returned plan_id."}},
		{Name: "get_next_task", UseWhen: "When you need the next recommended unit of work.", Notes: []string{"Use after get_status.", "Prefer dependency and phase-order flags unless you have a reason not to."}},
		{Name: "update_task", UseWhen: "When task state changes or you need to add evidence, notes, or reasons.", Notes: []string{"Use note/evidence for auditability.", "Default actor_type should be agent."}},
		{Name: "validate_plan", UseWhen: "After plan edits, reconcile flows, or before final completion claims."},
		{Name: "request_finish", UseWhen: "Immediately before claiming the plan is done.", Notes: []string{"Treat a denied finish as authoritative remaining work."}},
		{Name: "archive_plan", UseWhen: "Only after request_finish allows completion and you are performing closure work."},
		{Name: "edit_plan / reconcile_plan", UseWhen: "When the requested work changes plan structure rather than only task state.", Notes: []string{"Prefer bounded edits.", "Validate afterward."}},
	}
	guide.FinishGateRules = []string{
		"Do not claim completion from intuition; request_finish is the gate.",
		"Required unfinished tasks block finish even if optional tasks do not.",
		"If request_finish denies completion, surface the blocking reasons instead of summarizing optimistically.",
		"Archive only after the finish gate passes and closure is intentional.",
	}
	return guide
}

func (s *Server) buildAgentGuideLiveContext(planPath, requestedPlanID string, warnings *[]domain.ValidationIssue) *contracts.AgentGuideLiveContext {
	ctx := &contracts.AgentGuideLiveContext{PlanPath: planPath, SuggestedNextCalls: []string{"health_check", "get_status", "get_next_task"}}
	health := s.API.Health(planPath)
	*warnings = append(*warnings, health.Warnings...)
	if health.Data.PlanID == "" {
		*warnings = append(*warnings, domain.ValidationIssue{Severity: "warning", Code: contracts.ErrPlanNotFound, Message: fmt.Sprintf("No active plan detected at %s.", planPath), Path: planPath})
		return ctx
	}
	ctx.PlanDetected = true
	ctx.PlanID = health.Data.PlanID
	if requestedPlanID != "" {
		match := requestedPlanID == health.Data.PlanID
		ctx.IdentityMatch = &match
		if !match {
			*warnings = append(*warnings, domain.ValidationIssue{Severity: "warning", Code: contracts.ErrSyncConflict, Message: fmt.Sprintf("Requested plan_id %q does not match active plan_id %q.", requestedPlanID, health.Data.PlanID), Path: planPath})
			ctx.SuggestedNextCalls = []string{"get_status", "health_check"}
			return ctx
		}
	}
	status := s.API.Status(planPath, contracts.GetStatusRequest{PlanID: requestedPlanID})
	*warnings = append(*warnings, status.Warnings...)
	if status.OK {
		ctx.CurrentPhaseID = status.Data.Plan.CurrentPhaseID
		ctx.PlanStatus = string(status.Data.Plan.Status)
		ctx.NextTaskID = status.Data.NextTaskID
		ctx.SuggestedNextCalls = []string{"get_status", "get_next_task", "update_task", "validate_plan", "request_finish"}
	}
	return ctx
}

func agentGuidePlanID(guide contracts.AgentGuideData) string {
	if guide.LiveContext == nil {
		return ""
	}
	return guide.LiveContext.PlanID
}

func dedupeValidationIssues(issues []domain.ValidationIssue) []domain.ValidationIssue {
	if len(issues) < 2 {
		return issues
	}
	seen := map[string]struct{}{}
	result := make([]domain.ValidationIssue, 0, len(issues))
	for _, issue := range issues {
		key := strings.Join([]string{issue.Severity, issue.Code, issue.Message, issue.Path}, "\x00")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, issue)
	}
	return result
}

func toolResult(envelope any) map[string]any {
	payload, _ := json.Marshal(envelope)
	var structured any
	_ = json.Unmarshal(payload, &structured)
	return map[string]any{"content": []map[string]any{{"type": "text", "text": string(payload)}}, "structuredContent": structured, "isError": !inferOK(envelope)}
}

func inferOK(envelope any) bool {
	data, _ := json.Marshal(envelope)
	return !bytes.Contains(data, []byte(`"ok":false`))
}

func readFrame(reader *bufio.Reader) ([]byte, transportMode, error) {
	contentLength := -1
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF && line == "" {
				return nil, transportModeUnknown, io.EOF
			}
			return nil, transportModeUnknown, err
		}
		trimmed := strings.TrimSpace(line)
		if contentLength < 0 && looksLikeJSON(trimmed) {
			return readLineDelimitedJSON(reader, line)
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(parts[0]), "Content-Length") {
			value, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, transportModeUnknown, fmt.Errorf("invalid content length: %w", err)
			}
			contentLength = value
		}
	}
	if contentLength < 0 {
		return nil, transportModeUnknown, fmt.Errorf("missing Content-Length header")
	}
	payload := make([]byte, contentLength)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, transportModeUnknown, err
	}
	return payload, transportModeFramed, nil
}

func readLineDelimitedJSON(reader *bufio.Reader, firstLine string) ([]byte, transportMode, error) {
	var buf bytes.Buffer
	buf.WriteString(strings.TrimSpace(firstLine))
	for {
		payload := bytes.TrimSpace(buf.Bytes())
		if json.Valid(payload) {
			return bytes.Clone(payload), transportModeLineDelimited, nil
		}
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF && len(strings.TrimSpace(line)) > 0 {
				buf.WriteByte('\n')
				buf.WriteString(strings.TrimSpace(line))
				payload = bytes.TrimSpace(buf.Bytes())
				if json.Valid(payload) {
					return bytes.Clone(payload), transportModeLineDelimited, nil
				}
			}
			return nil, transportModeUnknown, io.ErrUnexpectedEOF
		}
		buf.WriteByte('\n')
		buf.WriteString(strings.TrimSpace(line))
	}
}

func looksLikeJSON(line string) bool {
	return strings.HasPrefix(line, "{") || strings.HasPrefix(line, "[")
}

func writeFrame(writer io.Writer, payload any, mode transportMode) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if mode == transportModeLineDelimited {
		_, err = fmt.Fprintf(writer, "%s\n", data)
		return err
	}
	_, err = fmt.Fprintf(writer, "Content-Length: %d\r\n\r\n%s", len(data), data)
	return err
}

func fieldPresent(raw json.RawMessage, key string) bool {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return false
	}
	_, ok := obj[key]
	return ok
}

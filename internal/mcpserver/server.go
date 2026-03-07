package mcpserver

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"warden-mcp/internal/api"
	"warden-mcp/internal/domain"
	"warden-mcp/internal/mcp/contracts"
	"warden-mcp/internal/observe"
)

const ProtocolVersion = "2025-03-26"

type Server struct {
	API         api.API
	Recorder    observe.Recorder
	initialized bool
	negotiated  bool
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
	for {
		payload, err := readFrame(buffered)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		var req request
		if err := json.Unmarshal(payload, &req); err != nil {
			if err := writeFrame(writer, response{JSONRPC: "2.0", Error: &responseError{Code: -32700, Message: "parse error"}}); err != nil {
				return err
			}
			continue
		}
		resp := s.handle(req)
		if len(req.ID) == 0 || resp == nil {
			continue
		}
		if err := writeFrame(writer, resp); err != nil {
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
		s.negotiated = true
		return &response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"protocolVersion": ProtocolVersion, "capabilities": map[string]any{"tools": map[string]any{}}, "serverInfo": map[string]any{"name": "warden-mcp", "version": "0.1.0"}, "instructions": "Use Warden tools to inspect and update the active plan safely."}}
	case "notifications/initialized":
		s.initialized = true
		s.record(observe.Event{Kind: "mcp", Method: req.Method, Accepted: observe.Accepted(true), Message: "client initialized"})
		return nil
	case "ping":
		return &response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{}}
	case "tools/list":
		if errResp := s.requireInitialized(req); errResp != nil {
			return errResp
		}
		return &response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"tools": toolDefinitions()}}
	case "tools/call":
		if errResp := s.requireInitialized(req); errResp != nil {
			return errResp
		}
		return s.handleToolCall(req)
	default:
		return &response{JSONRPC: "2.0", ID: req.ID, Error: &responseError{Code: -32601, Message: "method not found"}}
	}
}

func (s *Server) requireInitialized(req request) *response {
	if !s.negotiated {
		return &response{JSONRPC: "2.0", ID: req.ID, Error: &responseError{Code: -32001, Message: "server not initialized"}}
	}
	if !s.initialized {
		return &response{JSONRPC: "2.0", ID: req.ID, Error: &responseError{Code: -32002, Message: "client must send notifications/initialized before tool requests"}}
	}
	return nil
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

func readFrame(reader *bufio.Reader) ([]byte, error) {
	contentLength := -1
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF && line == "" {
				return nil, io.EOF
			}
			return nil, err
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
				return nil, fmt.Errorf("invalid content length: %w", err)
			}
			contentLength = value
		}
	}
	if contentLength < 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}
	payload := make([]byte, contentLength)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func writeFrame(writer io.Writer, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
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

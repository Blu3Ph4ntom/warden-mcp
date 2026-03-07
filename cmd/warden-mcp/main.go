package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"warden-mcp/internal/api"
	"warden-mcp/internal/domain"
	"warden-mcp/internal/mcp/contracts"
	"warden-mcp/internal/mcpserver"
	"warden-mcp/internal/observe"
	"warden-mcp/internal/security"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	return runWithIO(args, os.Stdin, stdout, stderr)
}

func runWithIO(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: warden-mcp <status|next|finish|update|reset|prioritize|reconcile|edit|health|export|validate|serve> [-plan path]")
		return 2
	}
	command := args[0]
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.SetOutput(stderr)
	planPath := fs.String("plan", ".agent/PLAN.md", "path to plan markdown")
	planID := fs.String("plan-id", "", "optional identity guard for the active plan")
	includeTasks := fs.Bool("include-tasks", false, "include tasks in status output")
	actor := fs.String("actor", string(domain.ActorAgent), "actor type for finish requests")
	taskID := fs.String("task", "", "task ID for update operations")
	status := fs.String("status", "", "task status for update operations")
	reason := fs.String("reason", "", "reason for reset/edit/archive-style operations")
	format := fs.String("format", string(contracts.ExportMarkdown), "export format")
	writePath := fs.String("write", "", "optional workspace-relative export output path")
	updates := fs.String("updates", "", "comma-separated task_id=PRIORITY updates")
	contentFile := fs.String("content-file", "", "path to reconcile markdown content file")
	reconcileMode := fs.String("mode", string(contracts.ReconcileDryRun), "reconcile mode")
	operation := fs.String("operation", "", "structured edit operation")
	targetID := fs.String("target", "", "target phase/task id for edit operations")
	payloadJSON := fs.String("payload", "{}", "JSON payload for edit operations")
	logEvents := fs.Bool("log-events", false, "emit structured events to stderr")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	workspaceRoot, err := os.Getwd()
	if err != nil {
		return writeError(stdout, command, contracts.ErrInternal, err.Error())
	}
	var recorder observe.Recorder
	if *logEvents || os.Getenv("WARDEN_LOG_JSON") == "1" {
		recorder = observe.NewJSONRecorder(stderr)
	}
	app := api.New(workspaceRoot, recorder)
	if command == "serve" {
		server := &mcpserver.Server{API: app, Recorder: recorder}
		if err := server.Serve(stdin, stdout); err != nil {
			return writeError(stdout, command, contracts.ErrInternal, security.RedactSecretLikeText(err.Error()))
		}
		return 0
	}
	switch command {
	case "validate":
		return writeEnvelope(stdout, app.Validate(*planPath, contracts.ValidatePlanRequest{}))
	case "edit":
		payload, err := parseJSONPayload(*payloadJSON)
		if err != nil {
			return writeError(stdout, command, contracts.ErrPlanInvalid, err.Error())
		}
		return writeEnvelope(stdout, app.Edit(*planPath, contracts.EditPlanRequest{PlanID: *planID, Operation: contracts.EditOperation(*operation), TargetID: *targetID, Reason: *reason, Payload: payload}))
	case "reconcile":
		resolvedContentPath, err := security.ResolveWorkspacePath(workspaceRoot, *contentFile, ".md")
		if err != nil {
			return writeError(stdout, command, contracts.ErrPlanInvalid, err.Error())
		}
		content, err := os.ReadFile(resolvedContentPath)
		if err != nil {
			return writeError(stdout, command, contracts.ErrPlanInvalid, err.Error())
		}
		return writeEnvelope(stdout, app.Reconcile(*planPath, contracts.ReconcilePlanRequest{PlanID: *planID, MarkdownContent: string(content), Mode: contracts.ReconcileMode(*reconcileMode)}))
	case "prioritize":
		parsedUpdates, err := parsePriorityUpdates(*updates)
		if err != nil {
			return writeError(stdout, command, contracts.ErrPlanInvalid, err.Error())
		}
		return writeEnvelope(stdout, app.Prioritize(*planPath, contracts.PrioritizeTasksRequest{PlanID: *planID, Updates: parsedUpdates}))
	case "reset":
		return writeEnvelope(stdout, app.Reset(*planPath, contracts.ResetTaskRequest{PlanID: *planID, TaskID: *taskID, Status: domain.TaskStatus(*status), Reason: *reason}))
	case "export":
		return writeEnvelope(stdout, app.Export(*planPath, contracts.ExportPlanRequest{Format: contracts.ExportFormat(*format)}, *writePath))
	case "health":
		return writeEnvelope(stdout, app.Health(*planPath))
	case "status":
		return writeEnvelope(stdout, app.Status(*planPath, contracts.GetStatusRequest{PlanID: *planID, IncludeTasks: *includeTasks}))
	case "next":
		return writeEnvelope(stdout, app.Next(*planPath, contracts.GetNextTaskRequest{PlanID: *planID, RespectPhaseOrder: true, RespectDependencies: true}))
	case "finish":
		return writeEnvelope(stdout, app.Finish(*planPath, contracts.RequestFinishRequest{PlanID: *planID, ActorType: domain.ActorType(*actor)}))
	case "update":
		return writeEnvelope(stdout, app.Update(*planPath, contracts.UpdateTaskRequest{PlanID: *planID, TaskID: *taskID, Status: domain.TaskStatus(*status), ActorType: domain.ActorType(*actor), Reason: *reason}))
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n", command)
		return 2
	}
}

func writeError(stdout io.Writer, tool, code, message string) int {
	envelope := contracts.ToolResponseEnvelope[struct{}]{OK: false, Tool: tool, Timestamp: apiTimestamp(), Warnings: []domain.ValidationIssue{}, Error: &contracts.ErrorObject{Code: code, Message: message, Retryable: false, Details: map[string]any{}}, Data: struct{}{}}
	return writeEnvelope(stdout, envelope)
}

func writeEnvelope[T any](stdout io.Writer, envelope contracts.ToolResponseEnvelope[T]) int {
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(envelope); err != nil {
		return 1
	}
	return 0
}

func apiTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func parsePriorityUpdates(raw string) ([]contracts.PriorityUpdate, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("updates flag is required")
	}
	parts := strings.Split(raw, ",")
	result := make([]contracts.PriorityUpdate, 0, len(parts))
	for _, part := range parts {
		taskID, priority, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok || strings.TrimSpace(taskID) == "" || strings.TrimSpace(priority) == "" {
			return nil, fmt.Errorf("invalid update: %s", part)
		}
		result = append(result, contracts.PriorityUpdate{TaskID: strings.TrimSpace(taskID), Priority: domain.Priority(strings.ToUpper(strings.TrimSpace(priority)))})
	}
	return result, nil
}

func parseJSONPayload(raw string) (json.RawMessage, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		trimmed = "{}"
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(trimmed), &decoded); err != nil {
		return nil, err
	}
	return json.RawMessage([]byte(trimmed)), nil
}

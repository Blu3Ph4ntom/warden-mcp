package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"warden-mcp/internal/domain"
	"warden-mcp/internal/mcp/contracts"
	"warden-mcp/internal/planfile"
	"warden-mcp/internal/security"
	"warden-mcp/internal/service"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: warden-mcp <status|next|finish> [-plan path]")
		return 2
	}
	command := args[0]
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.SetOutput(stderr)
	planPath := fs.String("plan", ".agent/PLAN.md", "path to plan markdown")
	includeTasks := fs.Bool("include-tasks", false, "include tasks in status output")
	actor := fs.String("actor", string(domain.ActorAgent), "actor type for finish requests")
	taskID := fs.String("task", "", "task ID for update operations")
	status := fs.String("status", "", "task status for update operations")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	workspaceRoot, err := os.Getwd()
	if err != nil {
		return writeError(stdout, command, contracts.ErrInternal, err.Error())
	}
	resolvedPlanPath, err := security.ResolveWorkspacePath(workspaceRoot, *planPath, ".md")
	if err != nil {
		return writeError(stdout, command, contracts.ErrPlanInvalid, security.RedactSecretLikeText(err.Error()))
	}
	plan, warnings, err := planfile.Load(resolvedPlanPath)
	if err != nil {
		return writeError(stdout, command, contracts.ErrPlanNotFound, security.RedactSecretLikeText(err.Error()))
	}
	switch command {
	case "status":
		return writeEnvelope(stdout, contracts.ToolResponseEnvelope[contracts.GetStatusData]{OK: true, Tool: "get_status", Timestamp: timestamp(), PlanID: plan.PlanID, Warnings: warnings, Data: service.GetStatus(plan, *includeTasks)})
	case "next":
		return writeEnvelope(stdout, contracts.ToolResponseEnvelope[contracts.GetNextTaskData]{OK: true, Tool: "get_next_task", Timestamp: timestamp(), PlanID: plan.PlanID, Warnings: warnings, Data: service.GetNextTask(plan, contracts.GetNextTaskRequest{PlanID: plan.PlanID, RespectPhaseOrder: true, RespectDependencies: true})})
	case "finish":
		return writeEnvelope(stdout, contracts.ToolResponseEnvelope[contracts.RequestFinishData]{OK: true, Tool: "request_finish", Timestamp: timestamp(), PlanID: plan.PlanID, Warnings: warnings, Data: service.RequestFinish(plan, contracts.RequestFinishRequest{PlanID: plan.PlanID, ActorType: domain.ActorType(*actor)})})
	case "update":
		if *taskID == "" || *status == "" {
			return writeError(stdout, command, contracts.ErrTaskNotFound, "update requires -task and -status")
		}
		data, updateWarnings, err := service.UpdateTask(resolvedPlanPath, contracts.UpdateTaskRequest{PlanID: plan.PlanID, TaskID: *taskID, Status: domain.TaskStatus(*status), ActorType: domain.ActorType(*actor)})
		if err != nil {
			return writeError(stdout, command, contracts.ErrTaskTransitionInvalid, security.RedactSecretLikeText(err.Error()))
		}
		warnings = append(warnings, updateWarnings...)
		return writeEnvelope(stdout, contracts.ToolResponseEnvelope[contracts.UpdateTaskData]{OK: true, Tool: "update_task", Timestamp: timestamp(), PlanID: plan.PlanID, Warnings: warnings, Data: data})
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n", command)
		return 2
	}
}

func writeError(stdout io.Writer, tool, code, message string) int {
	envelope := contracts.ToolResponseEnvelope[struct{}]{OK: false, Tool: tool, Timestamp: timestamp(), Warnings: []domain.ValidationIssue{}, Error: &contracts.ErrorObject{Code: code, Message: message, Retryable: false, Details: map[string]any{}}, Data: struct{}{}}
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

func timestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

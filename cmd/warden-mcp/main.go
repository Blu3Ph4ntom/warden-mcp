package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
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
		fmt.Fprintln(stderr, "usage: warden-mcp <status|next|finish|update|health|serve> [-plan path]")
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
	case "health":
		return writeEnvelope(stdout, app.Health(*planPath))
	case "status":
		return writeEnvelope(stdout, app.Status(*planPath, *includeTasks))
	case "next":
		return writeEnvelope(stdout, app.Next(*planPath, contracts.GetNextTaskRequest{RespectPhaseOrder: true, RespectDependencies: true}))
	case "finish":
		return writeEnvelope(stdout, app.Finish(*planPath, contracts.RequestFinishRequest{ActorType: domain.ActorType(*actor)}))
	case "update":
		return writeEnvelope(stdout, app.Update(*planPath, contracts.UpdateTaskRequest{TaskID: *taskID, Status: domain.TaskStatus(*status), ActorType: domain.ActorType(*actor)}))
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

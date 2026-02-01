package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
)

// ProcessInput is the input schema for zerops_process.
type ProcessInput struct {
	ProcessID string `json:"processId"`
	Action    string `json:"action,omitempty"` // "status" (default) or "cancel"
}

// RegisterProcess registers the zerops_process tool on the server.
func RegisterProcess(srv *mcp.Server, exec executor.Executor) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "zerops_process",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Check Process",
			ReadOnlyHint:   true,
			IdempotentHint: true,
		},
		Description: `Check or cancel an async process.

Actions:
- status (default): Check process status
- cancel: Cancel a running or pending process

Used internally by MCP for waitForCompletion polling.
Can also be used directly by agent to check operation status.

Process statuses: PENDING, RUNNING, FINISHED, FAILED, CANCELED`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ProcessInput) (*mcp.CallToolResult, any, error) {
		if input.ProcessID == "" {
			return errorResult("processId is required"), nil, nil
		}

		action := input.Action
		if action == "" {
			action = "status"
		}

		var args []string
		switch action {
		case "status":
			args = []string{"process", input.ProcessID}
		case "cancel":
			args = []string{"cancel", input.ProcessID}
		default:
			return errorResult("action must be 'status' or 'cancel'"), nil, nil
		}

		result, err := exec.RunZaia(ctx, args...)
		if err != nil {
			return cliErrorResult(err)
		}
		mcpResult, _ := ResultFromCLI(result)
		return mcpResult, nil, nil
	})
}

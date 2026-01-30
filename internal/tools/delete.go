package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
)

// DeleteInput is the input schema for zerops_delete.
type DeleteInput struct {
	ServiceHostname string `json:"serviceHostname"`
	Confirm         bool   `json:"confirm"`
}

// RegisterDelete registers the zerops_delete tool on the server.
func RegisterDelete(srv *mcp.Server, exec executor.Executor) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "zerops_delete",
		Description: `Delete a service from the project.

IMPORTANT: This is destructive. Deletes the service and all its data.

Only deletes services, NOT the project itself.

Parameters:
- serviceHostname (required)
- confirm (required, must be true)

Returns process ID for tracking via zerops_process.`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DeleteInput) (*mcp.CallToolResult, any, error) {
		if input.ServiceHostname == "" {
			return errorResult("serviceHostname is required"), nil, nil
		}
		if !input.Confirm {
			return errorResult("confirm must be true to delete a service"), nil, nil
		}

		args := []string{"delete", "--service", input.ServiceHostname, "--confirm"}
		result, err := exec.RunZaia(ctx, args...)
		if err != nil {
			return errorResult("CLI execution failed: " + err.Error()), nil, nil //nolint:nilerr // intentional: convert Go error to MCP error result
		}
		mcpResult, _ := ResultFromCLI(result)
		return mcpResult, nil, nil
	})
}

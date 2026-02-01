package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
)

// EventsInput is the input schema for zerops_events.
type EventsInput struct {
	ServiceHostname string `json:"serviceHostname,omitempty"`
	Limit           int    `json:"limit,omitempty"`
}

// RegisterEvents registers the zerops_events tool on the server.
func RegisterEvents(srv *mcp.Server, exec executor.Executor) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "zerops_events",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Project Activity Log",
			ReadOnlyHint:   true,
			IdempotentHint: true,
		},
		Description: `Fetch project activity timeline.

Aggregates processes (start/stop/restart/scale/import/delete/env) and build/deploy events into a unified timeline sorted by time.

Parameters:
- serviceHostname: Filter by service (optional)
- limit: Max events (default 50)

Returns:
- events: Unified timeline with timestamp, action, status, service, duration
- summary: Event counts`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input EventsInput) (*mcp.CallToolResult, any, error) {
		args := []string{"events"}
		if input.ServiceHostname != "" {
			args = append(args, "--service", input.ServiceHostname)
		}
		if input.Limit > 0 {
			args = append(args, "--limit", fmt.Sprintf("%d", input.Limit))
		}

		result, err := exec.RunZaia(ctx, args...)
		if err != nil {
			return cliErrorResult(err)
		}
		mcpResult, _ := ResultFromCLI(result)
		return mcpResult, nil, nil
	})
}

package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
)

// LogsInput is the input schema for zerops_logs.
type LogsInput struct {
	ServiceHostname string `json:"serviceHostname"`
	Severity        string `json:"severity,omitempty"`
	Since           string `json:"since,omitempty"`
	Limit           int    `json:"limit,omitempty"`
	Search          string `json:"search,omitempty"`
	BuildID         string `json:"buildId,omitempty"`
}

// RegisterLogs registers the zerops_logs tool on the server.
func RegisterLogs(srv *mcp.Server, exec executor.Executor) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "zerops_logs",
		Description: `Fetch logs from a Zerops service.

Parameters:
- serviceHostname (required)
- severity: Filter by severity (error, warning, info, debug)
- since: Time range (30m, 1h, 24h, 7d, or ISO 8601)
- limit: Max entries (default 100)
- search: Text search
- buildId: Get build logs`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input LogsInput) (*mcp.CallToolResult, any, error) {
		if input.ServiceHostname == "" {
			return errorResult("serviceHostname is required"), nil, nil
		}

		args := []string{"logs", "--service", input.ServiceHostname}
		if input.Severity != "" {
			args = append(args, "--severity", input.Severity)
		}
		if input.Since != "" {
			args = append(args, "--since", input.Since)
		}
		if input.Limit > 0 {
			args = append(args, "--limit", fmt.Sprintf("%d", input.Limit))
		}
		if input.Search != "" {
			args = append(args, "--search", input.Search)
		}
		if input.BuildID != "" {
			args = append(args, "--build", input.BuildID)
		}

		result, err := exec.RunZaia(ctx, args...)
		if err != nil {
			return cliErrorResult(err)
		}
		mcpResult, _ := ResultFromCLI(result)
		return mcpResult, nil, nil
	})
}

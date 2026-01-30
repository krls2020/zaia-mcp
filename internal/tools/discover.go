package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
)

// DiscoverInput is the input schema for zerops_discover.
type DiscoverInput struct {
	Service     string `json:"service,omitempty"`
	IncludeEnvs bool   `json:"includeEnvs,omitempty"`
}

// RegisterDiscover registers the zerops_discover tool on the server.
func RegisterDiscover(srv *mcp.Server, exec executor.Executor) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "zerops_discover",
		Description: `Discover services in your Zerops project.

Call this first to get service hostnames for other tools.

Returns:
- project: Current project info (id, name, status)
- services: List with hostname, type, status
- Optional: env vars per service`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DiscoverInput) (*mcp.CallToolResult, any, error) {
		args := []string{"discover"}
		if input.Service != "" {
			args = append(args, "--service", input.Service)
		}
		if input.IncludeEnvs {
			args = append(args, "--include-envs")
		}

		result, err := exec.RunZaia(ctx, args...)
		if err != nil {
			return errorResult("CLI execution failed: " + err.Error()), nil, nil //nolint:nilerr // intentional: convert Go error to MCP error result
		}
		mcpResult, _ := ResultFromCLI(result)
		return mcpResult, nil, nil
	})
}

func errorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: true,
	}
}

package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
)

// EnvInput is the input schema for zerops_env.
type EnvInput struct {
	Action          string   `json:"action"`
	ServiceHostname string   `json:"serviceHostname,omitempty"`
	Project         bool     `json:"project,omitempty"`
	Variables       []string `json:"variables,omitempty"`
}

// RegisterEnv registers the zerops_env tool on the server.
func RegisterEnv(srv *mcp.Server, exec executor.Executor) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "zerops_env",
		Description: `Manage environment variables.

Actions:
- get: Read env vars (sync response)
- set: Set env vars (async - returns process ID)
- delete: Delete env var (async - returns process ID)

Scope: Provide serviceHostname for service env, or project=true for project env.

Set format: ["KEY=value", "ANOTHER=value2"]
Delete format: ["KEY"]

Note: Use ${service_hostname} for cross-service references (underscore, not dash).`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input EnvInput) (*mcp.CallToolResult, any, error) {
		if input.Action == "" {
			return errorResult("action is required (get, set, delete)"), nil, nil
		}
		if input.ServiceHostname == "" && !input.Project {
			return errorResult("serviceHostname or project=true is required"), nil, nil
		}

		args := []string{"env", input.Action}
		if input.ServiceHostname != "" {
			args = append(args, "--service", input.ServiceHostname)
		}
		if input.Project {
			args = append(args, "--project")
		}
		args = append(args, input.Variables...)

		result, err := exec.RunZaia(ctx, args...)
		if err != nil {
			return errorResult("CLI execution failed: " + err.Error()), nil, nil //nolint:nilerr // intentional: convert Go error to MCP error result
		}
		mcpResult, _ := ResultFromCLI(result)
		return mcpResult, nil, nil
	})
}

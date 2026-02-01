package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
)

// DeployInput is the input schema for zerops_deploy.
type DeployInput struct {
	WorkingDir string `json:"workingDir,omitempty"`
	ServiceID  string `json:"serviceId,omitempty"`
}

// RegisterDeploy registers the zerops_deploy tool on the server.
// This tool calls zcli push directly (not via ZAIA).
func RegisterDeploy(srv *mcp.Server, exec executor.Executor) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "zerops_deploy",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Deploy Code",
			DestructiveHint: boolPtr(false),
		},
		Description: `Deploy code to a Zerops service using zcli push.

Uses the zerops.yml in the working directory.
Requires separate zcli authentication.

Parameters:
- workingDir: Directory with zerops.yml (optional)
- serviceId: Target service ID (optional, reads zerops.yml)

Returns deployment process info.`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DeployInput) (*mcp.CallToolResult, any, error) {
		args := []string{"push"}
		if input.ServiceID != "" {
			args = append(args, "--serviceId", input.ServiceID)
		}
		if input.WorkingDir != "" {
			args = append(args, "--workingDir", input.WorkingDir)
		}

		result, err := exec.RunZcli(ctx, args...)
		if err != nil {
			return zcliErrorResult(err)
		}
		mcpResult, _ := ResultFromCLI(result)
		return mcpResult, nil, nil
	})
}

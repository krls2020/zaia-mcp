package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
)

// SubdomainInput is the input schema for zerops_subdomain.
type SubdomainInput struct {
	ServiceHostname string `json:"serviceHostname"`
	Action          string `json:"action"` // "enable" or "disable"
}

// RegisterSubdomain registers the zerops_subdomain tool on the server.
func RegisterSubdomain(srv *mcp.Server, exec executor.Executor) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "zerops_subdomain",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Manage Subdomain",
			DestructiveHint: boolPtr(false),
			IdempotentHint:  true,
		},
		Description: `Enable or disable Zerops subdomain for a service.

Actions:
- enable: Create a *.zerops.app subdomain
- disable: Remove the subdomain

Idempotent: enabling an already enabled subdomain returns success.`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SubdomainInput) (*mcp.CallToolResult, any, error) {
		if input.ServiceHostname == "" {
			return errorResult("serviceHostname is required"), nil, nil
		}
		if input.Action != "enable" && input.Action != "disable" {
			return errorResult("action must be 'enable' or 'disable'"), nil, nil
		}

		args := []string{"subdomain", input.Action, "--service", input.ServiceHostname}
		result, err := exec.RunZaia(ctx, args...)
		if err != nil {
			return cliErrorResult(err)
		}
		mcpResult, _ := ResultFromCLI(result)
		return mcpResult, nil, nil
	})
}

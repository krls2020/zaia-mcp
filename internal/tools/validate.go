package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
)

// ValidateInput is the input schema for zerops_validate.
type ValidateInput struct {
	Content  string `json:"content,omitempty"`
	FilePath string `json:"filePath,omitempty"`
	Type     string `json:"type,omitempty"`
}

// RegisterValidate registers the zerops_validate tool on the server.
func RegisterValidate(srv *mcp.Server, exec executor.Executor) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "zerops_validate",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Validate Config",
			ReadOnlyHint:   true,
			IdempotentHint: true,
			OpenWorldHint:  boolPtr(false),
		},
		Description: `Validate YAML configuration.

Validates zerops.yml or import.yml for:
- YAML syntax
- Required fields
- Valid service types and versions
- Port ranges
- Best practices

In project-scoped context, import.yml must NOT contain 'project:' section.

Returns errors with fix suggestions.`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ValidateInput) (*mcp.CallToolResult, any, error) {
		args := []string{"validate"}
		if input.Content != "" {
			args = append(args, "--content", input.Content)
		} else if input.FilePath != "" {
			args = append(args, "--file", input.FilePath)
		}
		if input.Type != "" {
			args = append(args, "--type", input.Type)
		}

		result, err := exec.RunZaia(ctx, args...)
		if err != nil {
			return cliErrorResult(err)
		}
		mcpResult, _ := ResultFromCLI(result)
		return mcpResult, nil, nil
	})
}

package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
)

// ImportInput is the input schema for zerops_import.
type ImportInput struct {
	Content  string `json:"content,omitempty"`
	FilePath string `json:"filePath,omitempty"`
	DryRun   bool   `json:"dryRun,omitempty"`
}

// RegisterImport registers the zerops_import tool on the server.
func RegisterImport(srv *mcp.Server, exec executor.Executor) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "zerops_import",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Import Services",
			DestructiveHint: boolPtr(false),
		},
		Description: `Import services from YAML into the current project.

YAML format contains a 'services:' array. Do NOT include 'project:' section.

Use dryRun=true to preview what would be created (sync validation).

Example YAML:
  services:
    - hostname: api
      type: nodejs@22
      minContainers: 1
    - hostname: db
      type: postgresql@16

Returns process ID for tracking via zerops_process.`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ImportInput) (*mcp.CallToolResult, any, error) {
		if input.Content == "" && input.FilePath == "" {
			return errorResult("content or filePath is required"), nil, nil
		}
		if input.Content != "" && input.FilePath != "" {
			return errorResult("provide either content or filePath, not both"), nil, nil
		}

		args := []string{"import"}
		if input.Content != "" {
			args = append(args, "--content", input.Content)
		} else if input.FilePath != "" {
			args = append(args, "--file", input.FilePath)
		}
		if input.DryRun {
			args = append(args, "--dry-run")
		}

		result, err := exec.RunZaia(ctx, args...)
		if err != nil {
			return cliErrorResult(err)
		}
		mcpResult, _ := ResultFromCLI(result)
		return mcpResult, nil, nil
	})
}

package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
)

// KnowledgeInput is the input schema for zerops_knowledge.
type KnowledgeInput struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

// RegisterKnowledge registers the zerops_knowledge tool on the server.
func RegisterKnowledge(srv *mcp.Server, exec executor.Executor) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "zerops_knowledge",
		Description: `Search Zerops knowledge base.

This is KEYWORD search (BM25). Use specific terms:
- Service names: postgresql, nodejs, valkey
- Config: zerops.yml, import.yml, build
- Errors: redirect loop, connection refused

Returns topResult with full document content of best match.`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input KnowledgeInput) (*mcp.CallToolResult, any, error) {
		if input.Query == "" {
			return errorResult("query is required"), nil, nil
		}

		args := []string{"search", input.Query}
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

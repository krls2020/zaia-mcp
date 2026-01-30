package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
)

// RegisterKnowledgeResources registers the zerops://docs/{path} resource template.
// Documents are fetched via `zaia search --get <uri>`.
func RegisterKnowledgeResources(srv *mcp.Server, exec executor.Executor) {
	srv.AddResourceTemplate(
		&mcp.ResourceTemplate{
			URITemplate: "zerops://docs/{+path}",
			Name:        "zerops-docs",
			Description: "Zerops knowledge base documents. Use zerops_knowledge tool to search, then read specific docs via this resource.",
			MIMEType:    "text/markdown",
		},
		func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			uri := req.Params.URI
			if !strings.HasPrefix(uri, "zerops://docs/") {
				return nil, mcp.ResourceNotFoundError(uri)
			}

			result, err := exec.RunZaia(ctx, "search", "--get", uri)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch resource: %w", err)
			}

			// Parse the CLI response
			if len(result.Stdout) == 0 {
				return nil, mcp.ResourceNotFoundError(uri)
			}

			var resp struct {
				Type string          `json:"type"`
				Data json.RawMessage `json:"data"`
				Code string          `json:"code"`
			}
			if err := json.Unmarshal(result.Stdout, &resp); err != nil {
				return nil, fmt.Errorf("invalid CLI response: %w", err)
			}

			if resp.Type == "error" {
				return nil, mcp.ResourceNotFoundError(uri)
			}

			// Extract content from data
			var doc struct {
				URI     string `json:"uri"`
				Title   string `json:"title"`
				Content string `json:"content"`
			}
			if err := json.Unmarshal(resp.Data, &doc); err != nil {
				return nil, fmt.Errorf("invalid doc data: %w", err)
			}

			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{
					{
						URI:      uri,
						MIMEType: "text/markdown",
						Text:     doc.Content,
					},
				},
			}, nil
		},
	)
}

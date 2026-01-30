package tools

import (
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
)

// CLIResponse represents the parsed ZAIA CLI JSON output envelope.
type CLIResponse struct {
	Type       string          `json:"type"`                 // "sync", "async", "error"
	Status     string          `json:"status,omitempty"`     // "ok" or "initiated"
	Data       json.RawMessage `json:"data,omitempty"`       // sync payload
	Processes  json.RawMessage `json:"processes,omitempty"`  // async payload
	Code       string          `json:"code,omitempty"`       // error code
	Error      string          `json:"error,omitempty"`      // error message
	Suggestion string          `json:"suggestion,omitempty"` // error suggestion
	Context    json.RawMessage `json:"context,omitempty"`    // error context
}

// ParseCLIResponse parses the CLI stdout into a CLIResponse.
func ParseCLIResponse(result *executor.Result) (*CLIResponse, error) {
	if len(result.Stdout) == 0 {
		return nil, fmt.Errorf("empty CLI output (exit code %d, stderr: %s)", result.ExitCode, result.Stderr)
	}

	var resp CLIResponse
	if err := json.Unmarshal(result.Stdout, &resp); err != nil {
		return nil, fmt.Errorf("invalid CLI JSON output: %w (raw: %s)", err, result.Stdout)
	}

	return &resp, nil
}

// ToMCPResult converts a CLIResponse to an MCP CallToolResult.
func ToMCPResult(resp *CLIResponse) *mcp.CallToolResult {
	switch resp.Type {
	case "sync":
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(resp.Data)},
			},
		}
	case "async":
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(resp.Processes)},
			},
		}
	case "error":
		text := formatError(resp)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
			},
			IsError: true,
		}
	default:
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Unknown CLI response type: %s", resp.Type)},
			},
			IsError: true,
		}
	}
}

// ResultFromCLI is a convenience that parses and converts in one step.
func ResultFromCLI(result *executor.Result) (*mcp.CallToolResult, error) {
	resp, err := ParseCLIResponse(result)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("CLI execution error: %v", err)},
			},
			IsError: true,
		}, nil
	}
	return ToMCPResult(resp), nil
}

func formatError(resp *CLIResponse) string {
	result := map[string]interface{}{
		"code":  resp.Code,
		"error": resp.Error,
	}
	if resp.Suggestion != "" {
		result["suggestion"] = resp.Suggestion
	}
	if len(resp.Context) > 0 {
		var ctx interface{}
		if err := json.Unmarshal(resp.Context, &ctx); err == nil {
			result["context"] = ctx
		}
	}
	b, _ := json.Marshal(result)
	return string(b)
}

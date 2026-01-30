package integration

import (
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
	"github.com/zeropsio/zaia-mcp/internal/server"
)

// Harness provides a test harness for end-to-end MCP flows.
type Harness struct {
	t       *testing.T
	mock    *executor.MockExecutor
	srv     *server.MCPServer
	session *mcp.ClientSession
}

// NewHarness creates a new test harness with a mock executor.
func NewHarness(t *testing.T) *Harness {
	t.Helper()
	mock := executor.NewMockExecutor()
	srv := server.NewWithExecutor(mock)

	ctx := t.Context()
	t1, t2 := mcp.NewInMemoryTransports()
	if _, err := srv.Server().Connect(ctx, t1, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.1"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })

	return &Harness{
		t:       t,
		mock:    mock,
		srv:     srv,
		session: session,
	}
}

// Mock returns the mock executor for configuring responses.
func (h *Harness) Mock() *executor.MockExecutor {
	return h.mock
}

// Call calls an MCP tool and returns the result.
func (h *Harness) Call(name string, args map[string]interface{}) *mcp.CallToolResult {
	h.t.Helper()
	result, err := h.session.CallTool(h.t.Context(), &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		h.t.Fatalf("CallTool(%q): %v", name, err)
	}
	return result
}

// MustCallSuccess calls a tool and asserts no error.
func (h *Harness) MustCallSuccess(name string, args map[string]interface{}) string {
	h.t.Helper()
	result := h.Call(name, args)
	if result.IsError {
		h.t.Fatalf("expected success, got error: %s", h.GetText(result))
	}
	return h.GetText(result)
}

// MustCallError calls a tool and asserts it returns an error.
func (h *Harness) MustCallError(name string, args map[string]interface{}) string {
	h.t.Helper()
	result := h.Call(name, args)
	if !result.IsError {
		h.t.Fatalf("expected error, got success: %s", h.GetText(result))
	}
	return h.GetText(result)
}

// GetText extracts the text content from a result.
func (h *Harness) GetText(result *mcp.CallToolResult) string {
	h.t.Helper()
	if len(result.Content) == 0 {
		return ""
	}
	b, err := json.Marshal(result.Content[0])
	if err != nil {
		h.t.Fatal(err)
	}
	var obj struct {
		Text string `json:"text"`
	}
	_ = json.Unmarshal(b, &obj)
	return obj.Text
}

// ListTools returns all registered tool names.
func (h *Harness) ListTools() []string {
	h.t.Helper()
	names := make([]string, 0, 11)
	for tool, err := range h.session.Tools(h.t.Context(), nil) {
		if err != nil {
			h.t.Fatalf("ListTools: %v", err)
		}
		names = append(names, tool.Name)
	}
	return names
}

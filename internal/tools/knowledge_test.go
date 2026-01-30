package tools_test

import (
	"testing"

	"github.com/zeropsio/zaia-mcp/internal/executor"
	"github.com/zeropsio/zaia-mcp/internal/tools"
)

func TestKnowledge_Basic(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.SyncResult(`{"results":[]}`))
	srv := testServer(t, tools.RegisterKnowledge, mock)
	result := callTool(t, srv, "zerops_knowledge", map[string]interface{}{
		"query": "postgresql connection string",
	})
	if result.IsError {
		t.Errorf("unexpected error: %s", getTextContent(t, result))
	}
	assertArgs(t, mock.Calls[0].Args, "search", "postgresql connection string")
}

func TestKnowledge_MissingQuery(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := testServer(t, tools.RegisterKnowledge, mock)
	callToolExpectError(t, srv, "zerops_knowledge", nil)
}

func TestKnowledge_WithLimit(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.SyncResult(`{"results":[]}`))
	srv := testServer(t, tools.RegisterKnowledge, mock)
	callTool(t, srv, "zerops_knowledge", map[string]interface{}{
		"query": "valkey",
		"limit": float64(3),
	})
	assertContains(t, mock.Calls[0].Args, "--limit")
}

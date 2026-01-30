package tools_test

import (
	"testing"

	"github.com/zeropsio/zaia-mcp/internal/executor"
	"github.com/zeropsio/zaia-mcp/internal/tools"
)

func TestLogs_Basic(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithZaiaResponse("logs --service api", executor.SyncResult(`{"entries":[]}`))
	srv := testServer(t, tools.RegisterLogs, mock)
	result := callTool(t, srv, "zerops_logs", map[string]interface{}{
		"serviceHostname": "api",
	})
	if result.IsError {
		t.Errorf("unexpected error: %s", getTextContent(t, result))
	}
}

func TestLogs_MissingService(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := testServer(t, tools.RegisterLogs, mock)
	// SDK validates required fields — missing serviceHostname returns protocol error
	callToolExpectError(t, srv, "zerops_logs", nil)
}

func TestLogs_WithFilters(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.SyncResult(`{"entries":[]}`))
	srv := testServer(t, tools.RegisterLogs, mock)
	callTool(t, srv, "zerops_logs", map[string]interface{}{
		"serviceHostname": "api",
		"severity":        "error",
		"since":           "1h",
		"limit":           float64(50),
	})
	args := mock.Calls[0].Args
	assertContains(t, args, "--severity")
	assertContains(t, args, "--since")
	assertContains(t, args, "--limit")
}

func TestLogs_EmptyServiceHostname(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := testServer(t, tools.RegisterLogs, mock)
	result := callTool(t, srv, "zerops_logs", map[string]interface{}{
		"serviceHostname": "",
	})
	if !result.IsError {
		t.Error("expected error for empty serviceHostname")
	}
}

func TestLogs_ZeroLimit(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.SyncResult(`{"entries":[]}`))
	srv := testServer(t, tools.RegisterLogs, mock)
	callTool(t, srv, "zerops_logs", map[string]interface{}{
		"serviceHostname": "api",
		"limit":           float64(0),
	})
	// limit=0 should not add --limit flag
	for _, arg := range mock.Calls[0].Args {
		if arg == "--limit" {
			t.Error("--limit should not be present for limit=0")
		}
	}
}

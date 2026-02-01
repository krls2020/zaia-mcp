package tools_test

import (
	"testing"

	"github.com/zeropsio/zaia-mcp/internal/executor"
	"github.com/zeropsio/zaia-mcp/internal/tools"
)

func TestEvents_Basic(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithZaiaResponse("events", executor.SyncResult(`{"events":[],"summary":{"total":0}}`))
	srv := testServer(t, tools.RegisterEvents, mock)
	result := callTool(t, srv, "zerops_events", nil)

	if result.IsError {
		t.Errorf("unexpected error: %s", getTextContent(t, result))
	}
	text := getTextContent(t, result)
	if text != `{"events":[],"summary":{"total":0}}` {
		t.Errorf("got %q", text)
	}
}

func TestEvents_WithService(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithZaiaResponse("events --service api", executor.SyncResult(`{"events":[]}`))
	srv := testServer(t, tools.RegisterEvents, mock)
	result := callTool(t, srv, "zerops_events", map[string]interface{}{
		"serviceHostname": "api",
	})

	if result.IsError {
		t.Errorf("unexpected error: %s", getTextContent(t, result))
	}
	assertArgs(t, mock.Calls[0].Args, "events", "--service", "api")
}

func TestEvents_WithLimit(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithZaiaResponse("events --limit 10", executor.SyncResult(`{"events":[]}`))
	srv := testServer(t, tools.RegisterEvents, mock)
	callTool(t, srv, "zerops_events", map[string]interface{}{
		"limit": 10,
	})
	assertArgs(t, mock.Calls[0].Args, "events", "--limit", "10")
}

func TestEvents_Error(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithZaiaResponse("events", executor.ErrorResult("AUTH_REQUIRED", "Not authenticated", "Run: zaia login", 2))
	srv := testServer(t, tools.RegisterEvents, mock)
	result := callTool(t, srv, "zerops_events", nil)
	if !result.IsError {
		t.Error("expected error result")
	}
}

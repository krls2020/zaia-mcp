package tools_test

import (
	"testing"

	"github.com/zeropsio/zaia-mcp/internal/executor"
	"github.com/zeropsio/zaia-mcp/internal/tools"
)

func TestProcess_Status(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.SyncResult(`{"processId":"p1","status":"FINISHED"}`))
	srv := testServer(t, tools.RegisterProcess, mock)
	result := callTool(t, srv, "zerops_process", map[string]interface{}{
		"processId": "p1",
	})
	if result.IsError {
		t.Errorf("unexpected error: %s", getTextContent(t, result))
	}
	assertArgs(t, mock.Calls[0].Args, "process", "p1")
}

func TestProcess_Cancel(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.SyncResult(`{"processId":"p1","status":"CANCELED"}`))
	srv := testServer(t, tools.RegisterProcess, mock)
	callTool(t, srv, "zerops_process", map[string]interface{}{
		"processId": "p1",
		"action":    "cancel",
	})
	assertArgs(t, mock.Calls[0].Args, "cancel", "p1")
}

func TestProcess_MissingID(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := testServer(t, tools.RegisterProcess, mock)
	callToolExpectError(t, srv, "zerops_process", nil)
}

func TestProcess_InvalidAction(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := testServer(t, tools.RegisterProcess, mock)
	result := callTool(t, srv, "zerops_process", map[string]interface{}{
		"processId": "p1",
		"action":    "invalid",
	})
	if !result.IsError {
		t.Error("expected error for invalid action")
	}
}

package tools_test

import (
	"testing"

	"github.com/zeropsio/zaia-mcp/internal/executor"
	"github.com/zeropsio/zaia-mcp/internal/tools"
)

func TestDelete_Confirmed(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.AsyncResult(`[{"processId":"p1"}]`))
	srv := testServer(t, tools.RegisterDelete, mock)
	result := callTool(t, srv, "zerops_delete", map[string]interface{}{
		"serviceHostname": "old-service",
		"confirm":         true,
	})
	if result.IsError {
		t.Errorf("unexpected error: %s", getTextContent(t, result))
	}
	assertContains(t, mock.Calls[0].Args, "--confirm")
}

func TestDelete_NotConfirmed(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := testServer(t, tools.RegisterDelete, mock)
	// confirm=false means the handler validates it (not schema-level required)
	result := callTool(t, srv, "zerops_delete", map[string]interface{}{
		"serviceHostname": "api",
		"confirm":         false,
	})
	if !result.IsError {
		t.Error("expected error when confirm is false")
	}
}

func TestDelete_MissingService(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := testServer(t, tools.RegisterDelete, mock)
	callToolExpectError(t, srv, "zerops_delete", map[string]interface{}{
		"confirm": true,
	})
}

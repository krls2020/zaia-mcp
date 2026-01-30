package tools_test

import (
	"testing"

	"github.com/zeropsio/zaia-mcp/internal/executor"
	"github.com/zeropsio/zaia-mcp/internal/tools"
)

func TestDeploy_Basic(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.SyncResult(`{"deployed":true}`))
	srv := testServer(t, tools.RegisterDeploy, mock)
	result := callTool(t, srv, "zerops_deploy", nil)
	if result.IsError {
		t.Errorf("unexpected error: %s", getTextContent(t, result))
	}
	// Should call zcli, not zaia
	if mock.Calls[0].Binary != "zcli" {
		t.Errorf("expected zcli, got %s", mock.Calls[0].Binary)
	}
}

func TestDeploy_WithServiceID(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.SyncResult(`{"deployed":true}`))
	srv := testServer(t, tools.RegisterDeploy, mock)
	callTool(t, srv, "zerops_deploy", map[string]interface{}{
		"serviceId":  "svc-123",
		"workingDir": "/app",
	})
	assertContains(t, mock.Calls[0].Args, "--serviceId")
	assertContains(t, mock.Calls[0].Args, "--workingDir")
}

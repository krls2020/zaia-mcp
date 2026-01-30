package tools_test

import (
	"testing"

	"github.com/zeropsio/zaia-mcp/internal/executor"
	"github.com/zeropsio/zaia-mcp/internal/tools"
)

func TestEnv_Get(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.SyncResult(`{"envVars":[{"key":"PORT","value":"3000"}]}`))
	srv := testServer(t, tools.RegisterEnv, mock)
	result := callTool(t, srv, "zerops_env", map[string]interface{}{
		"action":          "get",
		"serviceHostname": "api",
	})
	if result.IsError {
		t.Errorf("unexpected error: %s", getTextContent(t, result))
	}
	assertArgs(t, mock.Calls[0].Args, "env", "get", "--service", "api")
}

func TestEnv_Set(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.AsyncResult(`[{"processId":"p1"}]`))
	srv := testServer(t, tools.RegisterEnv, mock)
	callTool(t, srv, "zerops_env", map[string]interface{}{
		"action":          "set",
		"serviceHostname": "api",
		"variables":       []interface{}{"KEY=value", "ANOTHER=val2"},
	})
	args := mock.Calls[0].Args
	assertContains(t, args, "KEY=value")
	assertContains(t, args, "ANOTHER=val2")
}

func TestEnv_ProjectScope(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.SyncResult(`{"envVars":[]}`))
	srv := testServer(t, tools.RegisterEnv, mock)
	callTool(t, srv, "zerops_env", map[string]interface{}{
		"action":  "get",
		"project": true,
	})
	assertContains(t, mock.Calls[0].Args, "--project")
}

func TestEnv_MissingScope(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := testServer(t, tools.RegisterEnv, mock)
	result := callTool(t, srv, "zerops_env", map[string]interface{}{
		"action": "get",
	})
	if !result.IsError {
		t.Error("expected error for missing scope")
	}
}

func TestEnv_EmptyAction(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := testServer(t, tools.RegisterEnv, mock)
	result := callTool(t, srv, "zerops_env", map[string]interface{}{
		"action":          "",
		"serviceHostname": "api",
	})
	if !result.IsError {
		t.Error("expected error for empty action")
	}
}

func TestEnv_Delete(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.AsyncResult(`[{"processId":"p1"}]`))
	srv := testServer(t, tools.RegisterEnv, mock)
	callTool(t, srv, "zerops_env", map[string]interface{}{
		"action":          "delete",
		"serviceHostname": "api",
		"variables":       []interface{}{"OLD_KEY"},
	})
	args := mock.Calls[0].Args
	assertArgs(t, args, "env", "delete", "--service", "api")
	assertContains(t, args, "OLD_KEY")
}

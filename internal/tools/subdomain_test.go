package tools_test

import (
	"testing"

	"github.com/zeropsio/zaia-mcp/internal/executor"
	"github.com/zeropsio/zaia-mcp/internal/tools"
)

func TestSubdomain_Enable(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.AsyncResult(`[{"processId":"p1"}]`))
	srv := testServer(t, tools.RegisterSubdomain, mock)
	result := callTool(t, srv, "zerops_subdomain", map[string]interface{}{
		"serviceHostname": "api",
		"action":          "enable",
	})
	if result.IsError {
		t.Errorf("unexpected error: %s", getTextContent(t, result))
	}
	assertArgs(t, mock.Calls[0].Args, "subdomain", "enable", "--service", "api")
}

func TestSubdomain_InvalidAction(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := testServer(t, tools.RegisterSubdomain, mock)
	result := callTool(t, srv, "zerops_subdomain", map[string]interface{}{
		"serviceHostname": "api",
		"action":          "toggle",
	})
	if !result.IsError {
		t.Error("expected error for invalid action")
	}
}

func TestSubdomain_MissingService(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := testServer(t, tools.RegisterSubdomain, mock)
	callToolExpectError(t, srv, "zerops_subdomain", map[string]interface{}{
		"action": "enable",
	})
}

func TestSubdomain_Disable(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.AsyncResult(`[{"processId":"p1"}]`))
	srv := testServer(t, tools.RegisterSubdomain, mock)
	result := callTool(t, srv, "zerops_subdomain", map[string]interface{}{
		"serviceHostname": "api",
		"action":          "disable",
	})
	if result.IsError {
		t.Errorf("unexpected error: %s", getTextContent(t, result))
	}
	assertArgs(t, mock.Calls[0].Args, "subdomain", "disable", "--service", "api")
}

func TestSubdomain_EmptyServiceHostname(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := testServer(t, tools.RegisterSubdomain, mock)
	result := callTool(t, srv, "zerops_subdomain", map[string]interface{}{
		"serviceHostname": "",
		"action":          "enable",
	})
	if !result.IsError {
		t.Error("expected error for empty serviceHostname")
	}
}

func TestSubdomain_EmptyAction(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := testServer(t, tools.RegisterSubdomain, mock)
	result := callTool(t, srv, "zerops_subdomain", map[string]interface{}{
		"serviceHostname": "api",
		"action":          "",
	})
	if !result.IsError {
		t.Error("expected error for empty action")
	}
}

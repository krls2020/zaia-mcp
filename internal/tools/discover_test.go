package tools_test

import (
	"testing"

	"github.com/zeropsio/zaia-mcp/internal/executor"
	"github.com/zeropsio/zaia-mcp/internal/tools"
)

func TestDiscover_Basic(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithZaiaResponse("discover", executor.SyncResult(`{"services":[]}`))
	srv := testServer(t, tools.RegisterDiscover, mock)
	result := callTool(t, srv, "zerops_discover", nil)

	if result.IsError {
		t.Errorf("unexpected error: %s", getTextContent(t, result))
	}
	text := getTextContent(t, result)
	if text != `{"services":[]}` {
		t.Errorf("got %q", text)
	}
}

func TestDiscover_WithService(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithZaiaResponse("discover --service api", executor.SyncResult(`{"service":{"hostname":"api"}}`))
	srv := testServer(t, tools.RegisterDiscover, mock)
	result := callTool(t, srv, "zerops_discover", map[string]interface{}{
		"service": "api",
	})

	if result.IsError {
		t.Errorf("unexpected error: %s", getTextContent(t, result))
	}
	if len(mock.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.Calls))
	}
	args := mock.Calls[0].Args
	assertArgs(t, args, "discover", "--service", "api")
}

func TestDiscover_WithEnvs(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithZaiaResponse("discover --include-envs", executor.SyncResult(`{"services":[]}`))
	srv := testServer(t, tools.RegisterDiscover, mock)
	callTool(t, srv, "zerops_discover", map[string]interface{}{
		"includeEnvs": true,
	})
	assertArgs(t, mock.Calls[0].Args, "discover", "--include-envs")
}

func TestDiscover_Error(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithZaiaResponse("discover", executor.ErrorResult("AUTH_REQUIRED", "Not authenticated", "Run: zaia login", 2))
	srv := testServer(t, tools.RegisterDiscover, mock)
	result := callTool(t, srv, "zerops_discover", nil)
	if !result.IsError {
		t.Error("expected error result")
	}
}

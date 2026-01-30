package tools_test

import (
	"testing"

	"github.com/zeropsio/zaia-mcp/internal/executor"
	"github.com/zeropsio/zaia-mcp/internal/tools"
)

func TestManage_Start(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.AsyncResult(`[{"processId":"p1","status":"PENDING"}]`))
	srv := testServer(t, tools.RegisterManage, mock)
	result := callTool(t, srv, "zerops_manage", map[string]interface{}{
		"action":          "start",
		"serviceHostname": "api",
	})
	if result.IsError {
		t.Errorf("unexpected error: %s", getTextContent(t, result))
	}
	assertArgs(t, mock.Calls[0].Args, "start", "--service", "api")
}

func TestManage_Scale(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.AsyncResult(`[{"processId":"p1"}]`))
	srv := testServer(t, tools.RegisterManage, mock)
	callTool(t, srv, "zerops_manage", map[string]interface{}{
		"action":          "scale",
		"serviceHostname": "api",
		"minCpu":          float64(1),
		"maxCpu":          float64(4),
		"minRam":          0.5,
		"maxRam":          4.0,
	})
	args := mock.Calls[0].Args
	assertContains(t, args, "--min-cpu")
	assertContains(t, args, "--max-cpu")
	assertContains(t, args, "--min-ram")
	assertContains(t, args, "--max-ram")
}

func TestManage_MissingAction(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := testServer(t, tools.RegisterManage, mock)
	callToolExpectError(t, srv, "zerops_manage", map[string]interface{}{
		"serviceHostname": "api",
	})
}

func TestManage_MissingService(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := testServer(t, tools.RegisterManage, mock)
	callToolExpectError(t, srv, "zerops_manage", map[string]interface{}{
		"action": "start",
	})
}

func TestManage_EmptyAction(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := testServer(t, tools.RegisterManage, mock)
	result := callTool(t, srv, "zerops_manage", map[string]interface{}{
		"action":          "",
		"serviceHostname": "api",
	})
	if !result.IsError {
		t.Error("expected error for empty action")
	}
}

func TestManage_EmptyServiceHostname(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := testServer(t, tools.RegisterManage, mock)
	result := callTool(t, srv, "zerops_manage", map[string]interface{}{
		"action":          "start",
		"serviceHostname": "",
	})
	if !result.IsError {
		t.Error("expected error for empty serviceHostname")
	}
}

func TestManage_ScaleWithDisk(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.AsyncResult(`[{"processId":"p1"}]`))
	srv := testServer(t, tools.RegisterManage, mock)
	callTool(t, srv, "zerops_manage", map[string]interface{}{
		"action":          "scale",
		"serviceHostname": "api",
		"minDisk":         1.0,
		"maxDisk":         10.0,
	})
	args := mock.Calls[0].Args
	assertContains(t, args, "--min-disk")
	assertContains(t, args, "--max-disk")
}

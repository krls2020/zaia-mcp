package tools_test

import (
	"testing"

	"github.com/zeropsio/zaia-mcp/internal/executor"
	"github.com/zeropsio/zaia-mcp/internal/tools"
)

func TestImport_Content(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.AsyncResult(`[{"processId":"p1"}]`))
	srv := testServer(t, tools.RegisterImport, mock)
	result := callTool(t, srv, "zerops_import", map[string]interface{}{
		"content": "services:\n  - hostname: api\n    type: nodejs@22",
	})
	if result.IsError {
		t.Errorf("unexpected error: %s", getTextContent(t, result))
	}
	assertContains(t, mock.Calls[0].Args, "--content")
}

func TestImport_DryRun(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.SyncResult(`{"valid":true,"services":[{"hostname":"api"}]}`))
	srv := testServer(t, tools.RegisterImport, mock)
	callTool(t, srv, "zerops_import", map[string]interface{}{
		"content": "services:\n  - hostname: api\n    type: nodejs@22",
		"dryRun":  true,
	})
	assertContains(t, mock.Calls[0].Args, "--dry-run")
}

func TestImport_MissingContentAndFile(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := testServer(t, tools.RegisterImport, mock)
	result := callTool(t, srv, "zerops_import", nil)
	if !result.IsError {
		t.Error("expected error when neither content nor filePath provided")
	}
}

func TestImport_BothContentAndFile(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := testServer(t, tools.RegisterImport, mock)
	result := callTool(t, srv, "zerops_import", map[string]interface{}{
		"content":  "services: []",
		"filePath": "/path/to/import.yml",
	})
	if !result.IsError {
		t.Error("expected error when both content and filePath provided")
	}
}

func TestImport_FilePath(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.AsyncResult(`[{"processId":"p1"}]`))
	srv := testServer(t, tools.RegisterImport, mock)
	result := callTool(t, srv, "zerops_import", map[string]interface{}{
		"filePath": "/path/to/import.yml",
	})
	if result.IsError {
		t.Errorf("unexpected error: %s", getTextContent(t, result))
	}
	assertContains(t, mock.Calls[0].Args, "--file")
}

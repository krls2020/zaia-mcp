package tools_test

import (
	"testing"

	"github.com/zeropsio/zaia-mcp/internal/executor"
	"github.com/zeropsio/zaia-mcp/internal/tools"
)

func TestValidate_Content(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.SyncResult(`{"valid":true}`))
	srv := testServer(t, tools.RegisterValidate, mock)
	result := callTool(t, srv, "zerops_validate", map[string]interface{}{
		"content": "services:\n  - hostname: api\n    type: nodejs@22",
		"type":    "import.yml",
	})
	if result.IsError {
		t.Errorf("unexpected error: %s", getTextContent(t, result))
	}
	assertContains(t, mock.Calls[0].Args, "--content")
	assertContains(t, mock.Calls[0].Args, "--type")
}

func TestValidate_File(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.SyncResult(`{"valid":true}`))
	srv := testServer(t, tools.RegisterValidate, mock)
	callTool(t, srv, "zerops_validate", map[string]interface{}{
		"filePath": "/path/to/zerops.yml",
	})
	assertContains(t, mock.Calls[0].Args, "--file")
}

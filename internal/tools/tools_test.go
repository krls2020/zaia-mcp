package tools_test

import (
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
	"github.com/zeropsio/zaia-mcp/internal/tools"
)

// testServer creates a server with one tool registered and a mock executor.
func testServer(t *testing.T, register func(*mcp.Server, executor.Executor), mock *executor.MockExecutor) *mcp.Server {
	t.Helper()
	srv := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "0.0.1"},
		nil,
	)
	register(srv, mock)
	return srv
}

// callTool creates an in-memory MCP session and calls the named tool.
func callTool(t *testing.T, srv *mcp.Server, name string, args map[string]interface{}) *mcp.CallToolResult {
	t.Helper()
	ctx := t.Context()
	t1, t2 := mcp.NewInMemoryTransports()
	if _, err := srv.Connect(ctx, t1, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.1"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer session.Close()

	params := &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	}
	result, err := session.CallTool(ctx, params)
	if err != nil {
		t.Fatalf("CallTool(%q): %v", name, err)
	}
	return result
}

// callToolExpectError calls a tool and expects a protocol-level error (e.g. missing required params).
func callToolExpectError(t *testing.T, srv *mcp.Server, name string, args map[string]interface{}) {
	t.Helper()
	ctx := t.Context()
	t1, t2 := mcp.NewInMemoryTransports()
	if _, err := srv.Connect(ctx, t1, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.1"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer session.Close()

	params := &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	}
	_, err = session.CallTool(ctx, params)
	if err == nil {
		t.Fatal("expected error for invalid params")
	}
}

func getTextContent(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("no content in result")
	}
	b, err := json.Marshal(result.Content[0])
	if err != nil {
		t.Fatal(err)
	}
	var obj struct {
		Text string `json:"text"`
	}
	_ = json.Unmarshal(b, &obj)
	return obj.Text
}

// --- DISCOVER ---

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

// --- LOGS ---

func TestLogs_Basic(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithZaiaResponse("logs --service api", executor.SyncResult(`{"entries":[]}`))
	srv := testServer(t, tools.RegisterLogs, mock)
	result := callTool(t, srv, "zerops_logs", map[string]interface{}{
		"serviceHostname": "api",
	})
	if result.IsError {
		t.Errorf("unexpected error: %s", getTextContent(t, result))
	}
}

func TestLogs_MissingService(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := testServer(t, tools.RegisterLogs, mock)
	// SDK validates required fields — missing serviceHostname returns protocol error
	callToolExpectError(t, srv, "zerops_logs", nil)
}

func TestLogs_WithFilters(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.SyncResult(`{"entries":[]}`))
	srv := testServer(t, tools.RegisterLogs, mock)
	callTool(t, srv, "zerops_logs", map[string]interface{}{
		"serviceHostname": "api",
		"severity":        "error",
		"since":           "1h",
		"limit":           float64(50),
	})
	args := mock.Calls[0].Args
	assertContains(t, args, "--severity")
	assertContains(t, args, "--since")
	assertContains(t, args, "--limit")
}

// --- VALIDATE ---

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

// --- KNOWLEDGE ---

func TestKnowledge_Basic(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.SyncResult(`{"results":[]}`))
	srv := testServer(t, tools.RegisterKnowledge, mock)
	result := callTool(t, srv, "zerops_knowledge", map[string]interface{}{
		"query": "postgresql connection string",
	})
	if result.IsError {
		t.Errorf("unexpected error: %s", getTextContent(t, result))
	}
	assertArgs(t, mock.Calls[0].Args, "search", "postgresql connection string")
}

func TestKnowledge_MissingQuery(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := testServer(t, tools.RegisterKnowledge, mock)
	callToolExpectError(t, srv, "zerops_knowledge", nil)
}

func TestKnowledge_WithLimit(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithDefault(executor.SyncResult(`{"results":[]}`))
	srv := testServer(t, tools.RegisterKnowledge, mock)
	callTool(t, srv, "zerops_knowledge", map[string]interface{}{
		"query": "valkey",
		"limit": float64(3),
	})
	assertContains(t, mock.Calls[0].Args, "--limit")
}

// --- PROCESS ---

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

// --- MANAGE ---

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

// --- ENV ---

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

// --- IMPORT ---

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

// --- DELETE ---

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

// --- SUBDOMAIN ---

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

// --- DEPLOY ---

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

// --- Helpers ---

func assertArgs(t *testing.T, got []string, want ...string) {
	t.Helper()
	if len(got) < len(want) {
		t.Errorf("got %v, want at least %v", got, want)
		return
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("arg[%d]: got %q, want %q (full: %v)", i, got[i], w, got)
			return
		}
	}
}

func assertContains(t *testing.T, args []string, want string) {
	t.Helper()
	for _, a := range args {
		if a == want {
			return
		}
	}
	t.Errorf("args %v does not contain %q", args, want)
}

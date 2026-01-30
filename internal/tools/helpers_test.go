package tools_test

import (
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
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

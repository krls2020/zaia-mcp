//go:build e2e

package e2e

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os/exec"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
	"github.com/zeropsio/zaia-mcp/internal/server"
)

// session wraps an MCP client session with test helpers.
type session struct {
	t      *testing.T
	client *mcp.ClientSession
}

// connectInMemory creates an in-memory MCP client connected to a real CLIExecutor server.
func connectInMemory(t *testing.T) *session {
	t.Helper()

	exec := executor.NewCLIExecutor("", "")
	srv := server.NewWithExecutor(exec)

	ctx := t.Context()
	t1, t2 := mcp.NewInMemoryTransports()

	if _, err := srv.Server().Connect(ctx, t1, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "e2e-client", Version: "0.0.1"}, nil)
	cs, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { cs.Close() })

	return &session{t: t, client: cs}
}

// callTool calls an MCP tool and returns the result.
func (s *session) callTool(name string, args map[string]interface{}) *mcp.CallToolResult {
	s.t.Helper()
	result, err := s.client.CallTool(s.t.Context(), &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		s.t.Fatalf("CallTool(%q): %v", name, err)
	}
	return result
}

// mustCallSuccess calls a tool and asserts it returns no error.
func (s *session) mustCallSuccess(name string, args map[string]interface{}) string {
	s.t.Helper()
	result := s.callTool(name, args)
	text := getTextContent(s.t, result)
	if result.IsError {
		s.t.Fatalf("expected success from %s, got error: %s", name, text)
	}
	return text
}

// getTextContent extracts text from an MCP CallToolResult.
func getTextContent(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		return ""
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

// deleteTestService attempts to delete a service using exec.Command directly.
// This bypasses the MCP session (whose context may already be canceled during cleanup)
// and calls zaia CLI directly with its own timeout context.
func deleteTestService(t *testing.T, hostname string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "zaia", "delete", "--service", hostname, "--confirm")
	out, err := cmd.Output()
	if err != nil {
		t.Logf("cleanup: delete %s failed: %v", hostname, err)
		return
	}
	t.Logf("cleanup: delete %s → %s", hostname, string(out))
}

// randomSuffix returns a short random hex string for unique hostnames.
func randomSuffix() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

package tools

import (
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
)

func TestParseCLIResponse_Sync(t *testing.T) {
	result := &executor.Result{
		Stdout: []byte(`{"type":"sync","status":"ok","data":{"project":{"id":"p1"}}}`),
	}
	resp, err := ParseCLIResponse(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Type != "sync" {
		t.Errorf("got type %q, want %q", resp.Type, "sync")
	}
	if resp.Status != "ok" {
		t.Errorf("got status %q, want %q", resp.Status, "ok")
	}
}

func TestParseCLIResponse_Async(t *testing.T) {
	result := &executor.Result{
		Stdout: []byte(`{"type":"async","status":"initiated","processes":[{"processId":"p1","status":"PENDING"}]}`),
	}
	resp, err := ParseCLIResponse(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Type != "async" {
		t.Errorf("got type %q, want %q", resp.Type, "async")
	}
}

func TestParseCLIResponse_Error(t *testing.T) {
	result := &executor.Result{
		Stdout:   []byte(`{"type":"error","code":"AUTH_REQUIRED","error":"Not authenticated","suggestion":"Run: zaia login <token>"}`),
		ExitCode: 2,
	}
	resp, err := ParseCLIResponse(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Type != "error" {
		t.Errorf("got type %q, want %q", resp.Type, "error")
	}
	if resp.Code != "AUTH_REQUIRED" {
		t.Errorf("got code %q, want %q", resp.Code, "AUTH_REQUIRED")
	}
}

func TestParseCLIResponse_EmptyOutput(t *testing.T) {
	result := &executor.Result{
		Stdout:   []byte{},
		ExitCode: 1,
	}
	_, err := ParseCLIResponse(result)
	if err == nil {
		t.Fatal("expected error for empty output")
	}
}

func TestParseCLIResponse_InvalidJSON(t *testing.T) {
	result := &executor.Result{
		Stdout: []byte("not json"),
	}
	_, err := ParseCLIResponse(result)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestToMCPResult_Sync(t *testing.T) {
	resp := &CLIResponse{
		Type:   "sync",
		Status: "ok",
		Data:   json.RawMessage(`{"services":[]}`),
	}
	result := ToMCPResult(resp)
	if result.IsError {
		t.Error("sync result should not be error")
	}
	if len(result.Content) != 1 {
		t.Fatalf("got %d content items, want 1", len(result.Content))
	}
	text := mustText(t, result)
	if text != `{"services":[]}` {
		t.Errorf("got text %q", text)
	}
}

func TestToMCPResult_Async(t *testing.T) {
	resp := &CLIResponse{
		Type:      "async",
		Status:    "initiated",
		Processes: json.RawMessage(`[{"processId":"p1"}]`),
	}
	result := ToMCPResult(resp)
	if result.IsError {
		t.Error("async result should not be error")
	}
	text := mustText(t, result)
	if text != `[{"processId":"p1"}]` {
		t.Errorf("got text %q", text)
	}
}

func TestToMCPResult_Error(t *testing.T) {
	resp := &CLIResponse{
		Type:       "error",
		Code:       "SERVICE_NOT_FOUND",
		Error:      "Service 'xyz' not found",
		Suggestion: "Available services: api, db",
	}
	result := ToMCPResult(resp)
	if !result.IsError {
		t.Error("error result should have IsError=true")
	}
	text := mustText(t, result)
	var errObj map[string]interface{}
	if err := json.Unmarshal([]byte(text), &errObj); err != nil {
		t.Fatalf("error text should be JSON: %v", err)
	}
	if errObj["code"] != "SERVICE_NOT_FOUND" {
		t.Errorf("got code %v", errObj["code"])
	}
}

func TestToMCPResult_UnknownType(t *testing.T) {
	resp := &CLIResponse{Type: "unknown"}
	result := ToMCPResult(resp)
	if !result.IsError {
		t.Error("unknown type should be error")
	}
}

func TestResultFromCLI_Success(t *testing.T) {
	cliResult := executor.SyncResult(`{"ok":true}`)
	result, err := ResultFromCLI(cliResult)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("should not be error")
	}
}

func TestResultFromCLI_ParseError(t *testing.T) {
	cliResult := &executor.Result{Stdout: []byte("garbage")}
	result, err := ResultFromCLI(cliResult)
	if err != nil {
		t.Fatalf("ResultFromCLI should not return Go error: %v", err)
	}
	if !result.IsError {
		t.Error("parse error should set IsError=true")
	}
}

func mustText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("no content")
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

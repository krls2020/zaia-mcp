package executor

import (
	"context"
	"testing"
)

func TestMockExecutor_ZaiaResponse(t *testing.T) {
	mock := NewMockExecutor().
		WithZaiaResponse("discover", SyncResult(`{"services":[]}`))

	result, err := mock.RunZaia(context.Background(), "discover")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("got exit code %d, want 0", result.ExitCode)
	}
	expected := `{"type":"sync","status":"ok","data":{"services":[]}}`
	if string(result.Stdout) != expected {
		t.Errorf("got %q, want %q", result.Stdout, expected)
	}
}

func TestMockExecutor_ZcliResponse(t *testing.T) {
	mock := NewMockExecutor().
		WithZcliResponse("push", SyncResult(`{"deployed":true}`))

	result, err := mock.RunZcli(context.Background(), "push")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("got exit code %d, want 0", result.ExitCode)
	}
}

func TestMockExecutor_ErrorResponse(t *testing.T) {
	mock := NewMockExecutor().
		WithZaiaResponse("discover", ErrorResult("AUTH_REQUIRED", "Not authenticated", "Run: zaia login", 2))

	result, err := mock.RunZaia(context.Background(), "discover")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExitCode != 2 {
		t.Errorf("got exit code %d, want 2", result.ExitCode)
	}
}

func TestMockExecutor_NoResponse(t *testing.T) {
	mock := NewMockExecutor()
	_, err := mock.RunZaia(context.Background(), "unknown")
	if err == nil {
		t.Fatal("expected error for unconfigured command")
	}
}

func TestMockExecutor_Default(t *testing.T) {
	mock := NewMockExecutor().
		WithDefault(SyncResult(`{}`))

	result, err := mock.RunZaia(context.Background(), "anything")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("got exit code %d, want 0", result.ExitCode)
	}
}

func TestMockExecutor_CallRecording(t *testing.T) {
	mock := NewMockExecutor().
		WithDefault(SyncResult(`{}`))

	mock.RunZaia(context.Background(), "discover")
	mock.RunZcli(context.Background(), "push", "--serviceId", "abc")

	if len(mock.Calls) != 2 {
		t.Fatalf("got %d calls, want 2", len(mock.Calls))
	}
	if mock.Calls[0].Binary != "zaia" {
		t.Errorf("call 0: got binary %q, want %q", mock.Calls[0].Binary, "zaia")
	}
	if mock.Calls[1].Binary != "zcli" {
		t.Errorf("call 1: got binary %q, want %q", mock.Calls[1].Binary, "zcli")
	}
}

func TestMockExecutor_ZaiaError(t *testing.T) {
	mock := NewMockExecutor().
		WithZaiaError("discover", context.DeadlineExceeded)

	_, err := mock.RunZaia(context.Background(), "discover")
	if err != context.DeadlineExceeded {
		t.Errorf("got error %v, want DeadlineExceeded", err)
	}
}

func TestAsyncResult(t *testing.T) {
	result := AsyncResult(`[{"processId":"p1","status":"PENDING"}]`)
	expected := `{"type":"async","status":"initiated","processes":[{"processId":"p1","status":"PENDING"}]}`
	if string(result.Stdout) != expected {
		t.Errorf("got %q, want %q", result.Stdout, expected)
	}
}

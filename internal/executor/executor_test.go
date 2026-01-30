package executor

import (
	"context"
	"testing"
	"time"
)

func TestCLIExecutor_Echo(t *testing.T) {
	exec := NewCLIExecutor("echo", "echo")
	result, err := exec.RunZaia(context.Background(), "hello", "world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result.Stdout) != "hello world\n" {
		t.Errorf("got stdout %q, want %q", result.Stdout, "hello world\n")
	}
	if result.ExitCode != 0 {
		t.Errorf("got exit code %d, want 0", result.ExitCode)
	}
}

func TestCLIExecutor_NonZeroExit(t *testing.T) {
	exec := NewCLIExecutor("false", "false")
	result, err := exec.RunZaia(context.Background())
	if err != nil {
		t.Fatalf("non-zero exit should not return Go error, got: %v", err)
	}
	if result.ExitCode == 0 {
		t.Error("expected non-zero exit code")
	}
}

func TestCLIExecutor_BinaryNotFound(t *testing.T) {
	exec := NewCLIExecutor("nonexistent-binary-xyz", "")
	_, err := exec.RunZaia(context.Background())
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
}

func TestCLIExecutor_ContextCancellation(t *testing.T) {
	exec := NewCLIExecutor("sleep", "sleep")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := exec.RunZaia(ctx, "10")
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestCLIExecutor_Stderr(t *testing.T) {
	exec := NewCLIExecutor("sh", "sh")
	result, err := exec.RunZaia(context.Background(), "-c", "echo stderr_msg >&2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result.Stderr) != "stderr_msg\n" {
		t.Errorf("got stderr %q, want %q", result.Stderr, "stderr_msg\n")
	}
}

func TestCLIExecutor_Defaults(t *testing.T) {
	exec := NewCLIExecutor("", "")
	if exec.ZaiaBinary != "zaia" {
		t.Errorf("got ZaiaBinary %q, want %q", exec.ZaiaBinary, "zaia")
	}
	if exec.ZcliBinary != "zcli" {
		t.Errorf("got ZcliBinary %q, want %q", exec.ZcliBinary, "zcli")
	}
}

func TestCLIExecutor_RunZcli(t *testing.T) {
	exec := NewCLIExecutor("", "echo")
	result, err := exec.RunZcli(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result.Stdout) != "test\n" {
		t.Errorf("got stdout %q, want %q", result.Stdout, "test\n")
	}
}

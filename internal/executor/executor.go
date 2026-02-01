package executor

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
)

// Result holds the output of a CLI subprocess execution.
type Result struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

// Executor defines how ZAIA-MCP calls CLI subprocesses.
type Executor interface {
	// RunZaia executes `zaia <args...>` and returns stdout/stderr/exit code.
	RunZaia(ctx context.Context, args ...string) (*Result, error)
	// RunZcli executes `zcli <args...>` and returns stdout/stderr/exit code.
	RunZcli(ctx context.Context, args ...string) (*Result, error)
}

const (
	defaultZaiaBinary = "zaia"
	defaultZcliBinary = "zcli"
)

// CLIExecutor implements Executor using exec.CommandContext.
type CLIExecutor struct {
	ZaiaBinary string   // path to zaia binary (default: "zaia")
	ZcliBinary string   // path to zcli binary (default: "zcli")
	env        []string // process environment with resolved PATH
}

// NewCLIExecutor creates a new CLIExecutor with the given binary paths.
// Empty strings use defaults ("zaia" and "zcli").
func NewCLIExecutor(zaiaBinary, zcliBinary string) *CLIExecutor {
	if zaiaBinary == "" {
		zaiaBinary = defaultZaiaBinary
	}
	if zcliBinary == "" {
		zcliBinary = defaultZcliBinary
	}
	env := buildEnvWithResolvedPATH()
	return &CLIExecutor{
		ZaiaBinary: zaiaBinary,
		ZcliBinary: zcliBinary,
		env:        env,
	}
}

// RunZaia executes the zaia CLI binary with the given arguments.
func (e *CLIExecutor) RunZaia(ctx context.Context, args ...string) (*Result, error) {
	return e.run(ctx, e.ZaiaBinary, args...)
}

// RunZcli executes the zcli CLI binary with the given arguments.
func (e *CLIExecutor) RunZcli(ctx context.Context, args ...string) (*Result, error) {
	return e.run(ctx, e.ZcliBinary, args...)
}

func (e *CLIExecutor) run(ctx context.Context, binary string, args ...string) (*Result, error) {
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Env = e.env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &Result{
		Stdout: stdout.Bytes(),
		Stderr: stderr.Bytes(),
	}

	if err != nil {
		// Check context cancellation first
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
			// Non-zero exit is not a Go error — CLI outputs JSON on stdout
			return result, nil
		}
		// Real error (binary not found, etc.)
		return result, err
	}

	return result, nil
}

// resolveShellPATH runs the user's login shell to get the full PATH,
// including paths added by tools like nvm, homebrew, etc. that are
// configured in shell profiles but not available to MCP servers.
func resolveShellPATH() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	out, err := exec.CommandContext(context.Background(), shell, "-lc", "echo $PATH").Output()
	if err != nil {
		return os.Getenv("PATH")
	}
	resolved := strings.TrimSpace(string(out))
	if resolved == "" {
		return os.Getenv("PATH")
	}
	return resolved
}

// buildEnvWithResolvedPATH returns a copy of the current process environment
// with PATH replaced by the value from the user's login shell.
func buildEnvWithResolvedPATH() []string {
	resolvedPATH := resolveShellPATH()
	env := os.Environ()
	result := make([]string, 0, len(env))
	found := false
	for _, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			result = append(result, "PATH="+resolvedPATH)
			found = true
		} else {
			result = append(result, e)
		}
	}
	if !found {
		result = append(result, "PATH="+resolvedPATH)
	}
	return result
}

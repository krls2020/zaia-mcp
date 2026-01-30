package executor

import (
	"context"
	"fmt"
	"strings"
)

// MockExecutor returns configurable responses for testing.
type MockExecutor struct {
	// responses maps "binary arg1 arg2 ..." to Result
	responses map[string]*Result
	// errors maps "binary arg1 arg2 ..." to error
	errors map[string]error
	// defaultResponse is returned when no specific mapping exists
	defaultResponse *Result
	// calls records all calls made (for assertions)
	Calls []MockCall
}

// MockCall records a single executor call.
type MockCall struct {
	Binary string
	Args   []string
}

// NewMockExecutor creates a new mock executor.
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		responses: make(map[string]*Result),
		errors:    make(map[string]error),
	}
}

// WithResponse configures a response for a specific command.
// key format: "arg1 arg2 ..." (without binary name)
func (m *MockExecutor) WithZaiaResponse(args string, result *Result) *MockExecutor {
	m.responses["zaia "+args] = result
	return m
}

// WithZcliResponse configures a response for a specific zcli command.
func (m *MockExecutor) WithZcliResponse(args string, result *Result) *MockExecutor {
	m.responses["zcli "+args] = result
	return m
}

// WithZaiaError configures an error for a specific zaia command.
func (m *MockExecutor) WithZaiaError(args string, err error) *MockExecutor {
	m.errors["zaia "+args] = err
	return m
}

// WithDefault sets a default response for unmatched commands.
func (m *MockExecutor) WithDefault(result *Result) *MockExecutor {
	m.defaultResponse = result
	return m
}

// RunZaia implements Executor.
func (m *MockExecutor) RunZaia(ctx context.Context, args ...string) (*Result, error) {
	return m.resolve(ctx, "zaia", args...)
}

// RunZcli implements Executor.
func (m *MockExecutor) RunZcli(ctx context.Context, args ...string) (*Result, error) {
	return m.resolve(ctx, "zcli", args...)
}

func (m *MockExecutor) resolve(_ context.Context, binary string, args ...string) (*Result, error) {
	m.Calls = append(m.Calls, MockCall{Binary: binary, Args: args})

	key := binary + " " + strings.Join(args, " ")

	if err, ok := m.errors[key]; ok {
		return nil, err
	}

	if result, ok := m.responses[key]; ok {
		return result, nil
	}

	// Try prefix matching: if the full key doesn't match, check if any registered
	// key is a prefix of the actual command. Note: Go map iteration order is
	// non-deterministic, so avoid registering multiple overlapping prefixes
	// (e.g. "zaia discover" and "zaia discover --service") — use exact keys instead.
	for k, result := range m.responses {
		if strings.HasPrefix(key, k) {
			return result, nil
		}
	}

	if m.defaultResponse != nil {
		return m.defaultResponse, nil
	}

	return nil, fmt.Errorf("mock: no response configured for %q", key)
}

// SyncResult creates a Result with sync JSON response.
func SyncResult(data string) *Result {
	return &Result{
		Stdout:   []byte(fmt.Sprintf(`{"type":"sync","status":"ok","data":%s}`, data)),
		ExitCode: 0,
	}
}

// AsyncResult creates a Result with async JSON response.
func AsyncResult(processes string) *Result {
	return &Result{
		Stdout:   []byte(fmt.Sprintf(`{"type":"async","status":"initiated","processes":%s}`, processes)),
		ExitCode: 0,
	}
}

// ErrorResult creates a Result with error JSON response.
func ErrorResult(code, msg, suggestion string, exitCode int) *Result {
	return &Result{
		Stdout:   []byte(fmt.Sprintf(`{"type":"error","code":"%s","error":"%s","suggestion":"%s"}`, code, msg, suggestion)),
		ExitCode: exitCode,
	}
}

package server

import (
	"testing"

	"github.com/zeropsio/zaia-mcp/internal/executor"
)

func TestNew(t *testing.T) {
	srv := New()
	if srv == nil {
		t.Fatal("New() returned nil")
	}
	if srv.server == nil {
		t.Fatal("server is nil")
	}
	if srv.executor == nil {
		t.Fatal("executor is nil")
	}
}

func TestNewWithExecutor(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := NewWithExecutor(mock)
	if srv == nil {
		t.Fatal("NewWithExecutor() returned nil")
	}
	if srv.executor != mock {
		t.Fatal("executor not set correctly")
	}
}

func TestInstructions(t *testing.T) {
	if Instructions == "" {
		t.Fatal("Instructions is empty")
	}
	// Verify key content is present
	checks := []string{
		"Zerops",
		"PaaS",
		"zerops.yml",
		"import.yml",
		"discover",
		"Critical Rules",
	}
	for _, check := range checks {
		found := false
		for i := 0; i < len(Instructions)-len(check)+1; i++ {
			if Instructions[i:i+len(check)] == check {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Instructions missing %q", check)
		}
	}
}

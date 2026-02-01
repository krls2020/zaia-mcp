package tools_test

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
	"github.com/zeropsio/zaia-mcp/internal/tools"
)

func TestAnnotations_AllToolsHaveTitleAndAnnotations(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "0.0.1"},
		nil,
	)

	// Register all 11 tools
	tools.RegisterDiscover(srv, mock)
	tools.RegisterLogs(srv, mock)
	tools.RegisterValidate(srv, mock)
	tools.RegisterKnowledge(srv, mock)
	tools.RegisterProcess(srv, mock)
	tools.RegisterManage(srv, mock)
	tools.RegisterEnv(srv, mock)
	tools.RegisterImport(srv, mock)
	tools.RegisterDelete(srv, mock)
	tools.RegisterSubdomain(srv, mock)
	tools.RegisterDeploy(srv, mock)

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

	toolMap := make(map[string]*mcp.Tool)
	for tool, err := range session.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		toolMap[tool.Name] = tool
	}

	type expected struct {
		title       string
		readOnly    bool
		destructive *bool
		idempotent  bool
		openWorld   *bool
	}

	boolPtr := func(b bool) *bool { return &b }

	tests := map[string]expected{
		"zerops_discover":  {title: "Discover Services", readOnly: true, idempotent: true, openWorld: nil},
		"zerops_logs":      {title: "Fetch Logs", readOnly: true, idempotent: true, openWorld: nil},
		"zerops_validate":  {title: "Validate Config", readOnly: true, idempotent: true, openWorld: boolPtr(false)},
		"zerops_knowledge": {title: "Search Knowledge", readOnly: true, idempotent: true, openWorld: boolPtr(false)},
		"zerops_process":   {title: "Check Process", readOnly: true, idempotent: true, openWorld: nil},
		"zerops_manage":    {title: "Manage Service", destructive: boolPtr(true)},
		"zerops_env":       {title: "Manage Env Vars", destructive: boolPtr(false)},
		"zerops_import":    {title: "Import Services", destructive: boolPtr(false)},
		"zerops_delete":    {title: "Delete Service", destructive: boolPtr(true)},
		"zerops_subdomain": {title: "Manage Subdomain", destructive: boolPtr(false), idempotent: true},
		"zerops_deploy":    {title: "Deploy Code", destructive: boolPtr(false)},
	}

	for name, exp := range tests {
		t.Run(name, func(t *testing.T) {
			tool, ok := toolMap[name]
			if !ok {
				t.Fatalf("tool %q not found", name)
			}
			if tool.Annotations == nil {
				t.Fatalf("tool %q has no annotations", name)
			}
			a := tool.Annotations

			if a.Title != exp.title {
				t.Errorf("Title: got %q, want %q", a.Title, exp.title)
			}
			if a.ReadOnlyHint != exp.readOnly {
				t.Errorf("ReadOnlyHint: got %v, want %v", a.ReadOnlyHint, exp.readOnly)
			}
			if a.IdempotentHint != exp.idempotent {
				t.Errorf("IdempotentHint: got %v, want %v", a.IdempotentHint, exp.idempotent)
			}
			if exp.destructive == nil {
				if a.DestructiveHint != nil {
					t.Errorf("DestructiveHint: got %v, want nil", *a.DestructiveHint)
				}
			} else {
				if a.DestructiveHint == nil {
					t.Errorf("DestructiveHint: got nil, want %v", *exp.destructive)
				} else if *a.DestructiveHint != *exp.destructive {
					t.Errorf("DestructiveHint: got %v, want %v", *a.DestructiveHint, *exp.destructive)
				}
			}
			if exp.openWorld == nil {
				if a.OpenWorldHint != nil {
					t.Errorf("OpenWorldHint: got %v, want nil", *a.OpenWorldHint)
				}
			} else {
				if a.OpenWorldHint == nil {
					t.Errorf("OpenWorldHint: got nil, want %v", *exp.openWorld)
				} else if *a.OpenWorldHint != *exp.openWorld {
					t.Errorf("OpenWorldHint: got %v, want %v", *a.OpenWorldHint, *exp.openWorld)
				}
			}
		})
	}
}

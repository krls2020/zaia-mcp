package resources_test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
	"github.com/zeropsio/zaia-mcp/internal/resources"
)

func TestKnowledgeResource_ReadDoc(t *testing.T) {
	docContent := `# PostgreSQL on Zerops

## TL;DR
PostgreSQL is the default database.`

	mock := executor.NewMockExecutor().
		WithZaiaResponse("search --get zerops://docs/services/postgresql",
			executor.SyncResult(`{"uri":"zerops://docs/services/postgresql","title":"PostgreSQL","content":"`+escapeJSON(docContent)+`"}`))

	srv := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "0.0.1"},
		&mcp.ServerOptions{HasResources: true},
	)
	resources.RegisterKnowledgeResources(srv, mock)

	ctx := context.Background()
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

	result, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
		URI: "zerops://docs/services/postgresql",
	})
	if err != nil {
		t.Fatalf("ReadResource: %v", err)
	}
	if len(result.Contents) != 1 {
		t.Fatalf("got %d contents, want 1", len(result.Contents))
	}
	if result.Contents[0].MIMEType != "text/markdown" {
		t.Errorf("got mimeType %q, want %q", result.Contents[0].MIMEType, "text/markdown")
	}
	if result.Contents[0].Text == "" {
		t.Error("content text is empty")
	}
}

func TestKnowledgeResource_NotFound(t *testing.T) {
	mock := executor.NewMockExecutor().
		WithZaiaResponse("search --get zerops://docs/nonexistent",
			executor.ErrorResult("NOT_FOUND", "Document not found", "", 4))

	srv := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "0.0.1"},
		&mcp.ServerOptions{HasResources: true},
	)
	resources.RegisterKnowledgeResources(srv, mock)

	ctx := context.Background()
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

	_, err = session.ReadResource(ctx, &mcp.ReadResourceParams{
		URI: "zerops://docs/nonexistent",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent resource")
	}
}

func TestKnowledgeResource_ListTemplates(t *testing.T) {
	mock := executor.NewMockExecutor()
	srv := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "0.0.1"},
		&mcp.ServerOptions{HasResources: true},
	)
	resources.RegisterKnowledgeResources(srv, mock)

	ctx := context.Background()
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

	var found bool
	for tmpl, err := range session.ResourceTemplates(ctx, nil) {
		if err != nil {
			t.Fatalf("ResourceTemplates: %v", err)
		}
		if tmpl.Name == "zerops-docs" {
			found = true
		}
	}
	if !found {
		t.Error("zerops-docs template not found")
	}
}

func escapeJSON(s string) string {
	// Simple escape for test strings
	result := ""
	for _, c := range s {
		switch c {
		case '"':
			result += `\"`
		case '\\':
			result += `\\`
		case '\n':
			result += `\n`
		case '\t':
			result += `\t`
		default:
			result += string(c)
		}
	}
	return result
}

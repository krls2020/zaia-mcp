package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
	"github.com/zeropsio/zaia-mcp/internal/resources"
	"github.com/zeropsio/zaia-mcp/internal/tools"
)

// Instructions is the MCP server instructions field.
// Token budget: ~250 tokens. Update carefully.
const Instructions = `Zerops

PaaS. Full Linux containers (Incus), bare-metal, SSH access. Not serverless.

Services
Runtime: nodejs php python go rust java dotnet elixir gleam bun deno
Container: alpine ubuntu docker(VM-based)
DB: postgresql(default) mariadb clickhouse
Cache: valkey(default, redis-compat) | keydb(deprecated)
Search: meilisearch(default) elasticsearch typesense qdrant(internal-only)
Queue: nats(default) kafka
Storage: object-storage(S3/MinIO) shared-storage(POSIX)
Web: nginx static(SPA-ready)

Files
zerops.yml = build + deploy + run config (per service)
import.yml = infrastructure-as-code (services array, NO project: section)

Critical Rules
- Internal networking: ALWAYS http://, NEVER https:// (SSL terminates at L7 balancer)
- Ports: 10-65435 only (0-9 and 65436+ reserved)
- HA mode: immutable after creation (cannot change single↔HA)
- prepareCommands: cached. initCommands: run every start
- Env var cross-ref: ${service_hostname} (underscore, not dash)
- Cloudflare: MUST use "Full (strict)" SSL mode
- No localhost — services communicate via hostname

Tools
discover → project info + service list (call first)
logs → service runtime/build logs
validate → check zerops.yml/import.yml before deploy
knowledge → BM25 search Zerops docs (use specific terms)
manage → start/stop/restart/scale (async)
configure → env vars + import infrastructure (async)
delete → remove service (requires confirm)
process → check async operation status

Defaults (use unless user specifies otherwise)
postgresql@16, valkey@8, meilisearch@1, nats, alpine base, NON_HA, SHARED CPU`

// MCPServer wraps the MCP server with ZAIA executor.
type MCPServer struct {
	server   *mcp.Server
	executor executor.Executor
}

// New creates a new ZAIA-MCP server with the default CLI executor.
func New() *MCPServer {
	return NewWithExecutor(executor.NewCLIExecutor("", ""))
}

// NewWithExecutor creates a new ZAIA-MCP server with a custom executor.
func NewWithExecutor(exec executor.Executor) *MCPServer {
	srv := mcp.NewServer(
		&mcp.Implementation{
			Name:    "zaia-mcp",
			Version: "0.1.0",
		},
		&mcp.ServerOptions{
			Instructions: Instructions,
		},
	)

	s := &MCPServer{
		server:   srv,
		executor: exec,
	}

	s.registerTools()
	s.registerResources()

	return s
}

// Run starts the MCP server over STDIO transport.
func (s *MCPServer) Run(ctx context.Context) error {
	return s.server.Run(ctx, &mcp.StdioTransport{})
}

// Server returns the underlying MCP server (for testing).
func (s *MCPServer) Server() *mcp.Server {
	return s.server
}

// registerTools registers all 11 MCP tools.
func (s *MCPServer) registerTools() {
	// Sync tools (5)
	tools.RegisterDiscover(s.server, s.executor)
	tools.RegisterLogs(s.server, s.executor)
	tools.RegisterValidate(s.server, s.executor)
	tools.RegisterKnowledge(s.server, s.executor)
	tools.RegisterProcess(s.server, s.executor)

	// Async tools (5)
	tools.RegisterManage(s.server, s.executor)
	tools.RegisterEnv(s.server, s.executor)
	tools.RegisterImport(s.server, s.executor)
	tools.RegisterDelete(s.server, s.executor)
	tools.RegisterSubdomain(s.server, s.executor)

	// Deploy (calls zcli, not zaia)
	tools.RegisterDeploy(s.server, s.executor)
}

// registerResources registers MCP resources.
func (s *MCPServer) registerResources() {
	resources.RegisterKnowledgeResources(s.server, s.executor)
}

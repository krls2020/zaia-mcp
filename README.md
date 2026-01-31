# ZAIA-MCP — MCP Server for AI Agents on Zerops

ZAIA-MCP is a thin [Model Context Protocol](https://modelcontextprotocol.io) (MCP) server that provides AI agents with structured access to [Zerops](https://zerops.io) PaaS operations. It contains no business logic — all logic lives in [ZAIA CLI](https://github.com/krls2020/zaia).

```
┌─────────────────────┐     ┌──────────────┐     ┌──────────────┐
│  Claude Code /      │     │  ZAIA-MCP    │     │  ZAIA CLI    │
│  Desktop            │────>│  (this repo) │────>│  (business   │
│                     │ MCP │  thin wrapper│exec │  logic)      │
│                     │STDIO│              │     └──────────────┘
└─────────────────────┘     │              │     ┌──────────────┐
                            │              │────>│  ZCLI        │
                            │              │exec │  (deploy)    │
                            └──────────────┘     └──────────────┘
```

| Property | Value |
|----------|-------|
| Transport | STDIO (not HTTP) |
| Auth | Pre-authenticated — ZAIA CLI handles auth |
| State | Stateless — each tool call = fresh CLI invocation |
| Business logic | None — all in ZAIA CLI |
| Tools | 11 MCP tools |
| Resources | `zerops://docs/{path}` via ResourceTemplate |
| Dependencies | 1 (MCP Go SDK v0.6.0) |

## MCP Tools

### Sync Tools (5)

| MCP Tool | CLI Command | Required Params |
|----------|-------------|-----------------|
| `zerops_discover` | `zaia discover` | — |
| `zerops_logs` | `zaia logs --service X` | serviceHostname |
| `zerops_validate` | `zaia validate` | content or filePath |
| `zerops_knowledge` | `zaia search "query"` | query |
| `zerops_process` | `zaia process <id>` / `zaia cancel <id>` | processId |

### Async Tools (5)

| MCP Tool | CLI Command | Required Params |
|----------|-------------|-----------------|
| `zerops_manage` | `zaia start/stop/restart/scale` | action, serviceHostname |
| `zerops_env` | `zaia env get/set/delete` | action, serviceHostname or project |
| `zerops_import` | `zaia import` | content or filePath |
| `zerops_delete` | `zaia delete --service X --confirm` | serviceHostname, confirm |
| `zerops_subdomain` | `zaia subdomain --service X --action Y` | serviceHostname, action |

### Deploy (via zcli)

| MCP Tool | CLI Command | Required Params |
|----------|-------------|-----------------|
| `zerops_deploy` | `zcli push` | — |

**Notes:**
- Deploy calls `zcli push` directly — not via ZAIA CLI
- `zerops_env` is sync for `get`, async for `set`/`delete`
- `zerops_process` supports `cancel` action (sync response)
- `zerops_subdomain` is idempotent — already enabled/disabled = sync success

## CLI Response Format

ZAIA CLI always outputs one of:

```json
{"type":"sync","status":"ok","data":{...}}
{"type":"async","status":"initiated","processes":[...]}
{"type":"error","code":"...","error":"...","suggestion":"..."}
```

The MCP server converts these to MCP results:
- `type=sync` → `TextContent{Text: data_json}`, `IsError: false`
- `type=async` → `TextContent{Text: processes_json}`, `IsError: false`
- `type=error` → `TextContent{Text: error_json}`, `IsError: true`

## MCP Resources

### `zerops://docs/{path}`

ResourceTemplate for knowledge docs. Calls `zaia search --get <uri>` internally.

## Instructions (System Prompt)

~250 token system prompt in `server.go` constant `Instructions`. Contains Zerops overview, tool summary, and defaults. Delivered automatically when the MCP server connects.

## Code Structure

```
zaia-mcp/
├── cmd/zaia-mcp/main.go           # Entry point — STDIO MCP server
├── internal/
│   ├── server/
│   │   └── server.go              # MCPServer — setup, Instructions, registration
│   ├── executor/
│   │   ├── executor.go            # Executor interface + CLIExecutor (exec.CommandContext)
│   │   └── mock.go                # MockExecutor for tests
│   ├── tools/
│   │   ├── convert.go             # ParseCLIResponse, ToMCPResult, ResultFromCLI
│   │   ├── discover.go ... subdomain.go  # 11 tool implementations
│   │   └── tools_test.go          # All tool tests (in-memory MCP sessions)
│   └── resources/
│       └── knowledge.go           # zerops://docs/{path} ResourceTemplate
└── integration/
    ├── harness.go                 # Test harness (in-memory MCP, mock executor)
    └── flow_test.go               # End-to-end flows (9 scenarios)
```

## Installation

### From GitHub Releases

Download the binary for your platform from [Releases](https://github.com/krls2020/zaia-mcp/releases) and place it on your PATH:

```bash
# Example: macOS Apple Silicon
curl -L https://github.com/krls2020/zaia-mcp/releases/latest/download/zaia-mcp-darwin-arm64 \
  -o ~/.local/bin/zaia-mcp
chmod +x ~/.local/bin/zaia-mcp
```

### macOS code signing fix

Release binaries from v0.4.0 and earlier were cross-compiled on Linux, which produces an invalid adhoc signature on macOS. If the binary gets killed immediately (SIGKILL), re-sign it:

```bash
codesign -f -s - ~/.local/bin/zaia-mcp
```

Releases from v0.5.0+ are built natively on macOS runners and do not need this step.

### Claude Code config

Add to your project's `.claude/settings.json` or `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "zaia-mcp": {
      "command": "zaia-mcp",
      "args": []
    }
  }
}
```

## Prerequisites

- **`zaia` binary** on PATH — handles all Zerops operations
- **`zcli` binary** on PATH — used only for `zerops_deploy` tool
- Both must be independently authenticated (`zaia login` + `zcli login`)

**PATH resolution:** The MCP server automatically resolves PATH from the user's login shell (`$SHELL -lc 'echo $PATH'`) at startup, so binaries installed via nvm, homebrew, or other profile-configured tools are found without manual PATH configuration.

## Dependencies

```
github.com/modelcontextprotocol/go-sdk v0.6.0  — MCP Go SDK
```

No other dependencies. ZAIA-MCP is intentionally lightweight.

## Build & Test

```bash
# Build
go build -o ./zaia-mcp ./cmd/zaia-mcp

# Run all tests (75)
go test ./... -count=1

# With race detection
go test ./... -race -count=1

# Integration tests only
go test ./integration/ -v -count=1

# Vet
go vet ./...
```

## Related

- **[ZAIA CLI](https://github.com/krls2020/zaia)** — Go CLI binary with all business logic
- **[Zerops](https://zerops.io)** — Developer-first PaaS with bare-metal infrastructure

# ZAIA-MCP — Isolated Development Guide

> Toto je **self-contained** reference pro vývoj ZAIA-MCP serveru.
> Claude Code se spouští přímo v tomto adresáři (`./zaia-mcp/`).

---

## ⚠️ POVINNÉ: Údržba CLAUDE.md

**CLAUDE.md MUSÍ být vždy aktuální.**

### Kdy aktualizovat

| Změna | Co aktualizovat v CLAUDE.md |
|-------|----------------------------|
| Nový/změněný tool v `tools/*.go` | Sekce **MCP Tools** |
| Změna executor interface | Sekce **Klíčové typy** |
| Změna server setup | Sekce **Architektura kódu** |
| Nová závislost v `go.mod` | Sekce **Závislosti** |
| Nový test pattern | Sekce **Vzory pro psaní testů** |
| Změna Instructions konstanty | Sekce **Instructions** |

---

## Co je ZAIA-MCP

**ZAIA-MCP** je tenká MCP server vrstva (Go binary) která volá ZAIA CLI a ZCLI jako subprocesy. Neobsahuje business logiku — veškerá logika je v ZAIA CLI.

```
┌─────────────────────┐     ┌──────────────┐     ┌──────────────┐
│  Claude Code /      │     │  ZAIA-MCP    │     │  ZAIA CLI    │
│  Desktop            │────▶│  (tento repo)│────▶│  (business   │
│                     │ MCP │  tenká vrstva│exec │  logika)     │
│                     │STDIO│              │────▶│              │
└─────────────────────┘     │              │     └──────────────┘
                            │              │     ┌──────────────┐
                            │              │────▶│  ZCLI        │
                            │              │exec │  (deploy)    │
                            └──────────────┘     └──────────────┘
```

### Klíčové vlastnosti

| Vlastnost | Hodnota |
|-----------|---------|
| Transport | STDIO (ne HTTP) |
| Auth | Pre-authenticated — ZAIA CLI handles auth |
| State | Stateless — each tool call = fresh CLI invocation |
| Business logic | None — all in ZAIA CLI |
| Tools | 11 MCP tools |
| Resources | zerops://docs/{path} via ResourceTemplate |
| Dependencies | 1 (MCP Go SDK v0.6.0) |

---

## Nadřazený repozitář

Tento projekt žije v `./zaia-mcp/` podadresáři repo `/Users/macbook/Documents/Zerops-MCP/`.

| Cesta (relativní k `../`) | Obsah |
|---|---|
| `spec/zaia-cli/` | ZAIA CLI specifikace (autoritativní) |
| `spec/zaia-mcp/` | ZAIA-MCP system prompt spec |
| `docs/decisions/` | Architektonická rozhodnutí |
| `zaia/` | ZAIA CLI implementace |
| `CLAUDE.md` | Nadřazené projektové instrukce |

---

## Architektura kódu

```
zaia-mcp/
├── cmd/zaia-mcp/main.go           # Entry point — STDIO MCP server
├── go.mod                          # github.com/zeropsio/zaia-mcp
│
├── internal/
│   ├── server/
│   │   ├── server.go               # MCPServer — setup, Instructions, tool+resource registration
│   │   └── server_test.go
│   │
│   ├── executor/
│   │   ├── executor.go             # Executor interface + CLIExecutor (exec.CommandContext)
│   │   ├── executor_test.go
│   │   ├── mock.go                 # MockExecutor + helper constructors (SyncResult, AsyncResult, ErrorResult)
│   │   └── mock_test.go
│   │
│   ├── tools/
│   │   ├── convert.go              # ParseCLIResponse, ToMCPResult, ResultFromCLI
│   │   ├── convert_test.go
│   │   ├── discover.go             # zerops_discover → zaia discover
│   │   ├── logs.go                 # zerops_logs → zaia logs
│   │   ├── validate.go             # zerops_validate → zaia validate
│   │   ├── knowledge.go            # zerops_knowledge → zaia search
│   │   ├── process.go              # zerops_process → zaia process / zaia cancel
│   │   ├── manage.go               # zerops_manage → zaia start/stop/restart/scale
│   │   ├── env.go                  # zerops_env → zaia env get/set/delete
│   │   ├── import.go               # zerops_import → zaia import
│   │   ├── delete.go               # zerops_delete → zaia delete
│   │   ├── subdomain.go            # zerops_subdomain → zaia subdomain
│   │   ├── deploy.go               # zerops_deploy → zcli push (only tool using zcli)
│   │   └── tools_test.go           # All tool tests (in-memory MCP sessions)
│   │
│   └── resources/
│       ├── knowledge.go            # zerops://docs/{path} ResourceTemplate
│       └── knowledge_test.go
│
└── integration/
    ├── harness.go                  # Test harness (in-memory MCP, mock executor)
    └── flow_test.go                # End-to-end flows (9 test scenarios)
```

---

## Klíčové typy a interface

### Executor (`internal/executor/executor.go`)

```go
type Result struct {
    Stdout   []byte
    Stderr   []byte
    ExitCode int
}

type Executor interface {
    RunZaia(ctx context.Context, args ...string) (*Result, error)
    RunZcli(ctx context.Context, args ...string) (*Result, error)
}
```

- `CLIExecutor` — real implementation using `exec.CommandContext`
- `MockExecutor` — configurable responses for tests

### CLI Response (`internal/tools/convert.go`)

ZAIA CLI always outputs one of:
```json
{"type":"sync","status":"ok","data":{...}}
{"type":"async","status":"initiated","processes":[...]}
{"type":"error","code":"...","error":"...","suggestion":"..."}
```

Conversion:
- `type=sync` → `mcp.TextContent{Text: data_json}`, `IsError: false`
- `type=async` → `mcp.TextContent{Text: processes_json}`, `IsError: false`
- `type=error` → `mcp.TextContent{Text: error_json}`, `IsError: true`

### MCPServer (`internal/server/server.go`)

```go
type MCPServer struct {
    server   *mcp.Server
    executor executor.Executor
}

func New() *MCPServer                              // default CLIExecutor
func NewWithExecutor(exec executor.Executor) *MCPServer  // custom executor (tests)
func (s *MCPServer) Run(ctx context.Context) error  // STDIO transport
```

---

## MCP Tools (11)

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

### Deploy (zcli)

| MCP Tool | CLI Command | Required Params |
|----------|-------------|-----------------|
| `zerops_deploy` | `zcli push` | — |

---

## MCP Resources

### zerops://docs/{path}

ResourceTemplate pro knowledge docs. Volá `zaia search --get <uri>`.

---

## Instructions

~250 tokenů system prompt v `server.go` konstantě `Instructions`. Obsahuje:
- Zerops overview (services, rules)
- Tool summary
- Defaults

---

## Vývoj: TDD Workflow

### Příkazy

```bash
# Unit tests
go test ./internal/<pkg> -v -count=1

# All tests
go test ./... -count=1

# With race detection
go test ./... -race -count=1

# Build
go build -o ./zaia-mcp ./cmd/zaia-mcp

# Vet
go vet ./...

# Integration tests
go test ./integration/ -v -count=1
```

---

## Vzory pro psaní testů

### Tool test (in-memory MCP session)

```go
func TestMyTool(t *testing.T) {
    mock := executor.NewMockExecutor().
        WithZaiaResponse("discover", executor.SyncResult(`{"services":[]}`))
    srv := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.1"}, nil)
    tools.RegisterDiscover(srv, mock)

    // Create in-memory transport
    t1, t2 := mcp.NewInMemoryTransports()
    srv.Connect(context.Background(), t1, nil)
    client := mcp.NewClient(&mcp.Implementation{Name: "c", Version: "0.0.1"}, nil)
    session, _ := client.Connect(context.Background(), t2, nil)
    defer session.Close()

    result, _ := session.CallTool(ctx, &mcp.CallToolParams{
        Name: "zerops_discover",
        Arguments: map[string]interface{}{},
    })
    // assert result
}
```

### Integration test (Harness)

```go
func TestFlow(t *testing.T) {
    h := integration.NewHarness(t)
    h.Mock().WithZaiaResponse("discover", executor.SyncResult(`{...}`))
    text := h.MustCallSuccess("zerops_discover", nil)
    // parse and assert
}
```

### Mock helpers

```go
executor.SyncResult(`{"key":"val"}`)           // → sync JSON envelope
executor.AsyncResult(`[{"processId":"p1"}]`)   // → async JSON envelope
executor.ErrorResult("CODE", "msg", "fix", 2)  // → error JSON envelope
```

---

## Aktuální stav implementace

### Hotovo

| Fáze | Co | Testy |
|------|----|-------|
| 0 | Scaffold (go.mod, main.go, server.go) | 3 |
| 1 | Executor interface + CLIExecutor + MockExecutor | 15 |
| 2 | CLI → MCP result conversion | 11 |
| 3 | Sync tools (5): discover, logs, validate, knowledge, process | 15 |
| 4 | Async tools (5): manage, env, import, delete, subdomain | 16 |
| 5 | Deploy tool (zcli push) | 2 |
| 6 | MCP Resources (zerops://docs/{path}) | 3 |
| 7 | Integration tests (9 flow scenarios) | 10 |

**Celkem: 75 testů, 0 failures, 0 race conditions.**

### Metriky

| Metrika | Cíl | Aktuální |
|---------|-----|----------|
| Unit tests | ~110+ | 75 |
| Coverage | >80% | TBD |
| Binary size | <10MB | 7.4MB |
| MCP tools | 11 | 11 |
| Dependencies | 1 | 1 (MCP SDK) |

---

## Závislosti

```
github.com/modelcontextprotocol/go-sdk v0.6.0   — MCP Go SDK
```

Žádné další závislosti. ZAIA-MCP je záměrně lightweight.

---

## Údržba

### Při přidání nového MCP tool

1. Vytvořit `tools/newtool.go`
2. Zaregistrovat v `server/server.go` → `registerTools()`
3. Přidat test v `tools/tools_test.go`
4. Přidat integration test v `integration/flow_test.go`
5. Aktualizovat sekci **MCP Tools** v CLAUDE.md

### Při změně Executor interface

1. Upravit `executor/executor.go`
2. Aktualizovat `executor/mock.go`
3. `go test ./... -count=1`

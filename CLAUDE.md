# ZAIA-MCP — Development Guide

Tenká MCP server vrstva (Go binary) která volá ZAIA CLI a ZCLI jako subprocesy. Neobsahuje business logiku — veškerá logika je v [ZAIA CLI](https://github.com/krls2020/zaia).

> **Public docs:** viz `README.md` (tools, architektura, prerekvizity)
> **Design docs:** viz `../design/zaia-mcp/` (historický záměr)

---

## Hierarchie zdrojů pravdy

```
1. Kód (Go types, interface, testy)  ← AUTORITATIVNÍ
2. CLAUDE.md                         ← PROVOZNÍ (workflow, konvence)
3. README.md                         ← PUBLIC DOCS (tools, API reference)
4. ../design/zaia-mcp/               ← HISTORICKÉ (původní spec)
```

---

## TDD Workflow

### Povinný workflow

```
1. RED: Napsat failing test PŘED implementací
2. GREEN: Minimální implementace
3. REFACTOR: Vyčistit, testy zůstávají zelené
```

### Pravidla

- NIKDY implementace bez odpovídajícího testu
- Table-driven testy (Go idiom)
- Popisné názvy: `TestDiscover_WithService_ReturnsEnvs`
- Max 300 řádků per soubor
- In-memory MCP transport pro tool testy — ne mock HTTP
- **Unit/integration testy = MockExecutor** (žádné síťové volání)
- **E2E testy (`e2e/`) = real Zerops API** — `go test ./e2e/ -tags e2e`; vyžadují `zaia` binary na PATH a platný token

### Příkazy

```bash
go test ./internal/<pkg> -v -count=1         # Package
go test ./internal/<pkg> -run TestName -v    # Jednotlivý test
go test ./... -count=1                        # Vše
go test ./... -race -count=1                  # S race detection
go build -o ./zaia-mcp ./cmd/zaia-mcp         # Build
go vet ./...                                  # Vet
go test ./integration/ -v -count=1            # Integration (mocked)
go test ./e2e/ -tags e2e -v -count=1         # E2E (real API, needs zaia on PATH)
```

---

## Architektura kódu

```
zaia-mcp/
├── cmd/zaia-mcp/main.go           # Entry point — STDIO MCP server
├── internal/
│   ├── server/
│   │   └── server.go              # MCPServer — setup, Instructions, registration
│   ├── executor/
│   │   ├── executor.go            # Executor interface + CLIExecutor
│   │   └── mock.go                # MockExecutor for tests
│   ├── tools/
│   │   ├── convert.go             # ParseCLIResponse, ToMCPResult
│   │   ├── discover.go ... subdomain.go  # 11 tool implementations
│   │   └── tools_test.go          # All tool tests (in-memory MCP)
│   └── resources/
│       └── knowledge.go           # zerops://docs/{path} ResourceTemplate
├── integration/
│   ├── harness.go                 # Test harness (MockExecutor)
│   └── flow_test.go               # 9 end-to-end flow scenarios
└── e2e/                           # Real API tests (build tag: e2e)
    ├── e2e_test.go                # 17-step full lifecycle test
    ├── helpers_test.go            # In-memory MCP session + cleanup
    └── process_test.go            # Process polling helpers
```

### Klíčové soubory (source of truth)

| Soubor | Co definuje |
|--------|-------------|
| `executor/executor.go` | Executor interface (RunZaia, RunZcli) |
| `server/server.go` | MCPServer, Instructions (~250 tokenů) |
| `tools/convert.go` | CLI → MCP result conversion |
| `tools/*.go` | 11 tool implementations |
| `resources/knowledge.go` | zerops://docs/{path} ResourceTemplate |

---

## Konvence

- **Stateless** — každý tool call = čerstvé CLI volání
- **STDIO transport** — ne HTTP
- **Pre-authenticated** — ZAIA CLI řeší auth
- **11 tools** — 5 sync, 5 async, 1 deploy (zcli)
- **Deploy = zcli push** — jediný tool který nevolá ZAIA
- **MockExecutor** pro testy — `SyncResult()`, `AsyncResult()`, `ErrorResult()`
- **JSON-only CLI output** — MCP tools parse stdout JSON
- **Error conversion** — Go error → MCP error result (nilerr pattern)
- **`errorResult()` helper** — all tools use same pattern
- **`ResultFromCLI()`** — parse + convert in one step
- **Mock keys** — `"binary arg1 arg2 ..."` (space-joined)

---

## Stav implementace

75 testů, 0 failures. 11 MCP tools. Instructions. MCP Resources. Integration testy (9 flows). CI/CD (3 workflows).

---

## Hooks (automatický TDD feedback)

```
.claude/
├── settings.json
└── hooks/
    ├── post-test.sh          # Po Edit/Write .go: go test + go vet
    ├── check-claude-md.sh    # Po Edit/Write klíčového souboru: reminder
    └── pre-commit-check.sh   # Před git commit: kontrola CLAUDE.md
```

---

## Údržba

### Kdy aktualizovat CLAUDE.md

| Změna | Akce |
|-------|------|
| Nový MCP tool | Aktualizuj "Architektura kódu" strom |
| Změna Executor interface | Aktualizuj "Klíčové soubory" |
| Změna stavu | Aktualizuj "Stav implementace" |

### Kdy aktualizovat README.md

| Změna | Akce |
|-------|------|
| Nový MCP tool | Aktualizuj tools reference |
| Změna prerekvizit | Aktualizuj prerekvizity |
| Změna architektury | Aktualizuj diagram |

Detailní tools reference → viz `README.md` a kód.

### Při přidání nového MCP tool

1. Vytvořit `tools/newtool.go`
2. Zaregistrovat v `server/server.go`
3. Přidat test v `tools/tools_test.go`
4. Přidat integration test v `integration/flow_test.go`

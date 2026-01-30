package integration

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/zeropsio/zaia-mcp/internal/executor"
)

func TestAllToolsRegistered(t *testing.T) {
	h := NewHarness(t)
	tools := h.ListTools()
	sort.Strings(tools)

	expected := []string{
		"zerops_delete",
		"zerops_deploy",
		"zerops_discover",
		"zerops_env",
		"zerops_import",
		"zerops_knowledge",
		"zerops_logs",
		"zerops_manage",
		"zerops_process",
		"zerops_subdomain",
		"zerops_validate",
	}

	if len(tools) != len(expected) {
		t.Fatalf("got %d tools, want %d: %v", len(tools), len(expected), tools)
	}
	for i, name := range expected {
		if tools[i] != name {
			t.Errorf("tool[%d]: got %q, want %q", i, tools[i], name)
		}
	}
}

func TestFlow_DiscoverThenManage(t *testing.T) {
	h := NewHarness(t)

	// Setup: discover returns services
	h.Mock().WithZaiaResponse("discover",
		executor.SyncResult(`{"project":{"id":"p1","name":"myapp"},"services":[{"hostname":"api","type":"nodejs@22","status":"ACTIVE"}]}`))

	// Step 1: Discover
	text := h.MustCallSuccess("zerops_discover", nil)
	var data map[string]interface{}
	_ = json.Unmarshal([]byte(text), &data)
	services := data["services"].([]interface{})
	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}

	// Setup: manage start
	h.Mock().WithZaiaResponse("restart --service api",
		executor.AsyncResult(`[{"processId":"proc-1","status":"PENDING"}]`))

	// Step 2: Restart
	text = h.MustCallSuccess("zerops_manage", map[string]interface{}{
		"action":          "restart",
		"serviceHostname": "api",
	})
	var processes []map[string]interface{}
	_ = json.Unmarshal([]byte(text), &processes)
	if len(processes) != 1 {
		t.Fatalf("expected 1 process, got %d", len(processes))
	}
	if processes[0]["processId"] != "proc-1" {
		t.Errorf("unexpected processId: %v", processes[0]["processId"])
	}

	// Setup: check process status
	h.Mock().WithZaiaResponse("process proc-1",
		executor.SyncResult(`{"processId":"proc-1","status":"FINISHED"}`))

	// Step 3: Check process
	text = h.MustCallSuccess("zerops_process", map[string]interface{}{
		"processId": "proc-1",
	})
	var proc map[string]interface{}
	_ = json.Unmarshal([]byte(text), &proc)
	if proc["status"] != "FINISHED" {
		t.Errorf("expected FINISHED, got %v", proc["status"])
	}
}

func TestFlow_ValidateThenImport(t *testing.T) {
	h := NewHarness(t)

	yaml := "services:\n  - hostname: cache\n    type: valkey@8"

	// Step 1: Validate
	h.Mock().WithDefault(executor.SyncResult(`{"valid":true,"warnings":[]}`))
	h.MustCallSuccess("zerops_validate", map[string]interface{}{
		"content": yaml,
		"type":    "import.yml",
	})

	// Step 2: Import (dry-run)
	h.Mock().WithDefault(executor.SyncResult(`{"valid":true,"services":[{"hostname":"cache","type":"valkey@8"}]}`))
	h.MustCallSuccess("zerops_import", map[string]interface{}{
		"content": yaml,
		"dryRun":  true,
	})

	// Step 3: Import (real)
	h.Mock().WithDefault(executor.AsyncResult(`[{"processId":"import-1","status":"PENDING"}]`))
	text := h.MustCallSuccess("zerops_import", map[string]interface{}{
		"content": yaml,
	})
	var processes []map[string]interface{}
	_ = json.Unmarshal([]byte(text), &processes)
	if processes[0]["processId"] != "import-1" {
		t.Errorf("unexpected processId: %v", processes[0]["processId"])
	}
}

func TestFlow_EnvGetSetGet(t *testing.T) {
	h := NewHarness(t)

	// Step 1: Get env (empty)
	h.Mock().WithZaiaResponse("env get --service api",
		executor.SyncResult(`{"envVars":[]}`))
	text := h.MustCallSuccess("zerops_env", map[string]interface{}{
		"action":          "get",
		"serviceHostname": "api",
	})
	var data map[string]interface{}
	_ = json.Unmarshal([]byte(text), &data)
	envVars := data["envVars"].([]interface{})
	if len(envVars) != 0 {
		t.Errorf("expected 0 env vars, got %d", len(envVars))
	}

	// Step 2: Set env
	h.Mock().WithDefault(executor.AsyncResult(`[{"processId":"env-1","status":"PENDING"}]`))
	h.MustCallSuccess("zerops_env", map[string]interface{}{
		"action":          "set",
		"serviceHostname": "api",
		"variables":       []interface{}{"PORT=3000", "NODE_ENV=production"},
	})

	// Step 3: Verify set call args
	calls := h.Mock().Calls
	lastCall := calls[len(calls)-1]
	if lastCall.Binary != "zaia" {
		t.Errorf("expected zaia, got %s", lastCall.Binary)
	}
	foundPort := false
	for _, arg := range lastCall.Args {
		if arg == "PORT=3000" {
			foundPort = true
		}
	}
	if !foundPort {
		t.Errorf("PORT=3000 not in args: %v", lastCall.Args)
	}
}

func TestFlow_ErrorPropagation(t *testing.T) {
	h := NewHarness(t)

	// Auth error should propagate
	h.Mock().WithZaiaResponse("discover",
		executor.ErrorResult("AUTH_REQUIRED", "Not authenticated", "Run: zaia login <token>", 2))
	errText := h.MustCallError("zerops_discover", nil)

	var errObj map[string]interface{}
	if err := json.Unmarshal([]byte(errText), &errObj); err != nil {
		t.Fatalf("error text not valid JSON: %v (text: %s)", err, errText)
	}
	if errObj["code"] != "AUTH_REQUIRED" {
		t.Errorf("expected AUTH_REQUIRED, got %v", errObj["code"])
	}
}

func TestFlow_DiscoverThenDelete(t *testing.T) {
	h := NewHarness(t)

	// Discover
	h.Mock().WithZaiaResponse("discover",
		executor.SyncResult(`{"services":[{"hostname":"old-cache","type":"keydb@7","status":"ACTIVE"}]}`))
	h.MustCallSuccess("zerops_discover", nil)

	// Delete (confirmed)
	h.Mock().WithZaiaResponse("delete --service old-cache --confirm",
		executor.AsyncResult(`[{"processId":"del-1","status":"PENDING"}]`))
	text := h.MustCallSuccess("zerops_delete", map[string]interface{}{
		"serviceHostname": "old-cache",
		"confirm":         true,
	})
	var processes []map[string]interface{}
	_ = json.Unmarshal([]byte(text), &processes)
	if processes[0]["processId"] != "del-1" {
		t.Errorf("unexpected processId: %v", processes[0]["processId"])
	}
}

func TestFlow_SubdomainEnableDisable(t *testing.T) {
	h := NewHarness(t)

	// Enable
	h.Mock().WithZaiaResponse("subdomain enable --service api",
		executor.AsyncResult(`[{"processId":"sub-1","status":"PENDING"}]`))
	h.MustCallSuccess("zerops_subdomain", map[string]interface{}{
		"serviceHostname": "api",
		"action":          "enable",
	})

	// Enable again (idempotent)
	h.Mock().WithZaiaResponse("subdomain enable --service api",
		executor.SyncResult(`{"status":"already_enabled","subdomain":"api-xyz.zerops.app"}`))
	h.MustCallSuccess("zerops_subdomain", map[string]interface{}{
		"serviceHostname": "api",
		"action":          "enable",
	})

	// Disable
	h.Mock().WithZaiaResponse("subdomain disable --service api",
		executor.AsyncResult(`[{"processId":"sub-2","status":"PENDING"}]`))
	h.MustCallSuccess("zerops_subdomain", map[string]interface{}{
		"serviceHostname": "api",
		"action":          "disable",
	})
}

func TestFlow_DeployViaZcli(t *testing.T) {
	h := NewHarness(t)

	// Deploy calls zcli, not zaia
	h.Mock().WithZcliResponse("push --serviceId svc-1",
		executor.SyncResult(`{"deployed":true,"buildId":"b-1"}`))
	text := h.MustCallSuccess("zerops_deploy", map[string]interface{}{
		"serviceId": "svc-1",
	})
	var data map[string]interface{}
	_ = json.Unmarshal([]byte(text), &data)
	if data["deployed"] != true {
		t.Errorf("expected deployed=true, got %v", data["deployed"])
	}

	// Verify it called zcli
	if h.Mock().Calls[0].Binary != "zcli" {
		t.Errorf("expected zcli, got %s", h.Mock().Calls[0].Binary)
	}
}

func TestFlow_KnowledgeSearch(t *testing.T) {
	h := NewHarness(t)

	h.Mock().WithZaiaResponse(`search postgresql connection string`,
		executor.SyncResult(`{"query":"postgresql connection string","results":[{"uri":"zerops://docs/services/postgresql","title":"PostgreSQL on Zerops","score":0.95}],"topResult":{"title":"PostgreSQL","content":"Port 5432"}}`))
	text := h.MustCallSuccess("zerops_knowledge", map[string]interface{}{
		"query": "postgresql connection string",
	})
	var data map[string]interface{}
	_ = json.Unmarshal([]byte(text), &data)
	if data["query"] != "postgresql connection string" {
		t.Errorf("unexpected query: %v", data["query"])
	}
}

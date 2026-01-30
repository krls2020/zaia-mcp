//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// TestE2E_FullLifecycle tests the full MCP → zaia CLI → Zerops API chain.
//
// Two test services are created:
//   - runtime (nodejs@22): for subdomain, env, logs testing
//   - managed (keydb@6): for start/stop lifecycle testing
//
// Subtests run sequentially — each depends on the previous state.
func TestE2E_FullLifecycle(t *testing.T) {
	s := connectInMemory(t)

	suffix := randomSuffix()
	rtHost := "e2ert" + suffix // runtime (nodejs@22)
	dbHost := "e2edb" + suffix // managed (keydb@6)

	t.Logf("test services: runtime=%s, managed=%s", rtHost, dbHost)

	// Cleanup on exit (even on failure).
	// Uses exec.Command directly — the MCP session context may be canceled by then.
	t.Cleanup(func() {
		t.Log("cleanup: deleting test services")
		deleteTestService(t, rtHost)
		deleteTestService(t, dbHost)
	})

	// --- Read-only operations ---

	t.Run("01_discover", func(t *testing.T) {
		text := s.mustCallSuccess("zerops_discover", nil)
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(text), &data); err != nil {
			t.Fatalf("parse discover response: %v", err)
		}
		if data["project"] == nil {
			t.Fatal("discover response missing 'project' field")
		}
		t.Logf("project found: %v", data["project"])
	})

	t.Run("02_knowledge", func(t *testing.T) {
		text := s.mustCallSuccess("zerops_knowledge", map[string]interface{}{
			"query": "postgresql",
		})
		if text == "" {
			t.Fatal("knowledge returned empty response")
		}
		t.Logf("knowledge response length: %d", len(text))
	})

	importYAML := fmt.Sprintf(`services:
  - hostname: %s
    type: nodejs@22
    mode: NON_HA
    buildBase: nodejs@22
    ports:
      - port: 3000
        scheme: http
  - hostname: %s
    type: keydb@6
    mode: NON_HA`, rtHost, dbHost)

	t.Run("03_validate", func(t *testing.T) {
		text := s.mustCallSuccess("zerops_validate", map[string]interface{}{
			"content": importYAML,
			"type":    "import.yml",
		})
		t.Logf("validate result: %s", text)
	})

	t.Run("04_import_dry_run", func(t *testing.T) {
		text := s.mustCallSuccess("zerops_import", map[string]interface{}{
			"content": importYAML,
			"dryRun":  true,
		})
		t.Logf("dry-run result: %s", text)
	})

	// --- Create services ---

	var importProcesses []processInfo

	t.Run("05_import_real", func(t *testing.T) {
		text := s.mustCallSuccess("zerops_import", map[string]interface{}{
			"content": importYAML,
		})
		importProcesses = parseProcesses(t, text)
		if len(importProcesses) == 0 {
			t.Fatal("import returned no processes")
		}
		t.Logf("import started: %d process(es)", len(importProcesses))
	})

	t.Run("06_wait_import", func(t *testing.T) {
		if len(importProcesses) == 0 {
			t.Skip("no import processes (05_import_real may have failed)")
		}
		for _, p := range importProcesses {
			status := waitForProcess(s, p.ProcessID)
			if status != "FINISHED" {
				t.Fatalf("import process %s: expected FINISHED, got %s", p.ProcessID, status)
			}
			t.Logf("import process %s: %s", p.ProcessID, status)
		}
	})

	t.Run("07_discover_both", func(t *testing.T) {
		text := s.mustCallSuccess("zerops_discover", nil)
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(text), &data); err != nil {
			t.Fatalf("parse discover: %v", err)
		}
		services, _ := data["services"].([]interface{})
		foundRT, foundDB := false, false
		for _, svc := range services {
			m, _ := svc.(map[string]interface{})
			hostname, _ := m["hostname"].(string)
			if hostname == rtHost {
				foundRT = true
			}
			if hostname == dbHost {
				foundDB = true
			}
		}
		if !foundRT {
			t.Errorf("runtime service %s not found in discover", rtHost)
		}
		if !foundDB {
			t.Errorf("managed service %s not found in discover", dbHost)
		}
	})

	// --- Managed service (keydb): lifecycle ---

	t.Run("08_stop_keydb", func(t *testing.T) {
		text := s.mustCallSuccess("zerops_manage", map[string]interface{}{
			"action":          "stop",
			"serviceHostname": dbHost,
		})
		processes := parseProcesses(t, text)
		for _, p := range processes {
			status := waitForProcess(s, p.ProcessID)
			if status != "FINISHED" {
				t.Fatalf("stop process %s: expected FINISHED, got %s", p.ProcessID, status)
			}
		}
		t.Logf("keydb stopped")
	})

	t.Run("09_start_keydb", func(t *testing.T) {
		text := s.mustCallSuccess("zerops_manage", map[string]interface{}{
			"action":          "start",
			"serviceHostname": dbHost,
		})
		processes := parseProcesses(t, text)
		for _, p := range processes {
			status := waitForProcess(s, p.ProcessID)
			if status != "FINISHED" {
				t.Fatalf("start process %s: expected FINISHED, got %s", p.ProcessID, status)
			}
		}
		t.Logf("keydb started")
	})

	// --- Runtime service (nodejs): full features ---

	t.Run("10_restart_nodejs", func(t *testing.T) {
		// Restart may fail if service is READY_TO_DEPLOY (no code deployed yet).
		// This is expected for a fresh runtime service — we treat it as non-fatal.
		result := s.callTool("zerops_manage", map[string]interface{}{
			"action":          "restart",
			"serviceHostname": rtHost,
		})
		text := getTextContent(t, result)
		if result.IsError {
			t.Logf("restart returned error (expected for undeployed service): %s", text)
			return
		}
		processes := parseProcesses(t, text)
		for _, p := range processes {
			status := waitForProcess(s, p.ProcessID)
			if status != "FINISHED" {
				t.Fatalf("restart process %s: expected FINISHED, got %s", p.ProcessID, status)
			}
		}
		t.Logf("nodejs restarted")
	})

	t.Run("11_env_set", func(t *testing.T) {
		text := s.mustCallSuccess("zerops_env", map[string]interface{}{
			"action":          "set",
			"serviceHostname": rtHost,
			"variables":       []interface{}{"TEST_KEY=test_value"},
		})
		processes := parseProcesses(t, text)
		for _, p := range processes {
			status := waitForProcess(s, p.ProcessID)
			if status != "FINISHED" {
				t.Fatalf("env set process %s: expected FINISHED, got %s", p.ProcessID, status)
			}
		}
		t.Logf("env var set")
	})

	t.Run("12_env_get", func(t *testing.T) {
		text := s.mustCallSuccess("zerops_env", map[string]interface{}{
			"action":          "get",
			"serviceHostname": rtHost,
		})
		if !strings.Contains(text, "TEST_KEY") {
			t.Errorf("TEST_KEY not found in env get response: %s", text)
		}
		t.Logf("env get contains TEST_KEY")
	})

	t.Run("13_subdomain_enable", func(t *testing.T) {
		// Subdomain may fail if service is READY_TO_DEPLOY (not yet serving HTTP).
		result := s.callTool("zerops_subdomain", map[string]interface{}{
			"serviceHostname": rtHost,
			"action":          "enable",
		})
		text := getTextContent(t, result)
		if result.IsError {
			t.Logf("subdomain enable returned error (expected for undeployed service): %s", text)
			return
		}
		t.Logf("subdomain enable: %s", text)
	})

	t.Run("14_subdomain_disable", func(t *testing.T) {
		// Subdomain may fail if enable was skipped.
		result := s.callTool("zerops_subdomain", map[string]interface{}{
			"serviceHostname": rtHost,
			"action":          "disable",
		})
		text := getTextContent(t, result)
		if result.IsError {
			t.Logf("subdomain disable returned error (expected if enable was skipped): %s", text)
			return
		}
		t.Logf("subdomain disable: %s", text)
	})

	t.Run("15_logs", func(t *testing.T) {
		text := s.mustCallSuccess("zerops_logs", map[string]interface{}{
			"serviceHostname": rtHost,
			"limit":           10,
		})
		// Logs may be empty for a fresh service — that's OK
		t.Logf("logs response length: %d", len(text))
	})

	// --- Cleanup ---

	t.Run("16_delete_both", func(t *testing.T) {
		// Delete runtime
		rtText := s.mustCallSuccess("zerops_delete", map[string]interface{}{
			"serviceHostname": rtHost,
			"confirm":         true,
		})
		rtProcs := parseProcesses(t, rtText)

		// Delete managed
		dbText := s.mustCallSuccess("zerops_delete", map[string]interface{}{
			"serviceHostname": dbHost,
			"confirm":         true,
		})
		dbProcs := parseProcesses(t, dbText)

		// Wait for all delete processes
		for _, p := range append(rtProcs, dbProcs...) {
			status := waitForProcess(s, p.ProcessID)
			if status != "FINISHED" {
				t.Fatalf("delete process %s: expected FINISHED, got %s", p.ProcessID, status)
			}
		}
		t.Logf("both services deleted")
	})

	t.Run("17_discover_gone", func(t *testing.T) {
		text := s.mustCallSuccess("zerops_discover", nil)
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(text), &data); err != nil {
			t.Fatalf("parse discover: %v", err)
		}
		services, _ := data["services"].([]interface{})
		for _, svc := range services {
			m, _ := svc.(map[string]interface{})
			hostname, _ := m["hostname"].(string)
			if hostname == rtHost || hostname == dbHost {
				t.Errorf("service %s still found after delete", hostname)
			}
		}
	})
}

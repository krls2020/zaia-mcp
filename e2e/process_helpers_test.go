//go:build e2e

package e2e

import (
	"encoding/json"
	"testing"
	"time"
)

const (
	maxPollAttempts = 40            // 40 × 3s = 120s max
	pollInterval    = 3 * time.Second
)

// processInfo represents a process returned by async operations.
type processInfo struct {
	ProcessID string `json:"processId"`
	Status    string `json:"status"`
}

// waitForProcess polls zerops_process until the process reaches a terminal state.
func waitForProcess(s *session, processID string) string {
	s.t.Helper()
	for i := 0; i < maxPollAttempts; i++ {
		text := s.mustCallSuccess("zerops_process", map[string]interface{}{
			"processId": processID,
		})
		var proc map[string]interface{}
		if err := json.Unmarshal([]byte(text), &proc); err != nil {
			s.t.Fatalf("parse process response: %v (text: %s)", err, text)
		}
		status, _ := proc["status"].(string)
		if status == "FINISHED" || status == "FAILED" || status == "CANCELED" {
			return status
		}
		s.t.Logf("process %s: %s (poll %d/%d)", processID, status, i+1, maxPollAttempts)
		time.Sleep(pollInterval)
	}
	s.t.Fatalf("process %s did not reach terminal state within %v", processID, time.Duration(maxPollAttempts)*pollInterval)
	return ""
}

// parseProcesses parses the JSON array of processes from an async tool response.
func parseProcesses(t *testing.T, text string) []processInfo {
	t.Helper()
	var processes []processInfo
	if err := json.Unmarshal([]byte(text), &processes); err != nil {
		t.Fatalf("parse processes: %v (text: %s)", err, text)
	}
	return processes
}

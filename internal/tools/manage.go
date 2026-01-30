package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zaia-mcp/internal/executor"
)

// ManageInput is the input schema for zerops_manage.
type ManageInput struct {
	Action          string  `json:"action"`
	ServiceHostname string  `json:"serviceHostname"`
	CPUMode         string  `json:"cpuMode,omitempty"`
	MinCPU          int     `json:"minCpu,omitempty"`
	MaxCPU          int     `json:"maxCpu,omitempty"`
	MinRAM          float64 `json:"minRam,omitempty"`
	MaxRAM          float64 `json:"maxRam,omitempty"`
	MinDisk         float64 `json:"minDisk,omitempty"`
	MaxDisk         float64 `json:"maxDisk,omitempty"`
	StartContainers int     `json:"startContainers,omitempty"`
	MinContainers   int     `json:"minContainers,omitempty"`
	MaxContainers   int     `json:"maxContainers,omitempty"`
}

// RegisterManage registers the zerops_manage tool on the server.
func RegisterManage(srv *mcp.Server, exec executor.Executor) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "zerops_manage",
		Description: `Manage service lifecycle and scaling.

Actions:
- start: Start a stopped service
- stop: Stop a running service
- restart: Restart a service
- scale: Change CPU/RAM/disk/container scaling

Scale parameters (only for action=scale):
- cpuMode: SHARED or DEDICATED
- minCpu/maxCpu, minRam/maxRam, minDisk/maxDisk
- startContainers, minContainers, maxContainers

Returns process ID for status tracking via zerops_process.`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ManageInput) (*mcp.CallToolResult, any, error) {
		if input.Action == "" {
			return errorResult("action is required (start, stop, restart, scale)"), nil, nil
		}
		if input.ServiceHostname == "" {
			return errorResult("serviceHostname is required"), nil, nil
		}

		args := []string{input.Action, "--service", input.ServiceHostname}

		if input.Action == "scale" {
			if input.CPUMode != "" {
				args = append(args, "--cpu-mode", input.CPUMode)
			}
			if input.MinCPU > 0 {
				args = append(args, "--min-cpu", fmt.Sprintf("%d", input.MinCPU))
			}
			if input.MaxCPU > 0 {
				args = append(args, "--max-cpu", fmt.Sprintf("%d", input.MaxCPU))
			}
			if input.MinRAM > 0 {
				args = append(args, "--min-ram", fmt.Sprintf("%g", input.MinRAM))
			}
			if input.MaxRAM > 0 {
				args = append(args, "--max-ram", fmt.Sprintf("%g", input.MaxRAM))
			}
			if input.MinDisk > 0 {
				args = append(args, "--min-disk", fmt.Sprintf("%g", input.MinDisk))
			}
			if input.MaxDisk > 0 {
				args = append(args, "--max-disk", fmt.Sprintf("%g", input.MaxDisk))
			}
			if input.StartContainers > 0 {
				args = append(args, "--start-containers", fmt.Sprintf("%d", input.StartContainers))
			}
			if input.MinContainers > 0 {
				args = append(args, "--min-containers", fmt.Sprintf("%d", input.MinContainers))
			}
			if input.MaxContainers > 0 {
				args = append(args, "--max-containers", fmt.Sprintf("%d", input.MaxContainers))
			}
		}

		result, err := exec.RunZaia(ctx, args...)
		if err != nil {
			return errorResult("CLI execution failed: " + err.Error()), nil, nil //nolint:nilerr // intentional: convert Go error to MCP error result
		}
		mcpResult, _ := ResultFromCLI(result)
		return mcpResult, nil, nil
	})
}

// main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mark3labs/mcp-go/server"
	"log"
	"log/slog"
	"os"
	"runtime"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

var tools = []mcp.Tool{
	{
		Name:        "get_host_info",
		Description: "Get host information (hostname, OS, platform, boot time)",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	},
	{
		Name:        "get_cpu_info",
		Description: "Get CPU usage percentage and core count",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	},
	{
		Name:        "get_memory_info",
		Description: "Get memory usage (total, available, used, percent)",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	},
	{
		Name:        "get_disk_usage",
		Description: "Get disk usage for root partition (or C: on Windows)",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Filesystem path to check (default: / or C:)",
					"default":     getDefaultDiskPath(),
				},
			},
			Required: []string{"path"},
		},
	},
	{
		Name:        "get_network_stats",
		Description: "Get basic network I/O statistics (bytes sent/received)",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	},
}

func getDefaultDiskPath() string {
	if runtime.GOOS == "windows" {
		return "C:"
	}
	return "/"
}

func handleToolCall(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var result interface{}
	var err error
	name := request.Params.Name
	slog.Info("Getting host info")
	switch name {
	case "get_host_info":
		result, err = collectHostInfo()
	case "get_cpu_info":
		result, err = collectCPUInfo()
	case "get_memory_info":
		result, err = collectMemoryInfo()
	case "get_disk_usage":
		path := getDefaultDiskPath()
		if p, ok := request.GetArguments()["path"].(string); ok && p != "" {
			path = p
		}
		result, err = collectDiskUsage(path)
	case "get_network_stats":
		result, err = collectNetworkStats()
	default:
		return &mcp.CallToolResult{}, mcp.ErrInternalError
	}

	if err != nil {
		return &mcp.CallToolResult{}, fmt.Errorf("failed to collect %s: %w", name, err)
	}
	data, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(data)), nil
}

func collectHostInfo() (map[string]interface{}, error) {
	info, err := host.Info()
	if err != nil {
		log.Printf("[ERROR] Failed to get host info: %v", err)
		return nil, err
	}
	return map[string]interface{}{
		"hostname":         info.Hostname,
		"os":               info.OS,
		"platform":         info.Platform,
		"platform_version": info.PlatformVersion,
		"kernel_version":   info.KernelVersion,
		"boot_time":        info.BootTime,
		//"uptime_seconds":   uint64(host.Uptime()),
	}, nil
}

func collectCPUInfo() (map[string]interface{}, error) {
	percent, err := cpu.Percent(0, false)
	if err != nil {
		return nil, err
	}
	cores := runtime.NumCPU()
	return map[string]interface{}{
		"cpu_percent": percent[0],
		"cpu_cores":   cores,
	}, nil
}

func collectMemoryInfo() (map[string]interface{}, error) {
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"total_bytes":     vmem.Total,
		"available_bytes": vmem.Available,
		"used_bytes":      vmem.Used,
		"used_percent":    vmem.UsedPercent,
	}, nil
}

func collectDiskUsage(path string) (map[string]interface{}, error) {
	usage, err := disk.Usage(path)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"path":            path,
		"total_bytes":     usage.Total,
		"free_bytes":      usage.Free,
		"used_bytes":      usage.Used,
		"used_percent":    usage.UsedPercent,
		"filesystem_type": usage.Fstype,
	}, nil
}

func collectNetworkStats() (map[string]interface{}, error) {
	counters, err := net.IOCounters(false)
	if err != nil || len(counters) == 0 {
		return nil, err
	}
	c := counters[0]
	return map[string]interface{}{
		"bytes_sent":   c.BytesSent,
		"bytes_recv":   c.BytesRecv,
		"packets_sent": c.PacketsSent,
		"packets_recv": c.PacketsRecv,
	}, nil
}

func main() {
	// Set up logging to stderr (since stdout is used for MCP communication)
	logFile, _ := os.Create("mcp-server.log")
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stderr)
	log.Printf("Starting gopsutil-mcp server...")
	log.Print(host.Info())
	// Create new MCP server with tools enabled
	s := server.NewMCPServer(
		"gopsutil-mcp",
		"0.1.0",
		server.WithToolCapabilities(true), // Enable tools
		server.WithLogging(),              // Add logging
	)

	for _, tool := range tools {
		s.AddTool(tool, handleToolCall)
	}

	// Run server using stdio
	log.Printf("Starting stdio server...")
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

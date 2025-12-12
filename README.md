# gopsutil-mcp — System Monitoring Tools via Model Context Protocol (MCP)

`gopsutil-mcp` is an [MCP (Model Context Protocol)](https://modelcontextprotocol.io/) server that exposes system monitoring capabilities using the [`gopsutil`](https://github.com/shirou/gopsutil) library. It allows LLM-powered applications (e.g., Cursor, Continue.dev, or other MCP-compatible clients) to securely retrieve real-time host, CPU, memory, disk, and network metrics from the machine where it runs.

---

## ✨ Features

- **Host Info**: Hostname, OS, platform, kernel version, boot time
- **CPU Stats**: Usage percentage and core count
- **Memory Stats**: Total, available, used memory and usage percent
- **Disk Usage**: Free/used space for any path (defaults to `/` or `C:`)
- **Network I/O**: Bytes and packets sent/received

All tools require **no authentication** and run locally—ideal for development, debugging, or internal tooling.

---

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- A MCP-compatible client (e.g., [Cursor](https://cursor.sh/), [Continue](https://continue.dev/))

### Build & Run

```bash
# Clone (if needed)
git clone https://github.com/your-username/gopsutil-mcp.git
cd gopsutil-mcp

# Build
go build -o gopsutil-mcp .

# Run (communicates via stdin/stdout)
./gopsutil-mcp
```

> 💡 The server uses **stdio** for MCP communication. Do not run it as a background daemon unless integrated into a parent process that manages stdio.

---

## 🔧 Tools Reference

| Tool Name | Description | Parameters |
|----------|-------------|-----------|
| `get_host_info` | Get OS and host metadata | None |
| `get_cpu_info` | Get CPU usage and core count | None |
| `get_memory_info` | Get memory usage stats | None |
| `get_disk_usage` | Get disk usage for a path | `path` (string, optional; defaults to `/` or `C:`) |
| `get_network_stats` | Get network I/O counters | None |

> ⚠️ Note: All responses are returned as **JSON strings** inside a `text` content block (per MCP spec).

### Example Response (`get_host_info`)

```json
{
  "hostname": "my-laptop",
  "os": "linux",
  "platform": "ubuntu",
  "platform_version": "22.04",
  "kernel_version": "5.15.0-101-generic",
  "boot_time": 1717020000
}
```

---

## 📦 Integration

To use in an MCP client:

1. Launch `gopsutil-mcp` as a subprocess.
2. Connect via stdio.
3. Call tools by name (e.g., `"get_memory_info"` with empty `{}` args).

Example prompt for an LLM:
> “Check how much free memory is available on the system.”

The LLM will invoke `get_memory_info`, parse the JSON response, and report results.

---

## 🛡️ Security Notes

- This tool **exposes system information**—only run in trusted environments.
- No sandboxing or access control is implemented.
- Avoid running on production servers or multi-tenant systems unless hardened.

---

## 📄 License

MIT License. See [LICENSE](LICENSE) for details.

---

## 🙌 Acknowledgements

- [gopsutil](https://github.com/shirou/gopsutil) – Cross-platform system monitoring in Go
- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) – Go implementation of MCP
- [Model Context Protocol](https://modelcontextprotocol.io/) – Standard for LLM tool interoperability

---

> 💡 **Tip**: Use this server to give your AI assistant "eyes" into your local machine’s health!
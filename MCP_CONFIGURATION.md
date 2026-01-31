# MCP Configuration Guide for NATS Documentation Server

This guide shows how to configure the NATS Documentation MCP Server for different MCP clients and environments.

## Quick Start

### Claude Desktop (Recommended)

1. **Build and install the server:**
```bash
goreleaser build --snapshot --clean
# Binary will be in dist/ directory
```

2. **Create a config file** (`~/.config/nats-docs/config.yaml`):
```yaml
log_level: info
docs_base_url: https://docs.nats.io
fetch_timeout: 30
max_concurrent: 5
transport_type: stdio

# Optional: Enable Syncp documentation support
syncp:
  enabled: false
```

3. **Add to Claude Desktop config:**

On **macOS**:
```json
{
  "mcpServers": {
    "nats-docs": {
      "command": "/path/to/nats-docs-mcp-server",
      "args": ["--config", "/Users/YOUR_USERNAME/.config/nats-docs/config.yaml"],
      "description": "NATS Documentation - search NATS and Syncp docs"
    }
  }
}
```

File location: `~/Library/Application Support/Claude/claude_desktop_config.json`

On **Windows**:
```json
{
  "mcpServers": {
    "nats-docs": {
      "command": "C:\\path\\to\\nats-docs-mcp-server.exe",
      "args": ["--config", "C:\\Users\\YOUR_USERNAME\\AppData\\Local\\nats-docs\\config.yaml"],
      "description": "NATS Documentation - search NATS and Syncp docs"
    }
  }
}
```

File location: `%APPDATA%\Claude\claude_desktop_config.json`

On **Linux**:
```json
{
  "mcpServers": {
    "nats-docs": {
      "command": "/home/username/.local/bin/nats-docs-mcp-server",
      "args": ["--config", "/home/username/.config/nats-docs/config.yaml"],
      "description": "NATS Documentation - search NATS and Syncp docs"
    }
  }
}
```

File location: `~/.config/Claude/claude_desktop_config.json`

---

## Configuration Examples

### Example 1: Basic NATS-Only (Recommended for Most Users)

```json
{
  "mcpServers": {
    "nats-docs": {
      "command": "/usr/local/bin/nats-docs-mcp-server",
      "args": ["--config", "/etc/nats-docs/config.yaml"],
      "description": "NATS Documentation Server"
    }
  }
}
```

**config.yaml:**
```yaml
log_level: info
docs_base_url: https://docs.nats.io
fetch_timeout: 30
max_concurrent: 5
transport_type: stdio
syncp:
  enabled: false
```

### Example 2: Dual Documentation Sources (NATS + Syncp)

```json
{
  "mcpServers": {
    "nats-docs": {
      "command": "/usr/local/bin/nats-docs-mcp-server",
      "args": ["--config", "/etc/nats-docs/config.yaml"],
      "description": "NATS & Syncp Documentation Server",
      "env": {
        "NATS_DOCS_LOG_LEVEL": "info"
      }
    }
  }
}
```

**config.yaml:**
```yaml
log_level: info
docs_base_url: https://docs.nats.io
fetch_timeout: 30
max_concurrent: 5
transport_type: stdio

syncp:
  enabled: true
  base_url: https://docs.synadia.com/control-plane
  fetch_timeout: 30

classification:
  syncp_keywords:
    - syncp
    - control-plane
    - synadia
    - namespace
    - managed
  nats_keywords:
    - jetstream
    - nats-server
    - nats-cli
    - subject
    - stream
    - consumer
```

### Example 3: Custom Keywords for Classification

```json
{
  "mcpServers": {
    "nats-docs": {
      "command": "/usr/local/bin/nats-docs-mcp-server",
      "args": ["--config", "/etc/nats-docs/config.yaml"],
      "description": "NATS & Syncp Documentation with Custom Keywords"
    }
  }
}
```

**config.yaml:**
```yaml
log_level: debug
docs_base_url: https://docs.nats.io
fetch_timeout: 30
max_concurrent: 5
transport_type: stdio

syncp:
  enabled: true
  base_url: https://docs.synadia.com/control-plane
  fetch_timeout: 30

classification:
  # Custom keywords optimized for your use case
  syncp_keywords:
    - syncp
    - control-plane
    - synadia
    - namespace
    - managed
    - account-control
    - dashboard
    - integrations
  nats_keywords:
    - jetstream
    - nats-server
    - nats-cli
    - subject
    - stream
    - consumer
    - pubsub
    - request-reply
    - queue
    - kvstore
    - objectstore
```

### Example 4: Development Environment (Debug Logging)

```json
{
  "mcpServers": {
    "nats-docs-dev": {
      "command": "/home/dev/projects/nats-docs-mcp-server/dist/nats-docs-mcp-server_linux_amd64/nats-docs-mcp-server",
      "args": [
        "--config", "/home/dev/.config/nats-docs/config.yaml",
        "--log-level", "debug"
      ],
      "description": "NATS Documentation Server (Development)",
      "env": {
        "RUST_LOG": "debug"
      }
    }
  }
}
```

**config.yaml:**
```yaml
log_level: debug
docs_base_url: https://docs.nats.io
fetch_timeout: 30
max_concurrent: 5
transport_type: stdio

syncp:
  enabled: true
  base_url: https://docs.synadia.com/control-plane
  fetch_timeout: 30
```

### Example 5: Command-Line Only (No Config File)

If you prefer to configure everything via command-line flags:

```json
{
  "mcpServers": {
    "nats-docs": {
      "command": "/usr/local/bin/nats-docs-mcp-server",
      "args": [
        "--log-level", "info",
        "--transport", "stdio"
      ],
      "description": "NATS Documentation (CLI configured)"
    }
  }
}
```

This uses all default values except for the specified flags.

### Example 6: Network Transport (SSE - Server-Sent Events)

```json
{
  "mcpServers": {
    "nats-docs-sse": {
      "command": "/usr/local/bin/nats-docs-mcp-server",
      "args": [
        "--config", "/etc/nats-docs/config.yaml",
        "--transport", "sse",
        "--host", "localhost",
        "--port", "8000"
      ],
      "description": "NATS Documentation Server (SSE Transport)"
    }
  }
}
```

**config.yaml:**
```yaml
log_level: info
docs_base_url: https://docs.nats.io
fetch_timeout: 30
max_concurrent: 5
transport_type: sse
host: localhost
port: 8000

syncp:
  enabled: false
```

### Example 7: Enterprise HTTP Transport

```json
{
  "mcpServers": {
    "nats-docs-http": {
      "command": "/usr/local/bin/nats-docs-mcp-server",
      "args": [
        "--config", "/etc/nats-docs/config.yaml",
        "--transport", "streamablehttp",
        "--host", "0.0.0.0",
        "--port", "8080"
      ],
      "description": "NATS Documentation Server (Enterprise HTTP)"
    }
  }
}
```

**config.yaml:**
```yaml
log_level: info
docs_base_url: https://docs.nats.io
fetch_timeout: 30
max_concurrent: 5
transport_type: streamablehttp
host: 0.0.0.0
port: 8080

syncp:
  enabled: true
  base_url: https://docs.synadia.com/control-plane
  fetch_timeout: 30
```

---

## Configuration Options

### Server Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `log_level` | string | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `transport_type` | string | `stdio` | Transport: `stdio`, `sse`, `streamablehttp` |
| `host` | string | `localhost` | Host for network transports |
| `port` | integer | `0` | Port for network transports |

### Documentation Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `docs_base_url` | string | `https://docs.nats.io` | NATS documentation base URL |
| `fetch_timeout` | integer | `30` | Fetch timeout in seconds |
| `max_concurrent` | integer | `5` | Maximum concurrent fetches |
| `max_search_results` | integer | `50` | Maximum search results per query |

### Syncp Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `syncp.enabled` | boolean | `false` | Enable Syncp documentation support |
| `syncp.base_url` | string | `https://docs.synadia.com/control-plane` | Syncp documentation URL |
| `syncp.fetch_timeout` | integer | `30` | Syncp fetch timeout in seconds |
| `classification.syncp_keywords` | array | [default list] | Keywords for Syncp classification |
| `classification.nats_keywords` | array | [default list] | Keywords for NATS classification |

---

## Environment Variables

You can override configuration with environment variables using the `NATS_DOCS_` prefix:

```bash
export NATS_DOCS_LOG_LEVEL=debug
export NATS_DOCS_DOCS_BASE_URL=https://docs.nats.io
export NATS_DOCS_FETCH_TIMEOUT=60
export NATS_DOCS_MAX_CONCURRENT=10
export NATS_DOCS_TRANSPORT_TYPE=sse
export NATS_DOCS_HOST=0.0.0.0
export NATS_DOCS_PORT=8080
```

---

## Troubleshooting

### Server fails to start

**Check logs:**
```bash
# Look for error messages in Claude's logs
# macOS: ~/Library/Application Support/Claude/Logs
# Windows: %APPDATA%\Claude\Logs
# Linux: ~/.config/Claude/logs
```

**Verify binary exists:**
```bash
ls -la /path/to/nats-docs-mcp-server
```

**Test the server directly:**
```bash
/path/to/nats-docs-mcp-server --version
/path/to/nats-docs-mcp-server --config /path/to/config.yaml
```

### Documentation fetch fails at startup

**Verify network connectivity:**
```bash
curl https://docs.nats.io/sitemap-pages.xml
```

**Increase timeout:**
```yaml
fetch_timeout: 60  # Increase from default 30
```

### Poor search results

**Debug query classification:**
```yaml
log_level: debug  # Enable debug logging to see classification decisions
```

**Review keyword lists:**
Make sure your classification keywords match your use case.

---

## Next Steps

1. **Copy config.example.yaml** to your preferred location
2. **Update file paths** in your MCP configuration
3. **Test the server** with your MCP client
4. **Restart** your MCP client to load the new server
5. **Verify** by searching NATS documentation

For more information, see [README.md](README.md) and [CONTRIBUTING.md](CONTRIBUTING.md).

# Setup Guide: NATS Documentation MCP Server

This guide walks you through setting up the NATS Documentation MCP Server with Claude Desktop.

## Quick Start (5 minutes)

### 1. Download or Build the Server

**Option A: Download Pre-built Binary**
- Go to [Releases](https://github.com/j4ng5y/nats-docs-mcp-server/releases)
- Download the binary for your platform
- Extract and move to a location in your PATH or note the path

**Option B: Build from Source**
```bash
git clone https://github.com/j4ng5y/nats-docs-mcp-server.git
cd nats-docs-mcp-server
goreleaser build --snapshot --clean
# Binary will be in dist/
```

### 2. Create Configuration Directory and File

**macOS/Linux:**
```bash
mkdir -p ~/.config/nats-docs
cp examples/config-nats-only.yaml ~/.config/nats-docs/config.yaml
```

**Windows:**
```cmd
mkdir %APPDATA%\nats-docs
copy examples\config-nats-only.yaml %APPDATA%\nats-docs\config.yaml
```

### 3. Update Claude Desktop Configuration

**macOS:**
1. Open `~/Library/Application Support/Claude/claude_desktop_config.json`
2. Add this to `mcpServers`:
```json
"nats-docs": {
  "command": "/usr/local/bin/nats-docs-mcp-server",
  "args": ["--config", "/Users/USERNAME/.config/nats-docs/config.yaml"]
}
```
3. Replace `/usr/local/bin` with your binary location
4. Replace `USERNAME` with your actual username

**Windows:**
1. Open `%APPDATA%\Claude\claude_desktop_config.json`
2. Add this to `mcpServers`:
```json
"nats-docs": {
  "command": "C:\\Program Files\\nats-docs\\nats-docs-mcp-server.exe",
  "args": ["--config", "C:\\Users\\USERNAME\\AppData\\Local\\nats-docs\\config.yaml"]
}
```
3. Update paths to match your installation

**Linux:**
1. Open `~/.config/Claude/claude_desktop_config.json`
2. Add this to `mcpServers`:
```json
"nats-docs": {
  "command": "/home/username/.local/bin/nats-docs-mcp-server",
  "args": ["--config", "/home/username/.config/nats-docs/config.yaml"]
}
```

### 4. Restart Claude Desktop

- Quit Claude Desktop completely
- Reopen Claude Desktop
- The server will start and fetch documentation on first use

### 5. Test It Works

Ask Claude to search NATS documentation:
> "Search for information about JetStream consumers"

## Configuration Options

### Basic Configuration (NATS Only)

```yaml
log_level: info
docs_base_url: https://docs.nats.io
fetch_timeout: 30
max_concurrent: 5
transport_type: stdio
syncp:
  enabled: false
```

See `examples/config-nats-only.yaml`

### Advanced Configuration (Dual Sources)

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
  nats_keywords:
    - jetstream
    - nats-server
    - stream
    - consumer
```

See `examples/config-dual-sources.yaml`

### Development Configuration (Debug Logging)

```yaml
log_level: debug
docs_base_url: https://docs.nats.io
...
```

See `examples/config-development.yaml`

## Troubleshooting

### Server doesn't appear in Claude Desktop

1. **Check configuration file exists:**
   ```bash
   cat ~/.config/nats-docs/config.yaml  # macOS/Linux
   type %APPDATA%\nats-docs\config.yaml  # Windows
   ```

2. **Test the server directly:**
   ```bash
   /path/to/nats-docs-mcp-server --version
   /path/to/nats-docs-mcp-server --config ~/.config/nats-docs/config.yaml
   ```

3. **Check Claude Desktop config syntax:**
   - Make sure JSON is valid (use online validator)
   - Paths are correct for your system
   - File exists at the path specified

4. **Check logs:**
   - macOS: `~/Library/Logs/Claude/` or `~/Library/Application Support/Claude/`
   - Windows: Check `%APPDATA%\Claude\` for logs
   - Linux: Check `~/.config/Claude/` for logs

### Documentation fetch fails

1. **Verify network access:**
   ```bash
   curl https://docs.nats.io/sitemap-pages.xml
   ```

2. **Increase timeout:**
   ```yaml
   fetch_timeout: 60  # Increase from 30
   ```

3. **Check firewall/proxy settings** - the server needs outbound HTTPS access

### Search results are missing

1. **Enable debug logging:**
   ```yaml
   log_level: debug
   ```

2. **Check query classification:**
   - Debug logs will show which documentation source was searched
   - Verify your keywords are correct

3. **Try broader search terms:**
   - "jetstream" instead of "jetstream consumer"
   - "configuration" instead of specific config option

## File Locations

### Configuration File

**macOS/Linux:**
- Default: `~/.config/nats-docs/config.yaml`
- Alternative: `/etc/nats-docs/config.yaml` (for system-wide)

**Windows:**
- Default: `%APPDATA%\nats-docs\config.yaml`
- Alternative: `C:\Program Files\nats-docs\config.yaml`

### Binary Installation

**macOS (Homebrew - if available):**
```bash
brew tap j4ng5y/nats-docs-mcp
brew install nats-docs-mcp-server
# Binary will be at /usr/local/bin/nats-docs-mcp-server
```

**macOS (Manual):**
```bash
# Copy to /usr/local/bin
sudo cp dist/nats-docs-mcp-server_darwin_amd64/nats-docs-mcp-server /usr/local/bin/
sudo chmod +x /usr/local/bin/nats-docs-mcp-server
```

**Linux:**
```bash
# Copy to ~/.local/bin
cp dist/nats-docs-mcp-server_linux_amd64/nats-docs-mcp-server ~/.local/bin/
chmod +x ~/.local/bin/nats-docs-mcp-server
```

**Windows:**
```cmd
REM Copy to Program Files
copy dist\nats-docs-mcp-server_windows_amd64\nats-docs-mcp-server.exe "C:\Program Files\nats-docs\"
```

## Advanced Setup

### Multiple Instances

You can run multiple instances for NATS-only and dual-source:

```json
{
  "mcpServers": {
    "nats-docs": {
      "command": "/usr/local/bin/nats-docs-mcp-server",
      "args": ["--config", "~/.config/nats-docs/config-nats-only.yaml"]
    },
    "nats-docs-syncp": {
      "command": "/usr/local/bin/nats-docs-mcp-server",
      "args": ["--config", "~/.config/nats-docs/config-dual-sources.yaml"]
    }
  }
}
```

### Environment Variables

You can use environment variables instead of/alongside config file:

```bash
export NATS_DOCS_LOG_LEVEL=debug
export NATS_DOCS_FETCH_TIMEOUT=60
export NATS_DOCS_MAX_CONCURRENT=10
```

Then run:
```bash
/path/to/nats-docs-mcp-server --config ~/.config/nats-docs/config.yaml
```

### Custom Keywords

To customize query classification keywords, edit your config:

```yaml
classification:
  syncp_keywords:
    - my-custom-keyword
    - another-keyword
  nats_keywords:
    - my-nats-keyword
    - another-nats-keyword
```

## Getting Help

- Check [README.md](README.md) for overview and features
- See [MCP_CONFIGURATION.md](MCP_CONFIGURATION.md) for detailed configuration options
- Check [CONTRIBUTING.md](CONTRIBUTING.md) for development information
- File an issue on [GitHub](https://github.com/j4ng5y/nats-docs-mcp-server/issues)

## Next Steps

1. ✅ Set up the server
2. ✅ Configure Claude Desktop
3. 🎉 Start searching NATS documentation!

**Example queries:**
- "How do I create a JetStream stream?"
- "What is a NATS consumer?"
- "How do I set up clustering?"
- "What are NATS subjects?"

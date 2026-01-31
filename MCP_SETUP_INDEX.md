# MCP Setup & Configuration - Complete Index

This document provides a complete index of all MCP configuration files and guides for the NATS Documentation Server.

## 📚 Documentation Files

### Primary Setup Guide
- **[SETUP_GUIDE.md](SETUP_GUIDE.md)** - **START HERE** ⭐
  - Quick start (5 minutes)
  - Platform-specific instructions (macOS, Windows, Linux)
  - Troubleshooting guide
  - File location reference

### Detailed Configuration Guide
- **[MCP_CONFIGURATION.md](MCP_CONFIGURATION.md)** - Advanced configurations
  - All configuration options explained
  - 7 different example configurations
  - Environment variable usage
  - Troubleshooting and debugging

### README Documentation
- **[README.md](README.md)** - Project overview
  - Features and capabilities
  - Syncp documentation support overview
  - Installation instructions
  - Architecture overview

---

## 🔧 MCP Configuration Files

### Generic Format
- **[.mcp.json](.mcp.json)** - Basic MCP format
- **[.mcp.json.example](.mcp.json.example)** - Example with comments

### Claude Desktop (Recommended)
- **[claude_desktop_config.json.example](claude_desktop_config.json.example)** - Generic example

### Platform-Specific

#### macOS
- **[examples/macos-claude-desktop-config.json](examples/macos-claude-desktop-config.json)**
  - Location: `~/Library/Application Support/Claude/claude_desktop_config.json`
  - Binary path: `/usr/local/bin/nats-docs-mcp-server`
  - Config path: `/Users/USERNAME/.config/nats-docs/config.yaml`

#### Windows
- **[examples/windows-claude-desktop-config.json](examples/windows-claude-desktop-config.json)**
  - Location: `%APPDATA%\Claude\claude_desktop_config.json`
  - Binary path: `C:\Program Files\nats-docs\nats-docs-mcp-server.exe`
  - Config path: `C:\Users\USERNAME\AppData\Local\nats-docs\config.yaml`

#### Linux
- **[examples/linux-claude-desktop-config.json](examples/linux-claude-desktop-config.json)**
  - Location: `~/.config/Claude/claude_desktop_config.json`
  - Binary path: `~/.local/bin/nats-docs-mcp-server`
  - Config path: `~/.config/nats-docs/config.yaml`

---

## ⚙️ Server Configuration Files (YAML)

### Pre-configured Options

#### NATS Only (Recommended)
- **[examples/config-nats-only.yaml](examples/config-nats-only.yaml)**
  - NATS documentation enabled
  - Syncp support disabled (optional)
  - Standard logging (info level)
  - **Best for:** Most users, production use

#### Dual Documentation Sources
- **[examples/config-dual-sources.yaml](examples/config-dual-sources.yaml)**
  - NATS documentation enabled
  - Syncp (Synadia Control Plane) documentation enabled
  - Query classification configured
  - **Best for:** Users who need both documentation sources

#### Development Configuration
- **[examples/config-development.yaml](examples/config-development.yaml)**
  - Debug logging enabled
  - Both documentation sources enabled
  - Extended keyword lists
  - **Best for:** Developers, testing, debugging

---

## 📋 Configuration File Reference

### Quick Comparison Table

| File | Purpose | Best For | Syncp |
|------|---------|----------|-------|
| config-nats-only.yaml | Standard setup | Most users | ❌ No |
| config-dual-sources.yaml | Dual sources | Advanced users | ✅ Yes |
| config-development.yaml | Development | Testing/debugging | ✅ Yes |

### Configuration Options

```yaml
# Server Logging
log_level: info|debug|warn|error

# NATS Documentation
docs_base_url: https://docs.nats.io
fetch_timeout: 30                    # seconds
max_concurrent: 5                    # parallel fetches
max_search_results: 10               # results per query

# Transport
transport_type: stdio|sse|streamablehttp
host: localhost|0.0.0.0              # for network transports
port: 8080                           # for network transports

# Syncp (Optional)
syncp:
  enabled: true|false
  base_url: https://docs.synadia.com/control-plane
  fetch_timeout: 30

# Query Classification (if syncp enabled)
classification:
  syncp_keywords: [list of keywords]
  nats_keywords: [list of keywords]
```

---

## 🚀 Quick Start Summary

### Step 1: Choose Configuration
Pick one:
- **Beginners:** Use `examples/config-nats-only.yaml`
- **Advanced:** Use `examples/config-dual-sources.yaml`
- **Developers:** Use `examples/config-development.yaml`

### Step 2: Choose Your Platform
- **macOS:** Copy `examples/macos-claude-desktop-config.json`
- **Windows:** Copy `examples/windows-claude-desktop-config.json`
- **Linux:** Copy `examples/linux-claude-desktop-config.json`

### Step 3: Customize
- Update binary path to your installation location
- Update username in config paths
- Adjust any configuration options as needed

### Step 4: Install
- Copy YAML config to appropriate location
- Copy JSON config to Claude Desktop config directory
- Restart Claude Desktop

### Step 5: Test
Ask Claude: "Search for jetstream consumer documentation"

---

## 📖 Documentation Hierarchy

```
MCP_SETUP_INDEX.md (You are here)
├── SETUP_GUIDE.md (Start here for setup)
├── MCP_CONFIGURATION.md (Detailed options)
├── README.md (Project overview)
├── CONTRIBUTING.md (Development guide)
└── Configuration Examples/
    ├── examples/
    │   ├── macos-claude-desktop-config.json
    │   ├── windows-claude-desktop-config.json
    │   ├── linux-claude-desktop-config.json
    │   ├── config-nats-only.yaml
    │   ├── config-dual-sources.yaml
    │   └── config-development.yaml
    └── Root/
        ├── .mcp.json
        ├── .mcp.json.example
        └── claude_desktop_config.json.example
```

---

## 🎯 Use Cases & Recommendations

### I want to...

**Search NATS documentation from Claude Desktop**
→ Use `examples/config-nats-only.yaml` and platform-specific JSON config

**Use both NATS and Syncp documentation**
→ Use `examples/config-dual-sources.yaml` and enable syncp

**Debug query classification**
→ Use `examples/config-development.yaml` with `log_level: debug`

**Run in production**
→ Use `examples/config-nats-only.yaml` with `log_level: info`

**Deploy on Linux server**
→ Use platform-independent YAML config + Linux JSON template

**Deploy on Windows**
→ Use platform-independent YAML config + Windows JSON template

**Deploy on macOS**
→ Use platform-independent YAML config + macOS JSON template

---

## ❓ Frequently Asked Questions

### Q: Which config should I use?
**A:** Start with `examples/config-nats-only.yaml` for standard usage.

### Q: How do I enable Syncp support?
**A:** Copy `examples/config-dual-sources.yaml` and set `syncp.enabled: true`.

### Q: Can I use environment variables?
**A:** Yes! See `MCP_CONFIGURATION.md` for `NATS_DOCS_*` environment variables.

### Q: Which platform-specific JSON should I use?
**A:** Choose based on your OS:
- macOS → `macos-claude-desktop-config.json`
- Windows → `windows-claude-desktop-config.json`
- Linux → `linux-claude-desktop-config.json`

### Q: What if I don't have Claude Desktop?
**A:** See `MCP_CONFIGURATION.md` for SSE and HTTP transport examples.

### Q: Can I run multiple instances?
**A:** Yes! See `MCP_CONFIGURATION.md` for multiple instance setup.

---

## 📍 File Locations

### Configuration Directory Paths

**macOS/Linux:**
```
~/.config/nats-docs/
  └── config.yaml
```

**Windows:**
```
%APPDATA%\nats-docs\
  └── config.yaml
```

### Claude Desktop Config Paths

**macOS:**
```
~/Library/Application Support/Claude/claude_desktop_config.json
```

**Windows:**
```
%APPDATA%\Claude\claude_desktop_config.json
```

**Linux:**
```
~/.config/Claude/claude_desktop_config.json
```

---

## ✨ Features Supported

✅ NATS Documentation Search
✅ Syncp (Optional) Documentation Search
✅ Query Classification & Routing
✅ Full-Text Search with Relevance
✅ Multiple Transport Types
✅ Configuration via YAML/Flags/Environment
✅ Debug Logging Support
✅ Graceful Degradation (if Syncp unavailable)

---

## 🔗 Related Resources

- [NATS Official Documentation](https://docs.nats.io)
- [Synadia Control Plane Documentation](https://docs.synadia.com/control-plane)
- [Model Context Protocol](https://modelcontextprotocol.io)
- [Claude Desktop](https://claude.ai)

---

## 📝 Summary

This MCP server provides:
- **Easy Claude Desktop integration** - Just copy config files
- **Flexible configuration** - YAML, CLI flags, or environment variables
- **Multiple platforms** - macOS, Windows, Linux support
- **Optional Syncp support** - Enable when needed
- **Production ready** - Tested and reliable

**Next Steps:**
1. Read `SETUP_GUIDE.md` for your platform
2. Copy appropriate configuration files
3. Update paths for your system
4. Restart Claude Desktop
5. Start searching NATS docs!

---

**Last Updated:** 2026-01-31
**Version:** 1.0.0

# Product Overview

NATS Documentation MCP Server is a Model Context Protocol (MCP) compliant server that provides LLMs with programmatic access to NATS documentation from https://docs.nats.io/.

## Purpose

Enable AI assistants to search and retrieve official NATS documentation through standardized MCP tools, allowing them to provide accurate, up-to-date information about NATS messaging system.

## Key Features

- MCP-compliant server exposing documentation search and retrieval tools
- Fast in-memory documentation indexing with TF-IDF relevance ranking
- Single binary distribution with no external dependencies
- Cross-platform support (Linux, macOS, Windows on AMD64 and ARM64)
- Automated build and release pipeline via GoReleaser and GitHub Actions

## Target Users

- LLM applications (Claude, etc.) needing access to NATS documentation
- Developers building AI-powered tools that require NATS knowledge
- System administrators deploying MCP servers for documentation access

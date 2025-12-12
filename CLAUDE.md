# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MCPProbe is a command-line utility for testing MCP (Model Context Protocol) servers. It's written in Go and provides capabilities for server discovery, tool enumeration, and direct tool execution.

## Build and Development Commands

```bash
# Build the application
go build -o mcp-probe

# Format code (required before committing)
gofmt -w .

# Run linting (fixes required issues)
go vet ./...

# Clean build
go clean
rm -f mcp-probe probe
```

## Architecture

The codebase is a single-file Go application (`main.go`) with the following key components:

1. **Transport Layer**: Supports both SSE and HTTP transports via the `github.com/mark3labs/mcp-go` library
2. **Client Management**: Creates and manages MCP client connections with proper initialization handshake
3. **Tool Discovery**: Enumerates server capabilities (tools, resources, prompts)  
4. **Tool Execution**: Executes MCP tools with JSON parameter support
5. **Interactive Mode**: Provides a REPL-like interface for exploring and testing tools

### Key Functions

- `main()`: Entry point, handles CLI flags and orchestrates the flow
- `performInitialization()`: Performs MCP protocol handshake
- `testServerCapabilities()`: Tests all server capabilities (tools, resources, prompts)
- `callSpecificTool()`: Executes a single tool with parameters
- `interactiveModeWithTimeout()`: Runs the interactive REPL interface
- `collectToolParameters()`: Parses tool JSON schemas and collects user input

## Code Style

- Use standard Go formatting (`gofmt`)
- Fix all `go vet` warnings before committing
- Avoid redundant newlines in `fmt.Println()` calls (common lint issue)
- Follow existing error handling patterns with descriptive messages

## Dependencies

- `github.com/mark3labs/mcp-go` v0.31.0 - Core MCP protocol implementation
- Go 1.24.3 or higher

## Common Issues

1. **Redundant newlines**: `go vet` will flag `fmt.Println("\n")` - use `fmt.Println()` instead
2. **Build artifacts**: Clean up `mcp-probe` and `probe` binaries when needed
# MCPProbe

This utility is for testing MCP (Model Context Protocol) servers. MCPProbe enumerates server capabilities and provides full tool calling functionality, allowing you to execute MCP tools directly from the command line and see their responses.

## Features

- **Multiple Transport Modes**: Supports HTTP, SSE (Server-Sent Events), and stdio connections
- **Initialization Handshake**: Performs proper MCP protocol initialization
- **Capability Discovery**: Reports on server capabilities (tools, resources, prompts)
- **Tool Calling**: Execute MCP tools directly from the command line with parameters
- **Interactive Mode**: Guided tool calling interface for exploration and testing
- **Verbose Output**: Provides detailed information about the connection and server responses
- **Custom Headers**: Supports custom HTTP headers for authentication or other purposes

## Installation

```bash
# Clone the repository
git clone https://github.com/PivotLLM/MCPProbe.git
cd MCPProbe

# Build the application
go build -o mcp-probe
```

## Quick Start

```bash
# 1. Test server connectivity and see all capabilities (SSE)
./mcp-probe -url http://localhost:8000/sse

# 2. Test a local MCP server via stdio
./mcp-probe -stdio ./my-mcp-server

# 3. List tool names only (minimal output)
./mcp-probe -url http://localhost:8000/sse -list

# 4. List available tools with details
./mcp-probe -url http://localhost:8000/sse -list-only

# 5. Call a specific tool
./mcp-probe -url http://localhost:8000/sse -call "echo" -params '{"message":"Hello MCP!"}'

# 6. Interactive exploration
./mcp-probe -url http://localhost:8000/sse -interactive
```

## Usage Modes

MCPProbe operates in five modes:

### 1. Discovery Mode (Default)
Tests the MCP server and reports all capabilities (tools, resources, prompts).
```bash
./mcp-probe -url <server-url>
```

### 2. List Mode (Minimal)
Lists tool names only with minimal output:
```bash
./mcp-probe -url <server-url> -list
```

### 3. List-Only Mode (Detailed)
Lists available tools with descriptions and schemas:
```bash
./mcp-probe -url <server-url> -list-only
```

### 4. Direct Tool Calling Mode
Executes a specific tool with provided parameters.
```bash
./mcp-probe -url <server-url> -call <tool-name> -params '<json>'
```

### 5. Interactive Mode
Provides a guided interface for exploring and testing tools.
```bash
./mcp-probe -url <server-url> -interactive
```

## Command-Line Options

| Option          | Description                                                                                                                                                                             | Default            |
|-----------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------|
| `-url`          | MCP server URL (required for SSE/HTTP)                                                                                                                                                  | -                  |
| `-stdio`        | Path to local MCP server executable (enables stdio transport)                                                                                                                           | -                  |
| `-args`         | Arguments for stdio server (comma-separated)                                                                                                                                            | -                  |
| `-env`          | Environment variables for stdio server (KEY=VALUE,...)                                                                                                                                  | -                  |
| `-transport`    | Transport mode: 'sse' or 'http' (for URL-based connections)                                                                                                                             | `sse`              |
| `-call`         | Name of the tool to call                                                                                                                                                                | -                  |
| `-params`       | JSON string of parameters for tool call                                                                                                                                                 | `{}`               |
| `-list`         | List tool names only (minimal output)                                                                                                                                                   | `false`            |
| `-list-only`    | List available tools with details                                                                                                                                                       | `false`            |
| `-interactive`  | Enable interactive mode                                                                                                                                                                 | `false`            |
| `-headers`      | Custom HTTP headers for authentication and other purposes. Format: 'key1:value1,key2:value2'. Common uses: 'Authorization:Bearer TOKEN' for bearer tokens, 'X-API-Key:KEY' for API keys | -                  |
| `-timeout`      | Connection timeout for initialization and listing                                                                                                                                       | `30s`              |
| `-call-timeout` | Timeout for tool call execution                                                                                                                                                         | `300s` (5 minutes) |
| `-verbose`      | Enable verbose output                                                                                                                                                                   | `true`             |

**Note:** Either `-url` or `-stdio` must be provided. The `-headers` and `-transport` options only apply to URL-based connections (SSE/HTTP).

## Detailed Examples

### Authentication

MCPProbe supports various authentication methods through the `-headers` flag:

#### Bearer Token Authentication
```bash
# Bearer token
./mcp-probe -url http://api.example.com/mcp \
  -headers "Authorization:Bearer YOUR_TOKEN_HERE"
```

#### API Key Authentication
```bash
# Standard X-API-Key header
./mcp-probe -url http://api.example.com/mcp \
  -headers "X-API-Key:sk-1234567890abcdef"

# Custom API key header
./mcp-probe -url http://api.example.com/mcp \
  -headers "Api-Token:your-secret-key"
```

#### Basic Authentication
```bash
# Basic auth (base64 encoded username:password)
./mcp-probe -url http://api.example.com/mcp \
  -headers "Authorization:Basic dXNlcm5hbWU6cGFzc3dvcmQ="
```

#### Multiple Headers
```bash
# Combine authentication with other headers
./mcp-probe -url http://api.example.com/mcp \
  -headers "Authorization:Bearer token123,X-API-Key:backup-key,Content-Type:application/json"

# API key with additional headers
./mcp-probe -url http://api.example.com/mcp \
  -headers "X-API-Key:secret,X-Client-ID:myclient,Accept:application/json"
```

#### Custom Authentication Headers
```bash
# Custom authentication scheme
./mcp-probe -url http://api.example.com/mcp \
  -headers "X-Custom-Auth:custom-token-format"

# Multiple authentication methods (if server supports fallback)
./mcp-probe -url http://api.example.com/mcp \
  -headers "Authorization:Bearer primary-token,X-API-Key:fallback-key"
```

### Basic Server Testing

```bash
# Test an SSE server (default transport)
./mcp-probe -url http://localhost:8000/sse

# Test an HTTP server
./mcp-probe -url http://localhost:8000/mcp -transport http

# Test with extended timeout for slow servers
./mcp-probe -url http://localhost:8000/sse -timeout 60s
```

### Stdio Transport (Local Servers)

The stdio transport allows you to test local MCP servers by spawning them as subprocesses and communicating over stdin/stdout.

```bash
# Test a local MCP server binary
./mcp-probe -stdio ./my-mcp-server

# Test with arguments (comma-separated)
./mcp-probe -stdio python -args "-m,my_mcp_module"

# Test with arguments and environment variables
./mcp-probe -stdio ./my-server -args "--port,8080,--debug" -env "LOG_LEVEL=debug,API_KEY=secret"

# Interactive mode with a local server
./mcp-probe -stdio ./my-server -interactive

# Call a specific tool on a local server
./mcp-probe -stdio ./my-server -call "get_data" -params '{"id": 1}'

# Complex example: local server with auth proxy arguments
./mcp-probe -stdio ./mcprelay -args "-url,http://127.0.0.1:8888/sse,-headers,{\"Authorization\":\"Bearer YOUR_TOKEN\"}" -verbose
```

**Stdio-specific options:**
- `-stdio <path>`: Path to the MCP server executable
- `-args <args>`: Comma-separated arguments to pass to the server
- `-env <vars>`: Comma-separated environment variables in KEY=VALUE format

### Tool Discovery

```bash
# List tool names only (minimal output)
./mcp-probe -url http://localhost:8000/sse -list

# List all available tools with descriptions
./mcp-probe -url http://localhost:8000/sse -list-only

# List tools with full schema information
./mcp-probe -url http://localhost:8000/sse -list-only -verbose

# List tools from authenticated server
./mcp-probe -url http://api.example.com/mcp -list-only \
  -headers "Authorization:Bearer YOUR_TOKEN"
```

### Direct Tool Calling

```bash
# Call a tool without parameters
./mcp-probe -url http://localhost:8000/sse -call "get_time"

# Call with simple parameters
./mcp-probe -url http://localhost:8000/sse \
  -call "echo" \
  -params '{"message":"Hello, MCP!"}'

# Call with authentication
./mcp-probe -url http://api.example.com/mcp \
  -headers "Authorization:Bearer YOUR_TOKEN" \
  -call "get_user_data" \
  -params '{"user_id":"12345"}'

# Call with API key authentication
./mcp-probe -url http://api.example.com/mcp \
  -headers "X-API-Key:your-api-key" \
  -call "fetch_records" \
  -params '{"limit":10}'

# Call with numeric parameters
./mcp-probe -url http://localhost:8000/sse \
  -call "calculate" \
  -params '{"operation":"multiply","x":7,"y":9}'

# Call with complex nested parameters
./mcp-probe -url http://localhost:8000/sse \
  -call "search_documents" \
  -params '{"query":"machine learning","filters":{"type":"pdf","date_range":{"start":"2024-01-01","end":"2024-12-31"},"max_results":10}}'

# Call with array parameters
./mcp-probe -url http://localhost:8000/sse \
  -call "process_batch" \
  -params '{"items":["item1","item2","item3"],"options":{"parallel":true}}'

# Call with extended timeout for long-running tools
./mcp-probe -url http://localhost:8000/sse \
  -call "analyze_large_dataset" \
  -params '{"dataset_id":"12345"}' \
  -call-timeout 10m
```

### Interactive Mode

```bash
# Start interactive mode
./mcp-probe -url http://localhost:8000/sse -interactive

# Interactive mode with authentication
./mcp-probe -url http://api.example.com/mcp -interactive \
  -headers "Authorization:Bearer YOUR_TOKEN"

# Interactive mode with extended tool timeout for long operations
./mcp-probe -url http://localhost:8000/sse -interactive -call-timeout 10m
```

#### Interactive Mode Commands:
- `list` or `ls` - Display all available tools
- `call` or `c` - Start guided tool calling process
- `1`, `2`, `3`... - Call tool by number directly
- `help` or `h` - Show available commands
- `exit` or `quit` - Exit interactive mode

#### Interactive Mode Example Session:
```
=== Interactive Tool Calling Mode ===
Type 'help' for commands, 'exit' to quit

> list

Available tools (3):
  1. echo - Returns the input message
  2. calculate - Performs arithmetic operations
  3. get_weather - Gets weather for a city

> 1

Calling tool: echo
Description: Returns the input message

Enter parameters (press Enter to skip optional parameters):
  message (The message to echo) [required]: Hello from interactive mode!

Calling tool 'echo'...

=== Tool Call Result ===
Tool call succeeded:

Hello from interactive mode!

> exit
Exiting interactive mode...
```

### Real-World Examples

#### Testing a Filesystem MCP Server
```bash
# List files in a directory
./mcp-probe -url http://localhost:8000/sse \
  -call "list_directory" \
  -params '{"path":"/home/user/documents","recursive":false}'

# Read a file
./mcp-probe -url http://localhost:8000/sse \
  -call "read_file" \
  -params '{"path":"/home/user/documents/readme.txt"}'
```

#### Testing a Database MCP Server
```bash
# Execute a query
./mcp-probe -url http://localhost:8000/sse \
  -call "execute_query" \
  -params '{"query":"SELECT * FROM users WHERE active = true","limit":10}'
```

#### Testing an AI/LLM MCP Server
```bash
# Generate text
./mcp-probe -url http://localhost:8000/sse \
  -call "generate_text" \
  -params '{"prompt":"Write a haiku about programming","max_tokens":50,"temperature":0.7}'
```

## Output Format

MCPProbe provides clear, structured output for different modes:

### Discovery Mode Output
```
=== MCP Server Test Tool ===
Server URL: http://localhost:8000/sse
Transport: sse
Timeout: 30s

Creating SSE client...
Starting client connection...
Client connection started successfully

Performing initialization handshake...
Server info: ExampleMCP v1.0.0
Protocol version: 2024-11-05

Server capabilities received:
  - Tools: supported (list_changed: true)
  - Resources: supported (subscribe: false, list_changed: true)

Initialization completed successfully

--- Tools Capability ---
Requesting list of available tools...
Found 2 tools:

  1. echo
     Description: Returns the input message
     Input Schema: {...}

  2. calculate
     Description: Performs basic arithmetic
     Input Schema: {...}

=== Finished ===
```

### Tool Call Output
```
=== Sending Tool Call ===
Tool: calculate
Parameters:
  operation: add (string)
  x: 5 (float64)
  y: 3 (float64)

Calling tool 'calculate'...

=== Tool Call Result ===
Tool call succeeded:

The result of 5 + 3 is 8
```

### Error Output
```
Failed to call tool 'nonexistent':
   Tool 'nonexistent' not found. Use -list-only to see available tools.
```

## Troubleshooting

### Common Issues

#### Connection Refused
```bash
# Error: dial tcp [::1]:8000: connect: connection refused
# Solution: Ensure the MCP server is running on the specified URL
```

#### Invalid Tool Name
```bash
# Error: Tool 'badname' not found
# Solution: Use -list-only to see available tools
./mcp-probe -url http://localhost:8000/sse -list-only
```

#### Invalid JSON Parameters
```bash
# Error: failed to parse parameters JSON: invalid character
# Solution: Ensure JSON is properly quoted and formatted
./mcp-probe -url http://localhost:8000/sse -call "echo" -params '{"message":"test"}'
```

#### Timeout Issues
```bash
# Error: context deadline exceeded during initialization/connection
# Solution: Increase connection timeout for slow servers
./mcp-probe -url http://localhost:8000/sse -timeout 60s

# Error: context deadline exceeded during tool call
# Solution: Increase tool call timeout for long-running tools
./mcp-probe -url http://localhost:8000/sse -call "long_task" -call-timeout 10m

# Both timeouts can be adjusted independently
./mcp-probe -url http://localhost:8000/sse -timeout 60s -call-timeout 15m -interactive
```

**Understanding Timeouts:**
- **`-timeout`** (default 30s): Controls connection, initialization, and listing operations
- **`-call-timeout`** (default 300s/5m): Controls how long tool execution can run

### Debugging Tips

1. **Use verbose mode** to see detailed request/response information
2. **Test connectivity first** with discovery mode before calling tools
3. **Validate JSON** parameters using a JSON validator
4. **Check server logs** for additional error information
5. **Use interactive mode** to explore tools safely

## JSON Parameter Guide

MCPProbe accepts tool parameters as JSON strings. Here are formatting guidelines:

### Basic Types
```bash
# String parameters
-params '{"name":"John Doe","city":"New York"}'

# Numeric parameters
-params '{"count":42,"price":19.99,"temperature":-5}'

# Boolean parameters
-params '{"enabled":true,"debug":false}'

# Null values
-params '{"optional_field":null}'
```

### Complex Structures
```bash
# Arrays
-params '{"tags":["urgent","important"],"scores":[85,92,78]}'

# Nested objects
-params '{"user":{"name":"Alice","age":30,"preferences":{"theme":"dark"}}}'

# Mixed complex structure
-params '{"query":"search term","filters":{"categories":["tech","science"],"date_range":{"start":"2024-01-01","end":"2024-12-31"}},"options":{"max_results":10,"sort_by":"relevance"}}'
```

### Shell Escaping
Different shells handle quotes differently:

#### Bash/Zsh (Linux/macOS)
```bash
# Use single quotes to wrap the JSON
-params '{"message":"Hello World"}'

# Escape inner quotes if needed
-params '{"message":"Say \"Hello\""}'
```

#### Windows Command Prompt
```cmd
# Use double quotes and escape inner quotes
-params "{\"message\":\"Hello World\"}"
```

#### PowerShell
```powershell
# Use single quotes or escape double quotes
-params '{"message":"Hello World"}'
# OR
-params "{`"message`":`"Hello World`"}"
```

## Dependencies

- [github.com/mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) - Go implementation of the Model Context Protocol

## Copyright and license

Copyright (c) 2025-2026 by Tenebris Technologies Inc. This software is licensed under the MIT License. Please see LICENSE for details.

## No Warranty (zilch, none, void, nil, null, "", {}, 0x00, 0b00000000, EOF)

THIS SOFTWARE IS PROVIDED “AS IS,” WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, AND NON-INFRINGEMENT. IN NO EVENT SHALL THE COPYRIGHT HOLDERS OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

Made in Canada

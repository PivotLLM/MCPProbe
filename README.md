# MCPProbe

A comprehensive Go application for testing and interacting with MCP (Model Context Protocol) servers. MCPProbe not only discovers server capabilities but also provides full tool calling functionality, allowing you to execute MCP tools directly from the command line with parameters and see their responses.

## Features

- **Multiple Transport Modes**: Supports both HTTP and SSE (Server-Sent Events) connections
- **Initialization Handshake**: Performs proper MCP protocol initialization
- **Capability Discovery**: Reports on server capabilities (tools, resources, prompts)
- **Tool Calling**: Execute MCP tools directly from the command line with parameters
- **Interactive Mode**: Guided tool calling interface for exploration and testing
- **Detailed Listing**: Lists all available tools, resources, and prompts with descriptions
- **Verbose Output**: Provides detailed information about the connection and server responses
- **Custom Headers**: Supports custom HTTP headers for authentication or other purposes

## Installation

```bash
# Clone the repository
git clone https://github.com/PivotLLM/MCPProbe.git
cd MCPProbe

# Build the application
go build -o mcp-probe

# Or run directly without building
go run main.go -url <server-url> [options]
```

## Quick Start

```bash
# 1. Test server connectivity and see all capabilities
./mcp-probe -url http://localhost:8000/sse

# 2. List available tools
./mcp-probe -url http://localhost:8000/sse -list-only

# 3. Call a specific tool
./mcp-probe -url http://localhost:8000/sse -call "echo" -params '{"message":"Hello MCP!"}'

# 4. Interactive exploration
./mcp-probe -url http://localhost:8000/sse -interactive
```

## Usage Modes

MCPProbe operates in four distinct modes:

### 1. Discovery Mode (Default)
Tests the MCP server and reports all capabilities (tools, resources, prompts).
```bash
./mcp-probe -url <server-url>
```

### 2. List-Only Mode
Quickly lists available tools without running full capability tests.
```bash
./mcp-probe -url <server-url> -list-only
```

### 3. Direct Tool Calling Mode
Executes a specific tool with provided parameters.
```bash
./mcp-probe -url <server-url> -call <tool-name> -params '<json>'
```

### 4. Interactive Mode
Provides a guided interface for exploring and testing tools.
```bash
./mcp-probe -url <server-url> -interactive
```

## Command-Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `-url` | MCP server URL (required) | - |
| `-transport` | Transport mode: 'sse' or 'http' | `sse` |
| `-call` | Name of the tool to call | - |
| `-params` | JSON string of parameters for tool call | `{}` |
| `-list-only` | Only list available tools | `false` |
| `-interactive` | Enable interactive mode | `false` |
| `-headers` | HTTP headers (format: 'key1:value1,key2:value2') | - |
| `-timeout` | Connection timeout | `30s` |
| `-verbose` | Enable verbose output | `true` |

## Detailed Examples

### Basic Server Testing

```bash
# Test an SSE server (default transport)
./mcp-probe -url http://localhost:8000/sse

# Test an HTTP server
./mcp-probe -url http://localhost:8000/mcp -transport http

# Test with authentication headers
./mcp-probe -url http://api.example.com/mcp \
  -headers "Authorization:Bearer token123,X-API-Key:secret"

# Test with extended timeout for slow servers
./mcp-probe -url http://localhost:8000/sse -timeout 60s
```

### Tool Discovery

```bash
# List all available tools with descriptions
./mcp-probe -url http://localhost:8000/sse -list-only

# List tools with full schema information
./mcp-probe -url http://localhost:8000/sse -list-only -verbose
```

### Direct Tool Calling

```bash
# Call a tool without parameters
./mcp-probe -url http://localhost:8000/sse -call "get_current_time"

# Call with simple parameters
./mcp-probe -url http://localhost:8000/sse \
  -call "echo" \
  -params '{"message":"Hello, MCP!"}'

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
```

### Interactive Mode

```bash
# Start interactive mode
./mcp-probe -url http://localhost:8000/sse -interactive
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
✅ Tool call succeeded:

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
✅ Tool call succeeded:

The result of 5 + 3 is 8
```

### Error Output
```
❌ Failed to call tool 'nonexistent':
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
# Error: context deadline exceeded
# Solution: Increase timeout for slow servers
./mcp-probe -url http://localhost:8000/sse -timeout 60s
```

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

Copyright (c) 2025 by Tenebris Technologies Inc. This software is licensed under the MIT License. Please see LICENSE for details.

## No Warranty (zilch, none, void, nil, null, "", {}, 0x00, 0b00000000, EOF)

THIS SOFTWARE IS PROVIDED “AS IS,” WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, AND NON-INFRINGEMENT. IN NO EVENT SHALL THE COPYRIGHT HOLDERS OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

Made in Canada

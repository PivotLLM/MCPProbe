# MCPTest

A Go application for testing MCP (Model Context Protocol) servers. This tool connects to MCP servers using either HTTP or SSE transport modes, performs the initialization handshake, and reports on the server's capabilities including tools, resources, and prompts.

## Features

- **Multiple Transport Modes**: Supports both HTTP and SSE (Server-Sent Events) connections
- **Initialization Handshake**: Performs proper MCP protocol initialization
- **Capability Discovery**: Reports on server capabilities (tools, resources, prompts)
- **Detailed Listing**: Lists all available tools, resources, and prompts with descriptions
- **Verbose Output**: Provides detailed information about the connection and server responses
- **Custom Headers**: Supports custom HTTP headers for authentication or other purposes

## Installation

```bash
# Clone or download the source code
git clonehttps://github.com/PivotLLM/MCPProbe.git

# Build the application
go build -o mcp-probe
```

## Usage

```bash
./mcp-probe -url <server-url> [options]
```

### Options

- `-url string`: MCP server URL (required)
- `-transport string`: Transport mode: 'sse' or 'http' (default "sse")
- `-headers string`: HTTP headers in format 'key1:value1,key2:value2'
- `-timeout duration`: Connection timeout (default 30s)
- `-verbose`: Enable verbose output (default true)

### Examples

#### Test an SSE MCP Server
```bash
./mcp-probe -url http://localhost:8000/sse -transport sse
```

#### Test an HTTP MCP Server with Custom Headers
```bash
./mcp-probe -url "http://localhost:8000/mcp" -transport http -headers "Authorization:Bearer token123,Content-Type:application/json"
```

#### Test with Custom Timeout
```bash
./mcp-probe -url "http://localhost:8000/sse" -timeout 60s
```

## Output

The tool provides comprehensive output including:

1. **Connection Information**: Server URL, transport mode, timeout settings
2. **Initialization Results**: Protocol version, server info, capabilities
3. **Tools Listing**: Available tools with descriptions and input schemas
4. **Resources Listing**: Available resources and resource templates
5. **Prompts Listing**: Available prompts with arguments and descriptions

## Dependencies

- [github.com/mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) - Go implementation of the Model Context Protocol

## Copyright and license

Copyright (c) 2025 by Tenebris Technologies Inc. This software is licensed under the MIT License. Please see LICENSE for details.

## No Warranty (zilch, none, void, nil, null, "", {}, 0x00, 0b00000000, EOF)

THIS SOFTWARE IS PROVIDED “AS IS,” WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, AND NON-INFRINGEMENT. IN NO EVENT SHALL THE COPYRIGHT HOLDERS OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

Made in Canada
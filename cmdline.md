# Command-Line MCP Tool Testing Implementation Plan

## Overview

This plan outlines the implementation of command-line based tool calling functionality for MCPProbe, enabling users to invoke specific MCP tools with parameters and display the results.

## Current State Analysis

### Existing Functionality
- **Tool Discovery**: The application already lists available tools via `testTools()` function (`main.go:242`)
- **MCP Client**: Uses `github.com/mark3labs/mcp-go` library with established connection patterns
- **Transport Support**: Both SSE and HTTP transport modes are supported
- **Tool Schema Display**: Shows tool descriptions and input schemas in verbose mode

### Limitations
- No ability to actually invoke/call tools
- No parameter input mechanism
- No tool result display functionality

## Implementation Plan

### Phase 1: Command-Line Interface Design

#### New Command-Line Flags
Add the following flags to the existing flag set:

```go
var (
    // Existing flags...
    callTool   = flag.String("call", "", "Name of the tool to call")
    toolParams = flag.String("params", "{}", "JSON string of parameters for the tool call")
    listOnly   = flag.Bool("list-only", false, "Only list available tools, don't test capabilities")
    interactive = flag.Bool("interactive", false, "Interactive mode for tool calling")
)
```

#### Usage Examples
```bash
# List available tools only
./mcp-probe -url http://localhost:8000/sse -list-only

# Call a specific tool with JSON parameters
./mcp-probe -url http://localhost:8000/sse -call "calculate" -params '{"operation":"add","x":5,"y":3}'

# Interactive mode for multiple tool calls
./mcp-probe -url http://localhost:8000/sse -interactive
```

### Phase 2: Core Implementation Architecture

#### 2.1 New Function: `callSpecificTool()`

```go
func callSpecificTool(ctx context.Context, mcpClient *client.Client, toolName string, paramsJSON string, verbose bool) error {
    // 1. Parse JSON parameters
    // 2. Create CallToolRequest
    // 3. Invoke mcpClient.CallTool()
    // 4. Display formatted results
    // 5. Handle errors gracefully
}
```

**Location**: Add after `testPrompts()` function (~line 387)

#### 2.2 Enhanced Main Function Logic

Modify the main function to support different execution modes:

```go
// After client initialization and before testServerCapabilities
if *listOnly {
    return listToolsOnly(ctx, mcpClient, *verbose)
}

if *callTool != "" {
    return callSpecificTool(ctx, mcpClient, *callTool, *toolParams, *verbose)
}

if *interactive {
    return interactiveMode(ctx, mcpClient, *verbose)
}

// Default behavior: existing capability testing
```

#### 2.3 New Function: `listToolsOnly()`

```go
func listToolsOnly(ctx context.Context, mcpClient *client.Client, verbose bool) error {
    // Simplified version of existing testTools() for list-only mode
    // Enhanced output formatting for better readability
}
```

#### 2.4 New Function: `interactiveMode()`

```go
func interactiveMode(ctx context.Context, mcpClient *client.Client, verbose bool) error {
    // 1. List available tools
    // 2. Prompt user for tool selection
    // 3. Prompt for parameters
    // 4. Execute tool call
    // 5. Display results
    // 6. Repeat until user exits
}
```

### Phase 3: Parameter Handling

#### 3.1 JSON Parameter Parsing

```go
func parseToolParameters(paramsJSON string) (map[string]interface{}, error) {
    var params map[string]interface{}
    if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
        return nil, fmt.Errorf("failed to parse parameters JSON: %w", err)
    }
    return params, nil
}
```

#### 3.2 Parameter Validation

```go
func validateToolParameters(tool mcp.Tool, params map[string]interface{}) error {
    // Validate required parameters exist
    // Type checking against tool schema
    // Return descriptive validation errors
}
```

### Phase 4: Result Display and Formatting

#### 4.1 Tool Call Result Formatting

```go
func formatToolResult(result *mcp.CallToolResult, verbose bool) {
    fmt.Println("\n=== Tool Call Result ===")
    
    if result.IsError {
        fmt.Printf("❌ Tool call failed:\n")
    } else {
        fmt.Printf("✅ Tool call succeeded:\n")
    }
    
    // Display content with proper formatting
    for i, content := range result.Content {
        fmt.Printf("Content %d:\n", i+1)
        // Handle different content types (text, structured, etc.)
    }
    
    // Verbose mode: show structured content, metadata
    if verbose && result.StructuredContent != nil {
        fmt.Printf("\nStructured Content:\n")
        // Pretty-print structured content
    }
}
```

#### 4.2 Request Display (for debugging)

```go
func displayToolRequest(toolName string, params map[string]interface{}, verbose bool) {
    if !verbose {
        return
    }
    
    fmt.Printf("\n=== Sending Tool Call ===\n")
    fmt.Printf("Tool: %s\n", toolName)
    fmt.Printf("Parameters:\n")
    
    for key, value := range params {
        fmt.Printf("  %s: %v (%T)\n", key, value, value)
    }
    fmt.Println()
}
```

### Phase 5: Interactive Mode Implementation

#### 5.1 Tool Selection Interface

```go
func selectTool(tools []mcp.Tool) (*mcp.Tool, error) {
    fmt.Println("\nAvailable Tools:")
    for i, tool := range tools {
        fmt.Printf("%d. %s", i+1, tool.Name)
        if tool.Description != "" {
            fmt.Printf(" - %s", tool.Description)
        }
        fmt.Println()
    }
    
    // Prompt for selection
    // Return selected tool
}
```

#### 5.2 Parameter Input Interface

```go
func collectToolParameters(tool mcp.Tool) (map[string]interface{}, error) {
    params := make(map[string]interface{})
    
    // Extract required parameters from tool schema
    // Prompt user for each parameter with type guidance
    // Validate input types
    // Return parameter map
}
```

### Phase 6: Error Handling and User Experience

#### 6.1 Comprehensive Error Messages

```go
func handleToolCallError(err error, toolName string) {
    fmt.Printf("❌ Failed to call tool '%s':\n", toolName)
    
    // Categorize error types
    switch {
    case strings.Contains(err.Error(), "not found"):
        fmt.Printf("   Tool '%s' not found. Use -list-only to see available tools.\n", toolName)
    case strings.Contains(err.Error(), "parameter"):
        fmt.Printf("   Parameter error: %v\n", err)
        fmt.Printf("   Check parameter format and required fields.\n")
    default:
        fmt.Printf("   %v\n", err)
    }
}
```

#### 6.2 Input Validation

```go
func validateInputs(toolName, paramsJSON string) error {
    if toolName == "" {
        return fmt.Errorf("tool name cannot be empty")
    }
    
    if paramsJSON != "" && paramsJSON != "{}" {
        var temp interface{}
        if err := json.Unmarshal([]byte(paramsJSON), &temp); err != nil {
            return fmt.Errorf("invalid JSON parameters: %w", err)
        }
    }
    
    return nil
}
```

### Phase 7: Integration Points

#### 7.1 Modified Main Function Structure

```go
func main() {
    // Existing flag parsing and validation...
    
    // New validation for tool calling flags
    if err := validateInputs(*callTool, *toolParams); err != nil {
        log.Fatalf("Input validation failed: %v", err)
    }
    
    // Existing client creation and initialization...
    
    // New execution flow branching
    switch {
    case *listOnly:
        if err := listToolsOnly(ctx, mcpClient, *verbose); err != nil {
            log.Fatalf("Failed to list tools: %v", err)
        }
    case *callTool != "":
        if err := callSpecificTool(ctx, mcpClient, *callTool, *toolParams, *verbose); err != nil {
            log.Fatalf("Tool call failed: %v", err)
        }
    case *interactive:
        if err := interactiveMode(ctx, mcpClient, *verbose); err != nil {
            log.Fatalf("Interactive mode failed: %v", err)
        }
    default:
        // Existing default behavior (capability testing)
        if err := testServerCapabilities(ctx, mcpClient, *verbose); err != nil {
            log.Fatalf("Failed to test capabilities: %v", err)
        }
    }
}
```

#### 7.2 Help Text Updates

Update the usage message to include new functionality:

```go
if *serverURL == "" {
    fmt.Println("Error: Server URL is required")
    fmt.Println("Usage:")
    fmt.Println("  Test MCP server capabilities:")
    fmt.Println("    go run main.go -url <server-url> [-transport sse|http] [-timeout 30s]")
    fmt.Println("  List available tools only:")
    fmt.Println("    go run main.go -url <server-url> -list-only")
    fmt.Println("  Call a specific tool:")
    fmt.Println("    go run main.go -url <server-url> -call <tool-name> -params '<json>'")
    fmt.Println("  Interactive tool calling:")
    fmt.Println("    go run main.go -url <server-url> -interactive")
    os.Exit(1)
}
```

## Implementation Sequence

### Step 1: Basic Tool Calling (Estimated: 2-4 hours)
1. Add new command-line flags
2. Implement `callSpecificTool()` function
3. Add JSON parameter parsing
4. Basic result display
5. Update main function logic

### Step 2: Enhanced UX (Estimated: 1-2 hours)
1. Implement `listToolsOnly()` function  
2. Add comprehensive error handling
3. Improve result formatting
4. Add request display for verbose mode

### Step 3: Interactive Mode (Estimated: 3-5 hours)
1. Implement tool selection interface
2. Add parameter collection interface
3. Implement interactive loop
4. Add exit mechanisms

### Step 4: Testing & Polish (Estimated: 1-2 hours)
1. Test with various MCP servers
2. Validate error scenarios
3. Refine user interface
4. Update documentation

## Dependencies

### Required Imports
```go
import (
    "bufio"      // For interactive input
    "encoding/json" // For parameter parsing
    "strconv"    // For type conversions
    // ... existing imports
)
```

### No Additional External Dependencies
The implementation uses only the existing `github.com/mark3labs/mcp-go` library and Go standard library.

## Backward Compatibility

- All existing functionality remains unchanged
- Default behavior (no new flags) maintains current capability testing
- New features are opt-in via command-line flags
- Existing command-line flags continue to work as expected

## Future Enhancements

1. **Configuration Files**: Support JSON/YAML config files for complex tool calls
2. **Batch Mode**: Execute multiple tool calls from a file
3. **Output Formats**: JSON, XML, or CSV output options
4. **Tool Call History**: Save and replay previous tool calls
5. **Schema Validation**: More sophisticated parameter validation
6. **Auto-completion**: Shell completion for tool names and parameters

This implementation plan provides a comprehensive approach to adding command-line tool calling functionality while maintaining the existing architecture and user experience.
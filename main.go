// Copyright (c) 2025 Tenebris Technologies Inc.
// This software is licensed under the MIT License (see LICENSE for details).

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	// Command line flags
	var (
		serverURL   = flag.String("url", "", "MCP server URL (required)")
		mode        = flag.String("transport", "sse", "Transport mode: 'sse' or 'http'")
		headers     = flag.String("headers", "", "HTTP headers in format 'key1:value1,key2:value2'")
		timeout     = flag.Duration("timeout", 30*time.Second, "Connection timeout for initialization and listing")
		callTimeout = flag.Duration("call-timeout", 300*time.Second, "Timeout for tool call execution")
		verbose     = flag.Bool("verbose", true, "Enable verbose output")
		callTool    = flag.String("call", "", "Name of the tool to call")
		toolParams  = flag.String("params", "{}", "JSON string of parameters for the tool call")
		listOnly    = flag.Bool("list-only", false, "Only list available tools, don't test capabilities")
		interactive = flag.Bool("interactive", false, "Interactive mode for tool calling")
	)
	flag.Parse()

	if *serverURL == "" {
		fmt.Println("Error: Server URL is required")
		fmt.Println("Usage:")
		fmt.Println("  Test MCP server capabilities:")
		fmt.Println("    go run main.go -url <server-url> [-transport sse|http] [-timeout 30s]")
		fmt.Println("  List available tools only:")
		fmt.Println("    go run main.go -url <server-url> -list-only")
		fmt.Println("  Call a specific tool:")
		fmt.Println("    go run main.go -url <server-url> -call <tool-name> -params '<json>' [-call-timeout 300s]")
		fmt.Println("  Interactive tool calling:")
		fmt.Println("    go run main.go -url <server-url> -interactive [-call-timeout 300s]")
		fmt.Println("\nTimeout options:")
		fmt.Println("  -timeout:      Connection/initialization timeout (default: 30s)")
		fmt.Println("  -call-timeout: Tool execution timeout (default: 300s)")
		os.Exit(1)
	}

	// Validate tool calling inputs
	if err := validateInputs(*callTool, *toolParams); err != nil {
		log.Fatalf("Input validation failed: %v", err)
	}

	fmt.Printf("=== MCP Server Test Tool ===\n")
	fmt.Printf("Server URL: %s\n", *serverURL)

	fmt.Printf("Transport: %s\n", *mode)
	fmt.Printf("Timeout: %s\n", *timeout)
	fmt.Println()

	// Parse headers
	headerMap := parseHeaders(*headers)
	if len(headerMap) > 0 && *verbose {
		fmt.Printf("Headers: %v\n", headerMap)
	}

	// Create client based on transport type
	var mcpClient *client.Client
	var err error

	switch strings.ToLower(*mode) {
	case "sse":
		fmt.Println("Creating SSE client...")
		mcpClient, err = createSSEClient(*serverURL, headerMap, *callTimeout)
	case "http":
		fmt.Println("Creating HTTP client...")
		mcpClient, err = createHTTPClient(*serverURL, headerMap, *callTimeout)
	default:
		fmt.Printf("Error: Unsupported transport type '%s'. Use 'sse' or 'http'\n", *mode)
		os.Exit(1)
	}

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer func(mcpClient *client.Client) {
		_ = mcpClient.Close()
	}(mcpClient)

	// Start the client connection with background context
	// The SSE/HTTP stream needs to stay alive for the duration of tool calls
	fmt.Println("Starting client connection...")
	if err := mcpClient.Start(context.Background()); err != nil {
		log.Fatalf("Failed to start client: %v", err)
	}
	fmt.Println("Client connection started successfully")

	// Display POST URL for SSE connections
	if strings.ToLower(*mode) == "sse" {
		if sseTransport, ok := mcpClient.GetTransport().(*transport.SSE); ok {
			endpoint := sseTransport.GetEndpoint()
			if endpoint != nil {
				fmt.Printf("SSE POST URL: %s\n", endpoint.String())
			}
		}
	}

	// Perform initialization handshake with timeout
	fmt.Println("\nPerforming initialization handshake...")
	initCtx, initCancel := context.WithTimeout(context.Background(), *timeout)
	defer initCancel()
	if err := performInitialization(initCtx, mcpClient, *verbose); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	fmt.Println("\nInitialization completed successfully")

	// Handle different execution modes with appropriate context management
	switch {
	case *listOnly:
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		defer cancel()
		if err := listToolsOnly(ctx, mcpClient, *verbose); err != nil {
			log.Fatalf("Failed to list tools: %v", err)
		}
	case *callTool != "":
		ctx, cancel := context.WithTimeout(context.Background(), *callTimeout)
		defer cancel()
		if err := callSpecificTool(ctx, mcpClient, *callTool, *toolParams, *verbose); err != nil {
			handleToolCallError(err, *callTool)
			os.Exit(1)
		}
	case *interactive:
		// Interactive mode manages its own contexts for each tool call
		// Connection uses background context to stay alive indefinitely
		if err := interactiveModeWithTimeout(mcpClient, *callTimeout, *verbose); err != nil {
			log.Fatalf("Interactive mode failed: %v", err)
		}
	default:
		// Default behavior: test server capabilities
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		defer cancel()
		if err := testServerCapabilities(ctx, mcpClient, *verbose); err != nil {
			log.Fatalf("Failed to test capabilities: %v", err)
		}
	}

	fmt.Println("\n=== Finished ===")
}

func parseHeaders(headerStr string) map[string]string {
	headers := make(map[string]string)
	if headerStr == "" {
		return headers
	}

	pairs := strings.Split(headerStr, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return headers
}

func createSSEClient(serverURL string, headers map[string]string, callTimeout time.Duration) (*client.Client, error) {
	// Create custom HTTP client with appropriate timeout for long-running tool calls
	// Add buffer to account for network overhead
	httpClient := &http.Client{
		Timeout: callTimeout + (30 * time.Second),
	}

	var options []transport.ClientOption
	options = append(options, transport.WithHTTPClient(httpClient))
	if len(headers) > 0 {
		options = append(options, client.WithHeaders(headers))
	}
	return client.NewSSEMCPClient(serverURL, options...)
}

func createHTTPClient(serverURL string, headers map[string]string, callTimeout time.Duration) (*client.Client, error) {
	var options []transport.StreamableHTTPCOption
	// Set HTTP timeout for tool call execution
	options = append(options, transport.WithHTTPTimeout(callTimeout))
	if len(headers) > 0 {
		options = append(options, transport.WithHTTPHeaders(headers))
	}
	return client.NewStreamableHttpClient(serverURL, options...)
}

func performInitialization(ctx context.Context, mcpClient *client.Client, verbose bool) error {
	// Create initialization request
	initRequest := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities: mcp.ClientCapabilities{
				Roots: &struct {
					ListChanged bool `json:"listChanged,omitempty"`
				}{
					ListChanged: true,
				},
				Sampling: &struct{}{},
			},
			ClientInfo: mcp.Implementation{
				Name:    "MCPTest",
				Version: "1.0.0",
			},
		},
	}

	if verbose {
		fmt.Printf("Sending initialization request with protocol version: %s\n", initRequest.Params.ProtocolVersion)
		fmt.Printf("Client info: %s v%s\n", initRequest.Params.ClientInfo.Name, initRequest.Params.ClientInfo.Version)
	}

	// Send initialization request
	initResult, err := mcpClient.Initialize(ctx, initRequest)
	if err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	if verbose {
		fmt.Printf("Server info: %s v%s\n", initResult.ServerInfo.Name, initResult.ServerInfo.Version)
		fmt.Printf("Protocol version: %s\n", initResult.ProtocolVersion)
		fmt.Printf("\nServer capabilities received:\n")
		printServerCapabilities(initResult.Capabilities)
	}

	return nil
}

func printServerCapabilities(caps mcp.ServerCapabilities) {
	if caps.Logging != nil {
		fmt.Printf("  - Logging: supported\n")
	}
	if caps.Prompts != nil {
		fmt.Printf("  - Prompts: supported (list_changed: %t)\n", caps.Prompts.ListChanged)
	}
	if caps.Resources != nil {
		fmt.Printf("  - Resources: supported (subscribe: %t, list_changed: %t)\n",
			caps.Resources.Subscribe, caps.Resources.ListChanged)
	}
	if caps.Tools != nil {
		fmt.Printf("  - Tools: supported (list_changed: %t)\n", caps.Tools.ListChanged)
	}
	if caps.Experimental != nil && len(caps.Experimental) > 0 {
		fmt.Printf("  - Experimental capabilities: %v\n", caps.Experimental)
	}
}

func testServerCapabilities(ctx context.Context, mcpClient *client.Client, verbose bool) error {

	// Get server capabilities
	serverCaps := mcpClient.GetServerCapabilities()

	// Test Tools capability
	fmt.Println("\n--- Tools Capability ---")
	if serverCaps.Tools != nil {
		if err := testTools(ctx, mcpClient, verbose); err != nil {
			fmt.Printf("Warning: Tools test failed: %v\n", err)
		}
	} else {

		fmt.Println("Tools capability not supported by server")
	}

	// Test Resources capability
	if serverCaps.Resources != nil {
		fmt.Println("--- Testing Resources Capability ---")
		if err := testResources(ctx, mcpClient, verbose); err != nil {
			fmt.Printf("Warning: Resources test failed: %v\n", err)
		}
	} else {
		fmt.Println("--- Resources Capability ---")
		fmt.Println("Resources capability not supported by server")
	}

	// Test Prompts capability
	if serverCaps.Prompts != nil {
		fmt.Println("--- Testing Prompts Capability ---")
		if err := testPrompts(ctx, mcpClient, verbose); err != nil {
			fmt.Printf("Warning: Prompts test failed: %v\n", err)
		}
	} else {
		fmt.Println("\n--- Prompts Capability ---")
		fmt.Println("Prompts capability not supported by server")
	}

	return nil
}

func formatToolInputSchema(schema mcp.ToolInputSchema, indent string) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("%sType: %s\n", indent, schema.Type))

	if len(schema.Required) > 0 {
		result.WriteString(fmt.Sprintf("%sRequired: %v\n", indent, schema.Required))
	} else {
		result.WriteString(fmt.Sprintf("%sRequired: (none)\n", indent))
	}

	if len(schema.Properties) > 0 {
		result.WriteString(fmt.Sprintf("%sProperties:\n", indent))
		for propName, propValue := range schema.Properties {
			result.WriteString(fmt.Sprintf("%s  - %s: ", indent, propName))

			// Pretty print the property value
			if propMap, ok := propValue.(map[string]interface{}); ok {
				// It's a property definition object
				if propType, hasType := propMap["type"]; hasType {
					result.WriteString(fmt.Sprintf("(type: %v", propType))
					if desc, hasDesc := propMap["description"]; hasDesc {
						result.WriteString(fmt.Sprintf(", description: %v", desc))
					}
					if enum, hasEnum := propMap["enum"]; hasEnum {
						result.WriteString(fmt.Sprintf(", enum: %v", enum))
					}
					if def, hasDef := propMap["default"]; hasDef {
						result.WriteString(fmt.Sprintf(", default: %v", def))
					}
					result.WriteString(")")
				} else {
					// Fallback to JSON representation
					jsonBytes, _ := json.MarshalIndent(propValue, "", "  ")
					result.WriteString(string(jsonBytes))
				}
			} else {
				// Simple value
				result.WriteString(fmt.Sprintf("%v", propValue))
			}
			result.WriteString("\n")
		}
	}

	if len(schema.Defs) > 0 {
		result.WriteString(fmt.Sprintf("%sDefinitions:\n", indent))
		for defName, defValue := range schema.Defs {
			result.WriteString(fmt.Sprintf("%s  - %s: ", indent, defName))
			jsonBytes, _ := json.MarshalIndent(defValue, indent+"    ", "  ")
			result.WriteString(string(jsonBytes))
			result.WriteString("\n")
		}
	}

	return result.String()
}

func testTools(ctx context.Context, mcpClient *client.Client, verbose bool) error {
	fmt.Println("Requesting list of available tools...")

	toolsRequest := mcp.ListToolsRequest{}
	toolsResult, err := mcpClient.ListTools(ctx, toolsRequest)
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	fmt.Printf("Found %d tools:\n\n", len(toolsResult.Tools))

	for i, tool := range toolsResult.Tools {
		fmt.Printf("  %d. %s\n", i+1, tool.Name)
		if verbose {
			if tool.Description != "" {
				fmt.Printf("     Description: %s\n", tool.Description)
			}
			fmt.Println("     Input Schema:")
			schemaOutput := formatToolInputSchema(tool.InputSchema, "       ")
			fmt.Print(schemaOutput)
			fmt.Println()
		}
	}

	if len(toolsResult.Tools) == 0 {
		fmt.Println("  (No tools available)\n")
	}

	return nil
}

func testResources(ctx context.Context, mcpClient *client.Client, verbose bool) error {
	fmt.Println("Requesting list of available resources...")

	resourcesRequest := mcp.ListResourcesRequest{}
	resourcesResult, err := mcpClient.ListResources(ctx, resourcesRequest)
	if err != nil {
		return fmt.Errorf("failed to list resources: %w", err)
	}

	fmt.Printf("Found %d resources:\n\n", len(resourcesResult.Resources))

	for i, resource := range resourcesResult.Resources {
		fmt.Printf("  %d. %s\n", i+1, resource.URI)
		if verbose {
			if resource.Name != "" {
				fmt.Printf("     Name: %s\n", resource.Name)
			}
			if resource.Description != "" {
				fmt.Printf("     Description: %s\n", resource.Description)
			}
			if resource.MIMEType != "" {
				fmt.Printf("     MIME Type: %s\n\n", resource.MIMEType)
			}
		}
	}

	if len(resourcesResult.Resources) == 0 {
		fmt.Println("  (No resources available)\n")
	}

	// Also test resource templates if available
	fmt.Println("Requesting list of available resource templates...")
	templatesRequest := mcp.ListResourceTemplatesRequest{}
	templatesResult, err := mcpClient.ListResourceTemplates(ctx, templatesRequest)
	if err != nil {
		fmt.Printf("Warning: Failed to list resource templates: %v\n", err)
		return nil
	}

	fmt.Printf("Found %d resource templates:\n\n", len(templatesResult.ResourceTemplates))

	for i, template := range templatesResult.ResourceTemplates {
		// Access the underlying template pattern using the template's MarshalJSON method
		var templateStr string
		if template.URITemplate != nil {
			// Use the template's MarshalJSON method
			jsonBytes, err := template.URITemplate.MarshalJSON()
			if err == nil {
				// Remove quotes from the JSON string
				templateStr = strings.Trim(string(jsonBytes), "\"")
			} else {
				templateStr = fmt.Sprintf("(Error marshaling template: %v)", err)
			}
		} else {
			templateStr = "(empty template)"
		}

		fmt.Printf("  %d. %s\n", i+1, templateStr)
		if verbose {
			if template.Name != "" {
				fmt.Printf("     Name: %s\n", template.Name)
			}
			if template.Description != "" {
				fmt.Printf("     Description: %s\n", template.Description)
			}
			if template.MIMEType != "" {
				fmt.Printf("     MIME Type: %s\n\n", template.MIMEType)
			}
		}
	}

	if len(templatesResult.ResourceTemplates) == 0 {
		fmt.Println("  (No resource templates available)\n")
	}

	return nil
}

func testPrompts(ctx context.Context, mcpClient *client.Client, verbose bool) error {
	fmt.Println("Requesting list of available prompts...")

	promptsRequest := mcp.ListPromptsRequest{}
	promptsResult, err := mcpClient.ListPrompts(ctx, promptsRequest)
	if err != nil {
		return fmt.Errorf("failed to list prompts: %w", err)
	}

	fmt.Printf("Found %d prompts:\n\n", len(promptsResult.Prompts))

	for i, prompt := range promptsResult.Prompts {
		fmt.Printf("  %d. %s\n", i+1, prompt.Name)
		if verbose {
			if prompt.Description != "" {
				fmt.Printf("     Description: %s\n", prompt.Description)
			}
			if len(prompt.Arguments) > 0 {
				fmt.Printf("     Arguments:\n")
				for _, arg := range prompt.Arguments {
					fmt.Printf("       - %s", arg.Name)
					if arg.Description != "" {
						fmt.Printf(": %s", arg.Description)
					}
					if arg.Required {
						fmt.Printf(" (required)")
					}
					fmt.Println("\n")
				}
			}
		}
	}

	if len(promptsResult.Prompts) == 0 {
		fmt.Println("  (No prompts available)\n")
	}

	return nil
}

// validateInputs validates command line inputs for tool calling
func validateInputs(toolName, paramsJSON string) error {
	if toolName != "" && paramsJSON != "" && paramsJSON != "{}" {
		var temp interface{}
		if err := json.Unmarshal([]byte(paramsJSON), &temp); err != nil {
			return fmt.Errorf("invalid JSON parameters: %w", err)
		}
	}
	return nil
}

// callSpecificTool calls a specific tool with the given parameters
func callSpecificTool(ctx context.Context, mcpClient *client.Client, toolName string, paramsJSON string, verbose bool) error {
	// Parse JSON parameters
	params, err := parseToolParameters(paramsJSON)
	if err != nil {
		return err
	}

	// Display request in verbose mode
	displayToolRequest(toolName, params, verbose)

	// Create the tool call request
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: params,
		},
	}

	// Call the tool
	fmt.Printf("Calling tool '%s'...\n", toolName)
	result, err := mcpClient.CallTool(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to call tool: %w", err)
	}

	// Format and display the result
	formatToolResult(result, verbose)

	return nil
}

// parseToolParameters parses JSON parameters for tool calls
func parseToolParameters(paramsJSON string) (map[string]interface{}, error) {
	var params map[string]interface{}
	if paramsJSON == "" || paramsJSON == "{}" {
		return make(map[string]interface{}), nil
	}

	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return nil, fmt.Errorf("failed to parse parameters JSON: %w", err)
	}
	return params, nil
}

// displayToolRequest displays the tool request in verbose mode
func displayToolRequest(toolName string, params map[string]interface{}, verbose bool) {
	if !verbose {
		return
	}

	fmt.Printf("\n=== Sending Tool Call ===\n")
	fmt.Printf("Tool: %s\n", toolName)
	if len(params) > 0 {
		fmt.Printf("Parameters:\n")
		for key, value := range params {
			fmt.Printf("  %s: %v (%T)\n", key, value, value)
		}
	} else {
		fmt.Printf("Parameters: (none)\n")
	}
	fmt.Println()
}

// formatToolResult formats and displays the tool call result
func formatToolResult(result *mcp.CallToolResult, verbose bool) {
	fmt.Println("\n=== Tool Call Result ===")

	if result.IsError {
		fmt.Printf("Tool call failed:\n")
	} else {
		fmt.Printf("Tool call succeeded:\n")
	}

	// Display content
	if len(result.Content) > 0 {
		for i, content := range result.Content {
			if len(result.Content) > 1 {
				fmt.Printf("\nContent %d:\n", i+1)
			} else {
				fmt.Printf("\n")
			}

			// Handle different content types using type assertion
			switch c := content.(type) {
			case mcp.TextContent:
				fmt.Printf("%s\n", c.Text)
			case mcp.ImageContent:
				if verbose {
					fmt.Printf("Image (MIME: %s)\n", c.MIMEType)
				}
			case mcp.AudioContent:
				if verbose {
					fmt.Printf("Audio (MIME: %s)\n", c.MIMEType)
				}
			default:
				if verbose {
					fmt.Printf("Unknown content type: %T\n", c)
				}
			}
		}
	}

	// Note: StructuredContent field doesn't exist in the current mcp-go version
	// This functionality may be added in future versions
}

// handleToolCallError handles errors from tool calls with user-friendly messages
func handleToolCallError(err error, toolName string) {
	fmt.Printf("Failed to call tool '%s':\n", toolName)

	// Categorize error types
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "not found"):
		fmt.Printf("   Tool '%s' not found. Use -list-only to see available tools.\n", toolName)
	case strings.Contains(errStr, "parameter") && strings.Contains(errStr, "required"):
		fmt.Printf("   Parameter validation error: %v\n", err)
		fmt.Printf("   The server requires parameters that weren't provided.\n")
		fmt.Printf("   💡 This may indicate the tool schema doesn't correctly mark required parameters.\n")
		fmt.Printf("   💡 Try calling the tool again and provide values for parameters that seem required.\n")
	case strings.Contains(errStr, "parameter"):
		fmt.Printf("   Parameter error: %v\n", err)
		fmt.Printf("   Check parameter format and required fields.\n")
	case strings.Contains(errStr, "timeout"):
		fmt.Printf("   Request timed out. Try increasing the timeout with -timeout flag.\n")
	case strings.Contains(errStr, "Invalid session ID"):
		fmt.Printf("   Session expired. Please restart MCPProbe.\n")
	default:
		fmt.Printf("   %v\n", err)
	}
}

// listToolsOnly lists available tools without running full capability tests
func listToolsOnly(ctx context.Context, mcpClient *client.Client, verbose bool) error {
	fmt.Println("\n--- Available Tools ---")

	// Check if tools capability is supported
	serverCaps := mcpClient.GetServerCapabilities()
	if serverCaps.Tools == nil {
		fmt.Println("Tools capability not supported by server")
		return nil
	}

	fmt.Println("Requesting list of available tools...")

	toolsRequest := mcp.ListToolsRequest{}
	toolsResult, err := mcpClient.ListTools(ctx, toolsRequest)
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	fmt.Printf("\nFound %d tools:\n\n", len(toolsResult.Tools))

	for i, tool := range toolsResult.Tools {
		fmt.Printf("%d. %s", i+1, tool.Name)
		if tool.Description != "" && verbose {
			fmt.Printf(" - %s", tool.Description)
		}
		fmt.Println()

		if verbose {
			// Pretty print the input schema
			schemaJSON, err := json.MarshalIndent(tool.InputSchema, "   ", "  ")
			if err == nil && string(schemaJSON) != "{}" && string(schemaJSON) != "null" {
				fmt.Printf("   Input Schema:\n")
				lines := strings.Split(string(schemaJSON), "\n")
				for _, line := range lines {
					fmt.Printf("   %s\n", line)
				}

				fmt.Println()
			}
		}
	}

	if len(toolsResult.Tools) == 0 {
		fmt.Println("  (No tools available)")
	}

	return nil
}

// interactiveModeWithTimeout provides an interactive interface for tool calling with timeout management
func interactiveModeWithTimeout(mcpClient *client.Client, timeout time.Duration, verbose bool) error {
	fmt.Println("\n=== Interactive Tool Calling Mode ===")
	fmt.Println("Type 'help' for commands, 'exit' to quit")

	// Check if tools capability is supported
	serverCaps := mcpClient.GetServerCapabilities()
	if serverCaps.Tools == nil {
		fmt.Println("Tools capability not supported by server")
		return nil
	}

	// Get list of available tools with fresh context
	listCtx, listCancel := context.WithTimeout(context.Background(), timeout)
	defer listCancel()
	toolsRequest := mcp.ListToolsRequest{}
	toolsResult, err := mcpClient.ListTools(listCtx, toolsRequest)
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	if len(toolsResult.Tools) == 0 {
		fmt.Println("No tools available on this server")
		return nil
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Split command and arguments
		parts := strings.Fields(input)
		command := parts[0]
		var args []string
		if len(parts) > 1 {
			args = parts[1:]
		}

		switch command {
		case "exit", "quit", "q":
			fmt.Println("Exiting interactive mode...")
			return nil
		case "help", "h", "?":
			printInteractiveHelp()
		case "list", "ls", "l":
			listToolsInteractive(toolsResult.Tools)
		case "call", "c":
			// Handle "call 3" or "c 3" syntax
			if len(args) > 0 {
				if num, err := strconv.Atoi(args[0]); err == nil && num > 0 && num <= len(toolsResult.Tools) {
					tool := toolsResult.Tools[num-1]
					if err := callToolDirectlyWithTimeout(mcpClient, &tool, scanner, timeout, verbose); err != nil {
						fmt.Printf("Error: %v\n", err)
					}
				} else {
					fmt.Printf("Invalid tool number: %s\n", args[0])
				}
			} else {
				// No arguments, show guided selection
				if err := callToolInteractiveWithTimeout(mcpClient, toolsResult.Tools, scanner, timeout, verbose); err != nil {
					fmt.Printf("Error: %v\n", err)
				}
			}
		default:
			// Try to interpret as a tool number
			if num, err := strconv.Atoi(command); err == nil && num > 0 && num <= len(toolsResult.Tools) {
				tool := toolsResult.Tools[num-1]
				if err := callToolDirectlyWithTimeout(mcpClient, &tool, scanner, timeout, verbose); err != nil {
					fmt.Printf("Error: %v\n", err)
				}
			} else {
				fmt.Printf("Unknown command: %s (type 'help' for commands)\n", command)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}

// printInteractiveHelp prints help for interactive mode
func printInteractiveHelp() {
	fmt.Println("\nAvailable commands:")
	fmt.Println("  list, ls, l     - List available tools")
	fmt.Println("  call, c         - Call a tool (guided selection)")
	fmt.Println("  call 3, c 3     - Call tool number 3 directly")
	fmt.Println("  3               - Call tool number 3 directly")
	fmt.Println("  help, h, ?      - Show this help")
	fmt.Println("  exit, quit, q   - Exit interactive mode")
}

// listToolsInteractive lists tools in interactive mode
func listToolsInteractive(tools []mcp.Tool) {
	fmt.Printf("\nAvailable tools (%d):\n", len(tools))
	for i, tool := range tools {
		fmt.Printf("  %d. %s", i+1, tool.Name)
		if tool.Description != "" {
			fmt.Printf(" - %s", tool.Description)
		}
		fmt.Println()
	}
}

// callToolInteractiveWithTimeout calls a tool in interactive mode with guided selection and timeout management
func callToolInteractiveWithTimeout(mcpClient *client.Client, tools []mcp.Tool, scanner *bufio.Scanner, timeout time.Duration, verbose bool) error {
	// List tools
	listToolsInteractive(tools)

	// Select tool
	fmt.Print("\nEnter tool number (or 'cancel'): ")
	if !scanner.Scan() {
		return nil
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "cancel" || input == "" {
		return nil
	}

	toolNum, err := strconv.Atoi(input)
	if err != nil || toolNum < 1 || toolNum > len(tools) {
		return fmt.Errorf("invalid tool number: %s", input)
	}

	tool := &tools[toolNum-1]
	return callToolDirectlyWithTimeout(mcpClient, tool, scanner, timeout, verbose)
}

// callToolDirectlyWithTimeout calls a specific tool with parameter collection and timeout management
func callToolDirectlyWithTimeout(mcpClient *client.Client, tool *mcp.Tool, scanner *bufio.Scanner, timeout time.Duration, verbose bool) error {
	// Create fresh context for this tool call
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return callToolDirectly(ctx, mcpClient, tool, scanner, verbose)
}

// callToolDirectly calls a specific tool with parameter collection
func callToolDirectly(ctx context.Context, mcpClient *client.Client, tool *mcp.Tool, scanner *bufio.Scanner, verbose bool) error {
	fmt.Printf("\nCalling tool: %s\n", tool.Name)
	if tool.Description != "" {
		fmt.Printf("Description: %s\n", tool.Description)
	}

	// Collect parameters
	params, err := collectToolParameters(tool, scanner)
	if err != nil {
		return err
	}

	// Display request in verbose mode
	displayToolRequest(tool.Name, params, verbose)

	// Create and send the request
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      tool.Name,
			Arguments: params,
		},
	}

	fmt.Printf("\nCalling tool '%s'...\n", tool.Name)
	result, err := mcpClient.CallTool(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to call tool: %w", err)
	}

	// Display result
	formatToolResult(result, verbose)

	return nil
}

// collectToolParameters collects parameters for a tool call interactively
func collectToolParameters(tool *mcp.Tool, scanner *bufio.Scanner) (map[string]interface{}, error) {
	params := make(map[string]interface{})

	// Marshal InputSchema to JSON for parsing
	schemaJSON, err := json.Marshal(tool.InputSchema)
	if err != nil || string(schemaJSON) == "null" || string(schemaJSON) == "{}" {
		// No schema or empty schema means no parameters
		return params, nil
	}

	// Try to parse the schema as a map
	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schemaJSON, &schemaMap); err != nil {
		// If we can't parse the schema, ask for JSON input
		fmt.Println("Enter parameters as JSON (or press Enter for no parameters):")
		if !scanner.Scan() {
			return params, nil
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			return params, nil
		}
		return parseToolParameters(input)
	}

	// Extract properties from schema
	properties, ok := schemaMap["properties"].(map[string]interface{})
	if !ok || len(properties) == 0 {
		fmt.Println("No parameters required for this tool")
		return params, nil
	}

	required := make(map[string]bool)
	if reqArray, ok := schemaMap["required"].([]interface{}); ok {
		for _, req := range reqArray {
			if reqStr, ok := req.(string); ok {
				required[reqStr] = true
			}
		}
	}

	// Debug: Show schema information in verbose mode
	if len(required) > 0 {
		fmt.Printf("Schema indicates required parameters: %v\n", getRequiredParamsList(required))
	} else {
		fmt.Println("Schema indicates no required parameters")
	}

	fmt.Println("\nParameter input:")
	fmt.Println("• Required parameters must have a value")
	fmt.Println("• Optional parameters can be skipped by pressing Enter")
	fmt.Println()

	// Collect each parameter
	for propName, propSchema := range properties {
		propMap, _ := propSchema.(map[string]interface{})
		propType := "string"
		if t, ok := propMap["type"].(string); ok {
			propType = t
		}

		description := ""
		if desc, ok := propMap["description"].(string); ok {
			description = fmt.Sprintf(" (%s)", desc)
		}

		requiredStr := ""
		if required[propName] {
			requiredStr = " [required]"
		} else {
			requiredStr = " [optional]"
		}

		fmt.Printf("  %s%s%s (type: %s): ", propName, description, requiredStr, propType)

		if !scanner.Scan() {
			return params, nil
		}

		input := strings.TrimSpace(scanner.Text())

		// Handle empty input
		if input == "" {
			if required[propName] {
				fmt.Printf("    This parameter is required. Please enter a value.\n")
				fmt.Printf("  %s%s%s (type: %s): ", propName, description, requiredStr, propType)
				if !scanner.Scan() {
					return params, nil
				}
				input = strings.TrimSpace(scanner.Text())
				if input == "" {
					return nil, fmt.Errorf("required parameter '%s' cannot be empty", propName)
				}
			} else {
				// Optional parameter, skip it
				fmt.Printf("    ✓ Skipped (optional)\n")
				continue
			}
		}

		// Parse based on type
		switch propType {
		case "number", "integer":
			if num, err := strconv.ParseFloat(input, 64); err == nil {
				if propType == "integer" {
					params[propName] = int(num)
					fmt.Printf("    ✓ Set to: %d\n", int(num))
				} else {
					params[propName] = num
					fmt.Printf("    ✓ Set to: %g\n", num)
				}
			} else {
				return nil, fmt.Errorf("invalid number for %s: %s", propName, input)
			}
		case "boolean":
			lower := strings.ToLower(input)
			value := lower == "true" || lower == "yes" || lower == "y" || lower == "1"
			params[propName] = value
			fmt.Printf("    ✓ Set to: %t\n", value)
		case "array":
			// Try to parse as JSON array
			var arr []interface{}
			if err := json.Unmarshal([]byte(input), &arr); err != nil {
				// If not JSON, treat as comma-separated
				splitArr := strings.Split(input, ",")
				params[propName] = splitArr
				fmt.Printf("    ✓ Set to: %v (comma-separated)\n", splitArr)
			} else {
				params[propName] = arr
				fmt.Printf("    ✓ Set to: %v (JSON array)\n", arr)
			}
		case "object":
			// Parse as JSON object
			var obj map[string]interface{}
			if err := json.Unmarshal([]byte(input), &obj); err != nil {
				return nil, fmt.Errorf("invalid JSON object for %s: %s", propName, input)
			}
			params[propName] = obj
			fmt.Printf("    ✓ Set to: %v\n", obj)
		default:
			params[propName] = input
			fmt.Printf("    ✓ Set to: \"%s\"\n", input)
		}
	}

	// Show summary of collected parameters
	if len(params) > 0 {
		fmt.Printf("\n📋 Parameter summary:\n")
		for key, value := range params {
			fmt.Printf("  • %s: %v\n", key, value)
		}
	} else {
		fmt.Printf("\n📋 No parameters provided\n")
	}

	return params, nil
}

// getRequiredParamsList returns a slice of required parameter names for display
func getRequiredParamsList(required map[string]bool) []string {
	var list []string
	for param := range required {
		list = append(list, param)
	}
	return list
}

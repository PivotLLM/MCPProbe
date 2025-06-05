// Copyright (c) 2025 Tenebris Technologies Inc.
// This software is licensed under the MIT License (see LICENSE for details).

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	// Command line flags
	var (
		serverURL = flag.String("url", "", "MCP server URL (required)")
		mode      = flag.String("transport", "sse", "Transport mode: 'sse' or 'http'")
		headers   = flag.String("headers", "", "HTTP headers in format 'key1:value1,key2:value2'")
		timeout   = flag.Duration("timeout", 30*time.Second, "Connection timeout")
		verbose   = flag.Bool("verbose", true, "Enable verbose output")
	)
	flag.Parse()

	if *serverURL == "" {
		fmt.Println("Error: Server URL is required")
		fmt.Println("Usage: go run main.go -url <server-url> [-transport sse|http] [-headers key:value,key:value] [-timeout 30s]")
		os.Exit(1)
	}

	fmt.Printf("=== MCP Server Test Tool ===\n")
	fmt.Printf("Server URL: %s\n", *serverURL)

	fmt.Printf("Transport: %s\n", *mode)
	fmt.Printf("Timeout: %s\n", *timeout)
	fmt.Println()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

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
		mcpClient, err = createSSEClient(*serverURL, headerMap)
	case "http":
		fmt.Println("Creating HTTP client...")
		mcpClient, err = createHTTPClient(*serverURL, headerMap)
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

	// Start the client connection
	fmt.Println("Starting client connection...")
	if err := mcpClient.Start(ctx); err != nil {
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

	// Perform initialization handshake
	fmt.Println("\nPerforming initialization handshake...")
	if err := performInitialization(ctx, mcpClient, *verbose); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	fmt.Println("\nInitialization completed successfully")

	// Query server capabilities and list items
	if err := testServerCapabilities(ctx, mcpClient, *verbose); err != nil {
		log.Fatalf("Failed to test capabilities: %v", err)
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

func createSSEClient(serverURL string, headers map[string]string) (*client.Client, error) {
	var options []transport.ClientOption
	if len(headers) > 0 {
		options = append(options, client.WithHeaders(headers))
	}
	return client.NewSSEMCPClient(serverURL, options...)
}

func createHTTPClient(serverURL string, headers map[string]string) (*client.Client, error) {
	var options []transport.StreamableHTTPCOption
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
			fmt.Printf("     Input Schema: %v\n\n", tool.InputSchema)
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

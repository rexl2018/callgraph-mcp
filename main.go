package main

import (
	"context"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	
	"callgraph-mcp/handlers"
)

// callgraphTool implements the callgraph functionality
func callgraphTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Delegate to the handler in handlers package
	return handlers.HandleCallgraphRequest(ctx, request)
}

func main() {
	// Create a new MCP server
	mcpServer := server.NewMCPServer(
		"callgraph-mcp",
		"1.0.0",
	)

	// Register the callgraph tool
	mcpServer.AddTool(mcp.Tool{
		Name:        "callgraph",
		Description: "Generate call graph for Go packages in JSON format",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"moduleArgs": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "Package/module arguments (e.g., ['./...'])",
				},
				"dir": map[string]interface{}{
					"type":        "string",
					"description": "Working directory",
				},
				"focus": map[string]interface{}{
					"type":        "string",
					"description": "Focus specific package using name or import path",
				},
				"group": map[string]interface{}{
					"type":        "string",
					"description": "Grouping functions by packages and/or types [pkg, type] (separated by comma)",
				},
				"limit": map[string]interface{}{
					"type":        "string",
					"description": "Limit package paths to given prefixes (separated by comma)",
				},
				"ignore": map[string]interface{}{
					"type":        "string",
					"description": "Ignore package paths containing given prefixes (separated by comma)",
				},
				"include": map[string]interface{}{
					"type":        "string",
					"description": "Include package paths with given prefixes (separated by comma)",
				},
				"nostd": map[string]interface{}{
					"type":        "boolean",
					"description": "Omit calls to/from packages in standard library",
				},
				"nointer": map[string]interface{}{
					"type":        "boolean",
					"description": "Omit calls to unexported functions",
				},
				"tests": map[string]interface{}{
					"type":        "boolean",
					"description": "Include test code",
				},
				"algo": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"static", "cha", "rta"},
					"description": "The algorithm used to construct the call graph",
				},
				"tags": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "Build tags",
				},
				"debug": map[string]interface{}{
					"type":        "boolean",
					"description": "Enable verbose log",
				},
			},
			Required: []string{"moduleArgs"},
		},
	}, callgraphTool)

	// Set up logging
	if len(os.Args) > 1 && os.Args[1] == "--debug" {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(os.Stderr) // For now, always log to stderr for debugging
	}

	log.Printf("Starting callgraph-mcp server...")

	// Start the server using stdio transport
	if err := server.ServeStdio(mcpServer); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
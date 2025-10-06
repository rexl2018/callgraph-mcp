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
		"1.0.1",
	)

	// Register the unified callHierarchy tool
	// Build external (visible) and internal (hidden) properties
    externalProps := map[string]interface{}{
	    "moduleArgs": map[string]interface{}{
	        "type":        "array",
	        "items":       map[string]string{"type": "string"},
	        "description": "Go package paths to analyze (e.g., ['./...', './cmd/myapp', 'github.com/user/repo'])",
	    },
	    "dir": map[string]interface{}{
	        "type":        "string",
	        "description": "Working directory for resolving relative package paths (should be absolute path of {pwd}, mandatory for working on local codebase)",
	    },
	    "limit_keyword": map[string]interface{}{
	        "type":        "array",
	        "items":       map[string]string{"type": "string"},
	        "description": "Limit by package path keywords (caller AND callee must both match; normally use your project name)",
	        "default":     []string{},
	    },
	    "ignore": map[string]interface{}{
	        "type":        "array",
	        "items":       map[string]string{"type": "string"},
	        "description": "Ignore package paths containing given prefixes",
	        "default":     []string{},
	    },
	    "limit_prefix": map[string]interface{}{
	        "type":        "array",
	        "items":       map[string]string{"type": "string"},
	        "description": "Limit by import path prefixes (caller AND callee must both match), example: code.byted.org/tikcast/aaa_bbb",
	        "default":     []string{},
	    },
	    "symbol": map[string]interface{}{
	        "type":        "string",
	        "description": "Function symbol to start traversal from. Supports: function name ('hello'), package.function ('main.main'), or full path ('github.com/user/repo.function')",
	    },
	    "direction": map[string]interface{}{
	        "type":        "string",
	        "enum":        []string{"downstream", "upstream", "both"},
	        "description": "Traversal direction (default: downstream; only effective when 'symbol' is specified)",
	    },
	    "max_dep": map[string]interface{}{
            "type":        "integer",
            "description": "Max traversal depth (0 for unlimited; defaults: 7 when symbol is specified, 4 otherwise)",
            "default":     0,
        },
    }
    internalProps := map[string]interface{}{
        "focus": map[string]interface{}{
            "type":        "string",
            "description": "Focus specific package using name or import path",
        },
        "group": map[string]interface{}{
            "type":        "array",
            "items":       map[string]string{"type": "string"},
            "description": "Grouping functions by packages and/or types [pkg,type]",
            "default":     []string{"pkg"},
        },
        "nostd": map[string]interface{}{
            "type":        "boolean",
            "description": "Omit calls to/from packages in standard library",
            "default":     true,
        },
        "nointer": map[string]interface{}{
            "type":        "boolean",
            "description": "Omit calls to unexported functions",
            "default":     true,
        },
        "tests": map[string]interface{}{
            "type":        "boolean",
            "description": "Include test code",
            "default":     false,
        },
        "algo": map[string]interface{}{
            "type":        "string",
            "enum":        []string{"static", "cha", "rta"},
            "description": "The algorithm used to construct the call graph (default: rta)",
            "default":     "rta",
        },
        "tags": map[string]interface{}{
            "type":        "array",
            "items":       map[string]string{"type": "string"},
            "description": "Build tags",
            "default":     []string{},
        },
        "debug": map[string]interface{}{
            "type":        "boolean",
            "description": "Enable verbose log",
            "default":     false,
        },
    }
    // Prevent unused variable error while keeping internal map for server-side parsing semantics
    _ = internalProps
    // Only expose external properties to clients
    filtered := externalProps

	mcpServer.AddTool(mcp.Tool{
		Name:        "callHierarchy",
		Description: "Generate call hierarchy for Go packages/functions in Mermaid format" +
		"\nExample (package-level):\n{" +
		"\n  \"dir\": \"/path/to/project/parent/dir/project_name\"," +
		"\n  \"limit_keyword\": [\"project_name\"]," +
		"\n  \"moduleArgs\": [\"./...\"]\n}" +
		"\nExample (through specific function - add these two lines):" +
		"\n  \"symbol\": \"main.main\"," +
		"\n  \"direction\": \"downstream\"" +
		"\n\nWhat this tool is good for:" +
		"\n- Analyze project structure and module boundaries" +
		"\n- Inspect function call chains (downstream/upstream)" +
		"\n- Understand cross-package dependencies and hot paths" +
		"\n- Quickly scope analysis with limit_keyword/limit_prefix (suggest to set one of them)" ,
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: filtered,
			Required:   []string{"moduleArgs"},
		},
	}, callgraphTool)

	// Set up logging
	if len(os.Args) > 1 && os.Args[1] == "--debug" {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(os.Stderr) // For now, always log to stderr for debugging
	}

	// Choose transport based on environment
    transport := os.Getenv("MCP_TRANSPORT")
    if transport == "sse" {
        addr := os.Getenv("MCP_ADDR")
        if addr == "" {
            addr = ":11156"
        }
        log.Printf("Starting callgraph-mcp SSE server on %s", addr)
        sseServer := server.NewSSEServer(mcpServer)
        if err := sseServer.Start(addr); err != nil {
            log.Fatalf("Server error: %v", err)
        }
    } else {
        log.Printf("Starting callgraph-mcp stdio server...")
        if err := server.ServeStdio(mcpServer); err != nil {
            log.Fatalf("Server error: %v", err)
        }
    }
}
// Remove old 'include' property entirely per request
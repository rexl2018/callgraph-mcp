package unit

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	
	"callgraph-mcp/handlers"
)

// TestCallgraphToolBasic tests the basic functionality of the callgraph tool
func TestCallgraphToolBasic(t *testing.T) {
	// Test basic functionality
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "callHierarchy",
			Arguments: map[string]interface{}{
				"moduleArgs": []string{"../fixtures/simple"},
				"algo":       "static",
				"nostd":      false, // Include std lib to see actual output
			},
		},
	}

	result, err := handlers.HandleCallgraphRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("HandleCallgraphRequest failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	if len(result.Content) == 0 {
		t.Fatal("result content is empty")
	}

	// Validate Mermaid output
	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("result content is not TextContent")
	}

	if text := textContent.Text; text == "" {
		t.Fatal("Mermaid output is empty")
	} else if !(len(text) >= 10 && text[:10] == "flowchart ") {
		t.Fatalf("Mermaid output does not start with 'flowchart ': %q", text[:10])
	}

	// Print the actual output for inspection
	fmt.Printf("\n=== Mermaid Output (no grouping) ===\n%s\n=== End ===\n", textContent.Text)
}

// TestCallgraphWithGrouping tests package grouping
func TestCallgraphWithGrouping(t *testing.T) {
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "callHierarchy",
			Arguments: map[string]interface{}{
				"moduleArgs": []string{"../fixtures/simple"},
				"algo":       "static",
				"nostd":      false,
				"group":      []string{"pkg"},
			},
		},
	}

	result, err := handlers.HandleCallgraphRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("HandleCallgraphRequest failed: %v", err)
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("result content is not TextContent")
	}

	// Print the actual output for inspection
	fmt.Printf("\n=== Mermaid Output (pkg grouping) ===\n%s\n=== End ===\n", textContent.Text)
}

// TestMCPCallgraphTypes tests the MCP types structure (unchanged)
func TestMCPCallgraphTypes(t *testing.T) {
	// Test MCPCallgraphRequest
	req := handlers.MCPCallgraphRequest{
		ModuleArgs: []string{"./test"},
		Algo:       "static",
		Focus:      "main",
		NoStd:      true,
		NoInter:    false,
		Group:      []string{"pkg", "type"},
		LimitKeyword: []string{"github.com/test"},
		Ignore:     []string{"vendor"},
		LimitPrefix: []string{"internal"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	var req2 handlers.MCPCallgraphRequest
	if err := json.Unmarshal(data, &req2); err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}
}

// Remove JSON-based response tests because output is now Mermaid

// TestSymbolCallsBasic tests the basic functionality of the callHierarchy tool
func TestSymbolCallsBasic(t *testing.T) {
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "callHierarchy",
			Arguments: map[string]interface{}{
				"moduleArgs": []string{"../fixtures/simple"},
				"algo":       "static",
				"nostd":      true,
				"group":      []string{"pkg"},
				"symbol":     "main.main",
				"direction":  "downstream",
			},
		},
	}

	result, err := handlers.HandleCallgraphRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("HandleCallgraphRequest failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	if len(result.Content) == 0 {
		t.Fatal("result content is empty")
	}

	// Validate Mermaid output
	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("result content is not TextContent")
	}

	if text := textContent.Text; text == "" {
		t.Fatal("Mermaid output is empty")
	} else if !(len(text) >= 10 && text[:10] == "flowchart ") {
		t.Fatalf("Mermaid output does not start with 'flowchart ': %q", text[:10])
	}

	// Print the actual output for inspection
	fmt.Printf("\n=== Symbol Calls Mermaid Output (downstream from main.main) ===\n%s\n=== End ===\n", textContent.Text)
}

// TestSymbolCallsUpstream tests upstream traversal
func TestSymbolCallsUpstream(t *testing.T) {
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "callHierarchy",
			Arguments: map[string]interface{}{
				"moduleArgs": []string{"../fixtures/simple"},
				"algo":       "static",
				"nostd":      true,
				"group":      []string{"pkg"},
				"symbol":     "hello",
				"direction":  "upstream",
			},
		},
	}

	result, err := handlers.HandleCallgraphRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("HandleCallgraphRequest failed: %v", err)
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("result content is not TextContent")
	}

	// Print the actual output for inspection
	fmt.Printf("\n=== Symbol Calls Mermaid Output (upstream from hello) ===\n%s\n=== End ===\n", textContent.Text)
}

// TestSymbolCallsRequestTypes tests the callHierarchy request structure
func TestSymbolCallsRequestTypes(t *testing.T) {
	// Test MCPCallgraphRequest with symbol and direction fields
	req := handlers.MCPCallgraphRequest{
	ModuleArgs: []string{"./test"},
	Algo:       "static",
	Focus:      "main",
	NoStd:      true,
	NoInter:    false,
	Group:      []string{"pkg", "type"},
	LimitKeyword: []string{"github.com/test"},
	Ignore:     []string{"vendor"},
	LimitPrefix: []string{"internal"},
	Symbol:     "main.main",
	Direction:  "downstream",
}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal callHierarchy request: %v", err)
	}

	var req2 handlers.MCPCallgraphRequest
	if err := json.Unmarshal(data, &req2); err != nil {
		t.Fatalf("failed to unmarshal callHierarchy request: %v", err)
	}

	// Verify symbol and direction fields are preserved
	if req2.Symbol != "main.main" {
		t.Errorf("expected symbol 'main.main', got '%s'", req2.Symbol)
	}
	if req2.Direction != "downstream" {
		t.Errorf("expected direction 'downstream', got '%s'", req2.Direction)
	}
}
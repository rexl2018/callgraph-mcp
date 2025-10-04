package unit

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	
	"callgraph-mcp/handlers"
)

// TestCallgraphToolBasic tests the basic functionality of the callgraph tool
func TestCallgraphToolBasic(t *testing.T) {
	// Test basic functionality
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "callgraph",
			Arguments: map[string]interface{}{
				"moduleArgs": []string{"../fixtures/simple"},
				"algo":       "static",
				"nostd":      true,
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

	// Check if the result contains valid JSON
	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("result content is not TextContent")
	}

	var response handlers.MCPCallgraphResponse
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	// Basic validation
	if response.Algorithm == "" {
		t.Error("algorithm is empty")
	}

	if response.Stats.DurationMs <= 0 {
		t.Error("duration should be positive")
	}
}

// TestMCPCallgraphTypes tests the MCP types structure
func TestMCPCallgraphTypes(t *testing.T) {
	// Test MCPCallgraphRequest
	req := handlers.MCPCallgraphRequest{
		ModuleArgs: []string{"./test"},
		Dir:        "/tmp",
		Focus:      "main",
		Group:      "pkg",
		Limit:      "github.com/test",
		Ignore:     "vendor",
		Include:    "internal",
		NoStd:      true,
		NoInter:    false,
		Tests:      true,
		Algo:       "static",
		Tags:       []string{"integration"},
		Debug:      false,
	}

	// Test JSON marshaling/unmarshaling
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	var req2 handlers.MCPCallgraphRequest
	if err := json.Unmarshal(data, &req2); err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}

	// Verify fields
	if req2.ModuleArgs[0] != req.ModuleArgs[0] {
		t.Errorf("ModuleArgs mismatch: got %v, want %v", req2.ModuleArgs, req.ModuleArgs)
	}

	if req2.Algo != req.Algo {
		t.Errorf("Algo mismatch: got %s, want %s", req2.Algo, req.Algo)
	}

	if req2.NoStd != req.NoStd {
		t.Errorf("NoStd mismatch: got %v, want %v", req2.NoStd, req.NoStd)
	}
}

// TestMCPCallgraphResponse tests the response structure
func TestMCPCallgraphResponse(t *testing.T) {
	focus := "main"
	response := handlers.MCPCallgraphResponse{
		Algorithm: "static",
		Focus:     &focus,
		Filters: handlers.MCPCallgraphFilters{
			Limit:   []string{"github.com/test"},
			Ignore:  []string{"vendor"},
			Include: []string{"internal"},
			NoStd:   true,
			NoInter: false,
			Group:   []string{"pkg"},
		},
		Stats: handlers.MCPCallgraphStats{
			NodeCount:  10,
			EdgeCount:  5,
			DurationMs: 100,
		},
		Graph: handlers.MCPCallgraphData{
			Nodes: []handlers.MCPCallgraphNode{
				{
					ID:          "main.main",
					Func:        "main",
					PackagePath: "main",
					PackageName: "main",
					File:        "main.go",
					Line:        10,
					IsStd:       false,
					Exported:    true,
				},
			},
			Edges: []handlers.MCPCallgraphEdge{
				{
					Caller:    "main.main",
					Callee:    "main.hello",
					File:      "main.go",
					Line:      11,
					Synthetic: false,
				},
			},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	// Test JSON unmarshaling
	var response2 handlers.MCPCallgraphResponse
	if err := json.Unmarshal(data, &response2); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Verify structure
	if response2.Algorithm != response.Algorithm {
		t.Errorf("Algorithm mismatch: got %s, want %s", response2.Algorithm, response.Algorithm)
	}

	if response2.Stats.NodeCount != response.Stats.NodeCount {
		t.Errorf("NodeCount mismatch: got %d, want %d", response2.Stats.NodeCount, response.Stats.NodeCount)
	}

	if len(response2.Graph.Nodes) != len(response.Graph.Nodes) {
		t.Errorf("Nodes length mismatch: got %d, want %d", len(response2.Graph.Nodes), len(response.Graph.Nodes))
	}

	if len(response2.Graph.Edges) != len(response.Graph.Edges) {
		t.Errorf("Edges length mismatch: got %d, want %d", len(response2.Graph.Edges), len(response.Graph.Edges))
	}
}
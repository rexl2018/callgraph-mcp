package integration

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	
	"callgraph-mcp/handlers"
)

func TestCallgraphToolIntegration(t *testing.T) {
	// Test basic functionality with real code
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

	// Should have some stats (even if filtered out, we should have processed something)
	if response.Stats.DurationMs <= 0 {
		t.Error("duration should be positive, indicating processing occurred")
	}
	
	// Log the actual stats for debugging
	t.Logf("Stats: NodeCount=%d, EdgeCount=%d, DurationMs=%d", 
		response.Stats.NodeCount, response.Stats.EdgeCount, response.Stats.DurationMs)
}

func TestCallgraphToolWithDifferentAlgorithms(t *testing.T) {
	algorithms := []string{"static", "cha", "rta"}
	
	for _, algo := range algorithms {
		t.Run(algo, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "callgraph",
					Arguments: map[string]interface{}{
						"moduleArgs": []string{"../fixtures/simple"},
						"algo":       algo,
						"nostd":      true,
					},
				},
			}

			result, err := handlers.HandleCallgraphRequest(context.Background(), request)
			if err != nil {
				t.Fatalf("HandleCallgraphRequest failed for %s: %v", algo, err)
			}

			if result == nil {
				t.Fatalf("result is nil for %s", algo)
			}

			// Parse response
			textContent, ok := result.Content[0].(mcp.TextContent)
			if !ok {
				t.Fatalf("result content is not TextContent for %s", algo)
			}

			var response handlers.MCPCallgraphResponse
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("failed to parse response JSON for %s: %v", algo, err)
			}

			if response.Algorithm != algo {
				t.Errorf("expected algorithm %s, got %s", algo, response.Algorithm)
			}
		})
	}
}

func TestCallgraphToolErrorHandling(t *testing.T) {
	tests := []struct {
		name    string
		request mcp.CallToolRequest
		wantErr bool
	}{
		{
			name: "missing module args",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "callgraph",
					Arguments: map[string]interface{}{
						"algo": "static",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid module path",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "callgraph",
					Arguments: map[string]interface{}{
						"moduleArgs": []string{"./nonexistent"},
						"algo":       "static",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid algorithm",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "callgraph",
					Arguments: map[string]interface{}{
						"moduleArgs": []string{"../fixtures/simple"},
						"algo":       "invalid",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handlers.HandleCallgraphRequest(context.Background(), tt.request)
			
			if tt.wantErr {
				if err == nil && (result == nil || !result.IsError) {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
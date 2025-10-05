package integration

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	
	"callgraph-mcp/handlers"
)

func TestCallgraphToolIntegration(t *testing.T) {
	// Test basic functionality with real code
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "callHierarchy",
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

	// Validate Mermaid flowchart string
	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("result content is not TextContent")
	}
	if text := textContent.Text; text == "" {
		t.Fatal("Mermaid output is empty")
	} else if !(len(text) >= 10 && text[:10] == "flowchart ") {
		t.Fatalf("Mermaid output does not start with 'flowchart ': %q", text[:10])
	}
}

func TestCallgraphToolWithDifferentAlgorithms(t *testing.T) {
	algorithms := []string{"static", "cha", "rta"}
	
	for _, algo := range algorithms {
		t.Run(algo, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "callHierarchy",
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

			// Validate Mermaid output
			textContent, ok := result.Content[0].(mcp.TextContent)
			if !ok {
				t.Fatalf("result content is not TextContent for %s", algo)
			}
			if text := textContent.Text; text == "" {
				t.Fatalf("Mermaid output is empty for %s", algo)
			} else if !(len(text) >= 10 && text[:10] == "flowchart ") {
				t.Fatalf("Mermaid output does not start with 'flowchart ' for %s: %q", algo, text[:10])
			}
		})
	}
}

func TestCallgraphToolErrorHandling(t *testing.T) {
	// Keep error handling tests, e.g., invalid algo should still error from DoAnalysis
	tests := []struct {
		name    string
		request mcp.CallToolRequest
		wantErr bool
	}{
		{
			name: "missing moduleArgs",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "callHierarchy",
					Arguments: map[string]interface{}{
						"algo":  "static",
						"nostd": true,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid algorithm",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "callHierarchy",
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
				if err != nil || result == nil || result.IsError {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSymbolCallsToolIntegration(t *testing.T) {
	// Test basic symbol_calls functionality
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "callHierarchy",
			Arguments: map[string]interface{}{
				"moduleArgs": []string{"../fixtures/simple"},
				"algo":       "static",
				"nostd":      true,
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

	// Validate Mermaid flowchart string
	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("result content is not TextContent")
	}
	if text := textContent.Text; text == "" {
		t.Fatal("Mermaid output is empty")
	} else if !(len(text) >= 10 && text[:10] == "flowchart ") {
		t.Fatalf("Mermaid output does not start with 'flowchart ': %q", text[:10])
	}
}

func TestSymbolCallsErrorHandling(t *testing.T) {
	tests := []struct {
		name    string
		request mcp.CallToolRequest
		wantErr bool
	}{
		{
			name: "missing symbol",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "callHierarchy",
					Arguments: map[string]interface{}{
						"moduleArgs": []string{"../fixtures/simple"},
						"algo":       "static",
						"nostd":      true,
						"direction":  "downstream",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid symbol",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "callHierarchy",
					Arguments: map[string]interface{}{
						"moduleArgs": []string{"../fixtures/simple"},
						"algo":       "static",
						"nostd":      true,
						"symbol":     "nonexistent.function",
						"direction":  "downstream",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing moduleArgs",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "callHierarchy",
					Arguments: map[string]interface{}{
						"algo":      "static",
						"nostd":     true,
						"symbol":    "main.main",
						"direction": "downstream",
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
				if err != nil || result == nil || result.IsError {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"callgraph-mcp/handlers"
)

func TestSymbolCallsDownstream(t *testing.T) {
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "callHierarchy",
			Arguments: map[string]interface{}{
				"moduleArgs": []string{"../fixtures/simple"},
				"algo":       "static",
				"nostd":      true,
				"nointer":    false,
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

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("result content is not TextContent")
	}

	text := textContent.Text
	if !strings.HasPrefix(text, "flowchart ") {
		t.Fatalf("Mermaid output does not start with 'flowchart ': %q", text[:10])
	}

	// Expect edges from main to hello and goodbye in simple fixture
	if !strings.Contains(text, "callgraph_mcp_tests_fixtures_simple_main --> callgraph_mcp_tests_fixtures_simple_hello") {
		t.Fatalf("Mermaid output missing expected edge main->hello. Output snippet: %q", text)
	}
	if !strings.Contains(text, "callgraph_mcp_tests_fixtures_simple_main --> callgraph_mcp_tests_fixtures_simple_goodbye") {
		t.Fatalf("Mermaid output missing expected edge main->goodbye. Output snippet: %q", text)
	}
}

func TestSymbolCallsUpstream(t *testing.T) {
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "callHierarchy",
			Arguments: map[string]interface{}{
				"moduleArgs": []string{"../fixtures/simple"},
				"algo":       "static",
				"nostd":      true,
				"nointer":    false,
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

	text := textContent.Text
	if !strings.HasPrefix(text, "flowchart ") {
		t.Fatalf("Mermaid output does not start with 'flowchart ': %q", text[:10])
	}

	// Expect edge from main to hello when traversing upstream from hello
	if !strings.Contains(text, "callgraph_mcp_tests_fixtures_simple_main --> callgraph_mcp_tests_fixtures_simple_hello") {
		t.Fatalf("Mermaid output missing expected upstream edge main->hello. Output snippet: %q", text)
	}
}

func TestSymbolCallsUpstreamGoodbye(t *testing.T) {
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "callHierarchy",
			Arguments: map[string]interface{}{
				"moduleArgs": []string{"../fixtures/simple"},
				"algo":       "static",
				"nostd":      true,
				"nointer":    false,
				"group":      []string{"pkg"},
				"symbol":     "goodbye",
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

	text := textContent.Text
	if !strings.HasPrefix(text, "flowchart ") {
		t.Fatalf("Mermaid output does not start with 'flowchart ': %q", text[:10])
	}

	// Expect edge from main to goodbye when traversing upstream from goodbye
	if !strings.Contains(text, "callgraph_mcp_tests_fixtures_simple_main --> callgraph_mcp_tests_fixtures_simple_goodbye") {
		t.Fatalf("Mermaid output missing expected upstream edge main->goodbye. Output snippet: %q", text)
	}
}

func TestSymbolCallsWorker(t *testing.T) {
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "callHierarchy",
			Arguments: map[string]interface{}{
				"moduleArgs": []string{"../fixtures/simple"},
				"algo":       "static",
				"nostd":      true,
				"nointer":    false,
				"group":      []string{"pkg"},
				"symbol":     "worker",
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

	text := textContent.Text
	if !strings.HasPrefix(text, "flowchart ") {
		t.Fatalf("Mermaid output does not start with 'flowchart ': %q", text[:10])
	}

	// 至少应包含 worker 节点
	if !strings.Contains(text, "callgraph_mcp_tests_fixtures_simple_worker[") {
		t.Fatalf("Mermaid output missing worker node. Output snippet: %q", text)
	}
}


func TestSymbolCallsWorkerDownstreamAlgorithms(t *testing.T) {
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
						"nointer":    false,
						"group":      []string{"pkg"},
						"symbol":     "main.main",
						"direction":  "downstream",
					},
				},
			}

			result, err := handlers.HandleCallgraphRequest(context.Background(), request)
			if err != nil {
				t.Fatalf("HandleCallgraphRequest failed (%s): %v", algo, err)
			}

			textContent, ok := result.Content[0].(mcp.TextContent)
			if !ok {
				t.Fatal("result content is not TextContent")
			}

			text := textContent.Text
			if !strings.HasPrefix(text, "flowchart ") {
				t.Fatalf("Mermaid output does not start with 'flowchart ' (%s): %q", algo, text[:10])
			}

			// 断言：存在 worker 节点
			if !strings.Contains(text, "callgraph_mcp_tests_fixtures_simple_worker[") {
				t.Fatalf("Mermaid output missing worker node (%s). Output snippet: %q", algo, text)
			}
			// 断言：存在 main -> worker 的边
			if !strings.Contains(text, "callgraph_mcp_tests_fixtures_simple_main --> callgraph_mcp_tests_fixtures_simple_worker") {
				t.Fatalf("Mermaid output missing edge main->worker (%s). Output snippet: %q", algo, text)
			}
		})
	}
}
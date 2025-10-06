package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"callgraph-mcp/handlers"

	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	fmt.Println("=== Testing improved nostd=true filtering ===")
	
	// Test with nostd=true - should now filter out io/fs, math/bits etc.
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "callHierarchy",
			Arguments: map[string]interface{}{
				"moduleArgs": []string{"./tests/fixtures/simple"},
				"algo":       "static",
				"nostd":      true,
				"group":      []string{"pkg"},
			},
		},
	}
	
	result, err := handlers.HandleCallgraphRequest(context.Background(), request)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	if textContent, ok := result.Content[0].(mcp.TextContent); ok {
		fmt.Printf("=== Mermaid Output (improved nostd=true, group=pkg) ===\n")
		fmt.Println(textContent.Text)
		fmt.Printf("=== End (Length: %d characters) ===\n", len(textContent.Text))
	}
	
	fmt.Println("\nExpected:")
	fmt.Println("- Should see main() -> hello() call")
	fmt.Println("- Should NOT see hello() -> fmt.Println() call (filtered by nostd)")
	fmt.Println("- Should NOT see io/fs, math/bits etc. (filtered by improved nostd)")
	fmt.Println("- Should have subgraph for main package")
	
	fmt.Println("\n=== Demo: callHierarchy (upstream from hello) ===")
	reqUpHello := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "callHierarchy",
			Arguments: map[string]interface{}{
				"moduleArgs": []string{"./tests/fixtures/simple"},
				"algo":       "static",
				"nostd":      true,
				"group":      []string{"pkg"},
				"symbol":     "hello",
				"direction":  "upstream",
			},
		},
	}
	if res, err := handlers.HandleCallgraphRequest(context.Background(), reqUpHello); err != nil {
		log.Printf("Error: %v", err)
	} else if tc, ok := res.Content[0].(mcp.TextContent); ok {
		fmt.Println(tc.Text)
	}

	fmt.Println("\n=== Demo: callHierarchy (upstream from goodbye) ===")
	reqUpGoodbye := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "callHierarchy",
			Arguments: map[string]interface{}{
				"moduleArgs": []string{"./tests/fixtures/simple"},
				"algo":       "static",
				"nostd":      true,
				"group":      []string{"pkg"},
				"symbol":     "goodbye",
				"direction":  "upstream",
			},
		},
	}
	if res, err := handlers.HandleCallgraphRequest(context.Background(), reqUpGoodbye); err != nil {
		log.Printf("Error: %v", err)
	} else if tc, ok := res.Content[0].(mcp.TextContent); ok {
		fmt.Println(tc.Text)
	}

	fmt.Println("\n=== Demo: callHierarchy (downstream main.main) across algorithms ===")
	algos := []string{"static", "cha", "rta"}
	for _, a := range algos {
		fmt.Printf("-- algo: %s --\n", a)
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "callHierarchy",
				Arguments: map[string]interface{}{
					"moduleArgs": []string{"./tests/fixtures/simple"},
					"algo":       a,
					"nostd":      true,
					"group":      []string{"pkg"},
					"symbol":     "main.main",
					"direction":  "downstream",
				},
			},
		}
		res, err := handlers.HandleCallgraphRequest(context.Background(), req)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		if tc, ok := res.Content[0].(mcp.TextContent); ok {
			text := tc.Text
			foundWorker := strings.Contains(text, "callgraph_mcp_tests_fixtures_simple_worker[")
			foundMainWorker := strings.Contains(text, "callgraph_mcp_tests_fixtures_simple_main --> callgraph_mcp_tests_fixtures_simple_worker")
			fmt.Printf("worker node: %v, main->worker edge: %v\n", foundWorker, foundMainWorker)
		}
	}
    fmt.Println("\n=== Demo: battle_task downstream CreateTask (max_dep=5) ===")
    reqBattle := mcp.CallToolRequest{
        Params: mcp.CallToolParams{
            Name: "callHierarchy",
            Arguments: map[string]interface{}{
                "moduleArgs": []string{"./..."},
                "dir":        "/Users/bytedance/work/git/battle_task",
				"limit_keyword": []string{"battle_task"},
				//"limit_prefix": []string{"code.byted.org/tikcast/battle_task"},
                //"symbol":     "CreateTask",
                //"direction":  "downstream",
                "max_dep":    5,
				"algo":       "rta",
            },
        },
    }
    resBattle, err := handlers.HandleCallgraphRequest(context.Background(), reqBattle)
    if err != nil {
        log.Printf("Error running battle_task demo: %v", err)
    } else if textContent, ok := resBattle.Content[0].(mcp.TextContent); ok {
        fmt.Println(textContent.Text)
        fmt.Printf("=== End (Length: %d characters) ===\n", len(textContent.Text))
    }
}
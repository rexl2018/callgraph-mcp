package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"callgraph-mcp/handlers"
)

func main() {
	fmt.Println("=== Testing improved nostd=true filtering ===")
	
	// Test with nostd=true - should now filter out io/fs, math/bits etc.
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "callgraph",
			Arguments: map[string]interface{}{
				"moduleArgs": []string{"./tests/fixtures/simple"},
				"algo":       "static",
				"nostd":      true,
				"group":      "pkg",
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
}
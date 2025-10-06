package integration

import (
    "context"
    "fmt"
    "regexp"
    "strings"
    "testing"

    "github.com/mark3labs/mcp-go/mcp"

    "callgraph-mcp/handlers"
)

// Helper: extract compact Mermaid node ID (e.g. N1) for a given function name
func extractID(text string, funcName string) (string, bool) {
    // Pattern: N123["<funcName><br/>"]
    // We only match the start of the label with funcName<br/>
    re := regexp.MustCompile(`(?m)^(N\d+)\["` + regexp.QuoteMeta(funcName) + `<br/>`)
    m := re.FindStringSubmatch(text)
    if m == nil || len(m) < 2 {
        return "", false
    }
    return m[1], true
}

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
    mainID, ok := extractID(text, "main")
    if !ok { t.Fatalf("main node ID not found. Output: %q", text) }
    helloID, ok := extractID(text, "hello")
    if !ok { t.Fatalf("hello node ID not found. Output: %q", text) }
    goodbyeID, ok := extractID(text, "goodbye")
    if !ok { t.Fatalf("goodbye node ID not found. Output: %q", text) }

    if !strings.Contains(text, fmt.Sprintf("%s --> %s", mainID, helloID)) {
        t.Fatalf("Mermaid output missing expected edge main->hello. Output snippet: %q", text)
    }
    if !strings.Contains(text, fmt.Sprintf("%s --> %s", mainID, goodbyeID)) {
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

    mainID, ok := extractID(text, "main")
    if !ok { t.Fatalf("main node ID not found. Output: %q", text) }
    helloID, ok := extractID(text, "hello")
    if !ok { t.Fatalf("hello node ID not found. Output: %q", text) }

    if !strings.Contains(text, fmt.Sprintf("%s --> %s", mainID, helloID)) {
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

    mainID, ok := extractID(text, "main")
    if !ok { t.Fatalf("main node ID not found. Output: %q", text) }
    goodbyeID, ok := extractID(text, "goodbye")
    if !ok { t.Fatalf("goodbye node ID not found. Output: %q", text) }

    if !strings.Contains(text, fmt.Sprintf("%s --> %s", mainID, goodbyeID)) {
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
    if !strings.Contains(text, "worker<br/>") {
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
            if !strings.Contains(text, "worker<br/>") {
                t.Fatalf("Mermaid output missing worker node (%s). Output snippet: %q", algo, text)
            }
            // 断言：存在 main -> worker 的边
            mainID, ok := extractID(text, "main")
            if !ok { t.Fatalf("main node ID not found (%s). Output: %q", algo, text) }
            workerID, ok := extractID(text, "worker")
            if !ok { t.Fatalf("worker node ID not found (%s). Output: %q", algo, text) }
            if !strings.Contains(text, fmt.Sprintf("%s --> %s", mainID, workerID)) {
                t.Fatalf("Mermaid output missing edge main->worker (%s). Output snippet: %q", algo, text)
            }
        })
    }
}
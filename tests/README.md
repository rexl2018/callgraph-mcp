# Tests

This directory contains organized tests for the callgraph-mcp project.

## Structure

```
tests/
├── unit/                    # Unit tests
│   ├── handlers_test.go     # Tests for handlers package
│   └── main_test.go         # Main functionality tests
├── integration/             # Integration tests
│   ├── callgraph_integration_test.go  # Go integration tests
│   └── mcp_server_test.py   # Python MCP server tests
└── fixtures/                # Test data and fixtures
    ├── simple/              # Simple Go program for testing
    │   └── main.go
    └── test_request.json    # Sample MCP request for testing
```

## Running Tests

### Unit Tests

Run all unit tests:
```bash
go test ./tests/unit/...
```

Run specific unit test:
```bash
go test ./tests/unit/ -run TestMCPRequestMapping
```

### Integration Tests

Run Go integration tests:
```bash
go test ./tests/integration/...
```

Run Python integration tests:
```bash
cd tests/integration
python mcp_server_test.py
```

### All Tests

Run all Go tests:
```bash
go test ./tests/...
```

## Test Categories

### Unit Tests
- Test individual functions and methods
- Mock external dependencies
- Fast execution
- No external resources required

### Integration Tests
- Test complete workflows
- Use real dependencies
- Test MCP protocol integration
- May require building the binary first

### Fixtures
- **simple/**: Simple Go programs used as test data for callgraph generation
- **test_request.json**: Sample MCP JSON-RPC request for manual testing
- Minimal examples for testing callgraph generation
- Should be kept simple and focused

## Adding New Tests

1. **Unit tests**: Add to `tests/unit/` directory
2. **Integration tests**: Add to `tests/integration/` directory
3. **Test data**: Add to `tests/fixtures/` directory

Make sure to follow Go testing conventions and use descriptive test names.

## Symbol Calls 测试

`callHierarchy` 用于从指定函数符号出发，按方向遍历生成调用图；当不指定 `symbol` 时，生成包级调用图。我们在集成测试中覆盖了 downstream 与 upstream 两种方向，并对 static、cha、rta 三种算法进行了对比。

运行示例：
```bash
# 下游遍历（main.main 出发），算法对比
go test ./tests/integration -run SymbolCallsWorkerDownstreamAlgorithms -v

# 上游遍历（hello 出发）
go test ./tests/integration -run SymbolCallsUpstream -v

# 上游遍历（goodbye 出发）
go test ./tests/integration -run SymbolCallsUpstreamGoodbye -v

# 上游遍历（worker 触发）
go test ./tests/integration -run SymbolCallsWorker -v
```

参数提示：
- `symbol`: 可使用函数名（如 `hello`）、包名+函数名（如 `main.main`）、或完整路径+函数名（如 `callgraph-mcp/tests/fixtures/simple.main`）
- `direction`: 支持 `downstream`、`upstream`、`both`，默认 `downstream`
- 其它参数与包级调用图一致（如 `nostd`、`group`、`algo` 等）
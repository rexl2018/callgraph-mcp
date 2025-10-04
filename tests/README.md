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
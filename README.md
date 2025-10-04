# callgraph-mcp

**callgraph-mcp** 是一个基于 [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) 的 Go 调用图分析工具。

## 功能特性

- 🔍 **静态分析**：支持 `static`、`cha`、`rta` 三种调用图算法
- 🎯 **精确过滤**：支持包路径过滤、标准库过滤、未导出函数过滤等
- 📊 **JSON 输出**：返回结构化的 JSON 数据，包含节点和边信息
- 🔌 **MCP 协议**：完全兼容 Model Context Protocol，可与支持 MCP 的客户端集成
- ⚡ **高性能**：基于 Go 的 SSA 中间表示进行分析

## 安装

```bash
go build -o callgraph-mcp
```

## 使用方法

### 作为 MCP 服务器

callgraph-mcp 主要设计为 MCP 服务器使用。启动服务器：

```bash
./callgraph-mcp
```

### MCP 工具调用

服务器提供一个名为 `callgraph` 的工具，支持以下参数：

#### 必需参数
- `moduleArgs` ([]string): 要分析的包路径，例如 `["./..."]` 或 `["./cmd/myapp"]`

#### 可选参数
- `algo` (string): 分析算法，可选值：`static`（默认）、`cha`、`rta`
- `dir` (string): 工作目录
- `focus` (string): 聚焦特定包（按名称或导入路径）
- `group` (string): 分组方式，可选值：`pkg`（默认）、`type`，用逗号分隔
- `limit` (string): 限制包路径前缀，用逗号分隔
- `ignore` (string): 忽略包路径前缀，用逗号分隔
- `include` (string): 包含包路径前缀，用逗号分隔
- `nostd` (boolean): 忽略标准库调用
- `nointer` (boolean): 忽略未导出函数调用
- `tests` (boolean): 包含测试代码
- `tags` ([]string): 构建标签
- `debug` (boolean): 启用详细日志

### 示例请求

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "callgraph",
    "arguments": {
      "moduleArgs": ["./examples/main"],
      "algo": "static",
      "nostd": true,
      "focus": "main"
    }
  }
}
```

### 响应格式

```json
{
  "algorithm": "static",
  "focus": "main",
  "filters": {
    "limit": [],
    "ignore": [],
    "include": [],
    "nostd": true,
    "nointer": false,
    "group": ["pkg"]
  },
  "stats": {
    "nodeCount": 5,
    "edgeCount": 8,
    "durationMs": 123
  },
  "graph": {
    "nodes": [
      {
        "id": "main.main",
        "func": "main",
        "packagePath": "main",
        "packageName": "main",
        "file": "main.go",
        "line": 10,
        "isStd": false,
        "exported": true,
        "receiverType": null
      }
    ],
    "edges": [
      {
        "caller": "main.main",
        "callee": "main.hello",
        "file": "main.go",
        "line": 11,
        "synthetic": false
      }
    ]
  }
}
```

## 算法说明

### Static Analysis (`static`)
- 最快的分析方法
- 基于静态调用关系
- 不考虑动态调用和接口调用

### Class Hierarchy Analysis (`cha`)
- 考虑类型层次结构
- 处理接口调用
- 比 static 更精确，但速度较慢

### Rapid Type Analysis (`rta`)
- 最精确的分析方法
- 考虑实际可达的类型
- 分析速度最慢，但结果最准确

## 过滤选项

### 包路径过滤
- `limit`: 只包含指定前缀的包
- `ignore`: 排除指定前缀的包
- `include`: 强制包含指定前缀的包（优先级高于其他过滤器）

### 函数过滤
- `nostd`: 排除标准库函数调用
- `nointer`: 排除未导出函数调用

### 聚焦分析
- `focus`: 只分析与指定包相关的调用关系

## 开发

### 测试

项目包含完整的测试套件，分为单元测试和集成测试：

```
tests/
├── unit/                    # 单元测试
├── integration/             # 集成测试
└── fixtures/                # 测试数据
```

#### 运行测试

```bash
# 运行所有测试
go test ./tests/...

# 只运行单元测试
go test ./tests/unit/...

# 只运行集成测试
go test ./tests/integration/...

# 使用测试工具
cd tests && make test-all
```

更多测试相关信息请参考 [tests/README.md](tests/README.md)。

### 构建

```bash
go build -o callgraph-mcp
```


## 参考项目
- [go-callvis](https://github.com/ofabry/go-callvis)


---

**版本**: v0.1.0

# kwcli v0.2.0 开发指南：新特性、使用与注意事项

> 让 KWDB 从 CLI 工具进化为 AI 生态的执行底座

## 一、回顾与展望

kwcli 最初是一个 KWDB 生态的命令行工具，提供 Playground 一键启动、KWDB 安装部署、SQL 连接等基础能力。经过持续迭代，kwcli 现在已经从单纯的 CLI 工具演进为 **KWDB AI 生态的标准执行底座**。

从 v0.1.1 到 v0.2.0，kwcli 引入了三大核心能力：
1. **Skill 管理系统** — 统一管理 KWDB Agent Skills
2. **MCP 协议支持** — 成为 AI Agent 的标准执行器
3. **AI 交互模式** — 内置轻量 Agent，支持自然语言操作数据库

## 二、新特性详解

### 2.1 全局 JSON 输出模式

所有核心命令支持 `--json` 标志，输出统一的 JSON 结构，便于脚本和 Agent 解析：

```bash
kwcli kwdb status --json
kwcli sql -e "SELECT count(*) FROM meter_data" --json
kwcli sampledb list --json
kwcli tsbs list --json
```

输出格式统一为：

```json
{
  "success": true,
  "command": "kwdb status",
  "data": { ... },
  "error": null
}
```

> **实现位置**：`pkg/output/formatter.go` 中的 `Response` 结构体和 `PrintJSON` 函数。

### 2.2 Schema 导出

支持将数据库 Schema 导出为 SQL DDL 或 JSON 格式：

```bash
# 导出为 SQL DDL
kwcli schema dump --db iot_db --output schema.sql

# 导出为 JSON（含表结构、索引、标签信息）
kwcli schema dump --db all --format json
```

> **实现位置**：`pkg/schema/schema.go`

### 2.3 批量 SQL 执行与结果导出

```bash
# 执行 SQL 脚本文件
kwcli sql -f init.sql

# 事务模式执行
kwcli sql -f migrate.sql --transaction

# 查询结果导出为 CSV
kwcli sql -e "SELECT * FROM meter_data" --export csv --output data.csv

# 查询结果导出为 JSON
kwcli sql -e "SELECT * FROM meter_data" --export json
```

> **实现位置**：`cmd/sql.go` 中的 `executeSQLFile` 和 `executeAndExportSQL` 函数，`pkg/output/export.go`

### 2.4 数据库巡检

一键式健康检查，覆盖 6 项核心指标：

```bash
# 运行标准巡检
kwcli inspect run

# 指定巡检项
kwcli inspect run --items status,logs,connections

# 输出报告
kwcli inspect run --output report.md --format markdown
kwcli inspect run --output report.html --format html
```

巡检项包括：服务状态、日志异常、连接数、慢查询、存储空间、SampleDB 健康。

> **实现位置**：`cmd/inspect.go`、`pkg/inspect/inspect.go`

### 2.5 性能快照

一键收集当前性能指标，快速定位性能瓶颈：

```bash
kwcli perf snapshot

# JSON 格式输出
kwcli perf snapshot --json

# 保存到文件
kwcli perf snapshot --output snapshot.json --format json
```

采集指标：QPS/TPS、活跃连接数、TOP 10 慢查询、存储空间使用。

> **实现位置**：`cmd/perf.go`、`pkg/perf/perf.go`

---

### 2.6 Skill 管理系统（Phase 2 - 全新）

Skill 是 KWDB 的知识包，包含元数据、参考文档和命令模板，帮助 AI Agent 理解和操作 KWDB。

```bash
# 列出已安装的 Skills
kwcli skill list

# 安装 Skill（从 GitHub/AtomGit）
kwcli skill install kwdb-text2sql-aiot
kwcli skill install kwdb-text2sql-aiot --version v1.0.1 --source github

# 查看 Skill 详情
kwcli skill info kwdb-text2sql-aiot

# 查看 Skill 支持的场景
kwcli skill scenarios kwdb-text2sql-aiot

# 预览 Skill 参考文件
kwcli skill preview kwdb-text2sql-aiot --ref ts-downsample.md

# 更新 Skill
kwcli skill update kwdb-text2sql-aiot

# 卸载 Skill
kwcli skill uninstall kwdb-text2sql-aiot

# 桥接到 AI Agent
kwcli skill link kwdb-text2sql-aiot --agent claude
kwcli skill link kwdb-text2sql-aiot --agent codex
```

**Skill 存储结构**：

```
~/.kwcli/skills/
└── kwdb-text2sql-aiot/
    ├── skill.yaml         # 元数据
    ├── README.md
    └── references/
        ├── ts-ddl.md
        ├── ts-downsample.md
        ├── ts-interpolate.md
        └── ts-window.md
```

**skill.yaml 元数据格式**：

```yaml
name: kwdb-text2sql-aiot
version: 1.0.0
description: "面向 AIoT 场景的自然语言转 KWDB SQL"
author: KWDB
triggers:
  - "查询 KWDB"
  - "时序分析"
  - "生成 SQL"
references:
  - ts-ddl.md
  - ts-downsample.md
commands:                      # 可选：命令模板
  - name: downsample
    description: "时序数据降采样"
    template: |
      SELECT time_bucket('{{interval}}', ts) AS bucket,
             avg({{metric}}) AS avg_{{metric}}
      FROM {{table}}
      WHERE ts > now() - INTERVAL '{{range}}'
      GROUP BY bucket
      ORDER BY bucket;
```

> **实现位置**：
> - `cmd/skill.go` — CLI 命令入口
> - `pkg/skill/manager.go` — Skill 安装/卸载/列表/信息
> - `pkg/skill/parser.go` — skill.yaml 解析
> - `pkg/skill/linker.go` — Agent 桥接
> - `pkg/skill/runner.go` — 命令模板渲染执行

---

### 2.7 MCP Server（Phase 3 - 全新）

MCP（Model Context Protocol）服务器，将 kwcli 的 10 个工具暴露给 AI Agent 调用。

```bash
# stdio 模式（本地 Agent 调用，默认）
kwcli mcp serve

# SSE 模式（远程 Agent 连接）
kwcli mcp serve --transport sse --port 8080

# 查看暴露的工具清单
kwcli mcp tools
```

**暴露的工具清单**：

| 工具 | 说明 |
|------|------|
| `kwcli_sql_exec` | 执行 KWDB SQL 查询 |
| `kwcli_schema_dump` | 导出数据库 Schema |
| `kwcli_status` | 获取 KWDB 服务状态 |
| `kwcli_logs` | 获取 KWDB 日志 |
| `kwcli_tsbs_run` | 运行 TSBS 基准测试 |
| `kwcli_sampledb_init` | 初始化 SampleDB |
| `kwcli_skill_list` | 列出已安装的 Skills |
| `kwcli_skill_info` | 获取 Skill 详情 |
| `kwcli_perf_snapshot` | 收集性能快照 |
| `kwcli_inspect_run` | 运行数据库巡检 |

**典型调用链路**：

```
用户自然语言 "查询最近一小时温度异常的设备"
    ↓
Claude Code + kwdb-text2sql-aiot Skill
    ↓
Skill 规划: "1. 发现 schema  2. 生成 SQL  3. 执行查询"
    ↓
通过 MCP 调用 kwcli_schema_dump → kwcli_sql_exec
    ↓
结果返回给用户
```

> **实现位置**：
> - `cmd/mcp.go` — CLI 命令入口
> - `pkg/mcp/server.go` — MCP 协议实现（stdio + SSE）
> - `pkg/mcp/tools.go` — 工具注册

---

### 2.8 AI 模式（Phase 4 - 全新）

kwcli 内置轻量 Agent，直接消费 Skills，让用户通过自然语言操作数据库。

```bash
# 单次提问
kwcli ai "查询最近一小时温度超过 40 度的设备"
kwcli ai "帮我部署一个单节点的 KWDB"

# 指定 Skill
kwcli ai --skill kwdb-text2sql-aiot "分析性能瓶颈"

# 只生成命令不执行（dry-run）
kwcli ai --dry-run "创建一个时序库 iot_db 和表 sensor_data"

# 多轮对话模式
kwcli ai --interactive
```

**配置 LLM**：

```bash
# 配置 OpenAI
kwcli config set llm.provider openai
kwcli config set llm.api_key sk-xxx
kwcli config set llm.model gpt-4.1
kwcli config set llm.base_url https://api.openai.com/v1

# 配置 DeepSeek
kwcli config set llm.provider deepseek
kwcli config set llm.api_key sk-xxx
kwcli config set llm.model deepseek-chat

# 查看当前配置
kwcli config show llm
```

**护栏设计**：

1. **Schema 发现优先** — 执行 SQL 前先获取真实表结构
2. **读写分离** — SELECT/SHOW/EXPLAIN 自动执行；INSERT/CREATE/ALTER/DROP 需用户确认
3. **危险操作双重确认** — DROP DATABASE/TRUNCATE 等操作需用户输入 "yes" 确认
4. **失败不自动重试** — 执行失败后读取日志，报告原因，不擅自重试
5. **参数不猜测** — 端口、IP、目录等必须由用户确认

**执行流程**：

```
用户输入: "查询最近一小时温度异常的设备"
    ↓
kwcli ai 加载 ~/.kwcli/skills/kwdb-text2sql-aiot/references/*.md
    ↓
构造 system prompt（Skill 知识 + kwcli 能力说明）
    ↓
调用 LLM API（OpenAI / DeepSeek）
    ↓
LLM 返回: 计划 + SQL
    ↓
kwcli 解析 SQL → 应用护栏
    ↓
只读 SQL 自动执行 | 写 SQL 询问用户确认 | dry-run 仅展示
    ↓
结果格式化输出
```

> **实现位置**：
> - `cmd/ai.go` — CLI 命令入口
> - `pkg/ai/client.go` — LLM API 客户端（兼容 OpenAI 格式）
> - `pkg/ai/prompt.go` — Prompt 构造器（自动加载 Skill 参考文件）
> - `pkg/ai/executor.go` — 护栏与命令执行控制

---

### 2.9 配置扩展

`~/.kwcli/config.yaml` 新增字段：

```yaml
source: atomgit          # 代码源
registry: auto           # 镜像源

# Phase 4 - LLM 配置
llm:
  provider: openai
  api_key: sk-...
  model: gpt-4.1
  base_url: https://api.openai.com/v1
  timeout: 60s

# Phase 3 - MCP 配置
mcp:
  transport: stdio       # stdio | sse
  port: 8080

# Phase 5 - 巡检配置
inspect:
  default_template: standard
  output_format: markdown
```

## 三、项目结构

```
kwcli/
├── cmd/
│   ├── root.go           # 根命令，全局配置
│   ├── general.go        # 组件管理（list/install/update/uninstall）
│   ├── sql.go            # SQL 连接与查询
│   ├── schema.go         # Schema 导出
│   ├── kwdb.go           # KWDB 服务管理
│   ├── playground.go     # Playground 管理
│   ├── sampledb.go       # SampleDB 智能电表模型
│   ├── tsbs.go           # TSBS 基准测试
│   ├── perf.go           # 性能快照
│   ├── inspect.go        # 数据库巡检
│   ├── skill.go          # Skill 管理（Phase 2）
│   ├── mcp.go            # MCP Server（Phase 3）
│   ├── ai.go             # AI 模式（Phase 4）
│   ├── completion.go     # Shell 自动补全
│   └── cmd_test.go       # 命令测试
├── pkg/
│   ├── skill/
│   │   ├── manager.go    # Skill 安装/卸载/列表
│   │   ├── parser.go     # skill.yaml 解析
│   │   ├── linker.go     # Agent 桥接
│   │   └── runner.go     # 模板解析执行
│   ├── mcp/
│   │   ├── server.go     # MCP 协议实现
│   │   └── tools.go      # 工具注册
│   ├── ai/
│   │   ├── client.go     # LLM API 客户端
│   │   ├── prompt.go     # Prompt 构造
│   │   └── executor.go   # 护栏与命令执行
│   ├── output/
│   │   ├── output.go     # JSON/Text 格式化
│   │   └── export.go     # CSV/JSON 导出
│   ├── config/
│   ├── component/
│   ├── sampledb/
│   ├── schema/
│   ├── perf/
│   ├── inspect/
│   └── utils/
└── go.mod
```

## 四、技术实现要点

### 4.1 关键依赖

| 功能 | 依赖 |
|------|------|
| CLI 框架 | `github.com/spf13/cobra` |
| 配置管理 | `github.com/spf13/viper` |
| YAML 解析 | `gopkg.in/yaml.v3` |
| 数据导出 | 标准库 `encoding/csv`, `encoding/json` |

### 4.2 JSON 输出模式

在 Cobra 命令的 `RunE` 中增加输出格式化层：

```go
import "github.com/shawn0915/kwcli/pkg/output"

// 在命令中使用
if output.JSONOutput {
    output.PrintJSON("command name", data, nil)
    return
}
output.PrintJSON("command name", data, err) // err != nil 时失败
```

全局 `--json` 标志在 `root.go` 中定义：

```go
rootCmd.PersistentFlags().BoolVar(&output.JSONOutput, "json", false, "Output in JSON format")
```

### 4.3 MCP 协议

MCP 基于 JSON-RPC 2.0 协议，支持两种传输方式：

- **stdio 模式**：stdin/stdout 传输 JSON-RPC 消息，适用于本地 Agent
- **SSE 模式**：HTTP Server-Sent Events，适用于远程 Agent 连接

工具调用流程：

```json
// 请求（Agent → kwcli）
{"method": "call_tool", "params": {"name": "kwcli_sql_exec", "arguments": {"sql": "SELECT 1"}}}

// 响应（kwcli → Agent）
{"jsonrpc": "2.0", "result": {"tool": "kwcli_sql_exec", "command": "kwcli sql -e \"SELECT 1\" --json", "status": "simulated"}}
```

### 4.4 AI 护栏机制

所有通过 AI 模式执行的 SQL 都经过护栏检查：

| SQL 类型 | 操作 | 行为 |
|---------|------|------|
| SELECT / SHOW / EXPLAIN | 只读 | 自动执行 |
| INSERT / CREATE / ALTER | 写操作 | 展示 SQL → 用户确认 |
| DROP / TRUNCATE | 危险操作 | 展示警告 → 用户输入 "yes" 确认 |
| 其他 | 未知 | 展示 SQL → 用户确认 |

## 五、开发与测试

### 5.1 本地构建

```bash
# 克隆仓库
git clone https://github.com/shawn0915/kwcli.git
cd kwcli

# 构建
make

# 运行测试
make test
# 或
go test ./... -v
```

### 5.2 添加新命令

1. 在 `cmd/` 目录下创建新的 `.go` 文件
2. 使用 Cobra 定义命令结构
3. 在 `init()` 中注册到 `rootCmd`
4. 如果命令有输出数据，使用 `pkg/output` 的 JSON 格式化
5. 添加测试用例到 `cmd/cmd_test.go` 或创建独立的 `*_test.go`

### 5.3 测试覆盖

所有新功能都有对应的单元测试：

```bash
# 运行所有测试
go test ./...

# 查看测试覆盖
go test ./... -cover

# 运行特定包的测试
go test ./pkg/skill/... -v
go test ./pkg/mcp/... -v
go test ./pkg/ai/... -v
go test ./cmd/... -v
```

### 5.4 代码风格

项目使用 Go 标准格式化工具：

```bash
go fmt ./...
```

## 六、特别注意

1. **LLM API 配置**：使用 AI 模式前必须先配置 LLM 提供商和 API Key，否则会提示配置。建议将 API Key 保存在 `~/.kwcli/config.yaml` 中，避免每次输入。

2. **MCP 端口冲突**：SSE 模式默认使用 8080 端口，如果已被占用请通过 `--port` 指定其他端口。

3. **Skill 来源**：目前 Skill 安装使用模拟仓库结构（创建本地 skill.yaml）。未来计划直接集成 GitHub/AtomGit 仓库克隆，实现真正的 Skill 分发。

4. **AI 模式安全检查**：护栏机制虽然能防止误操作，但建议在开发/测试环境中先使用 `--dry-run` 验证生成的 SQL。

5. **Shell 补全**：安装新命令后，重新生成 Shell 补全脚本以确保自动补全功能正常。

6. **数据安全**：LLM API 调用会将用户问题和数据库 Schema 发送到 LLM 提供商服务器。请确保在使用 AI 模式时，不要在查询中附带敏感数据。

## 七、总结

kwcli v0.2.0 是一次重要的架构升级，从单纯的 CLI 工具进化为一套完整的 **KWDB AI 生态执行底座**。通过 Skill 管理、MCP 协议和 AI 模式三大能力，kwcli 能够：

- **作为 Agent 的标准执行器**：MCP Server 暴露 10 个工具，AI Agent 可零配置调用
- **统一管理技能包**：Skill 系统让 Agent 获得 KWDB 领域知识
- **支持自然语言交互**：AI 模式让不熟悉 SQL 的用户也能轻松操作数据库

未来计划包括：组件清单与版本索引、离线镜像与私有化部署支持、Homebrew 一键安装等。欢迎社区贡献！

---

*本文档对应 kwcli v0.2.0 及以上版本*

# kwcli 改进开发计划

> 基于 KWDB Agent Skills 生态与 CLI 工具演进需求，为继续开发 kwcli 提供路线图。

---

## 一、当前状态（v0.1.1）

kwcli 已具备基础能力：

| 模块 | 功能 |
|------|------|
| **全局配置** | 代码源切换（GitHub/AtomGit）、镜像源配置 |
| **Playground** | 一键安装/启动/停止/升级/日志 |
| **KWDB 服务** | Docker 方式安装、启停、状态、日志、配置管理 |
| **SQL 连接** | 交互式 Shell + 单条执行（`kwcli sql -e`） |
| **SampleDB** | 智能电表模型：初始化、数据生成、场景查询 |
| **TSBS 基准测试** | 内置时序压测工具链（init/load/run/list/clean） |
| **组件管理** | 组件化架构，统一入口 `kwcli install/update/uninstall` |
| **Shell 补全** | Bash/Zsh/Fish/PowerShell 自动补全 |

**当前缺失**：AI 能力集成、Skill 管理、MCP 协议支持、JSON 输出模式、Schema 导出等。

---

## 二、改进目标

1. **成为 KWDB AI 生态的执行底座**：让 kwcli 成为 Agent Skills 的标准执行器
2. **统一 Skill 管理入口**：安装、更新、查看、运行 Skills
3. **支持多模式 AI 调用**：内置 Agent、MCP Server、模板解析三模式并存
4. **增强可编程性**：JSON 输出、批量执行、Schema 导出，便于脚本和 Agent 消费
5. **完善数据库运维能力**：巡检、性能快照、日志分析

---

## 三、分阶段实施计划

### Phase 1：基础增强（1-2 周）

目标：让 kwcli 的输出可被机器稳定解析，为 Agent 集成做准备。

#### 3.1.1 全局 JSON 输出模式

所有子命令支持 `--format json` 或 `--json` 标志：

```bash
kwcli kwdb status --json
kwcli sql -e "SELECT 1" --json
kwcli sampledb list --json
kwcli tsbs list --json
```

输出格式统一：

```json
{
  "success": true,
  "command": "kwdb status",
  "data": { ... },
  "error": null
}
```

**实现位置**：在 Cobra 命令的 `RunE` 中增加输出格式化层，统一封装响应结构。

#### 3.1.2 Schema 导出

```bash
# 导出指定数据库的完整 DDL
kwcli schema dump --db iot_db --output iot_db_schema.sql

# 导出为 JSON 格式（含表结构、索引、标签信息）
kwcli schema dump --db iot_db --format json
```

**用途**：Agent Skills 做 schema 发现时，可以一次性获取完整结构，减少多次查询。

#### 3.1.3 批量 SQL 执行

```bash
# 执行 SQL 脚本文件
kwcli sql -f init.sql

# 支持事务控制
kwcli sql -f migrate.sql --transaction
```

#### 3.1.4 查询结果导出

```bash
kwcli sql -e "SELECT * FROM sensor_data LIMIT 100" --export csv --output result.csv
kwcli sql -e "SELECT * FROM sensor_data LIMIT 100" --export json
```

---

### Phase 2：Skill 管理（2-3 周）

目标：让 kwcli 成为 KWDB Agent Skills 的统一管理入口。

#### 3.2.1 Skill 子命令设计

```bash
# 查看已安装 Skills
kwcli skill list

# 安装 Skill（从 GitHub/AtomGit）
kwcli skill install kwdb-text2sql-aiot
kwcli skill install kwdb-text2sql-aiot --version v1.0.1
kwcli skill install kwdb-text2sql-aiot --source atomgit

# 更新 Skill
kwcli skill update kwdb-text2sql-aiot

# 卸载 Skill
kwcli skill uninstall kwdb-text2sql-aiot

# 查看 Skill 详情
kwcli skill info kwdb-text2sql-aiot

# 查看 Skill 支持的场景/命令模板
kwcli skill scenarios kwdb-text2sql-aiot

# 预览 Skill 参考文件
kwcli skill preview kwdb-text2sql-aiot --ref ts-downsample.md
```

#### 3.2.2 Skill 存储结构

```
~/.kwcli/
├── bin/
├── components/
├── skills/                    # 新增
│   ├── kwdb-text2sql-aiot/
│   │   ├── skill.yaml         # 元数据
│   │   ├── README.md
│   │   └── references/
│   │       ├── ts-ddl.md
│   │       ├── ts-downsample.md
│   │       ├── ts-interpolate.md
│   │       └── ts-window.md
│   └── kwdb-install-deploy/
│       └── ...
├── data/
└── config.yaml
```

#### 3.2.3 Skill 元数据解析

解析 `skill.yaml`：

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
  - ts-interpolate.md
commands:                       # 可选：命令模板（模式三用）
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

#### 3.2.4 与 Agent 工具的桥接

安装 Skill 后，自动注册到本地 Agent：

| Agent | 桥接方式 |
|-------|---------|
| Claude Code | 复制到 `~/.codex/skills/` 或调用 `claude plugin install` |
| Codex | 写入 `skills-lock.json` 或复制到 Codex skills 目录 |
| OpenClaw | 调用 `clawhub install` 或复制到 clawhub 目录 |
| Kimi Code/Claw | 复制到 `.kimi/skills/`（如支持） |

```bash
# 手动触发桥接
kwcli skill link kwdb-text2sql-aiot --agent claude
kwcli skill link kwdb-text2sql-aiot --agent codex
```

---

### Phase 3：MCP Server（2 周）

目标：让 kwcli 成为 Agent Skills 的标准执行底座。

#### 3.3.1 启动 MCP Server

```bash
# 标准输入输出模式（本地 Agent 调用）
kwcli mcp serve

# SSE 模式（远程 Agent 连接）
kwcli mcp serve --transport sse --port 8080

# 查看暴露的工具列表
kwcli mcp tools
```

#### 3.3.2 暴露的工具清单

```yaml
tools:
  kwcli_sql_exec:
    description: 执行 KWDB SQL 查询
    command: kwcli sql -e "{{sql}}" -d {{db}} -u {{user}} --json

  kwcli_schema_dump:
    description: 导出数据库 Schema
    command: kwcli schema dump --db {{db}} --format json

  kwcli_status:
    description: 获取 KWDB 服务状态
    command: kwcli kwdb status --json

  kwcli_logs:
    description: 获取 KWDB 日志
    command: kwcli kwdb logs --tail {{n}}

  kwcli_tsbs_run:
    description: 运行 TSBS 基准测试
    command: kwcli tsbs run --use-case {{case}} --json

  kwcli_sampledb_init:
    description: 初始化 SampleDB
    command: kwcli sampledb init

  kwcli_skill_list:
    description: 列出已安装的 Skills
    command: kwcli skill list --json

  kwcli_skill_info:
    description: 获取 Skill 详情
    command: kwcli skill info {{name}} --json
```

#### 3.3.3 调用链路

```
用户自然语言
    ↓
Claude Code + kwdb-text2sql-aiot Skill
    ↓
Skill 规划: "1. 发现 schema  2. 生成 SQL  3. 执行查询"
    ↓
Claude Code 调用 MCP: kwcli_schema_dump
    ↓
kwcli schema dump --db iot_db --format json
    ↓
结果返回 Agent → 生成 SQL → 调用 kwcli_sql_exec
    ↓
kwcli sql -e "SELECT ..." --json
    ↓
结果返回给用户
```

---

### Phase 4：AI 模式（3-4 周）

目标：kwcli 内置轻量 Agent，直接消费 Skills。

#### 3.4.1 kwcli ai 子命令

```bash
# 进入 AI 交互模式（自动加载所有已安装 Skills）
kwcli ai

# 单次提问
kwcli ai "查询最近一小时温度超过 40 度的设备"
kwcli ai "帮我部署一个单节点的 KWDB"
kwcli ai "分析一下当前数据库的慢查询"

# 指定 Skill
kwcli ai --skill kwdb-performance-review "分析性能瓶颈"

# 只生成命令不执行（dry-run）
kwcli ai --dry-run "创建一个时序库 iot_db 和表 sensor_data"

# 多轮对话模式
kwcli ai --interactive
```

#### 3.4.2 配置 LLM

```bash
# 配置 LLM 提供商
kwcli config set llm.provider openai
kwcli config set llm.api_key sk-...
kwcli config set llm.model gpt-4.1
kwcli config set llm.base_url https://api.openai.com/v1

# 支持多个提供商
kwcli config set llm.provider deepseek
kwcli config set llm.model deepseek-chat

# 查看当前配置
kwcli config show llm
```

#### 3.4.3 执行流程

```
用户输入: "查询最近一小时温度异常的设备"
    ↓
kwcli ai 加载 ~/.kwcli/skills/kwdb-text2sql-aiot/references/*.md
    ↓
构造 system prompt（Skill 知识 + kwcli 能力说明）
    ↓
调用 LLM API
    ↓
LLM 返回: 计划 + SQL
    ↓
kwcli 自动执行: kwcli sql -e "SELECT ..." --json
    ↓
结果格式化输出
```

#### 3.4.4 护栏设计（与 Skills 一致）

- **Schema 发现优先**：执行 SQL 前先获取真实表结构
- **读写分离**：`SELECT/SHOW/EXPLAIN` 直接执行；`INSERT/CREATE/ALTER/DROP` 需用户确认
- **失败不自动重试**：执行失败后读取日志，报告原因，不擅自重试
- **参数不猜测**：端口、IP、目录等必须由用户确认

---

### Phase 5：数据库运维增强（2-3 周）

#### 3.5.1 性能快照

```bash
# 一键收集当前性能指标
kwcli perf snapshot

# 输出包含：
# - QPS / TPS
# - 活跃连接数
# - 慢查询列表（TOP 10）
# - 存储空间使用
# - 缓存命中率
```

#### 3.5.2 智能巡检

```bash
# 运行标准巡检模板
kwcli inspect run --template standard

# 自定义巡检项
kwcli inspect run --items status,logs,connections,slow_queries

# 输出巡检报告（Markdown/JSON）
kwcli inspect run --output report.md --format markdown
```

巡检项对应：

| 巡检项 | kwcli 命令 | 说明 |
|--------|-----------|------|
| 服务状态 | `kwcli kwdb status` | 是否正常运行 |
| 日志异常 | `kwcli kwdb logs` | ERROR/FATAL 关键字扫描 |
| 连接数 | `kwcli sql -e "SHOW SESSIONS"` | 是否接近上限 |
| 慢查询 | `kwcli sql -e "SHOW QUERIES"` | 长时间运行查询 |
| 存储空间 | Docker/系统命令 | 磁盘使用率 |
| SampleDB 健康 | `kwcli sampledb status` | 测试环境可用性 |

#### 3.5.3 慢 SQL 分析

```bash
# 分析指定 SQL 的执行计划
kwcli sql -e "EXPLAIN SELECT ..." --json

# 或封装为专用命令
kwcli analyze sql "SELECT * FROM sensor_data WHERE device_id = 'xxx'"
```

---

## 四、技术实现要点

### 4.1 项目结构建议

```
kwcli/
├── cmd/
│   ├── root.go
│   ├── ai.go              # Phase 4
│   ├── skill.go           # Phase 2
│   ├── mcp.go             # Phase 3
│   ├── schema.go          # Phase 1
│   ├── perf.go            # Phase 5
│   ├── inspect.go         # Phase 5
│   └── ...
├── pkg/
│   ├── skill/
│   │   ├── manager.go     # Skill 安装/卸载/列表
│   │   ├── parser.go      # skill.yaml 解析
│   │   ├── linker.go      # Agent 桥接
│   │   └── runner.go      # 模板解析执行
│   ├── mcp/
│   │   ├── server.go      # MCP Server 实现
│   │   └── tools.go       # 工具注册
│   ├── ai/
│   │   ├── client.go      # LLM API 客户端
│   │   ├── prompt.go      # Prompt 构造
│   │   └── executor.go    # 命令执行与确认
│   ├── output/
│   │   └── formatter.go   # JSON/CSV/Table 格式化
│   └── ...
├── internal/
│   └── ...
└── go.mod
```

### 4.2 关键依赖

| 功能 | 推荐依赖 |
|------|---------|
| MCP Server | `github.com/mark3labs/mcp-go` |
| LLM 调用 | `github.com/sashabaranov/go-openai`（兼容 OpenAI 格式） |
| YAML 解析 | `gopkg.in/yaml.v3` |
| CSV 导出 | 标准库 `encoding/csv` |
| JSON 输出 | 标准库 `encoding/json` |

### 4.3 配置扩展

`~/.kwcli/config.yaml` 新增字段：

```yaml
source: atomgit
registry: auto

# Phase 4 新增
llm:
  provider: openai
  api_key: sk-...
  model: gpt-4.1
  base_url: https://api.openai.com/v1
  timeout: 60s

# Phase 3 新增
mcp:
  transport: stdio          # stdio | sse
  port: 8080

# Phase 5 新增
inspect:
  default_template: standard
  output_format: markdown
```

---

## 五、与 KWDB Agent Skills 的协同关系

```
┌─────────────────────────────────────────────────────────────┐
│                      用户层                                   │
│  自然语言提问  /  CLI 命令  /  脚本调用                         │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                   Agent 层（可选）                             │
│  Claude Code / Codex / Kimi Claw + KWDB Agent Skills          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ text2sql    │  │ install     │  │ performance-review  │  │
│  │ -aiot       │  │ -deploy     │  │ -troubleshooting    │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              ↓  MCP 协议
┌─────────────────────────────────────────────────────────────┐
│                   kwcli 执行层                                │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐   │
│  │ mcp      │ │ sql      │ │ schema   │ │ kwdb         │   │
│  │ serve    │ │ -e/-f    │ │ dump     │ │ install/...  │   │
│  └──────────┘ └──────────┘ └──────────┘ └──────────────┘   │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐   │
│  │ sampledb │ │ tsbs     │ │ skill    │ │ ai           │   │
│  │ init/run │ │ init/run │ │ install  │ │ ask/...      │   │
│  └──────────┘ └──────────┘ └──────────┘ └──────────────┘   │
│  ┌──────────┐ ┌──────────┐                                  │
│  │ perf     │ │ inspect  │                                  │
│  │ snapshot │ │ run      │                                  │
│  └──────────┘ └──────────┘                                  │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                   KWDB 实例层                                 │
│         Docker / 二进制 / Playground / 远程集群                 │
└─────────────────────────────────────────────────────────────┘
```

---

## 六、开发优先级建议

| 优先级 | 阶段 | 功能 | 价值 |
|--------|------|------|------|
| P0 | Phase 1 | JSON 输出模式 | Agent 解析必备 |
| P0 | Phase 1 | Schema 导出 | 减少 Agent 多次查询 |
| P0 | Phase 3 | MCP Server | 成为 Agent 标准执行底座 |
| P1 | Phase 2 | Skill 管理（install/list/info） | 统一 Skill 入口 |
| P1 | Phase 2 | Skill Agent 桥接 | 降低用户配置成本 |
| P1 | Phase 4 | kwcli ai 基础模式 | 直接体验 AI 能力 |
| P2 | Phase 1 | 批量 SQL / 结果导出 | 提升可编程性 |
| P2 | Phase 4 | LLM 配置多提供商 | 适配国内环境 |
| P2 | Phase 5 | perf snapshot | 性能诊断基础 |
| P3 | Phase 5 | inspect 巡检 | 自动化运维 |
| P3 | Phase 2 | Skill 模板解析（模式三） | 离线场景 |

---

## 七、验收标准

1. **Agent 可以零配置调用 kwcli**：安装 kwdb-text2sql-aiot Skill 后，Claude Code 能通过 MCP 直接调用 kwcli 完成 schema 发现 + SQL 执行
2. **用户可以通过 kwcli 管理 Skills**：`kwcli skill install/list/uninstall` 工作正常
3. **JSON 输出覆盖所有核心命令**：`status`、`sql`、`sampledb list`、`tsbs list` 均支持 `--json`
4. **AI 模式可用**：`kwcli ai "查询最近一小时温度异常的设备"` 能返回正确结果
5. **巡检可输出报告**：`kwcli inspect run` 生成结构化 Markdown 报告

---

*文档生成时间：2026-05-21*
*适用版本：kwcli v0.1.1 及后续版本*


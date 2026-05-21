# KWCLI

[![Version](https://img.shields.io/badge/version-v0.1.2-blue)](https://github.com/shawn0915/kwcli)

KWCLI 是 KWDB 生态的命令行工具，采用组件化架构设计，帮助你一键安装、运行和管理 KWDB 相关组件。

**当前版本**：v0.1.2  
**开发语言**：Go (v1.25+)

## 特性

- 🧩 **组件化管理**：每个功能都是独立组件，按需安装、灵活扩展
- 🚀 **一键启动**：`kwcli playground start` 秒级启动 Playground
- 🐳 **开箱即用**：自动处理依赖与环境检查
- 💻 **跨平台**：基于 Go 构建，支持 Linux / macOS / Windows
- 🌐 **国内加速**：默认优先阿里云镜像仓库，代码源默认 AtomGit
- ⚙️ **全局配置**：支持一键切换默认代码源和镜像源
- 📊 **SampleDB**：内置智能电表模型，一键初始化 schema、生成数据、运行场景查询
- 📈 **TSBS 基准测试**：内置 KWDB 时序数据库性能测试工具，无需额外安装
- 🔍 **数据库巡检**：一键健康检查，输出 Markdown / JSON / HTML 报告
- ⚡ **性能快照**：实时采集 QPS/TPS、连接数、慢查询等性能指标
- 📐 **Schema 导出**：导出数据库 DDL 或 JSON 格式的表结构、索引、标签信息
- 📤 **查询结果导出**：SQL 查询结果支持 CSV / JSON 导出
- 🤖 **JSON 输出**：全局 `--json` 标志，所有命令输出可被机器稳定解析
## 安装

### 一键安装（推荐）

```bash
# 下载并执行安装脚本
curl -sL https://raw.githubusercontent.com/shawn0915/kwcli/main/install.sh | bash

# 或保存后执行
curl -sL https://raw.githubusercontent.com/shawn0915/kwcli/main/install.sh -o install.sh
chmod +x install.sh
./install.sh
```

**安装说明：**
- **RedHat/CentOS 7+**：自动安装到 `~/.kwcli/bin`，并追加 PATH 到 `~/.bashrc`
- **其他 Linux/macOS**：可选择安装到 `/usr/local/bin`（需要 sudo）
- **TSBS 二进制**：安装包已内置 `tsbs_generate_data`、`tsbs_generate_queries`、`tsbs_load_kwdb`、`tsbs_run_queries_kwdb`，无需额外安装

### 使用 Makefile

```bash
# 克隆仓库
git clone https://github.com/shawn0915/kwcli.git
cd kwcli

# 构建（自动注入编译时间）
make

# 安装到 ~/.kwcli/bin
make install

# 或安装到系统目录（需要 sudo）
make install-g
```

## 快速开始

### 全局配置

```bash
# 查看当前默认代码源（默认 atomgit）
kwcli source

# 切换到 GitHub
kwcli source github
```

### Shell 自动补全

```bash
# Bash
source <(kwcli completion bash)

# Zsh
kwcli completion zsh > "${fpath[1]}/_kwcli"

# Fish
kwcli completion fish | source

# PowerShell
kwcli completion powershell | Out-String | Invoke-Expression
```

### 启动 Playground

```bash
# 一键启动 KWDB Playground
kwcli playground start

# 指定版本安装
kwcli playground install -v v1.1.0

# 升级 Playground（代码源和镜像源可分别指定）
kwcli playground upgrade --source atomgit --registry auto
```

启动完成后，打开浏览器访问 http://localhost:3006 即可开始学习。

### 安装 KWDB

```bash
# 安装 KWDB（自动选择 Docker 方式）
kwcli kwdb install

# 启动 KWDB
kwcli kwdb start
```

### 体验 SampleDB（智能电表模型）

```bash
# 初始化数据库和表结构
kwcli sampledb init

# 生成示例数据（100 个电表，10000 条读数）
kwcli sampledb generate

# 查看所有场景
kwcli sampledb list

# 运行指定场景
kwcli sampledb run top10-area-energy
kwcli sampledb run fault-meters

# 运行所有场景
kwcli sampledb run --all

# 清理所有 SampleDB 数据
kwcli sampledb clean
```

## 命令手册

### 全局配置

| 命令 | 说明 |
|------|------|
| `kwcli source` | 查看当前默认代码源 |
| `kwcli source github` | 设置默认代码源为 GitHub |
| `kwcli source atomgit` | 设置默认代码源为 AtomGit |

### Playground 相关

| 命令 | 说明 |
|------|------|
| `kwcli playground install` | 安装 Playground（支持 `-v` 指定版本，`--source` 代码源） |
| `kwcli playground start` | 启动 Playground 服务（首次会自动下载） |
| `kwcli playground stop` | 停止 Playground 服务 |
| `kwcli playground status` | 查看 Playground 运行状态 |
| `kwcli playground restart` | 重启 Playground 服务 |
| `kwcli playground upgrade` | 升级 Playground（支持 `--source` 代码源，`--registry` 镜像源） |
| `kwcli playground logs` | 查看 Playground 日志 |
| `kwcli playground versions` | 查看已安装的 Playground 版本 |
| `kwcli playground uninstall` | 卸载 Playground（加 `--remove-files` 彻底删除） |

### SampleDB 相关

| 命令 | 说明 |
|------|------|
| `kwcli sampledb init` | 创建智能电表数据库和表结构 |
| `kwcli sampledb generate` | 生成示例数据 |
| `kwcli sampledb list` | 列出所有场景查询 |
| `kwcli sampledb run <name>` | 运行指定场景查询 |
| `kwcli sampledb run --all` | 运行所有场景查询 |
| `kwcli sampledb clean` | 清理所有 SampleDB 数据 |
| `kwcli sampledb status` | 检查 SampleDB 是否存在 |

### TSBS 基准测试

TSBS (Time-Series Benchmark Suite) 是 KWDB 内置的时序数据库性能测试工具，已经完整集成到 kwcli 中，无需额外安装。

| 命令 | 说明 |
|------|------|
| `kwcli tsbs init` | 初始化基准测试（生成测试数据和查询） |
| `kwcli tsbs load` | 加载数据到 KWDB |
| `kwcli tsbs run` | 运行查询基准测试 |
| `kwcli tsbs list` | 列出可用的查询类型 |
| `kwcli tsbs clean` | 清理测试文件及数据库 |

**使用示例**：
```bash
# 初始化基准测试（生成数据和查询）
kwcli tsbs init --use-case=cpu-only --scale=10 --queries=1000

# 查看可用查询类型
kwcli tsbs list

# 加载数据到 KWDB
kwcli tsbs load --file=/tmp/tsbs_data --host=127.0.0.1 --port=26257 --user=root --password=root

# 运行查询基准测试
kwcli tsbs run --file=/tmp/tsbs_queries --host=127.0.0.1 --port=26257

# 清理测试文件及数据库
kwcli tsbs clean --drop-db
```

**参数说明**：
- `--use-case`: 测试场景 (cpu-only, devops, iot)
- `--scale`: 设备数量
- `--queries`: 生成的查询数量
- `--timestamp-start/end`: 测试数据时间范围

**内置测试场景**：

| Use Case | Query Type | 说明 |
|----------|------------|------|
| **cpu-only** | single-groupby-1-1-12 | 单分组聚合，1个主机，1个指标，12小时 |
| | single-groupby-5-1-12 | 单分组聚合，5个主机，1个指标，12小时 |
| | double-groupby-5 | 双分组聚合，5个指标 |
| | high-cpu-1 | 高CPU查询，1个主机 |
| | single-groupby-1-1-1 | 单分组聚合，1个主机，1个指标，1小时 |
| | cpu-max-all-1 | CPU最大值的最大值，1个主机 |
| | double-groupby-all | 双分组聚合，全部指标 |
| | single-groupby-5-8-1 | 单分组聚合，5个主机，8个指标，1小时 |
| | cpu-max-all-32-24 | CPU最大值，32个主机，24小时 |
| | double-groupby-1 | 双分组聚合，1个指标 |
| | lastpoint | 最后一点查询 |
| | single-groupby-1-8-1 | 单分组聚合，1个主机，8个指标，1小时 |
| | single-groupby-5-1-1 | 单分组聚合，5个主机，1个指标，1小时 |
| | cpu-max-all-8 | CPU最大值的最大值，8个主机 |
| | groupby-orderby-limit | 分组排序限制查询 |
| | high-cpu-all | 高CPU查询，全部主机 |
| **devops** | single-groupby-5-8-1 | 单分组聚合，5个主机，8个指标，1小时 |
| | cpu-max-all-32-24 | CPU最大值，32个主机，24小时 |
| | double-groupby-1 | 双分组聚合，1个指标 |
| | lastpoint | 最后一点查询 |
| | single-groupby-1-8-1 | 单分组聚合，1个主机，8个指标，1小时 |
| | single-groupby-5-1-1 | 单分组聚合，5个主机，1个指标，1小时 |
| | cpu-max-all-8 | CPU最大值的最大值，8个主机 |
| | groupby-orderby-limit | 分组排序限制查询 |
| | high-cpu-all | 高CPU查询，全部主机 |
| | single-groupby-1-1-12 | 单分组聚合，1个主机，1个指标，12小时 |
| | single-groupby-5-1-12 | 单分组聚合，5个主机，1个指标，12小时 |
| | double-groupby-5 | 双分组聚合，5个指标 |
| | high-cpu-1 | 高CPU查询，1个主机 |
| | single-groupby-1-1-1 | 单分组聚合，1个主机，1个指标，1小时 |
| | cpu-max-all-1 | CPU最大值的最大值，1个主机 |
| | double-groupby-all | 双分组聚合，全部指标 |
| **iot** | last-loc | 最后位置查询 |
| | high-load | 高负载查询 |
| | long-daily-sessions | 长时间日常会话 |
| | avg-daily-driving-duration | 平均每日驾驶时长 |
| | daily-activity | 日常活动 |
| | breakdown-frequency | 故障频率 |
| | single-last-loc | 单设备最后位置 |
| | low-fuel | 低燃料查询 |
| | stationary-trucks | 静止卡车 |
| | long-driving-sessions | 长时间驾驶会话 |
| | avg-vs-projected-fuel-consumption | 平均vs预计燃油消耗 |
| | avg-daily-driving-session | 平均每日驾驶会话 |
| | avg-load | 平均负载 |

### SQL 连接

| 命令 | 说明 |
|------|------|
| `kwcli sql` | 交互式连接 KWDB |
| `kwcli sql -e "SELECT 1"` | 执行单条 SQL |
| `kwcli sql -u <user> -d <db>` | 指定用户和数据库 |
| `kwcli sql -f init.sql` | 执行 SQL 脚本文件 |
| `kwcli sql -f migrate.sql --transaction` | 以事务方式执行 SQL 脚本 |
| `kwcli sql -e "SELECT ..." --export csv --output result.csv` | 导出查询结果为 CSV |
| `kwcli sql -e "SELECT ..." --export json` | 导出查询结果为 JSON |
| `kwcli sql -e "SELECT ..." --json` | 以 JSON 格式输出查询结果 |

### Schema 导出

| 命令 | 说明 |
|------|------|
| `kwcli schema dump --db rdb --output rdb_schema.sql` | 导出关系库 DDL |
| `kwcli schema dump --db tsdb --output tsdb_schema.sql` | 导出时序库 DDL |
| `kwcli schema dump --db all --output full_schema.sql` | 导出所有库 DDL |
| `kwcli schema dump --db rdb --format json` | 以 JSON 格式导出（含表结构、索引、标签信息） |

### 性能监控

| 命令 | 说明 |
|------|------|
| `kwcli perf snapshot` | 采集性能快照（QPS/TPS、连接数、慢查询、存储） |
| `kwcli perf snapshot --json` | 以 JSON 格式输出性能快照 |
| `kwcli perf snapshot --output snapshot.json --format json` | 保存性能快照到文件 |

### 数据库巡检

| 命令 | 说明 |
|------|------|
| `kwcli inspect run` | 运行标准巡检（6 项检查） |
| `kwcli inspect run --items status,logs,connections` | 运行指定巡检项 |
| `kwcli inspect run --output report.md --format markdown` | 输出 Markdown 巡检报告 |
| `kwcli inspect run --output report.html --format html` | 输出 HTML 巡检报告 |
| `kwcli inspect run --json` | 以 JSON 格式输出巡检结果 |

**巡检项**：

| 巡检项 | 说明 |
|--------|------|
| `status` | 服务状态（KWDB 是否正常运行） |
| `logs` | 日志异常（ERROR/FATAL 关键字扫描） |
| `connections` | 连接数（是否接近上限） |
| `slow_queries` | 慢查询 / 长时间运行查询 |
| `storage` | 存储空间使用率 |
| `sampledb` | SampleDB 健康检查 |

### KWDB 相关

| 命令 | 说明 |
|------|------|
| `kwcli kwdb install` | 安装 KWDB（Docker 方式） |
| `kwcli kwdb start` | 启动 KWDB 服务 |
| `kwcli kwdb stop` | 停止 KWDB 服务 |
| `kwcli kwdb restart` | 重启 KWDB 服务 |
| `kwcli kwdb status` | 查看 KWDB 状态 |
| `kwcli kwdb logs` | 查看 KWDB 日志 |
| `kwcli kwdb config show` | 查看当前配置 |
| `kwcli kwdb config edit` | 交互式编辑配置 |
| `kwcli kwdb config set <key> <value>` | 设置配置项 |
| `kwcli kwdb config path` | 显示配置文件路径 |

### 通用命令

| 命令 | 说明 |
|------|------|
| `kwcli list` | 查看可用组件列表 |
| `kwcli install <component>` | 安装指定组件（支持 `--source`） |
| `kwcli update <component>` | 升级指定组件（支持 `--source`） |
| `kwcli uninstall <component>` | 卸载指定组件 |
| `kwcli status` | 查看所有运行中的组件 |

## 架构简介

KWCLI 采用组件化架构设计：

1. **组件隔离**：所有组件安装在 `~/.kwcli/components/` 目录下，不影响系统环境
2. **按需获取**：组件在首次使用时自动从远程仓库拉取，无需预装
3. **统一入口**：通过 `kwcli <component>` 统一调度，简化操作
4. **多版本共存**：支持同一组件多个版本并存，灵活切换

```
~/.kwcli/
├── bin/                 # kwcli 及 TSBS 二进制
├── components/          # 组件安装目录
│   ├── playground/
│   │   └── versions/    # 多版本共存
│   └── kwdb/            # KWDB 安装目录
├── data/                # 运行时数据
└── config.yaml          # 全局配置（代码源等）
```

## 依赖要求

- [Go](https://golang.org/) 1.25+（仅构建时需要）
- [Docker](https://docs.docker.com/get-docker/) & Docker Compose（Playground 必需）
- Git（用于克隆组件仓库）

## Roadmap

- [x] Playground 一键启动 (`kwcli playground start`)
- [x] KWDB 单机版安装 (`kwcli kwdb install`)
  - 支持 Docker 部署
  - 二进制下载接口已预留（待 Gitee API 集成）
- [x] KWDB 服务管理 (`kwcli kwdb start/stop/restart/status/logs`)
- [x] 配置管理 (`kwcli kwdb config show/edit/set/path`)
- [x] SQL 连接 (`kwcli sql`)
- [x] Playground 版本管理 (`kwcli playground install/versions/uninstall`)
- [x] 全局代码源配置 (`kwcli source`)
- [x] 阿里云镜像加速 (`--registry auto`)
- [x] Makefile 构建脚本
- [x] SampleDB 智能电表模型 (`kwcli sampledb`)
- [x] TSBS 基准测试工具 (`kwcli tsbs init/load/run/list/clean`)
- [x] Shell 自动补全 (`kwcli completion [bash|zsh|fish|powershell]`)
- [x] 全局 JSON 输出模式 (`--json`)
- [x] Schema 导出 (`kwcli schema dump --db rdb/tsdb/all`)
- [x] 批量 SQL 执行 (`kwcli sql -f`、`--transaction`)
- [x] 查询结果导出 (`kwcli sql --export csv/json`)
- [x] 数据库巡检 (`kwcli inspect run`)
- [x] 性能快照 (`kwcli perf snapshot`)
- [ ] 组件清单与版本索引
- [ ] 离线镜像与私有化部署支持
- [ ] Homebrew / install.sh 一键安装

## 作者

**Shawn Yan**

- 公众号「少安事务所」主笔
- KaiwuDB 社区 KWDB MVP
- 个人主页：[shawnyan.cn](https://shawnyan.cn)
- 联系邮箱：admin@shawnyan.cn

<table align="center">
  <tr>
    <td align="center" style="padding-bottom: 10px;">公众号</td>
    <td align="center" style="padding-bottom: 10px;">微信群</td>
  </tr>
  <tr>
    <td align="center"><img src="img/wx.jpg" width="200" /></td>
    <td align="center"><img src="img/qun.jpg" width="200" /></td>
  </tr>
</table>

## ❤️ 特别感谢

特别感谢以下项目：

[KWDB](https://github.com/KWDB) 是一款专为 AIoT 场景设计的分布式多模数据库，由开放原子开源基金会孵化。该项目源自浪潮的 KaiwuDB 项目，支持在同一实例内并发创建时序数据库与关系型数据库，并支持对多模数据进行集成处理。

![](https://github.com/KWDB/.github/raw/main/profile/logo.png)

## ☕ 捐赠支持

如果 kwcli 对你有所帮助，欢迎支持项目持续迭代。点击下面文章链接，文末【稀罕作者】

- [KaiwuDB开源工具kwcli v0.1.1更新](https://mp.weixin.qq.com/s/v5eDgOmFaOeIgxdrXrYQhg)
- [kwcli：开源一个 KaiwuDB 社区版的 CLI 工具](https://mp.weixin.qq.com/s/od7h4yy7tzTFIv11wcjY0w)

> 捐赠后欢迎留下你的昵称和 Github 链接，我将记录到项目 README 里。 ❤️

## ⭐ Star History

如果 kwcli 对你有帮助，欢迎点一个 Star ⭐

你的支持，是项目持续迭代最大的动力。

## 贡献

欢迎提交 Issue 和 PR！

## License

Apache License 2.0

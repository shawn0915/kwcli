package mcp

// RegisterDefaultTools registers the default set of tools for KWDB operations
func RegisterDefaultTools(server *Server) {
	tools := []*Tool{
		{
			Name:        "kwcli_sql_exec",
			Description: "执行 KWDB SQL 查询",
			Command:     "kwcli sql -e \"{{sql}}\" -d {{db}} -u {{user}} --json",
			Parameters: map[string]string{
				"sql":  "SQL query to execute",
				"db":   "Database name",
				"user": "Database user",
			},
		},
		{
			Name:        "kwcli_schema_dump",
			Description: "导出数据库 Schema",
			Command:     "kwcli schema dump --db {{db}} --format json",
			Parameters: map[string]string{
				"db": "Database to dump (rdb, tsdb, all)",
			},
		},
		{
			Name:        "kwcli_status",
			Description: "获取 KWDB 服务状态",
			Command:     "kwcli kwdb status --json",
		},
		{
			Name:        "kwcli_logs",
			Description: "获取 KWDB 日志",
			Command:     "kwcli kwdb logs --tail {{n}}",
			Parameters: map[string]string{
				"n": "Number of log lines to show",
			},
		},
		{
			Name:        "kwcli_tsbs_run",
			Description: "运行 TSBS 基准测试",
			Command:     "kwcli tsbs run --use-case {{case}} --json",
			Parameters: map[string]string{
				"case": "Use case (cpu, iot)",
			},
		},
		{
			Name:        "kwcli_sampledb_init",
			Description: "初始化 SampleDB",
			Command:     "kwcli sampledb init",
		},
		{
			Name:        "kwcli_skill_list",
			Description: "列出已安装的 Skills",
			Command:     "kwcli skill list --json",
		},
		{
			Name:        "kwcli_skill_info",
			Description: "获取 Skill 详情",
			Command:     "kwcli skill info {{name}} --json",
			Parameters: map[string]string{
				"name": "Skill name",
			},
		},
		{
			Name:        "kwcli_perf_snapshot",
			Description: "收集性能快照",
			Command:     "kwcli perf snapshot --json",
		},
		{
			Name:        "kwcli_inspect_run",
			Description: "运行数据库巡检",
			Command:     "kwcli inspect run --items {{items}} --json",
			Parameters: map[string]string{
				"items": "Comma-separated inspection items",
			},
		},
	}

	for _, tool := range tools {
		server.RegisterTool(tool)
	}
}

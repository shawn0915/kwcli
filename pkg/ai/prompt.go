package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shawn0915/kwcli/pkg/skill"
)

// SystemPromptBuilder builds system prompts for the AI mode
type SystemPromptBuilder struct {
	SkillsEnabled bool
	SkillNames    []string
}

// NewSystemPromptBuilder creates a new system prompt builder
func NewSystemPromptBuilder() *SystemPromptBuilder {
	return &SystemPromptBuilder{
		SkillsEnabled: true,
	}
}

// BuildSystemPrompt constructs the system prompt for AI interactions
func (b *SystemPromptBuilder) BuildSystemPrompt() (string, error) {
	var sb strings.Builder

	sb.WriteString(`你是一个 KWDB 数据库专家助手。你可以帮助用户完成以下任务：

1. **SQL 查询与数据分析** - 编写和执行 SQL 查询
2. **数据库管理** - 创建数据库、表、索引等
3. **性能诊断** - 分析查询性能、识别瓶颈
4. **运维管理** - 巡检、备份、监控
5. **数据导入导出** - 批量数据处理

## 执行规则

- **只读操作** (SELECT/SHOW/EXPLAIN) 自动执行，无需用户确认
- **写操作** (INSERT/CREATE/ALTER/DROP/UPDATE/DELETE) 需展示 SQL 并询问用户确认
- **DANGEROUS 操作** (DROP DATABASE/TRUNCATE/DROP TABLE) 需明确警告并双重确认
- 不猜测数据库端口、IP、目录等参数
- 执行失败后读取日志报告原因，不擅自重试

## 可用命令

你通过以下 kwcli 命令与数据库交互：

- kwcli sql -e "<sql>" --json  - 执行 SQL 查询
- kwcli schema dump --db <rdb|tsdb|all> --format json - 导出 Schema
- kwcli kwdb status --json  - 查看服务状态
- kwcli kwdb logs --tail <n>  - 查看日志
- kwcli perf snapshot --json  - 性能快照
- kwcli inspect run --json    - 数据库巡检

## Schema 发现

在执行 SQL 前，优先获取真实表结构了解数据模型。
`)

	// Add skill references if available
	if b.SkillsEnabled {
		installedSkills, err := skill.ListInstalled()
		if err == nil && len(installedSkills) > 0 {
			sb.WriteString("\n## 已安装的 Skills\n\n")
			for _, s := range installedSkills {
				sb.WriteString(fmt.Sprintf("### %s (v%s)\n", s.Name, s.Version))
				sb.WriteString(fmt.Sprintf("%s\n\n", s.Description))
				if len(s.Triggers) > 0 {
					sb.WriteString(fmt.Sprintf("适用场景: %s\n\n", strings.Join(s.Triggers, ", ")))
				}

				// Load reference files
				refsDir := filepath.Join(skill.GetSkillsDir(), s.Name, "references")
				if refs, err := os.ReadDir(refsDir); err == nil {
					for _, ref := range refs {
						if !ref.IsDir() {
							data, err := os.ReadFile(filepath.Join(refsDir, ref.Name()))
							if err == nil {
								sb.WriteString(fmt.Sprintf("#### %s\n\n```\n%s\n```\n\n", ref.Name(), string(data)))
							}
						}
					}
				}
			}
		}
	}

	return sb.String(), nil
}

// BuildQueryPrompt builds a query-specific prompt
func (b *SystemPromptBuilder) BuildQueryPrompt(userQuery string, schemaContext string) string {
	var sb strings.Builder

	if schemaContext != "" {
		sb.WriteString("当前数据库 Schema:\n")
		sb.WriteString(schemaContext)
		sb.WriteString("\n\n")
	}

	sb.WriteString("用户问题: ")
	sb.WriteString(userQuery)

	return sb.String()
}

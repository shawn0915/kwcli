package inspect

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/shawn0915/kwcli/pkg/config"
)

// InspectItem represents a single inspection item
type InspectItem struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "pass", "warn", "fail", "error"
	Detail string `json:"detail"`
}

// InspectReport represents the full inspection report
type InspectReport struct {
	Timestamp string        `json:"timestamp"`
	Items     []InspectItem `json:"items"`
	Summary   Summary       `json:"summary"`
}

// Summary represents the inspection summary
type Summary struct {
	Total int `json:"total"`
	Pass  int `json:"pass"`
	Warn  int `json:"warn"`
	Fail  int `json:"fail"`
}

// RunInspection executes the inspection with the given items
func RunInspection(items []string) (*InspectReport, error) {
	report := &InspectReport{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Items:     []InspectItem{},
	}

	for _, item := range items {
		var inspectItem InspectItem
		switch item {
		case "status":
			inspectItem = checkStatus()
		case "logs":
			inspectItem = checkLogs()
		case "connections":
			inspectItem = checkConnections()
		case "slow_queries":
			inspectItem = checkSlowQueries()
		case "storage":
			inspectItem = checkStorage()
		case "sampledb":
			inspectItem = checkSampleDB()
		default:
			inspectItem = InspectItem{
				Name:   item,
				Status: "error",
				Detail: fmt.Sprintf("Unknown inspection item: %s", item),
			}
		}
		report.Items = append(report.Items, inspectItem)
	}

	// Calculate summary
	for _, item := range report.Items {
		report.Summary.Total++
		switch item.Status {
		case "pass":
			report.Summary.Pass++
		case "warn":
			report.Summary.Warn++
		case "fail":
			report.Summary.Fail++
		}
	}

	return report, nil
}

// StandardItems returns the default inspection items
func StandardItems() []string {
	return []string{"status", "logs", "connections", "slow_queries", "storage", "sampledb"}
}

func checkStatus() InspectItem {
	item := InspectItem{Name: "服务状态 (Service Status)"}

	installDir := filepath.Join(config.GetComponentsDir(), "kwdb")

	// Check Docker mode
	dockerMarker := installDir + "/.docker_mode"
	if data, err := os.ReadFile(dockerMarker); err == nil {
		runtimeCmd := strings.TrimSpace(string(data))
		if runtimeCmd == "" {
			runtimeCmd = "docker"
		}
		cmd := exec.Command(runtimeCmd, "ps", "--filter", "name=^/kwdb$", "--format", "{{.Status}}")
		output, err := cmd.Output()
		if err != nil {
			item.Status = "fail"
			item.Detail = "KWDB container not found"
			return item
		}
		status := strings.TrimSpace(string(output))
		if strings.Contains(status, "Up") {
			item.Status = "pass"
			item.Detail = fmt.Sprintf("KWDB container is running (%s)", status)
		} else {
			item.Status = "fail"
			item.Detail = fmt.Sprintf("KWDB container is not running (%s)", status)
		}
		return item
	}

	// Check Binary mode
	cmd := exec.Command("pgrep", "-a", "kwbase")
	output, _ := cmd.Output()
	if len(output) > 0 {
		item.Status = "pass"
		item.Detail = "KWDB process is running"
	} else {
		item.Status = "fail"
		item.Detail = "KWDB is not running"
	}

	return item
}

func checkLogs() InspectItem {
	item := InspectItem{Name: "日志异常 (Log Anomalies)"}

	installDir := filepath.Join(config.GetComponentsDir(), "kwdb")

	// Check Docker mode
	dockerMarker := installDir + "/.docker_mode"
	if data, err := os.ReadFile(dockerMarker); err == nil {
		runtimeCmd := strings.TrimSpace(string(data))
		if runtimeCmd == "" {
			runtimeCmd = "docker"
		}
		// Check recent logs for ERROR/FATAL
		cmd := exec.Command(runtimeCmd, "logs", "--since", "1h", "kwdb")
		output, err := cmd.Output()
		if err != nil {
			item.Status = "error"
			item.Detail = "Failed to read KWDB logs"
			return item
		}
		logStr := string(output)
		errorCount := strings.Count(logStr, "ERROR")
		fatalCount := strings.Count(logStr, "FATAL")

		if fatalCount > 0 {
			item.Status = "fail"
			item.Detail = fmt.Sprintf("Found %d FATAL and %d ERROR entries in the last hour", fatalCount, errorCount)
		} else if errorCount > 0 {
			item.Status = "warn"
			item.Detail = fmt.Sprintf("Found %d ERROR entries in the last hour", errorCount)
		} else {
			item.Status = "pass"
			item.Detail = "No ERROR or FATAL entries found in the last hour"
		}
		return item
	}

	// Check Binary mode logs
	cfg, err := config.LoadKWDBConfig()
	if err != nil || cfg.LogDir == "" {
		item.Status = "warn"
		item.Detail = "Log directory not configured"
		return item
	}

	logFiles, _ := filepath.Glob(filepath.Join(cfg.LogDir, "*.log"))
	if len(logFiles) == 0 {
		item.Status = "warn"
		item.Detail = "No log files found"
		return item
	}

	// Read the most recent log file
	var latestLog string
	for _, f := range logFiles {
		if f > latestLog {
			latestLog = f
		}
	}

	content, err := os.ReadFile(latestLog)
	if err != nil {
		item.Status = "error"
		item.Detail = "Failed to read log file"
		return item
	}

	logStr := string(content)
	errorCount := strings.Count(logStr, "ERROR")
	fatalCount := strings.Count(logStr, "FATAL")

	if fatalCount > 0 {
		item.Status = "fail"
		item.Detail = fmt.Sprintf("Found %d FATAL and %d ERROR entries in %s", fatalCount, errorCount, filepath.Base(latestLog))
	} else if errorCount > 0 {
		item.Status = "warn"
		item.Detail = fmt.Sprintf("Found %d ERROR entries in %s", errorCount, filepath.Base(latestLog))
	} else {
		item.Status = "pass"
		item.Detail = "No ERROR or FATAL entries found"
	}

	return item
}

func checkConnections() InspectItem {
	item := InspectItem{Name: "连接数 (Connections)"}

	// Try to execute SHOW SESSIONS via kwbase
	output, err := executeSQL("SHOW SESSIONS")
	if err != nil {
		item.Status = "error"
		item.Detail = fmt.Sprintf("Failed to query sessions: %v", err)
		return item
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	// Count active connections (exclude header and separator lines)
	connCount := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(trimmed, "active_connections") && !strings.HasPrefix(trimmed, "(") {
			connCount++
		}
	}

	// Typical KWDB default max connections is around 100
	if connCount > 80 {
		item.Status = "warn"
		item.Detail = fmt.Sprintf("Active connections: %d (approaching limit)", connCount)
	} else if connCount > 0 {
		item.Status = "pass"
		item.Detail = fmt.Sprintf("Active connections: %d (normal)", connCount)
	} else {
		item.Status = "pass"
		item.Detail = "No active connections"
	}

	return item
}

func checkSlowQueries() InspectItem {
	item := InspectItem{Name: "慢查询 (Slow Queries)"}

	output, err := executeSQL("SHOW QUERIES")
	if err != nil {
		item.Status = "error"
		item.Detail = fmt.Sprintf("Failed to query running queries: %v", err)
		return item
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	queryCount := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(trimmed, "query_id") && !strings.HasPrefix(trimmed, "(") {
			queryCount++
		}
	}

	if queryCount > 10 {
		item.Status = "warn"
		item.Detail = fmt.Sprintf("%d running queries detected", queryCount)
	} else if queryCount > 0 {
		item.Status = "pass"
		item.Detail = fmt.Sprintf("%d running queries (normal)", queryCount)
	} else {
		item.Status = "pass"
		item.Detail = "No running queries"
	}

	return item
}

func checkStorage() InspectItem {
	item := InspectItem{Name: "存储空间 (Storage)"}

	installDir := filepath.Join(config.GetComponentsDir(), "kwdb")

	// Check Docker mode
	dockerMarker := installDir + "/.docker_mode"
	if data, err := os.ReadFile(dockerMarker); err == nil {
		runtimeCmd := strings.TrimSpace(string(data))
		if runtimeCmd == "" {
			runtimeCmd = "docker"
		}
		// Get container size info
		cmd := exec.Command(runtimeCmd, "system", "df")
		output, err := cmd.Output()
		if err != nil {
			// Fallback: check disk usage of data directory
			item.Status = "warn"
			item.Detail = "Unable to get storage info from container"
			return item
		}
		item.Status = "pass"
		item.Detail = "Storage usage:\n" + strings.TrimSpace(string(output))
		return item
	}

	// Binary mode: check data directory size
	cfg, err := config.LoadKWDBConfig()
	if err != nil || cfg.DataDir == "" {
		item.Status = "warn"
		item.Detail = "Data directory not configured"
		return item
	}

	cmd := exec.Command("du", "-sh", cfg.DataDir)
	output, err := cmd.Output()
	if err != nil {
		item.Status = "warn"
		item.Detail = "Unable to check data directory size"
		return item
	}

	item.Status = "pass"
	item.Detail = fmt.Sprintf("Data directory: %s", strings.TrimSpace(string(output)))

	return item
}

func checkSampleDB() InspectItem {
	item := InspectItem{Name: "SampleDB 健康 (SampleDB Health)"}

	// Check if rdb and tsdb databases exist
	rdbOutput, err := executeSQL("SELECT database_name FROM [SHOW DATABASES] WHERE database_name = 'rdb'")
	if err != nil {
		item.Status = "error"
		item.Detail = "Failed to check SampleDB status"
		return item
	}

	tsdbOutput, err := executeSQL("SELECT database_name FROM [SHOW DATABASES] WHERE database_name = 'tsdb'")
	if err != nil {
		item.Status = "error"
		item.Detail = "Failed to check SampleDB status"
		return item
	}

	rdbExists := strings.Contains(rdbOutput, "rdb")
	tsdbExists := strings.Contains(tsdbOutput, "tsdb")

	if rdbExists && tsdbExists {
		item.Status = "pass"
		item.Detail = "SampleDB is initialized (rdb and tsdb exist)"
	} else if rdbExists || tsdbExists {
		item.Status = "warn"
		item.Detail = fmt.Sprintf("SampleDB partially initialized (rdb: %v, tsdb: %v)", rdbExists, tsdbExists)
	} else {
		item.Status = "warn"
		item.Detail = "SampleDB is not initialized. Run: kwcli sampledb init"
	}

	return item
}

// executeSQL runs a SQL command via kwbase and returns the output
func executeSQL(sql string) (string, error) {
	installDir := filepath.Join(config.GetComponentsDir(), "kwdb")

	// Check Docker mode
	dockerMarker := installDir + "/.docker_mode"
	if data, err := os.ReadFile(dockerMarker); err == nil {
		runtimeCmd := strings.TrimSpace(string(data))
		if runtimeCmd == "" {
			runtimeCmd = "docker"
		}
		cmd := exec.Command(runtimeCmd, "exec", "kwdb", "./kwbase", "sql", "--insecure", "-e", sql)
		output, err := cmd.CombinedOutput()
		return string(output), err
	}

	// Binary mode
	binaryMarker := installDir + "/.binary_mode"
	if _, err := os.Stat(binaryMarker); err == nil {
		kwbasePath := findKwbaseBinary(inspectInstallDir)
		if kwbasePath != "" {
			cmd := exec.Command(kwbasePath, "sql", "--insecure", "-e", sql)
			output, err := cmd.CombinedOutput()
			return string(output), err
		}
	}

	return "", fmt.Errorf("KWDB is not installed")
}

var inspectInstallDir = filepath.Join(config.GetComponentsDir(), "kwdb")

func findKwbaseBinary(searchDir string) string {
	var kwbasePath string
	filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && (info.Name() == "kwbase" || info.Name() == "kwbase.exe") {
			kwbasePath = path
		}
		return nil
	})
	return kwbasePath
}

// ReportToMarkdown converts the inspection report to Markdown format
func ReportToMarkdown(report *InspectReport) string {
	var sb strings.Builder

	sb.WriteString("# KWDB 巡检报告\n\n")
	sb.WriteString(fmt.Sprintf("**巡检时间**: %s\n\n", report.Timestamp))

	sb.WriteString("## 巡检结果\n\n")
	sb.WriteString("| 巡检项 | 状态 | 详情 |\n")
	sb.WriteString("|--------|------|------|\n")

	statusEmoji := map[string]string{
		"pass":  "✅",
		"warn":  "⚠️",
		"fail":  "❌",
		"error": "🔴",
	}

	for _, item := range report.Items {
		emoji := statusEmoji[item.Status]
		sb.WriteString(fmt.Sprintf("| %s | %s %s | %s |\n", item.Name, emoji, strings.ToUpper(item.Status), item.Detail))
	}

	sb.WriteString("\n## 汇总\n\n")
	sb.WriteString(fmt.Sprintf("- 总计: %d 项\n", report.Summary.Total))
	sb.WriteString(fmt.Sprintf("- ✅ 通过: %d 项\n", report.Summary.Pass))
	sb.WriteString(fmt.Sprintf("- ⚠️ 警告: %d 项\n", report.Summary.Warn))
	sb.WriteString(fmt.Sprintf("- ❌ 失败: %d 项\n", report.Summary.Fail))

	return sb.String()
}

// ReportToJSON converts the inspection report to JSON format
func ReportToJSON(report *InspectReport) (string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ReportToHTML converts the inspection report to HTML format
func ReportToHTML(report *InspectReport) string {
	statusClass := map[string]string{
		"pass":  "status-pass",
		"warn":  "status-warn",
		"fail":  "status-fail",
		"error": "status-error",
	}

	statusLabel := map[string]string{
		"pass":  "✅ 通过",
		"warn":  "⚠️ 警告",
		"fail":  "❌ 失败",
		"error": "🔴 错误",
	}

	var rows strings.Builder
	for _, item := range report.Items {
		cls := statusClass[item.Status]
		label := statusLabel[item.Status]
		rows.WriteString(fmt.Sprintf(
			`<tr><td>%s</td><td><span class="status-badge %s">%s</span></td><td>%s</td></tr>`,
			item.Name, cls, label, item.Detail,
		))
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>KWDB 巡检报告</title>
  <style>
    :root {
      --primary: #1a73e8;
      --primary-light: #e8f0fe;
      --success: #34a853;
      --success-bg: #e6f4ea;
      --warning: #fbbc04;
      --warning-bg: #fef7e0;
      --danger: #ea4335;
      --danger-bg: #fce8e6;
      --error: #9334e6;
      --error-bg: #f3e8fd;
      --bg: #f8f9fa;
      --card-bg: #ffffff;
      --text: #202124;
      --text-secondary: #5f6368;
      --border: #dadce0;
      --radius: 12px;
    }
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
      background: var(--bg);
      color: var(--text);
      line-height: 1.6;
      padding: 24px;
    }
    .container { max-width: 960px; margin: 0 auto; }
    .header {
      background: linear-gradient(135deg, var(--primary), #4285f4);
      color: white;
      padding: 32px;
      border-radius: var(--radius);
      margin-bottom: 24px;
      box-shadow: 0 4px 12px rgba(26, 115, 232, 0.3);
    }
    .header h1 { font-size: 24px; font-weight: 600; margin-bottom: 8px; }
    .header .timestamp { opacity: 0.9; font-size: 14px; }
    .summary-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
      gap: 16px;
      margin-bottom: 24px;
    }
    .summary-card {
      background: var(--card-bg);
      border-radius: var(--radius);
      padding: 20px;
      text-align: center;
      box-shadow: 0 1px 3px rgba(0,0,0,0.1);
      border: 1px solid var(--border);
    }
    .summary-card .number { font-size: 32px; font-weight: 700; }
    .summary-card .label { font-size: 14px; color: var(--text-secondary); margin-top: 4px; }
    .summary-card.total .number { color: var(--primary); }
    .summary-card.pass .number { color: var(--success); }
    .summary-card.warn .number { color: var(--warning); }
    .summary-card.fail .number { color: var(--danger); }
    .table-card {
      background: var(--card-bg);
      border-radius: var(--radius);
      overflow: hidden;
      box-shadow: 0 1px 3px rgba(0,0,0,0.1);
      border: 1px solid var(--border);
    }
    .table-card h2 {
      padding: 16px 20px;
      border-bottom: 1px solid var(--border);
      font-size: 18px;
      font-weight: 600;
    }
    table { width: 100%%; border-collapse: collapse; }
    th {
      background: var(--bg);
      padding: 12px 16px;
      text-align: left;
      font-weight: 600;
      font-size: 13px;
      color: var(--text-secondary);
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }
    td {
      padding: 12px 16px;
      border-top: 1px solid var(--border);
      font-size: 14px;
    }
    tr:hover td { background: var(--primary-light); }
    .status-badge {
      display: inline-block;
      padding: 4px 12px;
      border-radius: 20px;
      font-size: 13px;
      font-weight: 500;
    }
    .status-pass { background: var(--success-bg); color: var(--success); }
    .status-warn { background: var(--warning-bg); color: #e37400; }
    .status-fail { background: var(--danger-bg); color: var(--danger); }
    .status-error { background: var(--error-bg); color: var(--error); }
    .footer {
      text-align: center;
      margin-top: 24px;
      color: var(--text-secondary);
      font-size: 13px;
    }
    .footer a { color: var(--primary); text-decoration: none; }
    @media (max-width: 640px) {
      body { padding: 12px; }
      .header { padding: 20px; }
      th, td { padding: 8px 12px; font-size: 13px; }
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>🔍 KWDB 巡检报告</h1>
      <div class="timestamp">巡检时间: %s</div>
    </div>
    <div class="summary-grid">
      <div class="summary-card total">
        <div class="number">%d</div>
        <div class="label">总计</div>
      </div>
      <div class="summary-card pass">
        <div class="number">%d</div>
        <div class="label">✅ 通过</div>
      </div>
      <div class="summary-card warn">
        <div class="number">%d</div>
        <div class="label">⚠️ 警告</div>
      </div>
      <div class="summary-card fail">
        <div class="number">%d</div>
        <div class="label">❌ 失败</div>
      </div>
    </div>
    <div class="table-card">
      <h2>巡检详情</h2>
      <table>
        <thead><tr><th>巡检项</th><th>状态</th><th>详情</th></tr></thead>
        <tbody>%s</tbody>
      </table>
    </div>
    <div class="footer">
      Generated by <a href="https://github.com/KWDB/kwcli">kwcli</a> — KWDB Ecosystem CLI Tool
    </div>
  </div>
</body>
</html>`, report.Timestamp, report.Summary.Total, report.Summary.Pass, report.Summary.Warn, report.Summary.Fail, rows.String())
}

// WriteReport writes the inspection report to a file in the specified format
func WriteReport(report *InspectReport, format string, outputFile string) error {
	switch format {
	case "json":
		content, err := ReportToJSON(report)
		if err != nil {
			return err
		}
		if outputFile != "" {
			return os.WriteFile(outputFile, []byte(content), 0644)
		}
		fmt.Println(content)
	case "markdown", "md":
		content := ReportToMarkdown(report)
		if outputFile != "" {
			return os.WriteFile(outputFile, []byte(content), 0644)
		}
		fmt.Println(content)
	case "html":
		content := ReportToHTML(report)
		if outputFile != "" {
			return os.WriteFile(outputFile, []byte(content), 0644)
		}
		fmt.Println(content)
	default:
		return fmt.Errorf("unsupported format: %s (supported: json, markdown, html)", format)
	}
	return nil
}

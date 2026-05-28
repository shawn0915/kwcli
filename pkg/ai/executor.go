package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Executor handles command execution from AI responses with guardrails
type Executor struct {
	DryRun      bool
	Interactive bool
}

// NewExecutor creates a new AI executor
func NewExecutor() *Executor {
	return &Executor{
		DryRun:      false,
		Interactive: false,
	}
}

// GuardResult represents the result of a guard check
type GuardResult struct {
	Allowed       bool
	SQL           string
	Category      string // "read", "write", "dangerous"
	Warning       string
	NeedsApproval bool
}

// ClassificationResult holds the parsed AI response
type ClassificationResult struct {
	Intent     string // "query", "schema_discovery", "admin", "unknown"
	SQL        string
	Command    string
	Plan       string
	Confidence float64
}

// ParseAndGuard parses AI output and applies safety guardrails
func (e *Executor) ParseAndGuard(aiResponse string) *GuardResult {
	// Extract SQL from the response
	sql := extractSQL(aiResponse)
	if sql == "" {
		return &GuardResult{
			Allowed:  true,
			Category: "unknown",
			Warning:  "No SQL found in response",
		}
	}

	return e.classifyAndGuard(sql)
}

// classifyAndGuard classifies SQL and applies guardrails
func (e *Executor) classifyAndGuard(sql string) *GuardResult {
	upperSQL := strings.TrimSpace(strings.ToUpper(sql))
	result := &GuardResult{
		SQL:           sql,
		NeedsApproval: false,
	}

	// Classify the SQL
	switch {
	case strings.HasPrefix(upperSQL, "SELECT"):
		result.Category = "read"
		result.Allowed = true
		result.Warning = ""

	case strings.HasPrefix(upperSQL, "SHOW"):
		result.Category = "read"
		result.Allowed = true

	case strings.HasPrefix(upperSQL, "EXPLAIN"):
		result.Category = "read"
		result.Allowed = true

	case strings.HasPrefix(upperSQL, "INSERT"):
		result.Category = "write"
		result.NeedsApproval = true
		result.Warning = "This operation will modify data"

	case strings.HasPrefix(upperSQL, "CREATE"):
		result.Category = "write"
		result.NeedsApproval = true
		result.Warning = "This operation will create new database objects"

	case strings.HasPrefix(upperSQL, "ALTER"):
		result.Category = "write"
		result.NeedsApproval = true
		result.Warning = "This operation will modify database schema"

	case strings.HasPrefix(upperSQL, "DROP"):
		result.Category = "dangerous"
		result.NeedsApproval = true
		result.Warning = "DANGER: This operation will permanently delete database objects"
		// Check for very dangerous operations
		if strings.Contains(upperSQL, "DROP DATABASE") || strings.Contains(upperSQL, "TRUNCATE") {
			result.Warning = "⚠️ DANGER: This operation will permanently delete data! Double confirmation required!"
		}

	case strings.HasPrefix(upperSQL, "UPDATE"):
		result.Category = "write"
		result.NeedsApproval = true
		result.Warning = "This operation will modify data"

	case strings.HasPrefix(upperSQL, "DELETE"):
		result.Category = "write"
		result.NeedsApproval = true
		result.Warning = "This operation will delete data"

	default:
		result.Category = "unknown"
		result.NeedsApproval = true
		result.Warning = "Unknown operation type - approval required"
	}

	return result
}

// ConfirmAction asks the user to confirm an action
func (e *Executor) ConfirmAction(result *GuardResult) bool {
	if e.DryRun {
		fmt.Println("[DRY RUN] Generated command (not executed):")
		fmt.Println(result.SQL)
		return false
	}

	fmt.Println()
	fmt.Printf("[%s] %s\n", strings.ToUpper(result.Category), result.Warning)
	fmt.Println()
	fmt.Println("SQL to execute:")
	fmt.Println(result.SQL)
	fmt.Println()
	fmt.Print("Execute this SQL? [y/N]: ")

	var answer string
	fmt.Scanln(&answer)
	answer = strings.ToLower(strings.TrimSpace(answer))

	if answer == "y" || answer == "yes" {
		// For dangerous operations, double confirm
		if result.Category == "dangerous" {
			fmt.Print("\n⚠️ This is a DANGEROUS operation. Type 'yes' to confirm: ")
			fmt.Scanln(&answer)
			answer = strings.ToLower(strings.TrimSpace(answer))
			return answer == "yes"
		}
		return true
	}

	return false
}

// extractSQL extracts SQL from an AI response
// Looks for SQL code blocks or standalone SQL statements
func extractSQL(response string) string {
	// Try to find SQL in code blocks first
	lines := strings.Split(response, "\n")
	inCodeBlock := false
	var sqlLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```sql") {
			inCodeBlock = true
			continue
		}
		if strings.HasPrefix(trimmed, "```") && inCodeBlock {
			break
		}
		if inCodeBlock {
			sqlLines = append(sqlLines, line)
		}
	}

	if len(sqlLines) > 0 {
		return strings.TrimSpace(strings.Join(sqlLines, "\n"))
	}

	// Fallback: look for lines starting with SQL keywords
	var sqlParts []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		upper := strings.ToUpper(trimmed)
		if strings.HasPrefix(upper, "SELECT") ||
			strings.HasPrefix(upper, "INSERT") ||
			strings.HasPrefix(upper, "CREATE") ||
			strings.HasPrefix(upper, "ALTER") ||
			strings.HasPrefix(upper, "DROP") ||
			strings.HasPrefix(upper, "UPDATE") ||
			strings.HasPrefix(upper, "DELETE") ||
			strings.HasPrefix(upper, "SHOW") ||
			strings.HasPrefix(upper, "EXPLAIN") ||
			strings.HasPrefix(upper, "WITH") {
			sqlParts = append(sqlParts, line)
		}
	}

	if len(sqlParts) > 0 {
		return strings.TrimSpace(strings.Join(sqlParts, "\n"))
	}

	return ""
}

// WriteGuardrailViolation logs a guardrail violation
func WriteGuardrailViolation(sql string, reason string) {
	logDir := filepath.Join(getHomeDir(), "logs")
	os.MkdirAll(logDir, 0755)

	logEntry := fmt.Sprintf("[GUARDRAIL] %s\nSQL: %s\nReason: %s\n\n",
		time.Now().Format(time.RFC3339), sql, reason)

	logFile := filepath.Join(logDir, "guardrail.log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		f.WriteString(logEntry)
	}
}

// getHomeDir returns the home directory for logging
func getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".kwcli"
	}
	return filepath.Join(home, ".kwcli")
}

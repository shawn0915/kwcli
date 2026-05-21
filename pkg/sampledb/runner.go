package sampledb

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/shawn0915/kwcli/pkg/config"
)

// Runner executes SQL against KWDB
type Runner struct {
	dockerMode bool
	kwbasePath string
}

// NewRunner detects KWDB installation and returns a Runner
func NewRunner() (*Runner, error) {
	installDir := filepath.Join(config.GetComponentsDir(), "kwdb")

	// Check Docker mode
	dockerMarker := installDir + "/.docker_mode"
	if _, err := os.Stat(dockerMarker); err == nil {
		// Verify container is running
		checkCmd := exec.Command("docker", "ps", "--filter", "name=^/kwdb$", "--format", "{{.Names}}")
		output, err := checkCmd.Output()
		if err != nil || strings.TrimSpace(string(output)) == "" {
			return nil, fmt.Errorf("KWDB container is not running. Run 'kwcli kwdb start' first")
		}
		return &Runner{dockerMode: true}, nil
	}

	// Check Binary mode
	binaryMarker := installDir + "/.binary_mode"
	if _, err := os.Stat(binaryMarker); err == nil {
		kwbasePath := findKwbase(installDir)
		if kwbasePath == "" {
			return nil, fmt.Errorf("kwbase binary not found. Please reinstall KWDB")
		}
		return &Runner{dockerMode: false, kwbasePath: kwbasePath}, nil
	}

	// Try to find kwbase anywhere in components directory
	kwbasePath := findKwbase(installDir)
	if kwbasePath != "" {
		return &Runner{dockerMode: false, kwbasePath: kwbasePath}, nil
	}

	return nil, fmt.Errorf("KWDB is not installed. Run 'kwcli kwdb install' first")
}

func findKwbase(searchDir string) string {
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

// ExecCapture executes a single SQL statement and returns the output
func (r *Runner) ExecCapture(sql string) (string, error) {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return "", nil
	}

	if r.dockerMode {
		args := []string{"exec", "kwdb", "./kwbase", "sql", "--insecure", "-e", sql}
		cmd := exec.Command("docker", args...)
		output, err := cmd.CombinedOutput()
		return string(output), err
	}

	cmdArgs := []string{"sql", "--insecure", "-e", sql}
	cmd := exec.Command(r.kwbasePath, cmdArgs...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Exec executes a single SQL statement
func (r *Runner) Exec(sql string) error {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return nil
	}

	fmt.Printf("  Executing: %s\n", firstLine(sql))

	if r.dockerMode {
		return r.execDocker(sql)
	}
	return r.execBinary(sql)
}

// ExecBatch executes multiple SQL statements separated by semicolons
func (r *Runner) ExecBatch(sql string) error {
	// Split by semicolon but be careful with statements inside strings
	statements := splitSQLStatements(sql)
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if err := r.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) execDocker(sql string) error {
	cmd := exec.Command("docker", "exec", "kwdb", "./kwbase", "sql", "--insecure", "--host=127.0.0.1", "-e", sql)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (r *Runner) execBinary(sql string) error {
	cmdArgs := []string{r.kwbasePath, "sql", "--insecure", "--host=127.0.0.1", "-e", sql}
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = filepath.Dir(r.kwbasePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CheckDatabaseExists checks if a database exists
func (r *Runner) CheckDatabaseExists(dbName string) bool {
	sql := fmt.Sprintf("SELECT datname FROM pg_database WHERE datname = '%s';", dbName)
	var cmd *exec.Cmd
	if r.dockerMode {
		cmd = exec.Command("docker", "exec", "kwdb", "./kwbase", "sql", "--insecure", "--host=127.0.0.1", "-e", sql)
	} else {
		cmd = exec.Command(r.kwbasePath, "sql", "--insecure", "--host=127.0.0.1", "-e", sql)
		cmd.Dir = filepath.Dir(r.kwbasePath)
	}
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), dbName)
}

func firstLine(s string) string {
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "--") {
			if len(trimmed) > 60 {
				return trimmed[:60] + "..."
			}
			return trimmed
		}
	}
	return "[empty]"
}

func splitSQLStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	inString := false
	stringChar := rune(0)

	for i, ch := range sql {
		if !inString && (ch == '\'' || ch == '"') {
			inString = true
			stringChar = ch
		} else if inString && ch == stringChar {
			// Check for escaped quote
			nextIdx := i + 1
			if nextIdx < len(sql) && rune(sql[nextIdx]) == stringChar {
				current.WriteRune(ch)
				continue // skip next iteration
			}
			inString = false
		}

		current.WriteRune(ch)

		if !inString && ch == ';' {
			statements = append(statements, current.String())
			current.Reset()
		}
	}

	remainder := current.String()
	if strings.TrimSpace(remainder) != "" {
		statements = append(statements, remainder)
	}

	return statements
}

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shawn0915/kwcli/pkg/config"
	"github.com/shawn0915/kwcli/pkg/output"
	"github.com/spf13/cobra"
)

// sqlCmd represents the sql command
var sqlCmd = &cobra.Command{
	Use:   "sql",
	Short: "Connect to KWDB database",
	Long:  "Connect to KWDB database using kwbase CLI tool",
	Run: func(cmd *cobra.Command, args []string) {
		sqlFile, _ := cmd.Flags().GetString("file")
		if sqlFile != "" {
			executeSQLFile(sqlFile)
			return
		}
		// When using -e with --export or --json, use the capture/export path
		if sqlExecute != "" && (sqlExport != "" || output.JSONOutput) {
			executeAndExportSQL(sqlExecute)
			return
		}
		connectSQL(args)
	},
}

var (
	sqlHost        string
	sqlUser        string
	sqlDatabase    string
	sqlExecute     string
	insecureMode   bool
	sqlFile        string
	sqlTransaction bool
	sqlExport      string
	sqlOutputFile  string
)

func init() {
	rootCmd.AddCommand(sqlCmd)

	sqlCmd.Flags().StringVar(&sqlHost, "host", "127.0.0.1", "KWDB server host")
	sqlCmd.Flags().StringVarP(&sqlUser, "user", "u", "root", "KWDB user")
	sqlCmd.Flags().StringVarP(&sqlDatabase, "database", "d", "defaultdb", "Database name")
	sqlCmd.Flags().StringVarP(&sqlExecute, "execute", "e", "", "Execute SQL and exit")
	sqlCmd.Flags().BoolVar(&insecureMode, "insecure", true, "Use insecure mode (no TLS)")
	sqlCmd.Flags().StringVarP(&sqlFile, "file", "f", "", "Execute SQL from file and exit")
	sqlCmd.Flags().BoolVar(&sqlTransaction, "transaction", false, "Wrap file execution in a transaction (use with -f)")
	sqlCmd.Flags().StringVar(&sqlExport, "export", "", "Export query result to format: csv, json (use with -e)")
	sqlCmd.Flags().StringVar(&sqlOutputFile, "output", "", "Output file path for export (use with --export)")

	// Flag completion for --export
	sqlCmd.RegisterFlagCompletionFunc("export", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"csv", "json"}, cobra.ShellCompDirectiveNoFileComp
	})

	// Flag completion for --user
	sqlCmd.RegisterFlagCompletionFunc("user", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"root", "admin"}, cobra.ShellCompDirectiveNoFileComp
	})

	// Add analyze subcommand
	sqlCmd.AddCommand(sqlAnalyzeCmd)
}

// sqlAnalyzeCmd represents the sql analyze command
var sqlAnalyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze SQL execution plan",
	Long:  "Analyze the execution plan of a SQL statement for performance optimization",
	Run: func(cmd *cobra.Command, args []string) {
		analyzeSQL, _ := cmd.Flags().GetString("execute")
		if analyzeSQL == "" {
			fmt.Println("Error: --execute (-e) flag is required for sql analyze")
			os.Exit(1)
		}
		explainSQL := "EXPLAIN " + analyzeSQL
		executeAndExportSQL(explainSQL)
	},
}

func init() {
	sqlAnalyzeCmd.Flags().StringVarP(&sqlExecute, "execute", "e", "", "SQL statement to analyze")
}

func connectSQL(args []string) {
	installDir := filepath.Join(config.GetComponentsDir(), "kwdb")

	// Check Docker mode
	dockerMarker := installDir + "/.docker_mode"
	if _, err := os.Stat(dockerMarker); err == nil {
		connectSQLDocker(args)
		return
	}

	// Check Binary mode
	binaryMarker := installDir + "/.binary_mode"
	if _, err := os.Stat(binaryMarker); err == nil {
		connectSQLBinary(args)
		return
	}

	// Try to find kwbase anywhere in components directory
	kwbasePath := findKwbaseBinary(installDir)
	if kwbasePath != "" {
		connectSQLWithKwbase(kwbasePath, args)
		return
	}

	fmt.Println("KWDB is not installed. Run `kwcli kwdb install` first.")
	os.Exit(1)
}

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

func connectSQLWithKwbase(kwbasePath string, args []string) {
	cmdArgs := []string{"sql", "--insecure"}
	if sqlHost != "127.0.0.1" {
		cmdArgs = append(cmdArgs, "--host", sqlHost)
	}
	if sqlUser != "root" {
		cmdArgs = append(cmdArgs, "--user", sqlUser)
	}
	if sqlDatabase != "defaultdb" {
		cmdArgs = append(cmdArgs, "--database", sqlDatabase)
	}
	if sqlExecute != "" {
		cmdArgs = append(cmdArgs, "--execute", sqlExecute)
	}

	cmd := exec.Command(kwbasePath, cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to connect to KWDB: %v\n", err)
		os.Exit(1)
	}
}

func connectSQLDocker(args []string) {
	// Determine if we need TTY allocation
	// When executing SQL with -e, no TTY is needed (use -i only)
	// When entering interactive mode, allocate TTY (use -it)
	ttyFlag := "-i"
	if sqlExecute == "" {
		ttyFlag = "-it"
	}

	dockerArgs := []string{"exec", ttyFlag, "kwdb", "./kwbase", "sql", "--insecure"}

	if sqlHost != "127.0.0.1" {
		dockerArgs = append(dockerArgs, "--host", sqlHost)
	}
	if sqlUser != "root" {
		dockerArgs = append(dockerArgs, "--user", sqlUser)
	}
	if sqlDatabase != "defaultdb" {
		dockerArgs = append(dockerArgs, "--database", sqlDatabase)
	}
	if sqlExecute != "" {
		dockerArgs = append(dockerArgs, "--execute", sqlExecute)
	}

	cmd := exec.Command("docker", dockerArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to connect to KWDB: %v\n", err)
		os.Exit(1)
	}
}

func connectSQLBinary(args []string) {
	installDir := filepath.Join(config.GetComponentsDir(), "kwdb")

	// Find kwbase binary
	kwbasePath := findKwbaseBinary(installDir)
	if kwbasePath == "" {
		fmt.Println("kwbase binary not found. Please reinstall KWDB.")
		os.Exit(1)
	}

	connectSQLWithKwbase(kwbasePath, args)
}

// executeSQLFile reads and executes SQL from a file
func executeSQLFile(filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		output.PrintError("sql file", fmt.Errorf("failed to read file %s: %v", filePath, err))
		os.Exit(1)
	}

	sqlContent := string(content)

	if sqlTransaction {
		sqlContent = "BEGIN;\n" + sqlContent + "\nCOMMIT;"
	}

	if sqlExport != "" || output.JSONOutput {
		// Execute and capture output for export/JSON
		executeAndExportSQL(sqlContent)
	} else {
		executeSQLRaw(sqlContent)
	}
}

// executeAndExportSQL executes SQL and handles export/JSON formatting
func executeAndExportSQL(sql string) {
	outputBytes, err := executeSQLCapture(sql)
	if err != nil {
		output.PrintError("sql execute", fmt.Errorf("failed to execute SQL: %v", err))
		os.Exit(1)
	}

	resultText := string(outputBytes)

	if output.JSONOutput {
		parsed := output.ParseTableOutput(resultText)
		output.PrintJSON("sql execute", parsed, nil)
		return
	}

	if sqlExport != "" {
		format, err := output.ParseExportFormat(sqlExport)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		parsed := output.ParseTableOutput(resultText)
		if err := output.Export(parsed, format, sqlOutputFile); err != nil {
			fmt.Printf("Error exporting: %v\n", err)
			os.Exit(1)
		}
		if sqlOutputFile != "" {
			fmt.Printf("Query result exported to %s\n", sqlOutputFile)
		}
		return
	}

	fmt.Print(resultText)
}

// executeSQLRaw executes SQL directly to stdout/stderr
func executeSQLRaw(sql string) {
	installDir := filepath.Join(config.GetComponentsDir(), "kwdb")

	// Check Docker mode
	dockerMarker := installDir + "/.docker_mode"
	if _, err := os.Stat(dockerMarker); err == nil {
		cmd := exec.Command("docker", "exec", "-i", "kwdb", "./kwbase", "sql", "--insecure", "-e", sql)
		if sqlDatabase != "defaultdb" {
			cmd.Args = append(cmd.Args, "--database", sqlDatabase)
		}
		if sqlUser != "root" {
			cmd.Args = append(cmd.Args, "--user", sqlUser)
		}
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			os.Exit(1)
		}
		return
	}

	// Binary mode
	binaryMarker := installDir + "/.binary_mode"
	if _, err := os.Stat(binaryMarker); err == nil {
		kwbasePath := findKwbaseBinary(installDir)
		if kwbasePath == "" {
			fmt.Println("kwbase binary not found. Please reinstall KWDB.")
			os.Exit(1)
		}
		cmdArgs := []string{"sql", "--insecure", "-e", sql}
		if sqlDatabase != "defaultdb" {
			cmdArgs = append(cmdArgs, "--database", sqlDatabase)
		}
		if sqlUser != "root" {
			cmdArgs = append(cmdArgs, "--user", sqlUser)
		}
		cmd := exec.Command(kwbasePath, cmdArgs...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			os.Exit(1)
		}
		return
	}

	fmt.Println("KWDB is not installed. Run `kwcli kwdb install` first.")
	os.Exit(1)
}

// executeSQLCapture executes SQL and returns the output as bytes
func executeSQLCapture(sql string) ([]byte, error) {
	installDir := filepath.Join(config.GetComponentsDir(), "kwdb")

	// Check Docker mode
	dockerMarker := installDir + "/.docker_mode"
	if _, err := os.Stat(dockerMarker); err == nil {
		args := []string{"exec", "kwdb", "./kwbase", "sql", "--insecure", "-e", sql}
		if sqlDatabase != "defaultdb" {
			args = append(args, "--database", sqlDatabase)
		}
		if sqlUser != "root" {
			args = append(args, "--user", sqlUser)
		}
		cmd := exec.Command("docker", args...)
		return cmd.CombinedOutput()
	}

	// Binary mode
	binaryMarker := installDir + "/.binary_mode"
	if _, err := os.Stat(binaryMarker); err == nil {
		kwbasePath := findKwbaseBinary(installDir)
		if kwbasePath == "" {
			return nil, fmt.Errorf("kwbase binary not found")
		}
		cmdArgs := []string{"sql", "--insecure", "-e", sql}
		if sqlDatabase != "defaultdb" {
			cmdArgs = append(cmdArgs, "--database", sqlDatabase)
		}
		if sqlUser != "root" {
			cmdArgs = append(cmdArgs, "--user", sqlUser)
		}
		cmd := exec.Command(kwbasePath, cmdArgs...)
		return cmd.CombinedOutput()
	}

	return nil, fmt.Errorf("KWDB is not installed")
}

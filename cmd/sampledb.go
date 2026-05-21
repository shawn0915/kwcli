package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/shawn0915/kwcli/pkg/output"
	"github.com/shawn0915/kwcli/pkg/sampledb"
	"github.com/spf13/cobra"
)

// sampledbCmd represents the sampledb command
var sampledbCmd = &cobra.Command{
	Use:   "sampledb",
	Short: "Manage KWDB SampleDB (Smart Meter model)",
	Long:  "Initialize schema, generate data, and run scenario queries for the KWDB Smart Meter sample database.",
}

var sampledbInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create databases and tables for Smart Meter model",
	Run: func(cmd *cobra.Command, args []string) {
		runner, err := sampledb.NewRunner()
		if err != nil {
			output.PrintError("sampledb init", err)
			os.Exit(1)
		}

		if output.JSONOutput {
			if err := runner.ExecBatch(sampledb.RDBSchema); err != nil {
				output.PrintJSON("sampledb init", map[string]interface{}{
					"step":    "rdb",
					"success": false,
				}, err)
				os.Exit(1)
			}
			if err := runner.ExecBatch(sampledb.TSDBSchema); err != nil {
				output.PrintJSON("sampledb init", map[string]interface{}{
					"step":    "tsdb",
					"success": false,
				}, err)
				os.Exit(1)
			}
			output.PrintJSON("sampledb init", map[string]interface{}{
				"steps":   []string{"rdb", "tsdb"},
				"success": true,
			}, nil)
			return
		}

		fmt.Println("Creating databases and tables...")
		fmt.Println("  [1/2] Creating relational database (rdb)...")
		if err := runner.ExecBatch(sampledb.RDBSchema); err != nil {
			fmt.Printf("Failed to create RDB schema: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("  [2/2] Creating time-series database (tsdb)...")
		if err := runner.ExecBatch(sampledb.TSDBSchema); err != nil {
			fmt.Printf("Failed to create TSDB schema: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Schema initialized successfully!")
		fmt.Println("Run 'kwcli sampledb generate' to populate sample data.")
	},
}

var sampledbGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate sample data for Smart Meter model",
	Run: func(cmd *cobra.Command, args []string) {
		runner, err := sampledb.NewRunner()
		if err != nil {
			output.PrintError("sampledb generate", err)
			os.Exit(1)
		}

		if output.JSONOutput {
			if err := runner.ExecBatch(sampledb.GenerateRDBData()); err != nil {
				output.PrintJSON("sampledb generate", map[string]interface{}{
					"step":    "rdb_data",
					"success": false,
				}, err)
				os.Exit(1)
			}
			if err := runner.ExecBatch(sampledb.GenerateTSDBData()); err != nil {
				output.PrintJSON("sampledb generate", map[string]interface{}{
					"step":    "tsdb_data",
					"success": false,
				}, err)
				os.Exit(1)
			}
			output.PrintJSON("sampledb generate", map[string]interface{}{
				"steps":   []string{"rdb_data", "tsdb_data"},
				"success": true,
			}, nil)
			return
		}

		fmt.Println("Generating sample data...")
		fmt.Println("  [1/2] Generating relationship data (users, areas, meters, rules)...")
		if err := runner.ExecBatch(sampledb.GenerateRDBData()); err != nil {
			fmt.Printf("Failed to generate RDB data: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("  [2/2] Generating time-series data (10,000 readings)...")
		if err := runner.ExecBatch(sampledb.GenerateTSDBData()); err != nil {
			fmt.Printf("Failed to generate TSDB data: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Sample data generated successfully!")
		fmt.Println("Run 'kwcli sampledb list' to see available scenarios.")
	},
}

var sampledbRunAll bool

var sampledbRunCmd = &cobra.Command{
	Use:   "run [scenario]",
	Short: "Run a specific scenario or all scenarios",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runner, err := sampledb.NewRunner()
		if err != nil {
			output.PrintError("sampledb run", err)
			os.Exit(1)
		}

		var scenarios []sampledb.Scenario
		if sampledbRunAll {
			scenarios = sampledb.AllScenarios
		} else if len(args) > 0 {
			s := sampledb.FindScenario(args[0])
			if s == nil {
				output.PrintError("sampledb run", fmt.Errorf("scenario not found: %s", args[0]))
				os.Exit(1)
			}
			scenarios = []sampledb.Scenario{*s}
		} else {
			fmt.Println("Please specify a scenario name or use --all flag")
			os.Exit(1)
		}

		if output.JSONOutput {
			results := make([]map[string]interface{}, 0, len(scenarios))
			for _, s := range scenarios {
				err := runner.Exec(s.SQL)
				results = append(results, map[string]interface{}{
					"name":    s.Name,
					"title":   s.Title,
					"success": err == nil,
				})
			}
			output.PrintJSON("sampledb run", results, nil)
			return
		}

		for _, s := range scenarios {
			fmt.Printf("\n=== Running: %s ===\n", s.Title)
			fmt.Printf("Description: %s\n", s.Description)
			if err := runner.Exec(s.SQL); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}
	},
}

var sampledbListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available scenarios",
	Run: func(cmd *cobra.Command, args []string) {
		category, _ := cmd.Flags().GetString("category")

		scenarios := sampledb.AllScenarios
		if category != "" {
			var filtered []sampledb.Scenario
			for _, s := range scenarios {
				if s.Category == category {
					filtered = append(filtered, s)
				}
			}
			scenarios = filtered
		}

		if output.JSONOutput {
			items := make([]map[string]string, 0, len(scenarios))
			for _, s := range scenarios {
				items = append(items, map[string]string{
					"name":        s.Name,
					"title":       s.Title,
					"description": s.Description,
					"category":    s.Category,
				})
			}
			output.PrintJSON("sampledb list", items, nil)
			return
		}

		fmt.Println("Available Scenarios:")
		fmt.Println("====================")
		currentCategory := ""
		for _, s := range scenarios {
			if s.Category != currentCategory {
				currentCategory = s.Category
				fmt.Printf("\n[%s]\n", strings.Title(currentCategory))
			}
			fmt.Printf("  %-30s %s\n", s.Name, s.Title)
		}
	},
}

var sampledbCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove all sample data",
	Run: func(cmd *cobra.Command, args []string) {
		runner, err := sampledb.NewRunner()
		if err != nil {
			output.PrintError("sampledb clean", err)
			os.Exit(1)
		}

		cleanSQL := `
DELETE FROM tsdb.meter_data WHERE 1=1;
TRUNCATE TABLE rdb.meter_info CASCADE;
TRUNCATE TABLE rdb.user_info CASCADE;
TRUNCATE TABLE rdb.area_info CASCADE;
TRUNCATE TABLE rdb.alarm_rules CASCADE;
`
		if output.JSONOutput {
			if err := runner.ExecBatch(cleanSQL); err != nil {
				output.PrintJSON("sampledb clean", nil, err)
				os.Exit(1)
			}
			output.PrintJSON("sampledb clean", map[string]interface{}{"success": true}, nil)
			return
		}

		fmt.Println("Cleaning sample data...")
		if err := runner.ExecBatch(cleanSQL); err != nil {
			fmt.Printf("Failed to clean sample data: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Sample data cleaned successfully!")
	},
}

var sampledbStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check SampleDB status",
	Run: func(cmd *cobra.Command, args []string) {
		runner, err := sampledb.NewRunner()
		if err != nil {
			output.PrintError("sampledb status", err)
			os.Exit(1)
		}

		// Check RDB tables
		rdbResult, _ := runner.ExecCapture("SELECT COUNT(*) FROM rdb.meter_info")
		rdbExists := strings.Contains(rdbResult, "0") || strings.Contains(rdbResult, "COUNT")

		// Check TSDB tables
		tsdbResult, _ := runner.ExecCapture("SELECT COUNT(*) FROM tsdb.meter_data")
		tsdbExists := strings.Contains(tsdbResult, "0") || strings.Contains(tsdbResult, "COUNT")

		if output.JSONOutput {
			output.PrintJSON("sampledb status", map[string]interface{}{
				"rdb_exists":  rdbExists,
				"tsdb_exists": tsdbExists,
			}, nil)
			return
		}

		fmt.Println("SampleDB Status:")
		fmt.Printf("  rdb  database: %s\n", boolStr(rdbExists, "exists", "not found"))
		fmt.Printf("  tsdb database: %s\n", boolStr(tsdbExists, "exists", "not found"))

		if !rdbExists || !tsdbExists {
			fmt.Println()
			fmt.Println("SampleDB is not initialized. Run:")
			fmt.Println("  kwcli sampledb init")
			fmt.Println("  kwcli sampledb generate")
		}
	},
}

func boolStr(b bool, trueStr, falseStr string) string {
	if b {
		return trueStr
	}
	return falseStr
}

func init() {
	rootCmd.AddCommand(sampledbCmd)
	sampledbCmd.AddCommand(sampledbInitCmd)
	sampledbCmd.AddCommand(sampledbGenerateCmd)
	sampledbCmd.AddCommand(sampledbListCmd)
	sampledbCmd.AddCommand(sampledbRunCmd)
	sampledbCmd.AddCommand(sampledbCleanCmd)
	sampledbCmd.AddCommand(sampledbStatusCmd)

	sampledbRunCmd.Flags().BoolVar(&sampledbRunAll, "all", false, "Run all scenarios")
	sampledbListCmd.Flags().StringP("category", "c", "", "Filter scenarios by category (basic, cross-mode, window)")

	// Dynamic completion for sampledb run [scenario]
	sampledbRunCmd.ValidArgsFunction = sampledbScenarioCompletionFunc

	// Flag completion for --category
	sampledbListCmd.RegisterFlagCompletionFunc("category", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"basic", "cross-mode", "window"}, cobra.ShellCompDirectiveNoFileComp
	})
}

// sampledbScenarioCompletionFunc provides dynamic completion for scenario names
func sampledbScenarioCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var names []string
	for _, s := range sampledb.AllScenarios {
		names = append(names, s.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

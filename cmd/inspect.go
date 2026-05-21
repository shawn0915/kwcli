package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/shawn0915/kwcli/pkg/inspect"
	"github.com/shawn0915/kwcli/pkg/output"
	"github.com/spf13/cobra"
)

// inspectCmd represents the inspect command
var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Run KWDB health inspection",
	Long: `Run KWDB health inspection to check service status, log anomalies,
connection counts, slow queries, storage usage, and SampleDB health.

Inspection items:
  - status       Service status (is KWDB running?)
  - logs         Log anomalies (ERROR/FATAL keyword scan)
  - connections  Connection count (approaching limit?)
  - slow_queries Slow/running queries
  - storage      Storage space usage
  - sampledb     SampleDB health check

Examples:
  kwcli inspect run --template standard
  kwcli inspect run --items status,logs,connections
  kwcli inspect run --output report.md --format markdown
  kwcli inspect run --output report.html --format html`,
}

var (
	inspectTemplate string
	inspectItems    string
	inspectFormat   string
	inspectOutput   string
)

// inspectRunCmd represents the inspect run command
var inspectRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run inspection",
	Long:  "Run a KWDB health inspection and generate a report",
	Run: func(cmd *cobra.Command, args []string) {
		var items []string

		if inspectItems != "" {
			items = strings.Split(inspectItems, ",")
			for i := range items {
				items[i] = strings.TrimSpace(items[i])
			}
		} else {
			switch inspectTemplate {
			case "standard":
				items = inspect.StandardItems()
			case "":
				items = inspect.StandardItems()
			default:
				output.PrintError("inspect run", fmt.Errorf("unknown template: %s (supported: standard)", inspectTemplate))
				os.Exit(1)
			}
		}

		report, err := inspect.RunInspection(items)
		if err != nil {
			output.PrintError("inspect run", err)
			os.Exit(1)
		}

		if output.JSONOutput {
			output.PrintJSON("inspect run", report, nil)
			return
		}

		if inspectFormat == "" && inspectOutput == "" {
			// Default: print to console
			fmt.Println("=== KWDB 巡检报告 ===")
			fmt.Printf("巡检时间: %s\n\n", report.Timestamp)

			for _, item := range report.Items {
				var statusIcon string
				switch item.Status {
				case "pass":
					statusIcon = "✅"
				case "warn":
					statusIcon = "⚠️"
				case "fail":
					statusIcon = "❌"
				case "error":
					statusIcon = "🔴"
				}
				fmt.Printf("%s [%s] %s\n", statusIcon, strings.ToUpper(item.Status), item.Name)
				fmt.Printf("   %s\n\n", item.Detail)
			}

			fmt.Println("--- 汇总 ---")
			fmt.Printf("总计: %d | 通过: %d | 警告: %d | 失败: %d\n",
				report.Summary.Total, report.Summary.Pass, report.Summary.Warn, report.Summary.Fail)
			return
		}

		// Determine format from output file extension if not specified
		if inspectFormat == "" && inspectOutput != "" {
			if strings.HasSuffix(inspectOutput, ".json") {
				inspectFormat = "json"
			} else if strings.HasSuffix(inspectOutput, ".html") || strings.HasSuffix(inspectOutput, ".htm") {
				inspectFormat = "html"
			} else {
				inspectFormat = "markdown"
			}
		}

		if inspectFormat == "" {
			inspectFormat = "markdown"
		}

		if err := inspect.WriteReport(report, inspectFormat, inspectOutput); err != nil {
			output.PrintError("inspect run", fmt.Errorf("failed to write report: %v", err))
			os.Exit(1)
		}

		if inspectOutput != "" {
			fmt.Printf("Inspection report saved to %s (%s format)\n", inspectOutput, inspectFormat)
		}
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)
	inspectCmd.AddCommand(inspectRunCmd)

	inspectRunCmd.Flags().StringVar(&inspectTemplate, "template", "standard", "Inspection template (standard)")
	inspectRunCmd.Flags().StringVar(&inspectItems, "items", "", "Comma-separated inspection items (status,logs,connections,slow_queries,storage,sampledb)")
	inspectRunCmd.Flags().StringVar(&inspectFormat, "format", "", "Output format: markdown, json, html")
	inspectRunCmd.Flags().StringVar(&inspectOutput, "output", "", "Output file path")

	// Flag completion for --template
	inspectRunCmd.RegisterFlagCompletionFunc("template", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"standard"}, cobra.ShellCompDirectiveNoFileComp
	})

	// Flag completion for --items
	inspectRunCmd.RegisterFlagCompletionFunc("items", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"status", "logs", "connections", "slow_queries", "storage", "sampledb"}, cobra.ShellCompDirectiveNoFileComp
	})

	// Flag completion for --format
	inspectRunCmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"markdown", "json", "html"}, cobra.ShellCompDirectiveNoFileComp
	})
}

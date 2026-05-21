package cmd

import (
	"fmt"
	"os"

	"github.com/shawn0915/kwcli/pkg/output"
	"github.com/shawn0915/kwcli/pkg/perf"
	"github.com/spf13/cobra"
)

// perfCmd represents the perf command
var perfCmd = &cobra.Command{
	Use:   "perf",
	Short: "KWDB performance monitoring",
	Long: `KWDB performance monitoring and diagnostics.

Commands:
  snapshot    Collect a performance snapshot

The snapshot collects:
  - QPS / TPS (from KWDB metrics endpoint)
  - Active connections
  - Running queries (TOP 10)
  - Storage space usage

Examples:
  kwcli perf snapshot
  kwcli perf snapshot --json
  kwcli perf snapshot --output snapshot.json --format json`,
}

var (
	perfFormat string
	perfOutput string
)

// perfSnapshotCmd represents the perf snapshot command
var perfSnapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Collect a performance snapshot",
	Long:  "Collect a one-time performance snapshot of the KWDB instance",
	Run: func(cmd *cobra.Command, args []string) {
		snapshot, err := perf.TakeSnapshot()
		if err != nil {
			output.PrintError("perf snapshot", err)
			os.Exit(1)
		}

		if output.JSONOutput {
			output.PrintJSON("perf snapshot", snapshot, nil)
			return
		}

		if perfFormat == "json" {
			if perfOutput != "" {
				if err := perf.WriteSnapshot(snapshot, "json", perfOutput); err != nil {
					output.PrintError("perf snapshot", fmt.Errorf("failed to write snapshot: %v", err))
					os.Exit(1)
				}
				fmt.Printf("Performance snapshot saved to %s\n", perfOutput)
			} else {
				content, _ := perf.SnapshotToJSON(snapshot)
				fmt.Println(content)
			}
			return
		}

		perf.PrintSnapshot(snapshot)
	},
}

func init() {
	rootCmd.AddCommand(perfCmd)
	perfCmd.AddCommand(perfSnapshotCmd)

	perfSnapshotCmd.Flags().StringVar(&perfFormat, "format", "", "Output format: json")
	perfSnapshotCmd.Flags().StringVar(&perfOutput, "output", "", "Output file path")

	// Flag completion for --format
	perfSnapshotCmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json"}, cobra.ShellCompDirectiveNoFileComp
	})
}

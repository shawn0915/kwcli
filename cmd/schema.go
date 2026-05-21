package cmd

import (
	"fmt"
	"os"

	"github.com/shawn0915/kwcli/pkg/output"
	"github.com/shawn0915/kwcli/pkg/schema"
	"github.com/spf13/cobra"
)

// schemaCmd represents the schema command
var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Manage KWDB database schemas",
	Long:  "Export and inspect KWDB database schemas (DDL, table structure, indexes, tags)",
}

var (
	schemaDumpDb     string
	schemaDumpOutput string
	schemaDumpFormat string
)

// schemaDumpCmd represents the schema dump command
var schemaDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Export database schema",
	Long: `Export the schema of a specified database in DDL or JSON format.

Supported databases:
  - rdb    Relational database (meter_info, user_info, area_info, alarm_rules)
  - tsdb   Time-series database (meter_data)
  - all    Export all databases (rdb + tsdb)

Examples:
  kwcli schema dump --db rdb --output rdb_schema.sql
  kwcli schema dump --db tsdb --format json
  kwcli schema dump --db all --output full_schema.sql`,
	Run: func(cmd *cobra.Command, args []string) {
		if schemaDumpDb == "" {
			fmt.Println("Error: --db flag is required. Specify: rdb, tsdb, or all")
			os.Exit(1)
		}

		info, err := schema.DumpSchema(schemaDumpDb)
		if err != nil {
			output.PrintError("schema dump", err)
			os.Exit(1)
		}

		if output.JSONOutput {
			output.PrintJSON("schema dump", info, nil)
			return
		}

		switch schemaDumpFormat {
		case "json":
			if schemaDumpOutput != "" {
				if err := schema.WriteSchemaJSON(info, schemaDumpOutput); err != nil {
					output.PrintError("schema dump", fmt.Errorf("failed to write JSON: %v", err))
					os.Exit(1)
				}
				fmt.Printf("Schema exported to %s (JSON format)\n", schemaDumpOutput)
			} else {
				schema.WriteSchemaJSON(info, "")
			}
		case "sql", "ddl":
			if schemaDumpOutput != "" {
				if err := schema.WriteSchemaDDL(info, schemaDumpOutput); err != nil {
					output.PrintError("schema dump", fmt.Errorf("failed to write DDL: %v", err))
					os.Exit(1)
				}
				fmt.Printf("Schema exported to %s (SQL format)\n", schemaDumpOutput)
			} else {
				fmt.Print(info.DDL)
			}
		default:
			fmt.Printf("Unsupported format: %s (supported: sql, json)\n", schemaDumpFormat)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(schemaCmd)
	schemaCmd.AddCommand(schemaDumpCmd)

	schemaDumpCmd.Flags().StringVar(&schemaDumpDb, "db", "", "Database to dump (rdb, tsdb, all)")
	schemaDumpCmd.Flags().StringVar(&schemaDumpOutput, "output", "", "Output file path (default: stdout)")
	schemaDumpCmd.Flags().StringVar(&schemaDumpFormat, "format", "sql", "Output format: sql, json")

	// Flag completion for --db
	schemaDumpCmd.RegisterFlagCompletionFunc("db", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"rdb", "tsdb", "all"}, cobra.ShellCompDirectiveNoFileComp
	})

	// Flag completion for --format
	schemaDumpCmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"sql", "json"}, cobra.ShellCompDirectiveNoFileComp
	})
}

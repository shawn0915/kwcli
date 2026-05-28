package cmd

import (
	"fmt"
	"os"

	"github.com/shawn0915/kwcli/pkg/mcp"
	"github.com/shawn0915/kwcli/pkg/output"
	"github.com/spf13/cobra"
)

// mcpCmd represents the mcp command
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "MCP Server for AI Agent integration",
	Long: `MCP (Model Context Protocol) server that exposes kwcli as tools for AI agents.

Two transport modes:
  - stdio (default) - JSON-RPC over stdin/stdout, for local agents
  - sse - HTTP Server-Sent Events, for remote agents

Examples:
  kwcli mcp serve                   Start stdio MCP server
  kwcli mcp serve --transport sse    Start SSE MCP server on :8080
  kwcli mcp serve --port 9090       Start SSE on custom port
  kwcli mcp tools                   List exposed tools`,
}

// mcpServeCmd represents the mcp serve command
var mcpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server",
	Long: `Start the MCP server to expose kwcli tools to AI agents.

In stdio mode, the server reads JSON-RPC requests from stdin and writes
responses to stdout. This is the standard mode for local agents like
Claude Code.

In SSE mode, the server listens on an HTTP port for remote connections.

Examples:
  kwcli mcp serve
  kwcli mcp serve --transport sse --port 8080`,
	Run: func(cmd *cobra.Command, args []string) {
		transport, _ := cmd.Flags().GetString("transport")
		port, _ := cmd.Flags().GetInt("port")

		server := mcp.NewServer(transport, port)
		mcp.RegisterDefaultTools(server)

		if transport == "sse" {
			addr := fmt.Sprintf(":%d", port)
			fmt.Printf("MCP Server starting in SSE mode on %s\n", addr)
			fmt.Println("Available tools:")
			for _, t := range server.GetTools() {
				fmt.Printf("  - %s: %s\n", t.Name, t.Description)
			}
		} else {
			fmt.Fprintln(os.Stderr, "MCP Server starting in stdio mode")
			fmt.Fprintln(os.Stderr, "Listening for JSON-RPC requests on stdin...")
		}

		if err := server.Start(); err != nil {
			output.PrintError("mcp serve", err)
			os.Exit(1)
		}
	},
}

// mcpToolsCmd represents the mcp tools command
var mcpToolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "List available MCP tools",
	Run: func(cmd *cobra.Command, args []string) {
		server := mcp.NewServer("stdio", 0)
		mcp.RegisterDefaultTools(server)
		tools := server.GetTools()

		if output.JSONOutput {
			output.PrintJSON("mcp tools", tools, nil)
			return
		}

		fmt.Println("Available MCP Tools:")
		fmt.Println("")
		for _, t := range tools {
			fmt.Printf("  %s\n", t.Name)
			fmt.Printf("    Description: %s\n", t.Description)
			fmt.Printf("    Command: %s\n", t.Command)
			if len(t.Parameters) > 0 {
				fmt.Println("    Parameters:")
				for k, v := range t.Parameters {
					fmt.Printf("      - %s: %s\n", k, v)
				}
			}
			fmt.Println()
		}
		fmt.Printf("Total: %d tool(s)\n", len(tools))
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.AddCommand(mcpServeCmd)
	mcpCmd.AddCommand(mcpToolsCmd)

	mcpServeCmd.Flags().String("transport", "stdio", "Transport mode: stdio, sse")
	mcpServeCmd.Flags().Int("port", 8080, "Port for SSE mode")

	// Flag completion
	mcpServeCmd.RegisterFlagCompletionFunc("transport", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"stdio", "sse"}, cobra.ShellCompDirectiveNoFileComp
	})
}

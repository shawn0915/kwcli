package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/shawn0915/kwcli/pkg/ai"
	"github.com/shawn0915/kwcli/pkg/output"
	"github.com/shawn0915/kwcli/pkg/schema"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// aiCmd represents the ai command
var aiCmd = &cobra.Command{
	Use:   "ai [question]",
	Short: "AI-powered KWDB assistant",
	Long: `AI-powered assistant for KWDB operations.

Uses installed Skills and LLM to understand natural language queries
and generate/execute KWDB SQL commands.

Modes:
  Direct query:    kwcli ai "query description"
  Interactive:     kwcli ai --interactive
  Dry-run:         kwcli ai --dry-run "query description"
  With Skill:      kwcli ai --skill <name> "query description"

Examples:
  kwcli ai "查询最近一小时温度超过 40 度的设备"
  kwcli ai --dry-run "创建一个时序库 iot_db 和表 sensor_data"
  kwcli ai --interactive`,
}

var (
	aiSkill       string
	aiDryRun      bool
	aiInteractive bool
)

// aiRun represents the main AI command execution
var aiRun = &cobra.Command{
	Use:   "ask [question]",
	Short: "Ask a question to the AI assistant",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if LLM is configured
		if !isLLMConfigured() {
			fmt.Println("AI mode requires LLM configuration.")
			fmt.Println("Please configure using:")
			fmt.Println("  kwcli config set llm.provider <provider>")
			fmt.Println("  kwcli config set llm.api_key <key>")
			fmt.Println("  kwcli config set llm.model <model>")
			fmt.Println()
			fmt.Println("Supported providers: openai, deepseek")
			return
		}

		// Interactive mode
		if aiInteractive {
			startInteractiveSession()
			return
		}

		// Single query mode
		query := strings.Join(args, " ")
		if query == "" {
			fmt.Println("Usage: kwcli ai \"your question\"")
			fmt.Println("  or:  kwcli ai --interactive")
			os.Exit(1)
		}

		processQuery(query)
	},
}

// configSetCmd handles setting configuration values (including llm.*)
var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		viper.Set(key, value)

		configFile := viper.ConfigFileUsed()
		if configFile == "" {
			configFile = GetHomeDir() + "/config.yaml"
		}

		if err := viper.WriteConfigAs(configFile); err != nil {
			// Try to create the config directory and file
			os.MkdirAll(GetHomeDir(), 0755)
			if err := viper.SafeWriteConfigAs(configFile); err != nil {
				if err := viper.WriteConfigAs(configFile); err != nil {
					fmt.Printf("Error: Failed to save config: %v\n", err)
					os.Exit(1)
				}
			}
		}

		fmt.Printf("Configuration updated: %s = %s\n", key, value)
		fmt.Printf("Config file: %s\n", configFile)
	},
}

// configShowCmd shows configuration values
var configShowCmd = &cobra.Command{
	Use:   "show [key]",
	Short: "Show configuration values",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// Show all config
			settings := viper.AllSettings()
			if output.JSONOutput {
				output.PrintJSON("config show", settings, nil)
				return
			}
			for k, v := range settings {
				if k == "llm" {
					if llmMap, ok := v.(map[string]interface{}); ok {
						fmt.Println("llm:")
						for lk, lv := range llmMap {
							if lk == "api_key" {
								fmt.Printf("  %s = %s\n", lk, maskKey(fmt.Sprintf("%v", lv)))
							} else {
								fmt.Printf("  %s = %v\n", lk, lv)
							}
						}
					}
				} else {
					fmt.Printf("%s = %v\n", k, v)
				}
			}
		} else {
			key := args[0]
			value := viper.Get(key)
			if output.JSONOutput {
				output.PrintJSON("config show", map[string]interface{}{key: value}, nil)
				return
			}
			fmt.Printf("%s = %v\n", key, value)
		}
	},
}

func init() {
	rootCmd.AddCommand(aiCmd)
	aiCmd.AddCommand(aiRun)

	aiRun.Flags().StringVar(&aiSkill, "skill", "", "Specify a skill to use")
	aiRun.Flags().BoolVar(&aiDryRun, "dry-run", false, "Generate commands without executing")
	aiRun.Flags().BoolVarP(&aiInteractive, "interactive", "i", false, "Start interactive session")

	// Register config set/show as subcommands of root for llm config
	// (they already exist if general.go has them, but we add them here for completeness)
	_ = configSetCmd
	_ = configShowCmd
}

// isLLMConfigured checks if the LLM configuration is available
func isLLMConfigured() bool {
	provider := viper.GetString("llm.provider")
	apiKey := viper.GetString("llm.api_key")
	return provider != "" && apiKey != ""
}

// getLLMConfig builds the LLM configuration from viper settings
func getLLMConfig() *ai.LLMConfig {
	return &ai.LLMConfig{
		Provider: viper.GetString("llm.provider"),
		Model:    viper.GetString("llm.model"),
		APIKey:   viper.GetString("llm.api_key"),
		BaseURL:  viper.GetString("llm.base_url"),
		Timeout:  viper.GetInt("llm.timeout"),
	}
}

// processQuery processes a single AI query
func processQuery(query string) {
	fmt.Printf("🤖 Processing query: %s\n\n", query)

	// Try to get schema context first
	schemaContext := ""
	schemaInfo, err := schema.DumpSchema("all")
	if err == nil && schemaInfo != nil {
		schemaContext = fmt.Sprintf("Database has %d tables", len(schemaInfo.Tables))
	}

	// Build system prompt
	builder := ai.NewSystemPromptBuilder()
	if aiSkill != "" {
		builder.SkillNames = []string{aiSkill}
	} else {
		builder.SkillsEnabled = true
	}

	systemPrompt, err := builder.BuildSystemPrompt()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to build system prompt: %v\n", err)
		systemPrompt = "You are a KWDB database expert assistant."
	}

	// Build query prompt
	userPrompt := builder.BuildQueryPrompt(query, schemaContext)

	// Create LLM client and send request
	config := getLLMConfig()
	client := ai.NewClient(config)

	msg := []ai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, err := client.Chat(msg)
	if err != nil {
		fmt.Printf("Error: AI request failed: %v\n", err)
		fmt.Println("Please check your LLM configuration and network connection.")
		return
	}

	fmt.Println("AI Response:")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println(response)
	fmt.Println(strings.Repeat("-", 40))

	// Apply guardrails and execute
	if !aiDryRun {
		executor := ai.NewExecutor()
		guardResult := executor.ParseAndGuard(response)

		if guardResult.Category == "read" && guardResult.SQL != "" {
			fmt.Println("\n⚡ Executing read query...")
			executeSQLRaw(guardResult.SQL)
		} else if guardResult.SQL != "" && guardResult.NeedsApproval {
			if executor.ConfirmAction(guardResult) {
				fmt.Println("\n⚡ Executing...")
				executeSQLRaw(guardResult.SQL)
			} else {
				fmt.Println("\n⚠️  Execution cancelled by user.")
			}
		}
	} else {
		fmt.Println("\n💡 DRY RUN mode - no commands were executed.")
	}
}

// startInteractiveSession starts an interactive AI session
func startInteractiveSession() {
	fmt.Println("🤖 KWDB AI Assistant - Interactive Mode")
	fmt.Println("Type 'exit' or 'quit' to end the session.")
	fmt.Println("Type 'clear' to clear the conversation context.")
	fmt.Println()

	config := getLLMConfig()
	client := ai.NewClient(config)
	builder := ai.NewSystemPromptBuilder()

	systemPrompt, err := builder.BuildSystemPrompt()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		systemPrompt = "You are a KWDB database expert assistant."
	}

	messages := []ai.Message{
		{Role: "system", Content: systemPrompt},
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		if input == "clear" {
			messages = []ai.Message{
				{Role: "system", Content: systemPrompt},
			}
			fmt.Println("Conversation context cleared.")
			continue
		}

		messages = append(messages, ai.Message{Role: "user", Content: input})

		response, err := client.Chat(messages)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Println(response)
		fmt.Println()

		messages = append(messages, ai.Message{Role: "assistant", Content: response})
	}
}

// maskKey masks sensitive key values for display
func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

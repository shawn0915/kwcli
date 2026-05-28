package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/shawn0915/kwcli/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version    = "0.1.2"
	commitHash = "unknown"
	buildTime  = "unknown"
	homeDir    string
	cfgFile    string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kwcli",
	Short: "KWDB ecosystem command-line tool",
	Long: fmt.Sprintf(`KWCLI - KWDB ecosystem command-line tool.
kwcli version %s (Git Hash: %s, Built at %s)
Go Version: %s

Adopts a component-based architecture to install,
run, and manage KWDB components with one click.

Quick Start:
    kwcli list              List available components
    kwcli install <comp>    Install a component
    kwcli <comp> start      Start a component

Example:
    kwcli playground start  Launch KWDB Playground interactive platform
    kwcli completion bash   Generate bash completion script
    kwcli completion zsh    Generate zsh completion script`,
		version, commitHash, buildTime, runtime.Version()),
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Set version template with build time and commit hash
	rootCmd.SetVersionTemplate(fmt.Sprintf("kwcli version {{.Version}} (Git Hash: %s, Built at %s)\n", commitHash, buildTime))

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kwcli/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&homeDir, "home", "", "KWCLI data directory (default is $HOME/.kwcli)")
	rootCmd.PersistentFlags().BoolVar(&output.JSONOutput, "json", false, "Output in JSON format")

	// Add completion command
	rootCmd.AddCommand(completionCmd)
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Set home directory
		if homeDir != "" {
			viper.Set("home", homeDir)
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error finding home directory:", err)
				os.Exit(1)
			}
			homeDir = home + "/.kwcli"
			viper.Set("home", homeDir)
		}

		// Set config search paths
		viper.AddConfigPath(homeDir)
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		// Update homeDir from config if not set
		if homeDir == "" && viper.IsSet("home") {
			homeDir = viper.GetString("home")
		}
	}

	// Set default source if not configured
	if !viper.IsSet("source") {
		viper.Set("source", "atomgit")
	}

	// Ensure directories exist
	ensureDirs()
}

func ensureDirs() {
	dirs := []string{
		homeDir,
		homeDir + "/components",
		homeDir + "/data",
		homeDir + "/bin",
		homeDir + "/skills",
	}
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}
}

// GetHomeDir returns the KWCLI home directory
func GetHomeDir() string {
	if homeDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ".kwcli"
		}
		homeDir = home + "/.kwcli"
	}
	return homeDir
}

// GetDefaultSource returns the default repository source from config
func GetDefaultSource() string {
	source := viper.GetString("source")
	if source == "" {
		return "atomgit"
	}
	return source
}

// GetComponentsDir returns the components directory
func GetComponentsDir() string {
	return GetHomeDir() + "/components"
}

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/shawn0915/kwcli/pkg/component"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var globalSource string

// listCmd lists all available components
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available components",
	Run: func(cmd *cobra.Command, args []string) {
		listComponents()
	},
}

// installCmd installs a component
var installCmd = &cobra.Command{
	Use:   "install [component]",
	Short: "Install a component",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		installComponent(args[0])
	},
}

// updateCmd updates a component
var updateCmd = &cobra.Command{
	Use:   "update [component]",
	Short: "Update a component to the latest version",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		updateComponent(args[0])
	},
}

// uninstallCmd uninstalls a component
var uninstallCmd = &cobra.Command{
	Use:   "uninstall [component]",
	Short: "Uninstall a component",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		uninstallComponent(args[0])
	},
}

// statusCmd shows running components status
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show running components status",
	Run: func(cmd *cobra.Command, args []string) {
		showStatus()
	},
}

// sourceCmd manages the global default repository source
var sourceCmd = &cobra.Command{
	Use:   "source [github|atomgit]",
	Short: "Show or set the global default repository source",
	Long: `Show or set the global default repository source.

Without arguments, shows the current default source.
With an argument, sets the default source (github or atomgit).

This affects all commands that use --source=auto (the default).`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			showSource()
		} else {
			setSource(args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(sourceCmd)

	installCmd.Flags().StringVar(&globalSource, "source", "auto", "Repository source: auto (use global config), github, atomgit")
	updateCmd.Flags().StringVar(&globalSource, "source", "auto", "Repository source: auto (use global config), github, atomgit")

	// Dynamic completion for component names
	installCmd.ValidArgsFunction = componentCompletionFunc
	updateCmd.ValidArgsFunction = componentCompletionFunc
	uninstallCmd.ValidArgsFunction = componentCompletionFunc

	// Flag completion for --source
	installCmd.RegisterFlagCompletionFunc("source", sourceCompletionFunc)
	updateCmd.RegisterFlagCompletionFunc("source", sourceCompletionFunc)

	// source command arg completion
	sourceCmd.ValidArgs = []string{"github", "atomgit"}
}

func listComponents() {
	components := component.ListComponents()

	fmt.Println("Available Components:")
	fmt.Printf("%-15s %-12s %s\n", "Component", "Status", "Repository")
	fmt.Println("------------------------------------------------------------")

	for _, c := range components {
		status := component.GetInstallStatus(c.Name)
		fmt.Printf("%-15s %-12s %s\n", c.Name, status, c.RepoURL)
	}
}

func installComponent(name string) {
	c := component.GetComponent(name)
	if c == nil {
		fmt.Printf("Unknown component: %s\n", name)
		fmt.Println("Use `kwcli list` to see available components.")
		os.Exit(1)
	}

	source := resolveSourceFlag(globalSource)
	var compSource component.Source
	switch source {
	case "github":
		compSource = component.SourceGitHub
	case "atomgit":
		compSource = component.SourceAtomGit
	default:
		compSource = component.SourceAuto
	}

	fmt.Printf("Installing component %s from %s...\n", name, source)
	if err := c.InstallWithVersionAndSource("", compSource); err != nil {
		fmt.Printf("Failed to install component %s: %v\n", name, err)
		os.Exit(1)
	}
}

func updateComponent(name string) {
	c := component.GetComponent(name)
	if c == nil {
		fmt.Printf("Unknown component: %s\n", name)
		os.Exit(1)
	}

	source := resolveSourceFlag(globalSource)
	var compSource component.Source
	switch source {
	case "github":
		compSource = component.SourceGitHub
	case "atomgit":
		compSource = component.SourceAtomGit
	default:
		compSource = component.SourceAuto
	}

	fmt.Printf("Updating component %s from %s...\n", name, source)
	if err := c.UpdateWithSource(compSource); err != nil {
		fmt.Printf("Failed to update component %s: %v\n", name, err)
		os.Exit(1)
	}
}

func uninstallComponent(name string) {
	c := component.GetComponent(name)
	if c == nil {
		fmt.Printf("Unknown component: %s\n", name)
		os.Exit(1)
	}

	if err := c.Uninstall(); err != nil {
		fmt.Printf("Failed to uninstall component %s: %v\n", name, err)
		os.Exit(1)
	}
}

func showStatus() {
	fmt.Println("Running Components Status (Feature in development)...")

	// TODO: Implement process/container level status tracking
}

func showSource() {
	source := GetDefaultSource()
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		configFile = GetHomeDir() + "/config.yaml"
	}
	fmt.Printf("Default repository source: %s\n", source)
	fmt.Printf("Config file: %s\n", configFile)
	fmt.Println("")
	fmt.Println("To change the default source, run:")
	fmt.Println("  kwcli source github")
	fmt.Println("  kwcli source atomgit")
}

func setSource(source string) {
	source = strings.ToLower(strings.TrimSpace(source))
	if source != "github" && source != "atomgit" {
		fmt.Println("Invalid source. Must be 'github' or 'atomgit'.")
		os.Exit(1)
	}

	viper.Set("source", source)
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		configFile = GetHomeDir() + "/config.yaml"
	}
	if err := viper.WriteConfigAs(configFile); err != nil {
		fmt.Printf("Failed to save config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Default repository source set to: %s\n", source)
	fmt.Printf("Config saved to: %s\n", configFile)
}

// componentCompletionFunc provides dynamic completion for component names
func componentCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	components := component.ListComponents()
	names := make([]string, len(components))
	for i, c := range components {
		names[i] = c.Name
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

// sourceCompletionFunc provides completion for --source flag values
func sourceCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"auto", "github", "atomgit"}, cobra.ShellCompDirectiveNoFileComp
}

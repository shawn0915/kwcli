package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/shawn0915/kwcli/pkg/output"
	"github.com/shawn0915/kwcli/pkg/skill"
	"github.com/spf13/cobra"
)

// skillCmd represents the skill command
var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage KWDB Agent Skills",
	Long: `Manage KWDB Agent Skills - install, list, update, uninstall, and view skills.

Skills are KWDB-specific knowledge packages that help AI agents understand
and interact with KWDB databases. They contain metadata, reference documents,
and command templates.

Examples:
  kwcli skill list                        List installed skills
  kwcli skill install kwdb-text2sql-aiot  Install a skill
  kwcli skill info kwdb-text2sql-aiot     View skill details`,
}

// skillListCmd represents the skill list command
var skillListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed skills",
	Run: func(cmd *cobra.Command, args []string) {
		skills, err := skill.ListInstalled()
		if err != nil {
			output.PrintError("skill list", err)
			os.Exit(1)
		}

		if output.JSONOutput {
			output.PrintJSON("skill list", skills, nil)
			return
		}

		if len(skills) == 0 {
			fmt.Println("No skills installed.")
			fmt.Println("Use `kwcli skill install <name>` to install a skill.")
			return
		}

		fmt.Println("Installed Skills:")
		fmt.Println(strings.Repeat("-", 60))
		for _, s := range skills {
			fmt.Printf("%-25s v%-8s %s\n", s.Name, s.Version, s.Description)
		}
		fmt.Println(strings.Repeat("-", 60))
		fmt.Printf("Total: %d skill(s)\n", len(skills))
	},
}

// skillInstallCmd represents the skill install command
var skillInstallCmd = &cobra.Command{
	Use:   "install [name]",
	Short: "Install a skill",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		version, _ := cmd.Flags().GetString("version")
		source, _ := cmd.Flags().GetString("source")

		if source == "auto" {
			source = GetDefaultSource()
		}

		if err := skill.Install(name, version, source); err != nil {
			output.PrintError("skill install", err)
			os.Exit(1)
		}
	},
}

// skillUpdateCmd represents the skill update command
var skillUpdateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update a skill to the latest version",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if err := skill.Update(name); err != nil {
			output.PrintError("skill update", err)
			os.Exit(1)
		}
	},
}

// skillUninstallCmd represents the skill uninstall command
var skillUninstallCmd = &cobra.Command{
	Use:   "uninstall [name]",
	Short: "Uninstall a skill",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if err := skill.Uninstall(name); err != nil {
			output.PrintError("skill uninstall", err)
			os.Exit(1)
		}
	},
}

// skillInfoCmd represents the skill info command
var skillInfoCmd = &cobra.Command{
	Use:   "info [name]",
	Short: "View skill details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		info, err := skill.GetInfo(name)
		if err != nil {
			output.PrintError("skill info", err)
			os.Exit(1)
		}

		if output.JSONOutput {
			output.PrintJSON("skill info", info, nil)
			return
		}

		fmt.Printf("Name:        %s\n", info.Name)
		fmt.Printf("Version:     %s\n", info.Version)
		fmt.Printf("Description: %s\n", info.Description)
		fmt.Printf("Author:      %s\n", info.Author)
		fmt.Printf("Install Dir: %s\n", info.InstallDir)
		fmt.Printf("Commands:    %d\n", info.Commands)

		if len(info.Triggers) > 0 {
			fmt.Printf("Triggers:    %s\n", strings.Join(info.Triggers, ", "))
		}
		if len(info.References) > 0 {
			fmt.Printf("References:  %s\n", strings.Join(info.References, ", "))
		}
	},
}

// skillScenariosCmd represents the skill scenarios command
var skillScenariosCmd = &cobra.Command{
	Use:   "scenarios [name]",
	Short: "View skill scenarios and command templates",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		scenarios, err := skill.GetScenarios(name)
		if err != nil {
			output.PrintError("skill scenarios", err)
			os.Exit(1)
		}

		if output.JSONOutput {
			output.PrintJSON("skill scenarios", scenarios, nil)
			return
		}

		if len(scenarios) == 0 {
			fmt.Printf("Skill %s has no command templates.\n", name)
			return
		}

		fmt.Printf("Scenarios for %s:\n\n", name)
		for _, s := range scenarios {
			fmt.Printf("  %-20s %s\n", s.Name, s.Description)
		}
	},
}

// skillPreviewCmd represents the skill preview command
var skillPreviewCmd = &cobra.Command{
	Use:   "preview [name]",
	Short: "Preview a skill reference file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		ref, _ := cmd.Flags().GetString("ref")

		if ref == "" {
			fmt.Println("Error: --ref flag is required. Specify a reference file name.")
			os.Exit(1)
		}

		content, err := skill.PreviewReference(name, ref)
		if err != nil {
			output.PrintError("skill preview", err)
			os.Exit(1)
		}

		fmt.Println(content)
	},
}

// skillLinkCmd represents the skill link command
var skillLinkCmd = &cobra.Command{
	Use:   "link [name]",
	Short: "Link a skill to an AI agent",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		agent, _ := cmd.Flags().GetString("agent")

		if agent == "" {
			fmt.Println("Error: --agent flag is required. Specify: claude, codex, openclaw, kimi")
			os.Exit(1)
		}

		var agentType skill.AgentType
		switch strings.ToLower(agent) {
		case "claude":
			agentType = skill.AgentClaude
		case "codex":
			agentType = skill.AgentCodex
		case "openclaw":
			agentType = skill.AgentOpenClaw
		case "kimi":
			agentType = skill.AgentKimi
		default:
			fmt.Printf("Error: unsupported agent: %s (supported: claude, codex, openclaw, kimi)\n", agent)
			os.Exit(1)
		}

		if err := skill.Link(name, agentType); err != nil {
			output.PrintError("skill link", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(skillCmd)
	skillCmd.AddCommand(skillListCmd)
	skillCmd.AddCommand(skillInstallCmd)
	skillCmd.AddCommand(skillUpdateCmd)
	skillCmd.AddCommand(skillUninstallCmd)
	skillCmd.AddCommand(skillInfoCmd)
	skillCmd.AddCommand(skillScenariosCmd)
	skillCmd.AddCommand(skillPreviewCmd)
	skillCmd.AddCommand(skillLinkCmd)

	skillInstallCmd.Flags().String("version", "", "Skill version to install")
	skillInstallCmd.Flags().String("source", "auto", "Repository source: auto, github, atomgit")
	skillPreviewCmd.Flags().String("ref", "", "Reference file to preview")
	skillLinkCmd.Flags().String("agent", "", "Agent to link to (claude, codex, openclaw, kimi)")

	// Dynamic completion for skill names (for commands that need installed skills)
	skillUpdateCmd.ValidArgsFunction = installedSkillCompletionFunc
	skillUninstallCmd.ValidArgsFunction = installedSkillCompletionFunc
	skillInfoCmd.ValidArgsFunction = installedSkillCompletionFunc
	skillScenariosCmd.ValidArgsFunction = installedSkillCompletionFunc
	skillPreviewCmd.ValidArgsFunction = installedSkillCompletionFunc
	skillLinkCmd.ValidArgsFunction = installedSkillCompletionFunc

	// Flag completion
	skillInstallCmd.RegisterFlagCompletionFunc("source", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"auto", "github", "atomgit"}, cobra.ShellCompDirectiveNoFileComp
	})
	skillLinkCmd.RegisterFlagCompletionFunc("agent", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"claude", "codex", "openclaw", "kimi"}, cobra.ShellCompDirectiveNoFileComp
	})
}

// installedSkillCompletionFunc provides completion for installed skill names
func installedSkillCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	skills, err := skill.ListInstalled()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

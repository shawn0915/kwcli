package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

// Helper to check if a command has a subcommand
func findSubCommand(parent *cobra.Command, name string) *cobra.Command {
	for _, cmd := range parent.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}

func TestSkillCmdSetup(t *testing.T) {
	skillCmd := findSubCommand(rootCmd, "skill")
	if skillCmd == nil {
		t.Fatal("skill command not found")
	}

	expectedSubs := []string{"list", "install", "update", "uninstall", "info", "scenarios", "preview", "link"}
	for _, name := range expectedSubs {
		sub := findSubCommand(skillCmd, name)
		if sub == nil {
			t.Errorf("skill subcommand %q not found", name)
		}
	}
}

func TestMCPSetup(t *testing.T) {
	mcpCmd := findSubCommand(rootCmd, "mcp")
	if mcpCmd == nil {
		t.Fatal("mcp command not found")
	}

	expectedSubs := []string{"serve", "tools"}
	for _, name := range expectedSubs {
		sub := findSubCommand(mcpCmd, name)
		if sub == nil {
			t.Errorf("mcp subcommand %q not found", name)
		}
	}
}

func TestAICmdSetup(t *testing.T) {
	aiCmd := findSubCommand(rootCmd, "ai")
	if aiCmd == nil {
		t.Fatal("ai command not found")
	}

	sub := findSubCommand(aiCmd, "ask")
	if sub == nil {
		t.Error("ai subcommand 'ask' not found")
	}
}

func TestMCPServeFlags(t *testing.T) {
	mcpCmd := findSubCommand(rootCmd, "mcp")
	serveCmd := findSubCommand(mcpCmd, "serve")

	transportFlag := serveCmd.Flags().Lookup("transport")
	if transportFlag == nil {
		t.Error("--transport flag not found on mcp serve command")
	}

	portFlag := serveCmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Error("--port flag not found on mcp serve command")
	}
}

func TestSkillInstallFlags(t *testing.T) {
	skillCmd := findSubCommand(rootCmd, "skill")
	installCmd := findSubCommand(skillCmd, "install")

	versionFlag := installCmd.Flags().Lookup("version")
	if versionFlag == nil {
		t.Error("--version flag not found on skill install command")
	}

	sourceFlag := installCmd.Flags().Lookup("source")
	if sourceFlag == nil {
		t.Error("--source flag not found on skill install command")
	}
}

func TestSkillLinkFlags(t *testing.T) {
	skillCmd := findSubCommand(rootCmd, "skill")
	linkCmd := findSubCommand(skillCmd, "link")

	agentFlag := linkCmd.Flags().Lookup("agent")
	if agentFlag == nil {
		t.Error("--agent flag not found on skill link command")
	}
}

func TestSkillPreviewFlags(t *testing.T) {
	skillCmd := findSubCommand(rootCmd, "skill")
	previewCmd := findSubCommand(skillCmd, "preview")

	refFlag := previewCmd.Flags().Lookup("ref")
	if refFlag == nil {
		t.Error("--ref flag not found on skill preview command")
	}
}

func TestAIRunFlags(t *testing.T) {
	aiCmd := findSubCommand(rootCmd, "ai")
	askCmd := findSubCommand(aiCmd, "ask")

	if askCmd == nil {
		t.Skip("ask command not found")
	}

	if flag := askCmd.Flags().Lookup("skill"); flag == nil {
		t.Error("--skill flag not found on ai ask command")
	}
	if flag := askCmd.Flags().Lookup("dry-run"); flag == nil {
		t.Error("--dry-run flag not found on ai ask command")
	}
	if flag := askCmd.Flags().Lookup("interactive"); flag == nil {
		t.Error("--interactive flag not found on ai ask command")
	}
}

func TestMCPServeUseLine(t *testing.T) {
	mcpCmd := findSubCommand(rootCmd, "mcp")
	if mcpCmd == nil {
		t.Skip("mcp command not found")
	}
	serveCmd := findSubCommand(mcpCmd, "serve")
	if serveCmd == nil {
		t.Skip("serve command not found")
	}
	if serveCmd.Use != "serve" {
		t.Errorf("expected Use 'serve', got %q", serveCmd.Use)
	}
}

func TestSkillInfoCmd(t *testing.T) {
	skillCmd := findSubCommand(rootCmd, "skill")
	infoCmd := findSubCommand(skillCmd, "info")
	if infoCmd == nil {
		t.Fatal("skill info command not found")
	}
	if infoCmd.Use != "info [name]" {
		t.Errorf("expected Use 'info [name]', got %q", infoCmd.Use)
	}
}

func TestSkillScenariosCmd(t *testing.T) {
	skillCmd := findSubCommand(rootCmd, "skill")
	scenariosCmd := findSubCommand(skillCmd, "scenarios")
	if scenariosCmd == nil {
		t.Fatal("skill scenarios command not found")
	}
	if scenariosCmd.Use != "scenarios [name]" {
		t.Errorf("expected Use 'scenarios [name]', got %q", scenariosCmd.Use)
	}
}

func TestSkillUninstallCmd(t *testing.T) {
	skillCmd := findSubCommand(rootCmd, "skill")
	uninstallCmd := findSubCommand(skillCmd, "uninstall")
	if uninstallCmd == nil {
		t.Fatal("skill uninstall command not found")
	}
	if uninstallCmd.Use != "uninstall [name]" {
		t.Errorf("expected Use 'uninstall [name]', got %q", uninstallCmd.Use)
	}
}

func TestMCPToolsCmd(t *testing.T) {
	mcpCmd := findSubCommand(rootCmd, "mcp")
	toolsCmd := findSubCommand(mcpCmd, "tools")
	if toolsCmd == nil {
		t.Fatal("mcp tools command not found")
	}
	if toolsCmd.Use != "tools" {
		t.Errorf("expected Use 'tools', got %q", toolsCmd.Use)
	}
}

func TestGlobalJSONFlagAffectsNewCommands(t *testing.T) {
	// Verify the global --json flag is set up for the new command groups
	if rootCmd.PersistentFlags().Lookup("json") == nil {
		t.Error("global --json flag not found")
	}
}

func TestSkillListSupportsJSON(t *testing.T) {
	// Verify the JSON flag is usable with skill list by checking persistent flags
	jsonFlag := rootCmd.PersistentFlags().Lookup("json")
	if jsonFlag == nil {
		t.Error("global --json flag not found on root command")
	}
}

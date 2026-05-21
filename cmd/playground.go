package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/shawn0915/kwcli/pkg/component"
	"github.com/shawn0915/kwcli/pkg/config"
	"github.com/shawn0915/kwcli/pkg/utils"
	"github.com/spf13/cobra"
)

const playgroundPort = 3006

var playgroundShowAll bool
var playgroundRemoveFiles bool
var playgroundPullSource string

// playgroundCmd represents the playground command
var playgroundCmd = &cobra.Command{
	Use:   "playground",
	Short: "Manage KWDB Playground interactive learning platform",
	Long:  "Manage KWDB Playground interactive learning platform",
}

// playgroundStartCmd starts the playground service
var playgroundStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Playground service",
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		startPlayground(port)
	},
}

// playgroundStopCmd stops the playground service
var playgroundStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop Playground service",
	Run: func(cmd *cobra.Command, args []string) {
		stopPlayground()
	},
}

// playgroundUninstallCmd uninstalls the playground service
var playgroundUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall Playground (remove containers, images and optionally files)",
	Run: func(cmd *cobra.Command, args []string) {
		uninstallPlayground(playgroundRemoveFiles)
	},
}

// playgroundRestartCmd restarts the playground service
var playgroundRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart Playground service",
	Run: func(cmd *cobra.Command, args []string) {
		stopPlayground()
		time.Sleep(1 * time.Second)
		startPlayground(playgroundPort)
	},
}

// playgroundStatusCmd checks the playground status
var playgroundStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check Playground running status",
	Run: func(cmd *cobra.Command, args []string) {
		statusPlayground(playgroundShowAll)
	},
}

// playgroundVersionsCmd lists all installed versions
var playgroundVersionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "List installed Playground versions",
	Run: func(cmd *cobra.Command, args []string) {
		comp := component.GetComponent("playground")
		versions := comp.ListInstalledVersions()
		if len(versions) == 0 {
			fmt.Println("No Playground versions installed.")
			fmt.Println("Run 'kwcli playground install' to install a version.")
			return
		}

		currentVersion := comp.GetCurrentVersion()
		fmt.Println("Installed Playground versions:")
		for _, v := range versions {
			if v == currentVersion {
				fmt.Printf("  %s (current)\n", v)
			} else {
				fmt.Printf("  %s\n", v)
			}
		}
	},
}

// playgroundUpgradeCmd upgrades the playground
var playgroundUpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade Playground to the latest version",
	Run: func(cmd *cobra.Command, args []string) {
		upgradePlayground()
	},
}

// playgroundLogsCmd shows playground logs
var playgroundLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View Playground logs",
	Run: func(cmd *cobra.Command, args []string) {
		logsPlayground()
	},
}

// playgroundInstallCmd installs the playground
var playgroundInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Playground",
	Run: func(cmd *cobra.Command, args []string) {
		installPlayground()
	},
}

var (
	playgroundVersion string
	playgroundSource  string
)

func init() {
	rootCmd.AddCommand(playgroundCmd)
	playgroundCmd.AddCommand(playgroundInstallCmd)
	playgroundCmd.AddCommand(playgroundStartCmd)
	playgroundCmd.AddCommand(playgroundStopCmd)
	playgroundCmd.AddCommand(playgroundUninstallCmd)
	playgroundCmd.AddCommand(playgroundRestartCmd)
	playgroundCmd.AddCommand(playgroundStatusCmd)
	playgroundCmd.AddCommand(playgroundUpgradeCmd)
	playgroundCmd.AddCommand(playgroundVersionsCmd)
	playgroundCmd.AddCommand(playgroundLogsCmd)

	playgroundStatusCmd.Flags().BoolVarP(&playgroundShowAll, "all", "a", false, "Show all containers (including stopped)")
	playgroundStatusCmd.Flags().StringVarP(&playgroundVersion, "version", "v", "", "Playground version to check status for (e.g., v1.0.0). Uses current if not specified.")
	playgroundUninstallCmd.Flags().BoolVarP(&playgroundRemoveFiles, "remove-files", "", false, "Also remove downloaded component files")

	playgroundInstallCmd.Flags().StringVarP(&playgroundVersion, "version", "v", "", "Playground version (e.g., v1.0.0). Only v1.0.0 and above are supported.")
	playgroundInstallCmd.Flags().StringVar(&playgroundSource, "source", "auto", "Repository source: auto (use global config), github, atomgit")
	playgroundUpgradeCmd.Flags().StringVar(&playgroundSource, "source", "auto", "Repository source: auto (use global config), github, atomgit")
	playgroundUpgradeCmd.Flags().StringVar(&playgroundPullSource, "registry", "auto", "Image registry: auto (try aliyun first, fallback to default), aliyun, default")
	playgroundStartCmd.Flags().Int("port", playgroundPort, "Playground service port")
	playgroundStartCmd.Flags().Bool("update", false, "Update playground to the latest version before starting")
	playgroundStartCmd.Flags().StringVar(&playgroundPullSource, "source", "auto", "Image source: auto (try aliyun first, fallback to default), default (Docker Hub), aliyun (Aliyun registry)")
	playgroundStartCmd.Flags().StringVarP(&playgroundVersion, "version", "v", "", "Playground version to start (e.g., v1.0.0). Uses latest if not specified.")

	// Flag completion for --source
	playgroundInstallCmd.RegisterFlagCompletionFunc("source", sourceCompletionFunc)
	playgroundUpgradeCmd.RegisterFlagCompletionFunc("source", sourceCompletionFunc)

	// Flag completion for --registry
	playgroundUpgradeCmd.RegisterFlagCompletionFunc("registry", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"auto", "aliyun", "default"}, cobra.ShellCompDirectiveNoFileComp
	})

	// Flag completion for start --source
	playgroundStartCmd.RegisterFlagCompletionFunc("source", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"auto", "default", "aliyun"}, cobra.ShellCompDirectiveNoFileComp
	})
}

func resolveSourceFlag(source string) string {
	if source == "auto" {
		return GetDefaultSource()
	}
	return source
}

func installPlayground() {
	comp := component.GetComponent("playground")

	// Resolve source from global config if auto
	playgroundSource = resolveSourceFlag(playgroundSource)

	// Check version requirement
	if playgroundVersion != "" {
		// Validate version format (v0.x or v0.x.x)
		if len(playgroundVersion) < 4 || playgroundVersion[0] != 'v' {
			fmt.Println("Invalid version format. Use format like: v1.0.0 or v1.0")
			os.Exit(1)
		}
		// Normalize version: v1.0 -> v1.0.0
		version := playgroundVersion[1:]
		parts := strings.Split(version, ".")
		if len(parts) == 2 {
			playgroundVersion = "v" + parts[0] + "." + parts[1] + ".0"
		}
		fmt.Printf("Installing Playground version %s...\n", playgroundVersion)
	} else {
		fmt.Println("Installing Playground (latest version)...")
	}

	// Check if already installed
	if playgroundVersion != "" {
		// If specific version requested, check if that version is already installed
		if comp.IsVersionInstalled(playgroundVersion) {
			fmt.Printf("Version %s is already installed. Use 'kwcli playground upgrade' to update.\n", playgroundVersion)
			fmt.Println("Or run 'kwcli playground uninstall --remove-files' first, then reinstall.")
			return
		}
	} else {
		// No version specified, check if any version is installed
		if comp.IsInstalled() {
			fmt.Println("Playground is already installed. Use 'kwcli playground upgrade' to update.")
			fmt.Println("Or run 'kwcli playground uninstall --remove-files' first, then reinstall.")
			return
		}
	}

	// Parse source
	var source component.Source
	switch playgroundSource {
	case "github":
		source = component.SourceGitHub
	case "atomgit":
		source = component.SourceAtomGit
	default:
		source = component.SourceAuto
	}

	fmt.Printf("  Using source: %s\n", playgroundSource)

	// Install with version and source
	if err := comp.InstallWithVersionAndSource(playgroundVersion, source); err != nil {
		fmt.Printf("Installation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Playground installed successfully!")
	fmt.Println("Run 'kwcli playground start' to start the service.")
}

func startPlayground(port int) {
	fmt.Println("Checking environment...")

	// Check Docker
	if !utils.CheckDocker() {
		fmt.Println("Docker is not installed or not running. Please install Docker first:")
		fmt.Println("   https://docs.docker.com/get-docker/")
		os.Exit(1)
	}

	if !utils.CheckDockerCompose() {
		fmt.Println("Docker Compose is not installed. Please install Docker Compose first:")
		fmt.Println("   https://docs.docker.com/compose/install/")
		os.Exit(1)
	}

	fmt.Println("Docker environment is ready")

	// Get or install playground component
	comp := component.GetComponent("playground")
	if comp == nil {
		fmt.Println("Playground component not found in registry")
		os.Exit(1)
	}

	// Set the version if specified
	if playgroundVersion != "" {
		comp.EnsureInstallDirWithVersion(playgroundVersion)
	}

	// Show version info
	if playgroundVersion != "" {
		fmt.Printf("Using version: %s\n", playgroundVersion)
	} else {
		currentVersion := comp.GetCurrentVersion()
		if currentVersion != "" {
			fmt.Printf("Using version: %s (current)\n", currentVersion)
		} else {
			latestVersion := comp.GetLatestVersion()
			if latestVersion != "" {
				fmt.Printf("Using version: %s (latest)\n", latestVersion)
			}
		}
	}

	// Check if update flag is set
	updateFlag, _ := rootCmd.Flags().GetBool("update")

	if updateFlag || !comp.IsInstalled() {
		if comp.IsInstalled() && updateFlag {
			if err := comp.Update(); err != nil {
				fmt.Printf("Update failed: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := comp.InstallWithVersion(playgroundVersion); err != nil {
				fmt.Printf("Installation failed: %v\n", err)
				os.Exit(1)
			}
		}
	}

	// Ensure install dir is set correctly based on version
	comp.EnsureInstallDirWithVersion(playgroundVersion)

	// Check port
	if !utils.IsPortAvailable(port) {
		fmt.Printf("Port %d is already in use. Please choose another port or stop the occupying process.\n", port)
		os.Exit(1)
	}

	// Start service
	composeFile := comp.GetComposeFile("docker/playground")
	if _, err := os.Stat(composeFile); os.IsNotExist(err) {
		fmt.Printf("Docker Compose file not found: %s\n", composeFile)
		fmt.Println("   The playground repository structure may have changed. Try `kwcli playground upgrade`.")
		os.Exit(1)
	}

	fmt.Printf("Starting Playground service (port: %d)...\n", port)

	composeCmd := utils.GetDockerComposeCmd()

	env := os.Environ()
	env = append(env, fmt.Sprintf("PLAYGROUND_PORT=%d", port))

	// Determine registry to use (auto defaults to aliyun for China users)
	var registry string
	switch playgroundPullSource {
	case "aliyun":
		registry = utils.AliyunRegistry
	case "default":
		registry = utils.DefaultRegistry
	default: // auto and anything else
		registry = utils.AliyunRegistry
	}

	// Save original compose file content
	originalContent, _ := utils.ReadDockerComposeFile(composeFile)

	// If using aliyun, rewrite docker-compose to use aliyun registry
	if registry != utils.DefaultRegistry {
		utils.RewriteDockerComposeImage(composeFile, registry)
	}

	// Track if we need to restore original
	restoreCompose := registry != utils.DefaultRegistry

	runDockerCompose := func() error {
		dockerCmd := exec.Command(composeCmd[0], append(composeCmd[1:], "-f", composeFile, "up", "-d")...)
		dockerCmd.Dir = comp.InstallDir
		dockerCmd.Env = env
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr
		return dockerCmd.Run()
	}

	err := runDockerCompose()

	// If failed with aliyun, try with default registry (for auto mode)
	if err != nil && playgroundPullSource == "auto" && registry == utils.AliyunRegistry {
		// Restore original and try default
		utils.RestoreDockerComposeImage(composeFile, originalContent)
		restoreCompose = false
		registry = utils.DefaultRegistry
		fmt.Println("  Aliyun registry failed, trying default registry...")
		err = runDockerCompose()
	}

	// Restore compose file if needed
	if restoreCompose {
		utils.RestoreDockerComposeImage(composeFile, originalContent)
	}

	if err != nil {
		fmt.Printf("Failed to start Playground: %v\n", err)
		os.Exit(1)
	}

	// Wait for service to be ready
	fmt.Println("Waiting for service to be ready...")
	time.Sleep(3)

	fmt.Println("")
	fmt.Println("KWDB Playground started successfully!")
	fmt.Printf("   Access URL: http://localhost:%d\n", port)
	fmt.Println("")
	fmt.Println("Common commands:")
	fmt.Println("   View logs:   kwcli playground logs")
	fmt.Println("   Stop service: kwcli playground stop")
}

func stopPlayground() {
	comp := component.GetComponent("playground")
	if !comp.IsInstalled() {
		fmt.Println("Playground is not installed.")
		return
	}

	composeFile := comp.GetComposeFile("docker/playground")
	if _, err := os.Stat(composeFile); os.IsNotExist(err) {
		fmt.Println("Docker Compose file not found.")
		return
	}

	composeCmd := utils.GetDockerComposeCmd()

	// Use "stop" instead of "down" to just stop containers without removing them
	dockerCmd := exec.Command(composeCmd[0], append(composeCmd[1:], "-f", composeFile, "stop")...)
	dockerCmd.Dir = comp.InstallDir
	dockerCmd.Stdout = os.Stdout
	dockerCmd.Stderr = os.Stderr

	if err := dockerCmd.Run(); err != nil {
		fmt.Printf("Stop failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Playground stopped.")
}

func uninstallPlayground(removeFiles bool) {
	comp := component.GetComponent("playground")

	composeFile := ""
	if comp.IsInstalled() {
		composeFile = comp.GetComposeFile("docker/playground")
		if _, err := os.Stat(composeFile); os.IsNotExist(err) {
			composeFile = ""
		}
	}

	// If docker compose file exists, remove containers and images
	if composeFile != "" {
		composeCmd := utils.GetDockerComposeCmd()

		// First, try to stop any running containers
		dockerCmd := exec.Command(composeCmd[0], append(composeCmd[1:], "-f", composeFile, "stop")...)
		dockerCmd.Dir = comp.InstallDir
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr
		dockerCmd.Run()

		// Then, remove containers and images
		dockerCmd = exec.Command(composeCmd[0], append(composeCmd[1:], "-f", composeFile, "down", "--rmi", "all")...)
		dockerCmd.Dir = comp.InstallDir
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr

		if err := dockerCmd.Run(); err != nil {
			fmt.Printf("Failed to remove containers/images: %v\n", err)
		}
	}

	// Remove component files if requested
	if removeFiles && comp.IsInstalled() {
		installDir := filepath.Join(config.GetComponentsDir(), "playground")
		if err := os.RemoveAll(installDir); err != nil {
			fmt.Printf("Failed to remove files: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Playground uninstalled (containers, images and files removed).")
	} else if composeFile != "" {
		fmt.Println("Playground uninstalled (containers and images removed).")
		fmt.Println("Component files preserved at ~/.kwcli/components/playground")
		fmt.Println("Use --remove-files to also remove the component files.")
	} else {
		fmt.Println("Playground is not installed.")
	}
}

func statusPlayground(showAll bool) {
	comp := component.GetComponent("playground")
	if !comp.IsInstalled() {
		fmt.Println("Playground is not installed. Run `kwcli playground start` to install and start.")
		return
	}

	// Set version if specified
	if playgroundVersion == "" {
		// Use current or latest version
		playgroundVersion = comp.GetCurrentVersion()
		if playgroundVersion == "" {
			playgroundVersion = comp.GetLatestVersion()
		}
	}

	// Set install dir based on version
	if playgroundVersion != "" {
		comp.EnsureInstallDirWithVersion(playgroundVersion)
		fmt.Printf("Version: %s\n", playgroundVersion)
	}

	composeFile := comp.GetComposeFile("docker/playground")
	if _, err := os.Stat(composeFile); os.IsNotExist(err) {
		fmt.Println("Docker Compose file not found.")
		return
	}

	composeCmd := utils.GetDockerComposeCmd()

	args := []string{}
	for _, c := range composeCmd {
		args = append(args, c)
	}
	args = append(args, "-f", composeFile, "ps")

	// Add -a flag to show all containers
	if showAll {
		args = append(args, "-a")
	}

	dockerCmd := exec.Command(args[0], args[1:]...)
	dockerCmd.Dir = comp.InstallDir
	dockerCmd.Stdout = os.Stdout
	dockerCmd.Stderr = os.Stderr
	dockerCmd.Run()
}

func upgradePlayground() {
	comp := component.GetComponent("playground")

	// Resolve source from global config if auto
	playgroundSource = resolveSourceFlag(playgroundSource)

	// Parse source
	var source component.Source
	switch playgroundSource {
	case "github":
		source = component.SourceGitHub
	case "atomgit":
		source = component.SourceAtomGit
	default:
		source = component.SourceAuto
	}

	fmt.Printf("  Using source: %s\n", playgroundSource)

	if !comp.IsInstalled() {
		fmt.Println("Playground is not installed, performing fresh install...")
		if err := comp.InstallWithVersionAndSource("", source); err != nil {
			fmt.Printf("Installation failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Update code
	fmt.Println("Updating Playground component...")
	if err := comp.UpdateWithSource(source); err != nil {
		fmt.Printf("Update failed: %v\n", err)
		os.Exit(1)
	}

	// Rebuild and start
	composeFile := comp.GetComposeFile("docker/playground")
	if _, err := os.Stat(composeFile); os.IsNotExist(err) {
		fmt.Println("Docker Compose file not found.")
		return
	}

	composeCmd := utils.GetDockerComposeCmd()

	// Determine registry to use (auto defaults to aliyun for China users)
	var registry string
	switch playgroundPullSource {
	case "aliyun":
		registry = utils.AliyunRegistry
	case "default":
		registry = utils.DefaultRegistry
	default: // auto and anything else
		registry = utils.AliyunRegistry
	}

	// Save original compose file content
	originalContent, _ := utils.ReadDockerComposeFile(composeFile)

	// If using aliyun, rewrite docker-compose to use aliyun registry
	if registry != utils.DefaultRegistry {
		utils.RewriteDockerComposeImage(composeFile, registry)
	}

	// Track if we need to restore original
	restoreCompose := registry != utils.DefaultRegistry

	runDockerCompose := func() error {
		dockerCmd := exec.Command(composeCmd[0], append(composeCmd[1:], "-f", composeFile, "up", "-d", "--build")...)
		dockerCmd.Dir = comp.InstallDir
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr
		return dockerCmd.Run()
	}

	fmt.Println("Rebuilding and starting Playground...")
	err := runDockerCompose()

	// If failed with aliyun, try with default registry (for auto mode)
	if err != nil && playgroundPullSource == "auto" && registry == utils.AliyunRegistry {
		// Restore original and try default
		utils.RestoreDockerComposeImage(composeFile, originalContent)
		restoreCompose = false
		registry = utils.DefaultRegistry
		fmt.Println("  Aliyun registry failed, trying default registry...")
		err = runDockerCompose()
	}

	// Restore compose file if needed
	if restoreCompose {
		utils.RestoreDockerComposeImage(composeFile, originalContent)
	}

	if err != nil {
		fmt.Printf("Start failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Playground upgraded successfully!")
}

func logsPlayground() {
	comp := component.GetComponent("playground")
	if !comp.IsInstalled() {
		fmt.Println("Playground is not installed.")
		return
	}

	composeFile := comp.GetComposeFile("docker/playground")
	if _, err := os.Stat(composeFile); os.IsNotExist(err) {
		fmt.Println("Docker Compose file not found.")
		return
	}

	composeCmd := utils.GetDockerComposeCmd()

	dockerCmd := exec.Command(composeCmd[0], append(composeCmd[1:], "-f", composeFile, "logs", "-f")...)
	dockerCmd.Dir = comp.InstallDir
	dockerCmd.Stdout = os.Stdout
	dockerCmd.Stderr = os.Stderr
	dockerCmd.Run()
}

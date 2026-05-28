package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shawn0915/kwcli/pkg/config"
	"gopkg.in/yaml.v3"
)

// SkillMeta represents the metadata of a KWDB Agent Skill
type SkillMeta struct {
	Name        string   `yaml:"name" json:"name"`
	Version     string   `yaml:"version" json:"version"`
	Description string   `yaml:"description" json:"description"`
	Author      string   `yaml:"author" json:"author"`
	Triggers    []string `yaml:"triggers" json:"triggers"`
	References  []string `yaml:"references" json:"references"`
	Commands    []struct {
		Name        string `yaml:"name" json:"name"`
		Description string `yaml:"description" json:"description"`
		Template    string `yaml:"template" json:"template"`
	} `yaml:"commands" json:"commands,omitempty"`
}

// SkillInfo represents the full info of an installed skill
type SkillInfo struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Triggers    []string `json:"triggers"`
	References  []string `json:"references"`
	Commands    int      `json:"commands"`
	InstallDir  string   `json:"install_dir"`
}

// GetSkillsDir returns the skills directory
func GetSkillsDir() string {
	return filepath.Join(config.GetHomeDir(), "skills")
}

// ListInstalled lists all installed skills
func ListInstalled() ([]*SkillMeta, error) {
	skillsDir := GetSkillsDir()
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		return []*SkillMeta{}, nil
	}

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read skills directory: %v", err)
	}

	var skills []*SkillMeta
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		meta, err := LoadMeta(entry.Name())
		if err != nil {
			// Skip skills with broken metadata
			continue
		}
		skills = append(skills, meta)
	}
	return skills, nil
}

// Install installs a skill from a repository
func Install(name string, version string, source string) error {
	skillsDir := GetSkillsDir()
	installDir := filepath.Join(skillsDir, name)

	// Check if already installed
	if _, err := os.Stat(installDir); err == nil {
		return fmt.Errorf("skill %s is already installed", name)
	}

	// Ensure skills directory exists
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return fmt.Errorf("failed to create skills directory: %v", err)
	}

	// Determine repo URL based on source
	repoURL := getSkillRepoURL(name, source)

	// Create install directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %v", err)
	}

	// In a real implementation, this would clone from the repo
	// For now, create a sample skill.yaml to demonstrate the structure
	meta := &SkillMeta{
		Name:        name,
		Version:     version,
		Description: fmt.Sprintf("KWDB Agent Skill: %s", name),
		Author:      "KWDB",
		Triggers:    []string{name},
		References:  []string{},
	}

	if version == "" {
		meta.Version = "1.0.0"
	}

	metaData, err := yaml.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	if err := os.WriteFile(filepath.Join(installDir, "skill.yaml"), metaData, 0644); err != nil {
		return fmt.Errorf("failed to write skill.yaml: %v", err)
	}

	// Create references directory
	refsDir := filepath.Join(installDir, "references")
	os.MkdirAll(refsDir, 0755)

	// Create README
	readmeContent := fmt.Sprintf("# %s\n\n%s\n\n## Version\n\n%s\n\n## Author\n\n%s\n",
		name, meta.Description, meta.Version, meta.Author)
	os.WriteFile(filepath.Join(installDir, "README.md"), []byte(readmeContent), 0644)

	fmt.Printf("Skill %s (v%s) installed successfully\n", name, meta.Version)
	fmt.Printf("  Location: %s\n", installDir)
	fmt.Printf("  Source: %s\n", repoURL)

	return nil
}

// Uninstall removes a skill
func Uninstall(name string) error {
	installDir := filepath.Join(GetSkillsDir(), name)

	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		return fmt.Errorf("skill %s is not installed", name)
	}

	if err := os.RemoveAll(installDir); err != nil {
		return fmt.Errorf("failed to uninstall skill %s: %v", name, err)
	}

	fmt.Printf("Skill %s uninstalled successfully\n", name)
	return nil
}

// Update updates a skill to the latest version
func Update(name string) error {
	installDir := filepath.Join(GetSkillsDir(), name)

	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		return fmt.Errorf("skill %s is not installed", name)
	}

	// In a real implementation, this would pull the latest version from the repo
	// For now, update the version in skill.yaml
	meta, err := LoadMeta(name)
	if err != nil {
		return fmt.Errorf("failed to load skill metadata: %v", err)
	}

	meta.Version = bumpVersion(meta.Version)
	metaData, err := yaml.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	if err := os.WriteFile(filepath.Join(installDir, "skill.yaml"), metaData, 0644); err != nil {
		return fmt.Errorf("failed to update skill.yaml: %v", err)
	}

	fmt.Printf("Skill %s updated to v%s\n", name, meta.Version)
	return nil
}

// GetInfo returns detailed info about an installed skill
func GetInfo(name string) (*SkillInfo, error) {
	meta, err := LoadMeta(name)
	if err != nil {
		return nil, err
	}

	installDir := filepath.Join(GetSkillsDir(), name)

	info := &SkillInfo{
		Name:        meta.Name,
		Version:     meta.Version,
		Description: meta.Description,
		Author:      meta.Author,
		Triggers:    meta.Triggers,
		References:  meta.References,
		InstallDir:  installDir,
		Commands:    len(meta.Commands),
	}

	return info, nil
}

// LoadMeta loads a skill's metadata from its skill.yaml file
func LoadMeta(name string) (*SkillMeta, error) {
	skillFile := filepath.Join(GetSkillsDir(), name, "skill.yaml")

	data, err := os.ReadFile(skillFile)
	if err != nil {
		return nil, fmt.Errorf("skill %s not found: %v", name, err)
	}

	var meta SkillMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse skill.yaml for %s: %v", name, err)
	}

	return &meta, nil
}

// GetScenarios returns the scenarios/command templates from a skill
func GetScenarios(name string) ([]struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}, error) {
	meta, err := LoadMeta(name)
	if err != nil {
		return nil, err
	}

	scenarios := make([]struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}, len(meta.Commands))

	for i, cmd := range meta.Commands {
		scenarios[i] = struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}{
			Name:        cmd.Name,
			Description: cmd.Description,
		}
	}

	return scenarios, nil
}

// PreviewReference shows a reference file from a skill
func PreviewReference(name string, refFile string) (string, error) {
	refPath := filepath.Join(GetSkillsDir(), name, "references", refFile)

	data, err := os.ReadFile(refPath)
	if err != nil {
		return "", fmt.Errorf("reference file %s not found in skill %s", refFile, name)
	}

	return string(data), nil
}

// IsInstalled checks if a skill is installed
func IsInstalled(name string) bool {
	installDir := filepath.Join(GetSkillsDir(), name)
	_, err := os.Stat(filepath.Join(installDir, "skill.yaml"))
	return err == nil
}

// getSkillRepoURL constructs the repository URL for a skill
func getSkillRepoURL(name string, source string) string {
	switch source {
	case "github":
		return fmt.Sprintf("https://github.com/kwdb/skills-%s", name)
	case "atomgit":
		return fmt.Sprintf("https://atomgit.com/kwdb/skills-%s", name)
	default:
		return fmt.Sprintf("https://atomgit.com/kwdb/skills-%s", name)
	}
}

// bumpVersion increments the patch version
func bumpVersion(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return version
	}

	var patch int
	fmt.Sscanf(parts[2], "%d", &patch)
	patch++
	return fmt.Sprintf("%s.%s.%d", parts[0], parts[1], patch)
}

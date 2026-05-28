package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// AgentType represents the type of agent to bridge to
type AgentType string

const (
	AgentClaude   AgentType = "claude"
	AgentCodex    AgentType = "codex"
	AgentOpenClaw AgentType = "openclaw"
	AgentKimi     AgentType = "kimi"
)

// Link bridges a skill to an agent by copying or linking its files
func Link(name string, agent AgentType) error {
	if !IsInstalled(name) {
		return fmt.Errorf("skill %s is not installed", name)
	}

	agentDir, err := getAgentDir(agent)
	if err != nil {
		return fmt.Errorf("cannot find agent directory for %s: %v", agent, err)
	}

	skillDir := filepath.Join(GetSkillsDir(), name)
	targetDir := filepath.Join(agentDir, "skills", name)

	// Ensure target directory exists
	if err := os.MkdirAll(filepath.Dir(targetDir), 0755); err != nil {
		return fmt.Errorf("failed to create agent skills directory: %v", err)
	}

	// Copy skill.yaml
	metaData, err := os.ReadFile(filepath.Join(skillDir, "skill.yaml"))
	if err != nil {
		return fmt.Errorf("failed to read skill.yaml: %v", err)
	}

	if err := os.WriteFile(filepath.Join(targetDir, "skill.yaml"), metaData, 0644); err != nil {
		return fmt.Errorf("failed to write skill.yaml to agent directory: %v", err)
	}

	// Copy references directory if it exists
	refsDir := filepath.Join(skillDir, "references")
	if refs, err := os.ReadDir(refsDir); err == nil {
		targetRefsDir := filepath.Join(targetDir, "references")
		os.MkdirAll(targetRefsDir, 0755)

		for _, ref := range refs {
			if !ref.IsDir() {
				data, err := os.ReadFile(filepath.Join(refsDir, ref.Name()))
				if err == nil {
					os.WriteFile(filepath.Join(targetRefsDir, ref.Name()), data, 0644)
				}
			}
		}
	}

	fmt.Printf("Skill %s linked to %s\n", name, agent)
	fmt.Printf("  Agent directory: %s\n", agentDir)
	return nil
}

// getAgentDir returns the agent's skills directory based on the OS
func getAgentDir(agent AgentType) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch agent {
	case AgentClaude:
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "Application Support", "Claude", "skills"), nil
		}
		return filepath.Join(home, ".claude", "skills"), nil

	case AgentCodex:
		return filepath.Join(home, ".codex", "skills"), nil

	case AgentOpenClaw:
		return filepath.Join(home, ".clawhub", "skills"), nil

	case AgentKimi:
		return filepath.Join(home, ".kimi", "skills"), nil

	default:
		return "", fmt.Errorf("unsupported agent: %s", agent)
	}
}

// ListLinked lists all skills linked to a specific agent
func ListLinked(agent AgentType) ([]string, error) {
	agentDir, err := getAgentDir(agent)
	if err != nil {
		return nil, err
	}

	skillsDir := filepath.Join(agentDir, "skills")
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, err
	}

	var skills []string
	for _, entry := range entries {
		if entry.IsDir() {
			if _, err := os.Stat(filepath.Join(skillsDir, entry.Name(), "skill.yaml")); err == nil {
				skills = append(skills, entry.Name())
			}
		}
	}

	return skills, nil
}

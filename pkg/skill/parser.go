package skill

// parser.go - Skill YAML metadata parsing
// The SkillMeta struct and LoadMeta function are defined in manager.go
// This file provides additional parsing utilities.

// ValidateMeta validates a skill's metadata
func ValidateMeta(meta *SkillMeta) error {
	if meta.Name == "" {
		return &ValidationError{"name is required"}
	}
	if meta.Version == "" {
		return &ValidationError{"version is required"}
	}
	return nil
}

// ValidationError represents a metadata validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return "skill metadata validation error: " + e.Message
}

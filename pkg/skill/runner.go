package skill

import (
	"bytes"
	"fmt"
	"text/template"
)

// CommandTemplate represents a parsed command template
type CommandTemplate struct {
	Name        string
	Description string
	Template    string
}

// RenderResult holds the result of rendering a command template
type RenderResult struct {
	Name     string
	SQL      string
	Metadata string
}

// RenderTemplate renders a command template with the given data
func RenderTemplate(cmdName string, meta *SkillMeta, data map[string]interface{}) (*RenderResult, error) {
	var commandTemplate string
	var description string

	for _, cmd := range meta.Commands {
		if cmd.Name == cmdName {
			commandTemplate = cmd.Template
			description = cmd.Description
			break
		}
	}

	if commandTemplate == "" {
		return nil, fmt.Errorf("command %s not found in skill %s", cmdName, meta.Name)
	}

	// Parse and execute the Go template
	tmpl, err := template.New(cmdName).Parse(commandTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %v", cmdName, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %v", cmdName, err)
	}

	return &RenderResult{
		Name:     cmdName,
		SQL:      buf.String(),
		Metadata: description,
	}, nil
}

// ListCommandTemplates returns all available command templates for a skill
func ListCommandTemplates(meta *SkillMeta) []CommandTemplate {
	templates := make([]CommandTemplate, len(meta.Commands))
	for i, cmd := range meta.Commands {
		templates[i] = CommandTemplate{
			Name:        cmd.Name,
			Description: cmd.Description,
			Template:    cmd.Template,
		}
	}
	return templates
}

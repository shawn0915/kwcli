package skill

import (
	"os"
	"testing"
)

func TestValidateMeta(t *testing.T) {
	tests := []struct {
		name    string
		meta    *SkillMeta
		wantErr bool
	}{
		{"valid", &SkillMeta{Name: "test", Version: "1.0.0"}, false},
		{"no name", &SkillMeta{Version: "1.0.0"}, true},
		{"no version", &SkillMeta{Name: "test"}, true},
		{"empty", &SkillMeta{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMeta(tt.meta)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMeta() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadMetaNotFound(t *testing.T) {
	_, err := LoadMeta("nonexistent-skill")
	if err == nil {
		t.Error("expected error for nonexistent skill")
	}
}

func TestInstallUninstall(t *testing.T) {
	// Save original skills dir and restore after test
	origSkillsDir := GetSkillsDir()
	defer func() {
		// Clean up
		os.RemoveAll(origSkillsDir)
	}()

	// Create a temp dir for testing
	tmpDir, err := os.MkdirTemp("", "skill-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Override via env
	os.Setenv("KWCLI_HOME", tmpDir)
	defer os.Unsetenv("KWCLI_HOME")

	// Test install
	err = Install("test-skill", "1.0.0", "github")
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	// Verify installed
	if !IsInstalled("test-skill") {
		t.Error("skill should be installed")
	}

	// Test info
	info, err := GetInfo("test-skill")
	if err != nil {
		t.Fatalf("GetInfo() error = %v", err)
	}
	if info.Name != "test-skill" {
		t.Errorf("expected name 'test-skill', got %q", info.Name)
	}
	if info.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", info.Version)
	}

	// Test list
	skills, err := ListInstalled()
	if err != nil {
		t.Fatalf("ListInstalled() error = %v", err)
	}
	if len(skills) != 1 {
		t.Errorf("expected 1 skill, got %d", len(skills))
	}

	// Test duplicate install
	err = Install("test-skill", "1.0.0", "github")
	if err == nil {
		t.Error("expected error for duplicate install")
	}

	// Test uninstall
	err = Uninstall("test-skill")
	if err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
	if IsInstalled("test-skill") {
		t.Error("skill should not be installed after uninstall")
	}
}

func TestSkillDir(t *testing.T) {
	dir := GetSkillsDir()
	if dir == "" {
		t.Error("expected non-empty skills directory")
	}
}

func TestRenderTemplate(t *testing.T) {
	meta := &SkillMeta{
		Name:    "test",
		Version: "1.0.0",
		Commands: []struct {
			Name        string `yaml:"name" json:"name"`
			Description string `yaml:"description" json:"description"`
			Template    string `yaml:"template" json:"template"`
		}{
			{
				Name:        "select-all",
				Description: "Select all from table",
				Template:    "SELECT * FROM {{.table}} WHERE ts > now() - INTERVAL '{{.range}}'",
			},
		},
	}

	data := map[string]interface{}{
		"table": "sensor_data",
		"range": "1 hour",
	}

	result, err := RenderTemplate("select-all", meta, data)
	if err != nil {
		t.Fatalf("RenderTemplate() error = %v", err)
	}
	if result.Name != "select-all" {
		t.Errorf("expected name 'select-all', got %q", result.Name)
	}
	if result.SQL != "SELECT * FROM sensor_data WHERE ts > now() - INTERVAL '1 hour'" {
		t.Errorf("unexpected SQL: %s", result.SQL)
	}
}

func TestRenderTemplateNotFound(t *testing.T) {
	meta := &SkillMeta{Name: "test", Version: "1.0.0"}
	_, err := RenderTemplate("nonexistent", meta, nil)
	if err == nil {
		t.Error("expected error for nonexistent command")
	}
}

func TestListCommandTemplates(t *testing.T) {
	meta := &SkillMeta{
		Name: "test",
		Commands: []struct {
			Name        string `yaml:"name" json:"name"`
			Description string `yaml:"description" json:"description"`
			Template    string `yaml:"template" json:"template"`
		}{
			{Name: "cmd1", Description: "Command 1", Template: "SELECT 1"},
			{Name: "cmd2", Description: "Command 2", Template: "SELECT 2"},
		},
	}

	templates := ListCommandTemplates(meta)
	if len(templates) != 2 {
		t.Errorf("expected 2 templates, got %d", len(templates))
	}
}

func TestUpdateSkill(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skill-update-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("KWCLI_HOME", tmpDir)
	defer os.Unsetenv("KWCLI_HOME")

	// Install
	err = Install("update-test", "1.0.0", "github")
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	// Update
	err = Update("update-test")
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Check version bumped
	info, err := GetInfo("update-test")
	if err != nil {
		t.Fatalf("GetInfo() error = %v", err)
	}
	if info.Version != "1.0.1" {
		t.Errorf("expected version '1.0.1', got %q", info.Version)
	}
}

func TestListInstalledEmpty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skill-list-empty")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("KWCLI_HOME", tmpDir)
	defer os.Unsetenv("KWCLI_HOME")

	skills, err := ListInstalled()
	if err != nil {
		t.Fatalf("ListInstalled() error = %v", err)
	}
	if len(skills) != 0 {
		t.Errorf("expected 0 skills, got %d", len(skills))
	}
}

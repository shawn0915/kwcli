package ai

import (
	"testing"
)

func TestDefaultLLMConfig(t *testing.T) {
	cfg := DefaultLLMConfig()
	if cfg == nil {
		t.Fatal("DefaultLLMConfig() returned nil")
	}
	if cfg.Provider != "openai" {
		t.Errorf("expected provider 'openai', got %q", cfg.Provider)
	}
	if cfg.Model != "gpt-4.1" {
		t.Errorf("expected model 'gpt-4.1', got %q", cfg.Model)
	}
	if cfg.Timeout != 60 {
		t.Errorf("expected timeout 60, got %d", cfg.Timeout)
	}
}

func TestAvailableProviders(t *testing.T) {
	providers := AvailableProviders()
	if len(providers) == 0 {
		t.Error("expected at least 1 provider")
	}

	providerMap := make(map[string]bool)
	for _, p := range providers {
		providerMap[p] = true
	}

	if !providerMap["openai"] {
		t.Error("expected 'openai' in providers")
	}
	if !providerMap["deepseek"] {
		t.Error("expected 'deepseek' in providers")
	}
}

func TestNewClient(t *testing.T) {
	cfg := DefaultLLMConfig()
	client := NewClient(cfg)
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
}

func TestBuildSystemPrompt(t *testing.T) {
	builder := NewSystemPromptBuilder()
	if builder == nil {
		t.Fatal("NewSystemPromptBuilder() returned nil")
	}

	prompt, err := builder.BuildSystemPrompt()
	if err != nil {
		t.Fatalf("BuildSystemPrompt() error = %v", err)
	}

	if prompt == "" {
		t.Error("expected non-empty system prompt")
	}

	// Should contain KWDB reference
	if !contains(prompt, "KWDB") {
		t.Error("expected system prompt to mention KWDB")
	}

	// Should contain execution rules
	if !contains(prompt, "只读") && !contains(prompt, "SELECT") {
		t.Error("expected system prompt to mention read-only operations")
	}
}

func TestBuildQueryPrompt(t *testing.T) {
	builder := NewSystemPromptBuilder()

	t.Run("with schema", func(t *testing.T) {
		prompt := builder.BuildQueryPrompt("test query", "schema context")
		if !contains(prompt, "test query") {
			t.Error("expected query prompt to contain user query")
		}
		if !contains(prompt, "schema context") {
			t.Error("expected query prompt to contain schema context")
		}
	})

	t.Run("without schema", func(t *testing.T) {
		prompt := builder.BuildQueryPrompt("test query", "")
		if !contains(prompt, "test query") {
			t.Error("expected query prompt to contain user query")
		}
	})
}

func TestNewExecutor(t *testing.T) {
	executor := NewExecutor()
	if executor == nil {
		t.Fatal("NewExecutor() returned nil")
	}
	if executor.DryRun {
		t.Error("expected DryRun to be false by default")
	}
}

func TestExtractSQL(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     string
	}{
		{
			name:     "sql code block",
			response: "Here's the query:\n```sql\nSELECT * FROM meter_data LIMIT 10;\n```\nDone",
			want:     "SELECT * FROM meter_data LIMIT 10;",
		},
		{
			name:     "no sql",
			response: "I can help you with that.",
			want:     "",
		},
		{
			name:     "inline sql keyword",
			response: "Try running:\nSELECT count(*) FROM meter_data;\nIt will count rows.",
			want:     "SELECT count(*) FROM meter_data;",
		},
		{
			name:     "multiple sql statements",
			response: "First:\nSELECT 1;\nThen:\nSELECT 2;",
			want:     "SELECT 1;\nSELECT 2;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractSQL(tt.response)
			if got != tt.want {
				t.Errorf("extractSQL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClassifyAndGuard(t *testing.T) {
	executor := NewExecutor()

	tests := []struct {
		name     string
		sql      string
		category string
		allowed  bool
	}{
		{"select", "SELECT * FROM meter_data", "read", true},
		{"show", "SHOW TABLES", "read", true},
		{"explain", "EXPLAIN SELECT * FROM meter_data", "read", true},
		{"insert", "INSERT INTO meter_data VALUES (1, 2)", "write", false},
		{"create", "CREATE TABLE test (id INT)", "write", false},
		{"alter", "ALTER TABLE test ADD COLUMN x INT", "write", false},
		{"drop", "DROP TABLE test", "dangerous", false},
		{"drop_database", "DROP DATABASE test", "dangerous", false},
		{"update", "UPDATE meter_data SET voltage=220 WHERE id=1", "write", false},
		{"delete", "DELETE FROM meter_data WHERE id=1", "write", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.classifyAndGuard(tt.sql)
			if result.Category != tt.category {
				t.Errorf("classifyAndGuard().category = %q, want %q", result.Category, tt.category)
			}
		})
	}
}

func TestParseAndGuard(t *testing.T) {
	executor := NewExecutor()

	t.Run("with sql in code block", func(t *testing.T) {
		response := "Here's the SQL:\n```sql\nSELECT * FROM data;\n```"
		result := executor.ParseAndGuard(response)
		if result == nil {
			t.Fatal("ParseAndGuard() returned nil")
		}
		if result.Category != "read" {
			t.Errorf("expected 'read' category, got %q", result.Category)
		}
	})

	t.Run("no sql found", func(t *testing.T) {
		response := "I don't understand the question"
		result := executor.ParseAndGuard(response)
		if result == nil {
			t.Fatal("ParseAndGuard() returned nil")
		}
		if result.Category != "unknown" {
			t.Errorf("expected 'unknown' category, got %q", result.Category)
		}
	})
}

func TestDangerousOperationWarning(t *testing.T) {
	executor := NewExecutor()

	result := executor.classifyAndGuard("DROP DATABASE test_db")
	if result.Category != "dangerous" {
		t.Errorf("expected 'dangerous' category, got %q", result.Category)
	}

	// Warning should contain the danger indicator
	if !contains(result.Warning, "DANGER") {
		t.Error("expected dangerous operation warning to contain 'DANGER'")
	}
}

// contains reports whether substr is within s
func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

// containsStr is a simple substring check
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

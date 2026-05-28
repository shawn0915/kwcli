package mcp

import (
	"testing"
)

func TestNewServer(t *testing.T) {
	server := NewServer("stdio", 0)
	if server == nil {
		t.Fatal("NewServer() returned nil")
	}
	if server.Transport != "stdio" {
		t.Errorf("expected transport 'stdio', got %q", server.Transport)
	}
}

func TestRegisterAndGetTools(t *testing.T) {
	server := NewServer("stdio", 0)
	tool := &Tool{
		Name:        "test_tool",
		Description: "Test tool",
		Command:     "kwcli sql -e \"{{sql}}\" --json",
	}

	server.RegisterTool(tool)

	// Test GetTools
	tools := server.GetTools()
	if len(tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(tools))
	}

	// Test GetTool
	got := server.GetTool("test_tool")
	if got == nil {
		t.Fatal("GetTool() returned nil")
	}
	if got.Name != "test_tool" {
		t.Errorf("expected name 'test_tool', got %q", got.Name)
	}
}

func TestGetToolNotFound(t *testing.T) {
	server := NewServer("stdio", 0)
	got := server.GetTool("nonexistent")
	if got != nil {
		t.Error("expected nil for nonexistent tool")
	}
}

func TestRegisterDefaultTools(t *testing.T) {
	server := NewServer("stdio", 0)
	RegisterDefaultTools(server)

	tools := server.GetTools()
	expectedTools := []string{
		"kwcli_sql_exec",
		"kwcli_schema_dump",
		"kwcli_status",
		"kwcli_logs",
		"kwcli_tsbs_run",
		"kwcli_sampledb_init",
		"kwcli_skill_list",
		"kwcli_skill_info",
		"kwcli_perf_snapshot",
		"kwcli_inspect_run",
	}

	if len(tools) != len(expectedTools) {
		t.Errorf("expected %d tools, got %d", len(expectedTools), len(tools))
	}

	for _, name := range expectedTools {
		if server.GetTool(name) == nil {
			t.Errorf("tool %q not registered", name)
		}
	}
}

func TestToolDescriptionNotEmpty(t *testing.T) {
	server := NewServer("stdio", 0)
	RegisterDefaultTools(server)

	for _, tool := range server.GetTools() {
		if tool.Description == "" {
			t.Errorf("tool %q has empty description", tool.Name)
		}
		if tool.Command == "" {
			t.Errorf("tool %q has empty command", tool.Name)
		}
	}
}

func TestHandleListTools(t *testing.T) {
	server := NewServer("stdio", 0)
	RegisterDefaultTools(server)

	request := map[string]interface{}{
		"method": "list_tools",
	}

	response := server.handleRequest(request)

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'result' in response")
	}

	tools, ok := result["tools"].([]*Tool)
	if !ok {
		t.Fatal("expected 'tools' in result")
	}

	if len(tools) == 0 {
		t.Error("expected at least 1 tool")
	}
}

func TestHandleUnknownMethod(t *testing.T) {
	server := NewServer("stdio", 0)

	request := map[string]interface{}{
		"method": "unknown_method",
	}

	response := server.handleRequest(request)

	_, hasError := response["error"]
	if !hasError {
		t.Error("expected error for unknown method")
	}
}

func TestHandleExecuteUnknownTool(t *testing.T) {
	server := NewServer("stdio", 0)
	RegisterDefaultTools(server)

	request := map[string]interface{}{
		"method": "call_tool",
		"params": map[string]interface{}{
			"name": "nonexistent_tool",
		},
	}

	response := server.handleRequest(request)

	_, hasError := response["error"]
	if !hasError {
		t.Error("expected error for unknown tool")
	}
}

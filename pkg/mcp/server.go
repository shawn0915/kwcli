package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// MCP Server implementation for kwcli
// Implements the Model Context Protocol (MCP) over stdio or SSE transport

// Tool represents an MCP tool definition
type Tool struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Command     string            `json:"command"`
	Parameters  map[string]string `json:"parameters,omitempty"`
}

// Server represents an MCP server instance
type Server struct {
	Transport string
	Port      int
	Tools     map[string]*Tool
}

// NewServer creates a new MCP server
func NewServer(transport string, port int) *Server {
	return &Server{
		Transport: transport,
		Port:      port,
		Tools:     make(map[string]*Tool),
	}
}

// RegisterTool registers a tool with the server
func (s *Server) RegisterTool(tool *Tool) {
	s.Tools[tool.Name] = tool
}

// GetTools returns all registered tools
func (s *Server) GetTools() []*Tool {
	tools := make([]*Tool, 0, len(s.Tools))
	for _, t := range s.Tools {
		tools = append(tools, t)
	}
	return tools
}

// GetTool returns a specific tool by name
func (s *Server) GetTool(name string) *Tool {
	return s.Tools[name]
}

// Start starts the MCP server based on the configured transport
func (s *Server) Start() error {
	switch s.Transport {
	case "sse":
		return s.startSSE()
	case "stdio":
		fallthrough
	default:
		return s.startStdio()
	}
}

// startStdio starts the MCP server in stdio mode (JSON-RPC over stdin/stdout)
func (s *Server) startStdio() error {
	log.SetOutput(os.Stderr)
	log.Printf("MCP server starting in stdio mode")

	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		var request map[string]interface{}
		if err := decoder.Decode(&request); err != nil {
			if err == io.EOF {
				return nil
			}
			log.Printf("Error decoding request: %v", err)
			continue
		}

		response := s.handleRequest(request)
		if err := encoder.Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}

// startSSE starts the MCP server in SSE mode
func (s *Server) startSSE() error {
	addr := fmt.Sprintf(":%d", s.Port)
	log.Printf("MCP server starting in SSE mode on %s", addr)

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", s.handleSSE)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	return http.ListenAndServe(addr, mux)
}

// handleSSE handles SSE connections
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Send initial tools list
	tools := s.GetTools()
	data, _ := json.Marshal(map[string]interface{}{
		"type":  "tools",
		"tools": tools,
	})
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()

	// Handle incoming messages
	if r.Method == http.MethodPost {
		var request map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			return
		}

		response := s.handleRequest(request)
		respData, _ := json.Marshal(response)
		fmt.Fprintf(w, "data: %s\n\n", respData)
		flusher.Flush()
	}

	// Keep connection alive
	<-r.Context().Done()
}

// handleRequest processes a JSON-RPC request
func (s *Server) handleRequest(request map[string]interface{}) map[string]interface{} {
	method, _ := request["method"].(string)
	params, _ := request["params"].(map[string]interface{})

	switch method {
	case "list_tools", "tools":
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"result": map[string]interface{}{
				"tools": s.GetTools(),
			},
		}

	case "call_tool", "execute":
		toolName, _ := params["name"].(string)
		args, _ := params["arguments"].(map[string]interface{})

		tool := s.GetTool(toolName)
		if tool == nil {
			return map[string]interface{}{
				"jsonrpc": "2.0",
				"error": map[string]interface{}{
					"code":    -32602,
					"message": fmt.Sprintf("Tool not found: %s", toolName),
				},
			}
		}

		// Build and execute the command
		result := s.executeTool(tool, args)
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"result":  result,
		}

	default:
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32601,
				"message": fmt.Sprintf("Method not found: %s", method),
			},
		}
	}
}

// executeTool executes a tool's command
func (s *Server) executeTool(tool *Tool, args map[string]interface{}) map[string]interface{} {
	// Build the command string with substitutions
	cmd := tool.Command
	for k, v := range args {
		placeholder := fmt.Sprintf("{{%s}}", k)
		valStr := fmt.Sprintf("%v", v)
		cmd = strings.ReplaceAll(cmd, placeholder, valStr)
	}

	return map[string]interface{}{
		"tool":    tool.Name,
		"command": cmd,
		"status":  "simulated",
		"message": fmt.Sprintf("Tool %s executed. Command: %s", tool.Name, cmd),
	}
}

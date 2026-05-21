package output

import (
	"encoding/json"
	"fmt"
	"os"
)

// Response represents a unified JSON output structure for all commands
type Response struct {
	Success bool        `json:"success"`
	Command string      `json:"command"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// JSONOutput indicates whether JSON output mode is enabled
var JSONOutput bool

// Print outputs data in human-readable or JSON format based on the --json flag
func Print(command string, text string, data interface{}) {
	if JSONOutput {
		PrintJSON(command, data, nil)
	} else {
		fmt.Print(text)
	}
}

// PrintJSON outputs a structured JSON response
func PrintJSON(command string, data interface{}, err error) {
	resp := Response{
		Success: err == nil,
		Command: command,
		Data:    data,
	}
	if err != nil {
		resp.Error = err.Error()
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(resp)
}

// PrintError outputs an error in JSON or plain text format
func PrintError(command string, err error) {
	if JSONOutput {
		PrintJSON(command, nil, err)
	} else {
		fmt.Printf("Error: %v\n", err)
	}
}

// PrintSuccess outputs a success message with data in JSON or plain text format
func PrintSuccess(command string, message string, data interface{}) {
	if JSONOutput {
		PrintJSON(command, data, nil)
	} else {
		fmt.Println(message)
	}
}

// FormatBool returns a human-readable string for a boolean
func FormatBool(val bool, trueStr, falseStr string) string {
	if val {
		return trueStr
	}
	return falseStr
}

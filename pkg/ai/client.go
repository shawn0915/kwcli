package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LLMConfig holds the configuration for an LLM provider
type LLMConfig struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	APIKey   string `json:"api_key"`
	BaseURL  string `json:"base_url"`
	Timeout  int    `json:"timeout"` // seconds
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents a request to the chat completion API
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// ChatResponse represents a response from the chat completion API
type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// DefaultLLMConfig returns the default LLM configuration
func DefaultLLMConfig() *LLMConfig {
	return &LLMConfig{
		Provider: "openai",
		Model:    "gpt-4.1",
		BaseURL:  "https://api.openai.com/v1",
		Timeout:  60,
	}
}

// Client represents an LLM API client
type Client struct {
	config *LLMConfig
	http   *http.Client
}

// NewClient creates a new LLM API client
func NewClient(config *LLMConfig) *Client {
	return &Client{
		config: config,
		http: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}
}

// Chat sends a chat completion request and returns the response
func (c *Client) Chat(messages []Message) (string, error) {
	// Determine the API endpoint
	url := c.getAPIURL()

	request := ChatRequest{
		Model:       c.config.Model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   4096,
		Stream:      false,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from API")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// ChatStream sends a streaming chat completion request
// Returns a channel that yields content chunks
func (c *Client) ChatStream(messages []Message) (<-chan string, error) {
	url := c.getAPIURL()

	request := ChatRequest{
		Model:       c.config.Model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   4096,
		Stream:      true,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %v", err)
	}

	ch := make(chan string)

	go func() {
		defer resp.Body.Close()
		defer close(ch)

		decoder := json.NewDecoder(resp.Body)
		for {
			var streamResp struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}

			if err := decoder.Decode(&streamResp); err != nil {
				break
			}

			if len(streamResp.Choices) > 0 {
				ch <- streamResp.Choices[0].Delta.Content
			}
		}
	}()

	return ch, nil
}

// getAPIURL returns the appropriate API endpoint URL based on the provider config
func (c *Client) getAPIURL() string {
	baseURL := c.config.BaseURL
	if baseURL == "" {
		switch c.config.Provider {
		case "deepseek":
			baseURL = "https://api.deepseek.com/v1"
		case "openai":
			baseURL = "https://api.openai.com/v1"
		default:
			baseURL = "https://api.openai.com/v1"
		}
	}

	return baseURL + "/chat/completions"
}

// AvailableProviders returns the list of supported LLM providers
func AvailableProviders() []string {
	return []string{"openai", "deepseek"}
}

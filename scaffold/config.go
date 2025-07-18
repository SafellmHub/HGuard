package scaffold

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// UserRole represents different user roles in the system
type UserRole string

const (
	RoleUser      UserRole = "user"
	RoleAdmin     UserRole = "admin"
	RoleManager   UserRole = "manager"
	RoleDeveloper UserRole = "developer"
	RoleGuest     UserRole = "guest"
)

// UserContext represents the current user's context
type UserContext struct {
	ID          string            `json:"id"`
	Role        UserRole          `json:"role"`
	Permissions []string          `json:"permissions"`
	IPAddress   string            `json:"ip_address"`
	Metadata    map[string]string `json:"metadata"`
}

// SessionContext represents the current session context
type SessionContext struct {
	ID            string    `json:"id"`
	StartTime     time.Time `json:"start_time"`
	PreviousCalls []string  `json:"previous_calls"`
}

// Config holds configuration for the AI agent
type Config struct {
	AnthropicAPIKey   string
	OpenWeatherAPIKey string
	MavapayAPIKey     string
	SystemPrompt      string
}

// LoadConfig initializes configuration with the provided API keys
func LoadConfig() Config {
	return Config{
		AnthropicAPIKey:   "",
		OpenWeatherAPIKey: getenvOrDefault("OPENWEATHER_API_KEY", ""),
		MavapayAPIKey:     getenvOrDefault("MAVAPAY_API_KEY", ""),
		SystemPrompt: `You are a helpful AI assistant with access to various tools. You can help with weather information, calculations, web searches, financial operations, and system management tasks. Always respond in a helpful and professional manner.

When using tools, format your response as JSON with the exact tool name and parameters as defined in the schemas. If you need to call a tool, respond ONLY with the JSON object. If you're having a conversation, respond normally.

Available tools include:
- weather: Get weather information for a location
- addition: Perform mathematical addition
- search: Search the web for information
- quote: Get financial quotes
- price: Get currency pricing
- file_operations: Manage files (admin only)
- system_info: Get system information
- user_management: Manage users (admin only)
- send_email: Send emails (requires permission)
- database_query: Query databases (requires permission)`,
	}
}

// ConversationMessage represents a message in the conversation
type ConversationMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CallAnthropicConversation makes a conversation call to Anthropic's API
func CallAnthropicConversation(ctx context.Context, apiKey string, messages []ConversationMessage, systemPrompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model":      "claude-sonnet-4-20250514",
		"max_tokens": 1024,
		"messages":   messages,
	}

	// Add system prompt as top-level parameter if provided
	if systemPrompt != "" {
		reqBody["system"] = systemPrompt
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", strings.NewReader(string(bodyBytes)))
	if err != nil {
		return "", err
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("Anthropic API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if contentArr, ok := result["content"].([]interface{}); ok && len(contentArr) > 0 {
		if contentBlock, ok := contentArr[0].(map[string]interface{}); ok {
			if text, ok := contentBlock["text"].(string); ok {
				return text, nil
			}
		}
	}

	return "", fmt.Errorf("unexpected response format: %v", result)
}

// CallAnthropic makes a simple call to Anthropic's API (kept for backward compatibility)
func CallAnthropic(ctx context.Context, apiKey, prompt string) (string, error) {
	messages := []ConversationMessage{
		{Role: "user", Content: prompt},
	}
	return CallAnthropicConversation(ctx, apiKey, messages, "")
}

// GetUserPermissions returns permissions for a given user role
func GetUserPermissions(role UserRole) []string {
	switch role {
	case RoleAdmin:
		return []string{"financial_access", "system_access", "user_management", "file_operations", "email_send", "database_query"}
	case RoleManager:
		return []string{"financial_access", "email_send", "database_query"}
	case RoleDeveloper:
		return []string{"system_access", "file_operations", "database_query"}
	case RoleUser:
		return []string{"email_send"}
	default:
		return []string{}
	}
}

// CreateUserContext creates a user context with the specified role
func CreateUserContext(userID string, role UserRole, ipAddress string) UserContext {
	return UserContext{
		ID:          userID,
		Role:        role,
		Permissions: GetUserPermissions(role),
		IPAddress:   ipAddress,
		Metadata:    make(map[string]string),
	}
}

// CreateSessionContext creates a new session context
func CreateSessionContext(sessionID string) SessionContext {
	return SessionContext{
		ID:            sessionID,
		StartTime:     time.Now(),
		PreviousCalls: make([]string, 0),
	}
}

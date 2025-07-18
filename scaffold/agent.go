// Package scaffold provides a comprehensive AI agent implementation using HallucinationGuard (hguard)
// for safe, policy-driven tool call validation and execution with role-based access control.
//
// This agent supports:
// - Conversation flows with context management
// - Role-based access control (admin, manager, developer, user, guest)
// - Session management and tracking
// - Advanced tool validation and execution
// - Integration with Anthropic's Claude API
//
// Example usage:
//
//	config := LoadConfig()
//	agent := NewStandardAgent("schemas.yaml", "policies.yaml", config)
//
//	userCtx := CreateUserContext("user123", RoleAdmin, "192.168.1.100")
//	sessionCtx := CreateSessionContext("session456")
//
//	response, err := agent.ProcessMessage(ctx, "What's the weather in London?", userCtx, sessionCtx)
package scaffold

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/SafellmHub/hguard-go/pkg/hallucinationguard"
)

// ToolCallResponse represents a tool call as output by the LLM (parsed from JSON).
// Contains the tool name and a map of parameters.
type ToolCallResponse struct {
	Name       string                 `json:"tool"`
	Parameters map[string]interface{} `json:"parameters"`
}

// ConversationHistory stores the conversation messages for context
type ConversationHistory struct {
	Messages []ConversationMessage `json:"messages"`
}

// StandardAgent is the core agent struct that uses HallucinationGuard with comprehensive
// role-based access control, session management, and conversation capabilities.
type StandardAgent struct {
	guard         *hallucinationguard.Guard
	tools         map[string]ToolFunc
	config        Config
	conversations map[string]*ConversationHistory // sessionID -> conversation history
}

// NewStandardAgent initializes a new comprehensive agent with role-based access control
func NewStandardAgent(schemaPath, policyPath string, config Config) *StandardAgent {
	ctx := context.Background()
	guard := hallucinationguard.New()

	if err := guard.LoadSchemasFromFile(ctx, schemaPath); err != nil {
		log.Fatalf("Schema load error: %v", err)
	}

	if err := guard.LoadPoliciesFromFile(ctx, policyPath); err != nil {
		log.Fatalf("Policy load error: %v", err)
	}

	return &StandardAgent{
		guard:         guard,
		tools:         ToolRegistry,
		config:        config,
		conversations: make(map[string]*ConversationHistory),
	}
}

// ProcessMessage processes a user message, determines if it requires tool calls,
// and executes them with proper validation and role-based access control
func (a *StandardAgent) ProcessMessage(ctx context.Context, message string, userCtx UserContext, sessionCtx SessionContext) (string, error) {
	// Update session context with time information
	sessionCtx.PreviousCalls = append(sessionCtx.PreviousCalls, message)

	// Get or create conversation history
	conversation, exists := a.conversations[sessionCtx.ID]
	if !exists {
		conversation = &ConversationHistory{Messages: make([]ConversationMessage, 0)}
		a.conversations[sessionCtx.ID] = conversation
	}

	// Add user message to conversation
	conversation.Messages = append(conversation.Messages, ConversationMessage{
		Role:    "user",
		Content: message,
	})

	// Create system message with role context
	systemMessage := fmt.Sprintf("You are an AI assistant. The user has role: %s with permissions: %s. Current time: %s. %s",
		userCtx.Role,
		strings.Join(userCtx.Permissions, ", "),
		time.Now().Format("2006-01-02 15:04:05"),
		a.config.SystemPrompt)

	// Call Anthropic API to get response with system prompt as separate parameter
	response, err := CallAnthropicConversation(ctx, a.config.AnthropicAPIKey, conversation.Messages, systemMessage)
	if err != nil {
		return "", fmt.Errorf("error calling Anthropic API: %v", err)
	}

	// Check if response is a tool call (JSON format)
	if strings.TrimSpace(response) != "{}" && strings.HasPrefix(strings.TrimSpace(response), "{") {
		var toolCall ToolCallResponse
		if err := json.Unmarshal([]byte(response), &toolCall); err == nil && toolCall.Name != "" {
			// Execute tool call
			toolResponse, err := a.ExecuteToolCall(ctx, toolCall, userCtx, sessionCtx)
			if err != nil {
				return fmt.Sprintf("Error executing tool: %v", err), nil
			}

			// Add tool response to conversation
			conversation.Messages = append(conversation.Messages, ConversationMessage{
				Role:    "assistant",
				Content: toolResponse,
			})

			return toolResponse, nil
		}
	}

	// Add assistant response to conversation
	conversation.Messages = append(conversation.Messages, ConversationMessage{
		Role:    "assistant",
		Content: response,
	})

	return response, nil
}

// ExecuteToolCall executes a tool call with proper validation and role-based access control
func (a *StandardAgent) ExecuteToolCall(ctx context.Context, toolCall ToolCallResponse, userCtx UserContext, sessionCtx SessionContext) (string, error) {
	// Create context for hguard validation
	metadata := make(map[string]interface{})
	for k, v := range userCtx.Metadata {
		metadata[k] = v
	}

	// Create tool call with context
	hguardToolCall := hallucinationguard.ToolCall{
		Name:       toolCall.Name,
		Parameters: toolCall.Parameters,
		Context: &hallucinationguard.CallContext{
			UserID:          userCtx.ID,
			UserRole:        string(userCtx.Role),
			SessionID:       sessionCtx.ID,
			PreviousCalls:   sessionCtx.PreviousCalls,
			UserPermissions: userCtx.Permissions,
			IPAddress:       userCtx.IPAddress,
			TimeOfDay:       time.Now().Hour(),
			Metadata:        metadata,
		},
	}

	// Validate tool call using hguard
	result := a.guard.ValidateToolCall(ctx, hguardToolCall)

	// Handle validation result
	switch result.PolicyAction {
	case hallucinationguard.PolicyActionALLOW:
		// Execute the tool
		toolFunc, exists := a.tools[toolCall.Name]
		if !exists {
			return fmt.Sprintf("Tool '%s' not found", toolCall.Name), nil
		}

		output, err := toolFunc(ctx, toolCall.Parameters)
		if err != nil {
			return fmt.Sprintf("Error executing tool '%s': %v", toolCall.Name, err), nil
		}

		// Update session context
		sessionCtx.PreviousCalls = append(sessionCtx.PreviousCalls, toolCall.Name)

		return output, nil

	case hallucinationguard.PolicyActionREJECT:
		return fmt.Sprintf("Tool call rejected: %s", result.Error), nil

	case hallucinationguard.PolicyActionREWRITE:
		// Rewrite the tool call and try again
		if result.SuggestedCorrection != nil {
			rewrittenCall := ToolCallResponse{
				Name:       result.SuggestedCorrection.Name,
				Parameters: result.SuggestedCorrection.Parameters,
			}
			return a.ExecuteToolCall(ctx, rewrittenCall, userCtx, sessionCtx)
		}
		return fmt.Sprintf("Tool call needs rewriting but no suggestion provided"), nil

	default:
		if result.ExecutionAllowed {
			// Execute the tool
			toolFunc, exists := a.tools[toolCall.Name]
			if !exists {
				return fmt.Sprintf("Tool '%s' not found", toolCall.Name), nil
			}

			output, err := toolFunc(ctx, toolCall.Parameters)
			if err != nil {
				return fmt.Sprintf("Error executing tool '%s': %v", toolCall.Name, err), nil
			}

			// Update session context
			sessionCtx.PreviousCalls = append(sessionCtx.PreviousCalls, toolCall.Name)

			return output, nil
		} else {
			return fmt.Sprintf("Tool call not allowed: %s", result.Error), nil
		}
	}
}

// ValidateToolCall validates a tool call using hguard (backward compatibility)
func (a *StandardAgent) ValidateToolCall(ctx context.Context, toolCall ToolCallResponse) hallucinationguard.ValidationResult {
	return a.guard.ValidateToolCall(ctx, hallucinationguard.ToolCall{
		Name:       toolCall.Name,
		Parameters: toolCall.Parameters,
	})
}

// ValidateToolCallWithContext validates a tool call with full context
func (a *StandardAgent) ValidateToolCallWithContext(ctx context.Context, toolCall ToolCallResponse, userCtx UserContext, sessionCtx SessionContext) hallucinationguard.ValidationResult {
	// Create context for hguard validation
	metadata := make(map[string]interface{})
	for k, v := range userCtx.Metadata {
		metadata[k] = v
	}

	// Create tool call with context
	hguardToolCall := hallucinationguard.ToolCall{
		Name:       toolCall.Name,
		Parameters: toolCall.Parameters,
		Context: &hallucinationguard.CallContext{
			UserID:          userCtx.ID,
			UserRole:        string(userCtx.Role),
			SessionID:       sessionCtx.ID,
			PreviousCalls:   sessionCtx.PreviousCalls,
			UserPermissions: userCtx.Permissions,
			IPAddress:       userCtx.IPAddress,
			TimeOfDay:       time.Now().Hour(),
			Metadata:        metadata,
		},
	}

	return a.guard.ValidateToolCall(ctx, hguardToolCall)
}

// ExecuteTool executes a validated tool call by looking up the tool function by name (backward compatibility)
func (a *StandardAgent) ExecuteTool(ctx context.Context, toolCall ToolCallResponse) (string, error) {
	toolFunc, ok := a.tools[toolCall.Name]
	if !ok {
		return "", fmt.Errorf("tool not found: %s", toolCall.Name)
	}
	return toolFunc(ctx, toolCall.Parameters)
}

// GetConversationHistory returns the conversation history for a session
func (a *StandardAgent) GetConversationHistory(sessionID string) *ConversationHistory {
	return a.conversations[sessionID]
}

// ClearConversationHistory clears the conversation history for a session
func (a *StandardAgent) ClearConversationHistory(sessionID string) {
	delete(a.conversations, sessionID)
}

// GetAvailableTools returns a list of available tools for a user based on their role
func (a *StandardAgent) GetAvailableTools(userCtx UserContext) []string {
	ctx := context.Background()
	var availableTools []string

	// Create context for hguard validation
	metadata := make(map[string]interface{})
	for k, v := range userCtx.Metadata {
		metadata[k] = v
	}

	// Check each tool against the user's context
	for toolName := range a.tools {
		hguardToolCall := hallucinationguard.ToolCall{
			Name:       toolName,
			Parameters: map[string]interface{}{},
			Context: &hallucinationguard.CallContext{
				UserID:          userCtx.ID,
				UserRole:        string(userCtx.Role),
				UserPermissions: userCtx.Permissions,
				IPAddress:       userCtx.IPAddress,
				TimeOfDay:       time.Now().Hour(),
				Metadata:        metadata,
			},
		}

		// Test with minimal parameters
		result := a.guard.ValidateToolCall(ctx, hguardToolCall)

		if result.ExecutionAllowed {
			availableTools = append(availableTools, toolName)
		}
	}

	return availableTools
}

// GetToolDescription returns a description of what each tool does
func (a *StandardAgent) GetToolDescription(toolName string) string {
	descriptions := map[string]string{
		"weather":         "Get current weather information for any city",
		"addition":        "Perform mathematical addition of two numbers",
		"search":          "Search the web for information",
		"quote":           "Get financial quotes for currency exchanges",
		"price":           "Get current currency pricing information",
		"file_operations": "Manage files and directories (admin only)",
		"system_info":     "Get system information and metrics",
		"user_management": "Manage user accounts and roles (admin only)",
		"send_email":      "Send email notifications (requires permission)",
		"database_query":  "Query databases (requires permission)",
		"calendar":        "Manage calendar events and appointments",
		"task_management": "Create and manage tasks",
		"analytics":       "Generate analytics reports and insights",
		"document_gen":    "Generate documents from templates",
		"notification":    "Send notifications to users",
	}

	if desc, exists := descriptions[toolName]; exists {
		return desc
	}
	return "Tool description not available"
}

// GetUserStats returns statistics about the user's tool usage
func (a *StandardAgent) GetUserStats(userCtx UserContext, sessionCtx SessionContext) map[string]interface{} {
	return map[string]interface{}{
		"user_id":         userCtx.ID,
		"role":            userCtx.Role,
		"permissions":     userCtx.Permissions,
		"session_id":      sessionCtx.ID,
		"session_start":   sessionCtx.StartTime,
		"calls_made":      len(sessionCtx.PreviousCalls),
		"available_tools": len(a.GetAvailableTools(userCtx)),
	}
}

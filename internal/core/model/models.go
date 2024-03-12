package model

import "time"

// ToolCall represents a tool call from the LLM output
type ToolCall struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
	Context    CallContext            `json:"context"`
	Timestamp  time.Time              `json:"timestamp"`
}

// CallContext represents the context of a tool call
type CallContext struct {
	UserID          string   `json:"user_id"`
	SessionID       string   `json:"session_id"`
	ConversationID  string   `json:"conversation_id"`
	PreviousCalls   []string `json:"previous_calls"`
	UserPermissions []string `json:"user_permissions"`
}

// ValidationResult represents the result of validating a tool call
type ValidationResult struct {
	ToolCallID          string                 `json:"tool_call_id"`
	Status              string                 `json:"status"` // approved, rejected, rewritten
	Confidence          float64                `json:"confidence"`
	Reason              string                 `json:"reason,omitempty"`
	Modifications       map[string]interface{} `json:"modifications,omitempty"`
	ExecutionAllowed    bool                   `json:"execution_allowed"`
	SuggestedCorrection *ToolCall              `json:"suggested_correction,omitempty"`
}

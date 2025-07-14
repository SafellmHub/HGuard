// Package scaffold provides a reference implementation of an LLM agent using HallucinationGuard (hguard)
// for safe, policy-driven tool call validation and execution.
//
// This agent bridges the LLM's tool call output (as JSON) with the hguard validation layer and the actual tool implementations.
//
// Example usage:
//
//	agent := NewHGuardAgent("schemas.yaml", "policies.yaml")
//	toolCall := ToolCallResponse{Name: "weather", Parameters: map[string]interface{}{ "city": "London" }}
//	result := agent.ValidateToolCall(ctx, toolCall)
//	if result.ExecutionAllowed {
//	    output, err := agent.ExecuteTool(ctx, toolCall)
//	    // handle output
//	}
package scaffold

import (
	"context"
	"log"

	"github.com/SafellmHub/hguard-go/pkg/hallucinationguard"
)

// ToolCallResponse represents a tool call as output by the LLM (parsed from JSON).
// Contains the tool name and a map of parameters.
//
// Example:
//
//	tc := ToolCallResponse{Name: "addition", Parameters: map[string]interface{}{ "a": 5, "b": 7 }}
type ToolCallResponse struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
}

// HGuardAgent is the core agent struct that uses HallucinationGuard to validate and execute tool calls.
// It holds a reference to the hguard Guard and a registry of available tool functions.
//
// Example initialization:
//
//	agent := NewHGuardAgent("schemas.yaml", "policies.yaml")
type HGuardAgent struct {
	guard *hallucinationguard.Guard
	tools map[string]ToolFunc
}

// NewHGuardAgent initializes a new agent, loading schemas and policies from the given YAML files.
// It panics if loading fails.
//
// Example:
//
//	agent := NewHGuardAgent("schemas.yaml", "policies.yaml")
func NewHGuardAgent(schemaPath, policyPath string) *HGuardAgent {
	ctx := context.Background()
	guard := hallucinationguard.New()
	if err := guard.LoadSchemasFromFile(ctx, schemaPath); err != nil {
		log.Fatalf("Schema load error: %v", err)
	}
	if err := guard.LoadPoliciesFromFile(ctx, policyPath); err != nil {
		log.Fatalf("Policy load error: %v", err)
	}
	return &HGuardAgent{guard: guard, tools: ToolRegistry}
}

// ValidateToolCall validates a tool call using hguard.
// Returns a ValidationResult indicating if the call is allowed, rejected, or needs correction.
//
// Example:
//
//	result := agent.ValidateToolCall(ctx, toolCall)
//	if result.ExecutionAllowed { /* ... */ }
func (a *HGuardAgent) ValidateToolCall(ctx context.Context, toolCall ToolCallResponse) hallucinationguard.ValidationResult {
	return a.guard.ValidateToolCall(ctx, hallucinationguard.ToolCall{
		Name:       toolCall.Name,
		Parameters: toolCall.Parameters,
	})
}

// ExecuteTool executes a validated tool call by looking up the tool function by name.
// Returns the tool's output or an error if the tool is not found or execution fails.
//
// Example:
//
//	output, err := agent.ExecuteTool(ctx, toolCall)
func (a *HGuardAgent) ExecuteTool(ctx context.Context, toolCall ToolCallResponse) (string, error) {
	toolFunc, ok := a.tools[toolCall.Name]
	if !ok {
		return "", nil
	}
	return toolFunc(ctx, toolCall.Parameters)
}

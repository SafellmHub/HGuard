package hallucinationguard

import (
	"github.com/fishonamos/hallucination-shield/internal/core/model"
	"github.com/fishonamos/hallucination-shield/internal/core/policy"
	"github.com/fishonamos/hallucination-shield/internal/schema"
)

// Guard is the main SDK struct for embedding HallucinationGuard
// in other Go projects.
type Guard struct{}

// New creates a new Guard instance
func New() *Guard {
	return &Guard{}
}

// LoadSchemasFromFile loads tool schemas from a YAML file
func (g *Guard) LoadSchemasFromFile(path string) error {
	return schema.LoadSchemasFromYAML(path)
}

// LoadPoliciesFromFile loads policies from a YAML file
func (g *Guard) LoadPoliciesFromFile(path string) error {
	return policy.LoadPoliciesFromYAML(path)
}

// ValidateToolCall validates a tool call using loaded schemas and policies
func (g *Guard) ValidateToolCall(tc model.ToolCall) model.ValidationResult {
	return ValidateToolCall(tc)
}

// ValidateToolCall is a wrapper for the internal validation logic
func ValidateToolCall(tc model.ToolCall) model.ValidationResult {
	// Use the same logic as the REST API/server
	return schema.ValidateAndPolicy(tc)
}

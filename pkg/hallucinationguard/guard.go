package hallucinationguard

import (
	"github.com/fishonamos/HGuard/pkg/internal/core/model"
	"github.com/fishonamos/HGuard/pkg/internal/core/policy"
	"github.com/fishonamos/HGuard/pkg/internal/schema"
)

// ToolCall represents a tool call request
type ToolCall struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
}

// ValidationResult represents the result of validating a tool call
type ValidationResult struct {
	ExecutionAllowed    bool      `json:"allowed"`
	Error               string    `json:"error,omitempty"`
	PolicyAction        string    `json:"policy_action,omitempty"`
	SuggestedCorrection *ToolCall `json:"suggested_correction,omitempty"`
}

// PolicyAction constants
const (
	PolicyActionALLOW          = "ALLOW"
	PolicyActionREJECT         = "REJECT"
	PolicyActionREWRITE        = "REWRITE"
	PolicyActionRATE_LIMIT     = "RATE_LIMIT"
	PolicyActionCONTEXT_REJECT = "CONTEXT_REJECT"
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
func (g *Guard) ValidateToolCall(tc ToolCall) ValidationResult {
	// Convert to internal model
	internalCall := model.ToolCall{
		Name:       tc.Name,
		Parameters: tc.Parameters,
	}

	// Validate using internal logic
	result := schema.ValidateAndPolicy(internalCall)

	// Convert back to public type
	validationResult := ValidationResult{
		ExecutionAllowed: result.ExecutionAllowed,
		Error:            result.Reason,
		PolicyAction:     string(result.PolicyAction),
	}

	if result.SuggestedCorrection != nil {
		validationResult.SuggestedCorrection = &ToolCall{
			Name:       result.SuggestedCorrection.Name,
			Parameters: result.SuggestedCorrection.Parameters,
		}
	}

	return validationResult
}

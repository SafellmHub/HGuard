package hallucinationguard

import (
	"context"
	"fmt"
	"sync"

	"github.com/SafellmHub/HGuard/pkg/internal/core/model"
	"github.com/SafellmHub/HGuard/pkg/internal/core/policy"
	"github.com/SafellmHub/HGuard/pkg/internal/schema"
)

// ToolCall represents a tool call request
// ToolCall is the public struct for tool call validation requests.
type ToolCall struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
}

// ValidationResult represents the result of validating a tool call
// ValidationResult contains the outcome, error, policy action, and any suggested correction.
type ValidationResult struct {
	ExecutionAllowed    bool      `json:"allowed"`
	Error               string    `json:"error,omitempty"`
	PolicyAction        string    `json:"policy_action,omitempty"`
	SuggestedCorrection *ToolCall `json:"suggested_correction,omitempty"`
	ToolCallID          string    `json:"tool_call_id,omitempty"`
	Status              string    `json:"status,omitempty"`
	Confidence          float64   `json:"confidence,omitempty"`
}

// PolicyAction constants
const (
	PolicyActionALLOW          = "ALLOW"
	PolicyActionREJECT         = "REJECT"
	PolicyActionREWRITE        = "REWRITE"
	PolicyActionRATE_LIMIT     = "RATE_LIMIT"
	PolicyActionCONTEXT_REJECT = "CONTEXT_REJECT"
)

// SchemaLoader defines the interface for loading schemas
// This allows for custom schema loading implementations.
type SchemaLoader interface {
	LoadSchemas(ctx context.Context, path string) error
}

// PolicyEngine defines the interface for loading and applying policies
// This allows for custom policy engine implementations.
type PolicyEngine interface {
	LoadPolicies(ctx context.Context, path string) error
}

// Guard is the main SDK struct for embedding HallucinationGuard
// in other Go projects. It is safe for concurrent use.
type Guard struct {
	mu           sync.RWMutex
	schemaLoader SchemaLoader
	policyEngine PolicyEngine
}

// GuardOption is a functional option for configuring Guard
// Example: WithSchemaLoader(customLoader)
type GuardOption func(*Guard)

// WithSchemaLoader sets a custom SchemaLoader
func WithSchemaLoader(loader SchemaLoader) GuardOption {
	return func(g *Guard) {
		g.schemaLoader = loader
	}
}

// WithPolicyEngine sets a custom PolicyEngine
func WithPolicyEngine(engine PolicyEngine) GuardOption {
	return func(g *Guard) {
		g.policyEngine = engine
	}
}

// New creates a new Guard instance with optional configuration
func New(opts ...GuardOption) *Guard {
	g := &Guard{
		schemaLoader: defaultSchemaLoader{},
		policyEngine: defaultPolicyEngine{},
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// LoadSchemasFromFile loads tool schemas from a YAML file using the configured loader
func (g *Guard) LoadSchemasFromFile(ctx context.Context, path string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if err := g.schemaLoader.LoadSchemas(ctx, path); err != nil {
		return fmt.Errorf("failed to load schemas from %s: %w", path, err)
	}
	return nil
}

// LoadPoliciesFromFile loads policies from a YAML file using the configured engine
func (g *Guard) LoadPoliciesFromFile(ctx context.Context, path string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if err := g.policyEngine.LoadPolicies(ctx, path); err != nil {
		return fmt.Errorf("failed to load policies from %s: %w", path, err)
	}
	return nil
}

// ValidateToolCall validates a tool call using loaded schemas and policies
func (g *Guard) ValidateToolCall(ctx context.Context, tc ToolCall) ValidationResult {
	g.mu.RLock()
	defer g.mu.RUnlock()
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
		PolicyAction:     result.PolicyAction,
		ToolCallID:       result.ToolCallID,
		Status:           result.Status,
		Confidence:       result.Confidence,
	}
	if result.SuggestedCorrection != nil {
		validationResult.SuggestedCorrection = &ToolCall{
			Name:       result.SuggestedCorrection.Name,
			Parameters: result.SuggestedCorrection.Parameters,
		}
	}
	return validationResult
}

// defaultSchemaLoader is the default implementation using the internal schema package
// Implements SchemaLoader
// ...
type defaultSchemaLoader struct{}

func (d defaultSchemaLoader) LoadSchemas(ctx context.Context, path string) error {
	return schema.LoadSchemasFromYAML(path)
}

type defaultPolicyEngine struct{}

func (d defaultPolicyEngine) LoadPolicies(ctx context.Context, path string) error {
	return policy.LoadPoliciesFromYAML(path)
}

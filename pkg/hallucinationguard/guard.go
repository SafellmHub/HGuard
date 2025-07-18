package hallucinationguard

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SafellmHub/hguard-go/pkg/internal/core/model"
	"github.com/SafellmHub/hguard-go/pkg/internal/core/policy"
	"github.com/SafellmHub/hguard-go/pkg/internal/schema"
)

// Package hallucinationguard provides a guardrail system for validating and enforcing policies on LLM tool calls.
//
// Example usage:
//
//	guard := hallucinationguard.New()
//	err := guard.LoadSchemasFromFile(ctx, "schemas.yaml")
//	err = guard.LoadPoliciesFromFile(ctx, "policies.yaml")
//	result := guard.ValidateToolCall(ctx, hallucinationguard.ToolCall{Name: "weather", Parameters: map[string]interface{}{...}})
//	if result.ExecutionAllowed { /* execute tool */ }
//
// ToolCall represents a tool call request.
// Use this struct to pass tool call data to the Guard for validation.
//
// Example:
//
//	tc := ToolCall{Name: "weather", Parameters: map[string]interface{}{ "city": "London" }}
type ToolCall struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
	Context    *CallContext           `json:"context,omitempty"` // Optional context for conditional policies
}

// CallContext represents the context information for conditional policy evaluation
type CallContext struct {
	UserID          string                 `json:"user_id,omitempty"`
	UserRole        string                 `json:"user_role,omitempty"`
	SessionID       string                 `json:"session_id,omitempty"`
	ConversationID  string                 `json:"conversation_id,omitempty"`
	PreviousCalls   []string               `json:"previous_calls,omitempty"`
	UserPermissions []string               `json:"user_permissions,omitempty"`
	IPAddress       string                 `json:"ip_address,omitempty"`
	TimeOfDay       int                    `json:"time_of_day,omitempty"` // Hour of day (0-23)
	Metadata        map[string]interface{} `json:"metadata,omitempty"`    // Arbitrary context data
}

// ValidationResult represents the result of validating a tool call.
//
// Example:
//
//	result := guard.ValidateToolCall(ctx, tc)
//	if result.ExecutionAllowed { /* ... */ }
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

// SchemaLoader defines the interface for loading schemas.
// Implement this interface to provide custom schema loading logic.
type SchemaLoader interface {
	LoadSchemas(ctx context.Context, path string) error
}

// PolicyEngine defines the interface for loading and applying policies.
// Implement this interface to provide custom policy engine logic.
type PolicyEngine interface {
	LoadPolicies(ctx context.Context, path string) error
}

// Guard is the main SDK struct for embedding HallucinationGuard in other Go projects.
// It is safe for concurrent use.
//
// Example usage:
//
//	guard := hallucinationguard.New()
//	err := guard.LoadSchemasFromFile(ctx, "schemas.yaml")
//	err = guard.LoadPoliciesFromFile(ctx, "policies.yaml")
type Guard struct {
	mu           sync.RWMutex
	schemaLoader SchemaLoader
	policyEngine PolicyEngine
}

// GuardOption is a functional option for configuring Guard.
//
// Example:
//
//	guard := hallucinationguard.New(hallucinationguard.WithSchemaLoader(myLoader))
type GuardOption func(*Guard)

// WithSchemaLoader sets a custom SchemaLoader for the Guard.
//
// Example:
//
//	guard := hallucinationguard.New(WithSchemaLoader(myLoader))
func WithSchemaLoader(loader SchemaLoader) GuardOption {
	return func(g *Guard) {
		g.schemaLoader = loader
	}
}

// WithPolicyEngine sets a custom PolicyEngine for the Guard.
//
// Example:
//
//	guard := hallucinationguard.New(WithPolicyEngine(myEngine))
func WithPolicyEngine(engine PolicyEngine) GuardOption {
	return func(g *Guard) {
		g.policyEngine = engine
	}
}

// New creates a new Guard instance with optional configuration.
//
// Example:
//
//	guard := New()
//	guard := New(WithSchemaLoader(myLoader), WithPolicyEngine(myEngine))
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

// LoadSchemasFromFile loads tool schemas from a YAML file using the configured loader.
//
// Example:
//
//	err := guard.LoadSchemasFromFile(ctx, "schemas.yaml")
func (g *Guard) LoadSchemasFromFile(ctx context.Context, path string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if err := g.schemaLoader.LoadSchemas(ctx, path); err != nil {
		return fmt.Errorf("failed to load schemas from %s: %w", path, err)
	}
	return nil
}

// LoadPoliciesFromFile loads policies from a YAML file using the configured engine.
//
// Example:
//
//	err := guard.LoadPoliciesFromFile(ctx, "policies.yaml")
func (g *Guard) LoadPoliciesFromFile(ctx context.Context, path string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if err := g.policyEngine.LoadPolicies(ctx, path); err != nil {
		return fmt.Errorf("failed to load policies from %s: %w", path, err)
	}
	return nil
}

// ValidateToolCall validates a tool call using loaded schemas and policies.
//
// Example:
//
//	result := guard.ValidateToolCall(ctx, ToolCall{Name: "weather", Parameters: map[string]interface{}{ "city": "London" }})
func (g *Guard) ValidateToolCall(ctx context.Context, tc ToolCall) ValidationResult {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Generate ID if not provided
	callID := fmt.Sprintf("call_%d", time.Now().UnixNano())

	// Convert public context to internal context
	var internalContext model.CallContext
	if tc.Context != nil {
		internalContext = model.CallContext{
			UserID:          tc.Context.UserID,
			UserRole:        tc.Context.UserRole,
			SessionID:       tc.Context.SessionID,
			ConversationID:  tc.Context.ConversationID,
			PreviousCalls:   tc.Context.PreviousCalls,
			UserPermissions: tc.Context.UserPermissions,
			IPAddress:       tc.Context.IPAddress,
			TimeOfDay:       tc.Context.TimeOfDay,
			Metadata:        tc.Context.Metadata,
		}
	}

	// Convert to internal model
	internalCall := model.ToolCall{
		ID:         callID,
		Name:       tc.Name,
		Parameters: tc.Parameters,
		Context:    internalContext,
		Timestamp:  time.Now(),
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
		publicContext := &CallContext{
			UserID:          result.SuggestedCorrection.Context.UserID,
			UserRole:        result.SuggestedCorrection.Context.UserRole,
			SessionID:       result.SuggestedCorrection.Context.SessionID,
			ConversationID:  result.SuggestedCorrection.Context.ConversationID,
			PreviousCalls:   result.SuggestedCorrection.Context.PreviousCalls,
			UserPermissions: result.SuggestedCorrection.Context.UserPermissions,
			IPAddress:       result.SuggestedCorrection.Context.IPAddress,
			TimeOfDay:       result.SuggestedCorrection.Context.TimeOfDay,
			Metadata:        result.SuggestedCorrection.Context.Metadata,
		}
		validationResult.SuggestedCorrection = &ToolCall{
			Name:       result.SuggestedCorrection.Name,
			Parameters: result.SuggestedCorrection.Parameters,
			Context:    publicContext,
		}
	}

	return validationResult
}

// defaultSchemaLoader is the default implementation using the internal schema package.
// Implements SchemaLoader.
type defaultSchemaLoader struct{}

func (d defaultSchemaLoader) LoadSchemas(ctx context.Context, path string) error {
	return schema.LoadSchemasFromYAML(path)
}

// defaultPolicyEngine is the default implementation using the internal policy package.
// Implements PolicyEngine.
type defaultPolicyEngine struct{}

func (d defaultPolicyEngine) LoadPolicies(ctx context.Context, path string) error {
	return policy.LoadPoliciesFromYAML(path)
}

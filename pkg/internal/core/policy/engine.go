package policy

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/SafellmHub/hguard-go/pkg/internal/core/model"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"gopkg.in/yaml.v3"
)

// Package policy provides policy definitions, registration, and enforcement for LLM tool calls.
//
// Example YAML policy:
//
//	policies:
//	  - tool_name: weather
//	    type: ALLOW
//	  - tool_name: transfer_money
//	    type: REJECT
//	    condition: "user.role != 'admin'"
//	    reason: "Only admins can transfer money"
//
// Example usage:
//
//	err := policy.LoadPoliciesFromYAML("policies.yaml")
//	result := policy.EvaluatePolicy(tc)
//	action := result.Action
//
// PolicyType defines the type of policy action (REJECT, REWRITE, LOG, ALLOW, etc.).
type PolicyType string

// PolicyAction is a string representing the result of a policy application.
type PolicyAction string

const (
	PolicyReject        PolicyType = "REJECT"
	PolicyRewrite       PolicyType = "REWRITE"
	PolicyLog           PolicyType = "LOG"
	PolicyAllow         PolicyType = "ALLOW"
	PolicyContextReject PolicyType = "CONTEXT_REJECT"
	PolicyRateLimit     PolicyType = "RATE_LIMIT"

	ActionNone      PolicyAction = "none"
	ActionRejected  PolicyAction = "rejected"
	ActionRewritten PolicyAction = "rewritten"
	ActionLogged    PolicyAction = "logged"
	ActionApproved  PolicyAction = "approved"
)

// Policy defines a guardrail policy for a tool.
type Policy struct {
	ToolName  string     `yaml:"tool_name"`
	Type      PolicyType `yaml:"type"`
	Condition string     `yaml:"condition,omitempty"` // Conditional expression to evaluate
	Reason    string     `yaml:"reason,omitempty"`    // Custom reason for rejection/rewrite
	Priority  int        `yaml:"priority,omitempty"`  // Priority for multiple matching policies (higher = more specific)
	Target    string     `yaml:"target,omitempty"`    // Target tool name for REWRITE policies
}

// PolicyResult represents the result of policy evaluation
type PolicyResult struct {
	Action   PolicyType
	Reason   string
	Target   string
	Matched  bool
	PolicyID string
}

// In-memory policy registry
var policies = map[string][]Policy{}

// RegisterPolicy adds a policy to the registry.
//
// Example:
//
//	policy.RegisterPolicy(Policy{ToolName: "weather", Type: policy.PolicyAllow})
func RegisterPolicy(p Policy) {
	policies[p.ToolName] = append(policies[p.ToolName], p)
	// Sort policies by priority (higher first)
	sort.Slice(policies[p.ToolName], func(i, j int) bool {
		return policies[p.ToolName][i].Priority > policies[p.ToolName][j].Priority
	})
}

// GetPolicy retrieves policies for a tool name.
//
// Example:
//
//	policies, ok := policy.GetPolicy("weather")
func GetPolicy(toolName string) ([]Policy, bool) {
	p, ok := policies[toolName]
	return p, ok
}

// GetAllPolicies returns all policies for a tool name (including wildcards)
func GetAllPolicies(toolName string) []Policy {
	var allPolicies []Policy

	// Add specific tool policies
	if toolPolicies, ok := policies[toolName]; ok {
		allPolicies = append(allPolicies, toolPolicies...)
	}

	// Add wildcard policies
	if wildcardPolicies, ok := policies["*"]; ok {
		allPolicies = append(allPolicies, wildcardPolicies...)
	}

	// Sort by priority (higher first)
	sort.Slice(allPolicies, func(i, j int) bool {
		return allPolicies[i].Priority > allPolicies[j].Priority
	})

	return allPolicies
}

// EvaluatePolicy evaluates all applicable policies for a tool call and returns the result
func EvaluatePolicy(tc model.ToolCall) PolicyResult {
	allPolicies := GetAllPolicies(tc.Name)

	for _, policy := range allPolicies {
		if policy.Condition == "" {
			// No condition, policy always applies
			return PolicyResult{
				Action:   policy.Type,
				Reason:   policy.Reason,
				Target:   policy.Target,
				Matched:  true,
				PolicyID: fmt.Sprintf("%s:%s", policy.ToolName, policy.Type),
			}
		}

		// Evaluate condition
		match, err := evaluateCondition(policy.Condition, tc)
		if err != nil {
			// Log error and continue to next policy
			fmt.Printf("Error evaluating condition for policy %s: %v\n", policy.ToolName, err)
			continue
		}

		if match {
			reason := policy.Reason
			if reason == "" {
				reason = fmt.Sprintf("Policy %s matched for tool %s", policy.Type, tc.Name)
			}
			return PolicyResult{
				Action:   policy.Type,
				Reason:   reason,
				Target:   policy.Target,
				Matched:  true,
				PolicyID: fmt.Sprintf("%s:%s", policy.ToolName, policy.Type),
			}
		}
	}

	// No matching policies, default to allow
	return PolicyResult{
		Action:   PolicyAllow,
		Reason:   "No matching policies found",
		Target:   "",
		Matched:  false,
		PolicyID: "default:allow",
	}
}

// Expression compilation cache for performance
var (
	exprCache = make(map[string]*vm.Program)
	cacheMu   sync.RWMutex
)

// evaluateCondition evaluates a conditional expression using the tool call context
func evaluateCondition(condition string, tc model.ToolCall) (bool, error) {
	// Create evaluation environment
	env := map[string]interface{}{
		"user": map[string]interface{}{
			"id":          tc.Context.UserID,
			"role":        tc.Context.UserRole,
			"permissions": tc.Context.UserPermissions,
		},
		"session": map[string]interface{}{
			"id":              tc.Context.SessionID,
			"conversation_id": tc.Context.ConversationID,
			"previous_calls":  tc.Context.PreviousCalls,
		},
		"params": tc.Parameters,
		"tool": map[string]interface{}{
			"name": tc.Name,
		},
		"time": map[string]interface{}{
			"hour": tc.Context.TimeOfDay,
		},
		"request": map[string]interface{}{
			"ip": tc.Context.IPAddress,
		},
		"metadata": tc.Context.Metadata,
		// Add helper functions directly to environment
		"len": func(arr []string) int {
			return len(arr)
		},
		"contains": func(arr []string, item string) bool {
			for _, a := range arr {
				if a == item {
					return true
				}
			}
			return false
		},
		"now": func() time.Time {
			return time.Now()
		},
	}

	// Check cache for compiled expression
	cacheMu.RLock()
	program, exists := exprCache[condition]
	cacheMu.RUnlock()

	if !exists {
		// Compile and cache the expression
		var err error
		program, err = expr.Compile(condition, expr.Env(env))
		if err != nil {
			return false, fmt.Errorf("failed to compile condition: %w", err)
		}

		// Cache the compiled program
		cacheMu.Lock()
		exprCache[condition] = program
		cacheMu.Unlock()
	}

	result, err := expr.Run(program, env)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate condition: %w", err)
	}

	boolResult, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("condition did not evaluate to boolean: %T", result)
	}

	return boolResult, nil
}

// ClearExpressionCache clears the compiled expression cache (useful for testing)
func ClearExpressionCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	exprCache = make(map[string]*vm.Program)
}

// ApplyPolicy applies the policy to a tool call and returns the policy type (legacy function for backward compatibility).
//
// Example:
//
//	action := policy.ApplyPolicy(tc)
func ApplyPolicy(tc model.ToolCall) PolicyType {
	result := EvaluatePolicy(tc)
	return result.Action
}

// LoadPoliciesFromYAML loads policies from a YAML file and registers them.
//
// Example:
//
//	err := policy.LoadPoliciesFromYAML("policies.yaml")
func LoadPoliciesFromYAML(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var data struct {
		Policies []Policy `yaml:"policies"`
	}
	if err := yaml.NewDecoder(f).Decode(&data); err != nil {
		return err
	}

	// Clear existing policies
	policies = make(map[string][]Policy)

	for _, p := range data.Policies {
		RegisterPolicy(p)
	}
	return nil
}

package policy

import (
	"github.com/fishonamos/hallucination-shield/internal/core/model"
)

// PolicyType defines the type of policy action
// (REJECT, REWRITE, LOG, ALLOW, etc.)
type PolicyType string

const (
	PolicyReject  PolicyType = "REJECT"
	PolicyRewrite PolicyType = "REWRITE"
	PolicyLog     PolicyType = "LOG"
	PolicyAllow   PolicyType = "ALLOW"
)

// Policy defines a guardrail policy for a tool
// For MVP, only basic fields are included
// (extend as needed for user/role/context-based policies)
type Policy struct {
	ToolName string
	Type     PolicyType
}

// In-memory policy registry
var policies = map[string]Policy{}

// RegisterPolicy adds a policy to the registry
func RegisterPolicy(p Policy) {
	policies[p.ToolName] = p
}

// GetPolicy retrieves a policy for a tool name
func GetPolicy(toolName string) (Policy, bool) {
	p, ok := policies[toolName]
	return p, ok
}

// ApplyPolicy applies the policy to a tool call and returns the policy type
func ApplyPolicy(tc model.ToolCall) PolicyType {
	if p, ok := GetPolicy(tc.Name); ok {
		return p.Type
	}
	return PolicyAllow // Default: allow if no policy is set
}

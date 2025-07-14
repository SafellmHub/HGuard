package policy

import (
	"os"

	"github.com/SafellmHub/hguard-go/pkg/internal/core/model"
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
//
// Example usage:
//
//	err := policy.LoadPoliciesFromYAML("policies.yaml")
//	p, ok := policy.GetPolicy("weather")
//	action := policy.ApplyPolicy(tc)
//
// PolicyType defines the type of policy action (REJECT, REWRITE, LOG, ALLOW, etc.).
type PolicyType string

// PolicyAction is a string representing the result of a policy application.
type PolicyAction string

const (
	PolicyReject  PolicyType = "REJECT"
	PolicyRewrite PolicyType = "REWRITE"
	PolicyLog     PolicyType = "LOG"
	PolicyAllow   PolicyType = "ALLOW"

	ActionNone      PolicyAction = "none"
	ActionRejected  PolicyAction = "rejected"
	ActionRewritten PolicyAction = "rewritten"
	ActionLogged    PolicyAction = "logged"
	ActionApproved  PolicyAction = "approved"
)

// Policy defines a guardrail policy for a tool.
type Policy struct {
	ToolName string     `yaml:"tool_name"`
	Type     PolicyType `yaml:"type"`
}

// In-memory policy registry
var policies = map[string]Policy{}

// RegisterPolicy adds a policy to the registry.
//
// Example:
//
//	policy.RegisterPolicy(Policy{ToolName: "weather", Type: policy.PolicyAllow})
func RegisterPolicy(p Policy) {
	policies[p.ToolName] = p
}

// GetPolicy retrieves a policy for a tool name.
//
// Example:
//
//	p, ok := policy.GetPolicy("weather")
func GetPolicy(toolName string) (Policy, bool) {
	p, ok := policies[toolName]
	return p, ok
}

// ApplyPolicy applies the policy to a tool call and returns the policy type.
//
// Example:
//
//	action := policy.ApplyPolicy(tc)
func ApplyPolicy(tc model.ToolCall) PolicyType {
	if p, ok := GetPolicy(tc.Name); ok {
		return p.Type
	}
	return PolicyAllow // Default: allow if no policy is set
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
	for _, p := range data.Policies {
		RegisterPolicy(p)
	}
	return nil
}

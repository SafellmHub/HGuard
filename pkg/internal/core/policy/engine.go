package policy

import (
	"os"

	"github.com/SafellmHub/HGuard/pkg/internal/core/model"
	"gopkg.in/yaml.v3"
)

// PolicyType defines the type of policy action
// (REJECT, REWRITE, LOG, ALLOW, etc.)
type PolicyType string

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

// Policy defines a guardrail policy for a tool
type Policy struct {
	ToolName string     `yaml:"tool_name"`
	Type     PolicyType `yaml:"type"`
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

// LoadPoliciesFromYAML
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

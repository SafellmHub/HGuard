package policy

import (
	"testing"

	"github.com/SafellmHub/hguard-go/pkg/internal/core/model"
)

func TestContextAwarePolicies(t *testing.T) {
	// Clear policies and add test policies
	policies = make(map[string][]Policy)

	// Role-based policy
	RegisterPolicy(Policy{
		ToolName:  "admin_tool",
		Type:      PolicyReject,
		Condition: "user.role != 'admin'",
		Reason:    "Only admins can use this tool",
		Priority:  10,
	})

	RegisterPolicy(Policy{
		ToolName:  "admin_tool",
		Type:      PolicyAllow,
		Condition: "user.role == 'admin'",
		Reason:    "Admin access granted",
		Priority:  20,
	})

	// Parameter-based policy
	RegisterPolicy(Policy{
		ToolName:  "transfer_money",
		Type:      PolicyReject,
		Condition: "params.amount > 1000",
		Reason:    "Amount too high",
		Priority:  15,
	})

	// Time-based policy
	RegisterPolicy(Policy{
		ToolName:  "maintenance",
		Type:      PolicyReject,
		Condition: "time.hour < 9 || time.hour > 17",
		Reason:    "Maintenance only during business hours",
		Priority:  5,
	})

	// Session-based policy using array indexing to check for presence
	RegisterPolicy(Policy{
		ToolName:  "sensitive_op",
		Type:      PolicyReject,
		Condition: "'sensitive_op' in session.previous_calls",
		Reason:    "Already performed in this session",
		Priority:  8,
	})

	// Permission-based policy using array indexing
	RegisterPolicy(Policy{
		ToolName:  "financial_data",
		Type:      PolicyAllow,
		Condition: "'read_financial' in user.permissions",
		Reason:    "User has financial permissions",
		Priority:  12,
	})

	tests := []struct {
		name     string
		toolCall model.ToolCall
		expected PolicyResult
	}{
		{
			name: "Admin access granted",
			toolCall: model.ToolCall{
				Name: "admin_tool",
				Context: model.CallContext{
					UserRole: "admin",
				},
			},
			expected: PolicyResult{
				Action:  PolicyAllow,
				Reason:  "Admin access granted",
				Matched: true,
			},
		},
		{
			name: "Non-admin access denied",
			toolCall: model.ToolCall{
				Name: "admin_tool",
				Context: model.CallContext{
					UserRole: "user",
				},
			},
			expected: PolicyResult{
				Action:  PolicyReject,
				Reason:  "Only admins can use this tool",
				Matched: true,
			},
		},
		{
			name: "Amount too high",
			toolCall: model.ToolCall{
				Name: "transfer_money",
				Parameters: map[string]interface{}{
					"amount": 5000,
				},
			},
			expected: PolicyResult{
				Action:  PolicyReject,
				Reason:  "Amount too high",
				Matched: true,
			},
		},
		{
			name: "Amount within limit",
			toolCall: model.ToolCall{
				Name: "transfer_money",
				Parameters: map[string]interface{}{
					"amount": 500,
				},
			},
			expected: PolicyResult{
				Action:  PolicyAllow,
				Matched: false,
			},
		},
		{
			name: "Outside business hours",
			toolCall: model.ToolCall{
				Name: "maintenance",
				Context: model.CallContext{
					TimeOfDay: 20, // 8 PM
				},
			},
			expected: PolicyResult{
				Action:  PolicyReject,
				Reason:  "Maintenance only during business hours",
				Matched: true,
			},
		},
		{
			name: "During business hours",
			toolCall: model.ToolCall{
				Name: "maintenance",
				Context: model.CallContext{
					TimeOfDay: 14, // 2 PM
				},
			},
			expected: PolicyResult{
				Action:  PolicyAllow,
				Matched: false,
			},
		},
		{
			name: "Session restriction triggered",
			toolCall: model.ToolCall{
				Name: "sensitive_op",
				Context: model.CallContext{
					PreviousCalls: []string{"weather", "sensitive_op", "search"},
				},
			},
			expected: PolicyResult{
				Action:  PolicyReject,
				Reason:  "Already performed in this session",
				Matched: true,
			},
		},
		{
			name: "Permission granted",
			toolCall: model.ToolCall{
				Name: "financial_data",
				Context: model.CallContext{
					UserPermissions: []string{"read_financial", "write_basic"},
				},
			},
			expected: PolicyResult{
				Action:  PolicyAllow,
				Reason:  "User has financial permissions",
				Matched: true,
			},
		},
		{
			name: "Permission denied",
			toolCall: model.ToolCall{
				Name: "financial_data",
				Context: model.CallContext{
					UserPermissions: []string{"read_basic"},
				},
			},
			expected: PolicyResult{
				Action:  PolicyAllow,
				Reason:  "No matching policies found",
				Matched: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EvaluatePolicy(tt.toolCall)

			if result.Action != tt.expected.Action {
				t.Errorf("Expected action %v, got %v", tt.expected.Action, result.Action)
			}

			if result.Matched != tt.expected.Matched {
				t.Errorf("Expected matched %v, got %v", tt.expected.Matched, result.Matched)
			}

			if tt.expected.Reason != "" && result.Reason != tt.expected.Reason {
				t.Errorf("Expected reason '%s', got '%s'", tt.expected.Reason, result.Reason)
			}
		})
	}
}

func TestPolicyPriority(t *testing.T) {
	// Clear policies and add test policies
	policies = make(map[string][]Policy)

	// Lower priority policy (should be overridden)
	RegisterPolicy(Policy{
		ToolName: "test_tool",
		Type:     PolicyReject,
		Reason:   "Default rejection",
		Priority: 5,
	})

	// Higher priority policy (should take precedence)
	RegisterPolicy(Policy{
		ToolName:  "test_tool",
		Type:      PolicyAllow,
		Condition: "user.role == 'admin'",
		Reason:    "Admin override",
		Priority:  10,
	})

	toolCall := model.ToolCall{
		Name: "test_tool",
		Context: model.CallContext{
			UserRole: "admin",
		},
	}

	result := EvaluatePolicy(toolCall)

	if result.Action != PolicyAllow {
		t.Errorf("Expected PolicyAllow, got %v", result.Action)
	}

	if result.Reason != "Admin override" {
		t.Errorf("Expected 'Admin override', got '%s'", result.Reason)
	}
}

func TestRewritePolicy(t *testing.T) {
	// Clear policies and add test policies
	policies = make(map[string][]Policy)

	RegisterPolicy(Policy{
		ToolName: "old_tool",
		Type:     PolicyRewrite,
		Target:   "new_tool",
		Reason:   "Tool renamed",
		Priority: 10,
	})

	toolCall := model.ToolCall{
		Name: "old_tool",
		Parameters: map[string]interface{}{
			"param1": "value1",
		},
	}

	result := EvaluatePolicy(toolCall)

	if result.Action != PolicyRewrite {
		t.Errorf("Expected PolicyRewrite, got %v", result.Action)
	}

	if result.Target != "new_tool" {
		t.Errorf("Expected target 'new_tool', got '%s'", result.Target)
	}

	if result.Reason != "Tool renamed" {
		t.Errorf("Expected 'Tool renamed', got '%s'", result.Reason)
	}
}

func TestComplexConditions(t *testing.T) {
	// Clear policies and add test policies
	policies = make(map[string][]Policy)

	RegisterPolicy(Policy{
		ToolName:  "complex_tool",
		Type:      PolicyAllow,
		Condition: "user.role == 'admin' && params.amount < 1000 && time.hour >= 9 && time.hour <= 17 && len(session.previous_calls) < 3",
		Reason:    "Complex condition met",
		Priority:  10,
	})

	// Test case that should pass
	toolCall := model.ToolCall{
		Name: "complex_tool",
		Parameters: map[string]interface{}{
			"amount": 500,
		},
		Context: model.CallContext{
			UserRole:      "admin",
			TimeOfDay:     14,
			PreviousCalls: []string{"tool1", "tool2"},
		},
	}

	result := EvaluatePolicy(toolCall)

	if result.Action != PolicyAllow {
		t.Errorf("Expected PolicyAllow, got %v", result.Action)
	}

	// Test case that should fail (too many previous calls)
	toolCall.Context.PreviousCalls = []string{"tool1", "tool2", "tool3", "tool4"}

	result = EvaluatePolicy(toolCall)

	if result.Action != PolicyAllow {
		t.Errorf("Expected PolicyAllow (default), got %v", result.Action)
	}

	if result.Matched {
		t.Error("Expected no policy match, but got match")
	}
}

func TestMetadataConditions(t *testing.T) {
	// Clear policies and add test policies
	policies = make(map[string][]Policy)

	RegisterPolicy(Policy{
		ToolName:  "premium_tool",
		Type:      PolicyAllow,
		Condition: "metadata.subscription_tier == 'premium'",
		Reason:    "Premium subscription verified",
		Priority:  10,
	})

	toolCall := model.ToolCall{
		Name: "premium_tool",
		Context: model.CallContext{
			Metadata: map[string]interface{}{
				"subscription_tier": "premium",
				"region":            "us-west",
			},
		},
	}

	result := EvaluatePolicy(toolCall)

	if result.Action != PolicyAllow {
		t.Errorf("Expected PolicyAllow, got %v", result.Action)
	}

	if result.Reason != "Premium subscription verified" {
		t.Errorf("Expected 'Premium subscription verified', got '%s'", result.Reason)
	}
}

func TestBackwardCompatibility(t *testing.T) {
	// Clear policies and add test policies
	policies = make(map[string][]Policy)

	RegisterPolicy(Policy{
		ToolName: "simple_tool",
		Type:     PolicyAllow,
		// No condition - should always match
	})

	toolCall := model.ToolCall{
		Name: "simple_tool",
	}

	// Test new function
	result := EvaluatePolicy(toolCall)
	if result.Action != PolicyAllow {
		t.Errorf("Expected PolicyAllow, got %v", result.Action)
	}

	// Test legacy function
	legacyResult := ApplyPolicy(toolCall)
	if legacyResult != PolicyAllow {
		t.Errorf("Expected PolicyAllow, got %v", legacyResult)
	}
}

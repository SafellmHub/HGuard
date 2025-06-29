package validation

import (
	"fmt"

	"github.com/SafellmHub/hguard-go/pkg/internal/core/fuzzy"
	"github.com/SafellmHub/hguard-go/pkg/internal/core/model"
	"github.com/SafellmHub/hguard-go/pkg/internal/core/policy"
	"github.com/SafellmHub/hguard-go/pkg/internal/schema"
)

// ValidateToolCall is the main entry point for validating a tool call
func ValidateToolCall(tc model.ToolCall) model.ValidationResult {
	ts, ok := schema.GetToolSchema(tc.Name)
	if !ok {
		// Fuzzy match for near-miss tool names
		knownNames := make([]string, 0, len(schema.ToolSchemas()))
		for name := range schema.ToolSchemas() {
			knownNames = append(knownNames, name)
		}
		if suggestion, _ := fuzzy.FuzzyMatchToolName(tc.Name, knownNames, 2); suggestion != "" {
			// Check if policy for this tool is REWRITE
			p, hasPolicy := policy.GetPolicy(suggestion)
			if hasPolicy && p.Type == policy.PolicyRewrite {
				// Rewrite and approve
				return model.ValidationResult{
					ToolCallID:       tc.ID,
					Status:           "rewritten",
					Confidence:       0.95,
					Reason:           fmt.Sprintf("Tool name rewritten to '%s' by policy", suggestion),
					ExecutionAllowed: true,
					PolicyAction:     string(policy.ActionRewritten),
					Modifications:    map[string]interface{}{"name": suggestion},
					SuggestedCorrection: &model.ToolCall{
						ID:         tc.ID,
						Name:       suggestion,
						Parameters: tc.Parameters,
						Context:    tc.Context,
						Timestamp:  tc.Timestamp,
					},
				}
			}
			return model.ValidationResult{
				ToolCallID: tc.ID,
				Status:     "rejected",
				Confidence: 0.9,
				Reason:     fmt.Sprintf("Unknown tool name. Did you mean '%s'?", suggestion),
				SuggestedCorrection: &model.ToolCall{
					ID:         tc.ID,
					Name:       suggestion,
					Parameters: tc.Parameters,
					Context:    tc.Context,
					Timestamp:  tc.Timestamp,
				},
				ExecutionAllowed: false,
				PolicyAction:     string(policy.ActionRejected),
			}
		}
		return model.ValidationResult{
			ToolCallID:       tc.ID,
			Status:           "rejected",
			Confidence:       1.0,
			Reason:           "Unknown tool name",
			ExecutionAllowed: false,
			PolicyAction:     string(policy.ActionRejected),
		}
	}

	err := schema.ValidateParameters(ts, tc.Parameters)
	if err != nil {
		return model.ValidationResult{
			ToolCallID:       tc.ID,
			Status:           "rejected",
			Confidence:       1.0,
			Reason:           fmt.Sprintf("Parameter validation failed: %v", err),
			ExecutionAllowed: false,
			PolicyAction:     string(policy.ActionRejected),
		}
	}

	// Policy engine integration
	policyType := policy.ApplyPolicy(tc)
	switch policyType {
	case policy.PolicyReject:
		return model.ValidationResult{
			ToolCallID:       tc.ID,
			Status:           "rejected",
			Confidence:       1.0,
			Reason:           "Policy engine: tool call rejected by policy",
			ExecutionAllowed: false,
			PolicyAction:     string(policy.ActionRejected),
		}
	case policy.PolicyRewrite:
		// For now i will just mark as rewritten
		return model.ValidationResult{
			ToolCallID:       tc.ID,
			Status:           "rewritten",
			Confidence:       1.0,
			Reason:           "Policy engine: tool call rewritten by policy",
			ExecutionAllowed: true,
			PolicyAction:     string(policy.ActionRewritten),
		}
	case policy.PolicyLog:
		return model.ValidationResult{
			ToolCallID:       tc.ID,
			Status:           "approved",
			Confidence:       1.0,
			ExecutionAllowed: true,
			PolicyAction:     string(policy.ActionLogged),
		}
	case policy.PolicyAllow:
		return model.ValidationResult{
			ToolCallID:       tc.ID,
			Status:           "approved",
			Confidence:       1.0,
			ExecutionAllowed: true,
			PolicyAction:     string(policy.ActionApproved),
		}
	default:
		return model.ValidationResult{
			ToolCallID:       tc.ID,
			Status:           "approved",
			Confidence:       1.0,
			ExecutionAllowed: true,
			PolicyAction:     string(policy.ActionApproved),
		}
	}
}

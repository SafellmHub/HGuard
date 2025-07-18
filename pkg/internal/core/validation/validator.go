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
			suggestedTC := model.ToolCall{
				ID:         tc.ID,
				Name:       suggestion,
				Parameters: tc.Parameters,
				Context:    tc.Context,
				Timestamp:  tc.Timestamp,
			}
			policyResult := policy.EvaluatePolicy(suggestedTC)
			if policyResult.Action == policy.PolicyRewrite {
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

	// Use the new policy evaluation with context-aware conditions
	policyResult := policy.EvaluatePolicy(tc)
	switch policyResult.Action {
	case policy.PolicyReject:
		return model.ValidationResult{
			ToolCallID:       tc.ID,
			Status:           "rejected",
			Confidence:       1.0,
			Reason:           policyResult.Reason,
			ExecutionAllowed: false,
			PolicyAction:     string(policy.ActionRejected),
		}
	case policy.PolicyContextReject:
		return model.ValidationResult{
			ToolCallID:       tc.ID,
			Status:           "rejected",
			Confidence:       1.0,
			Reason:           policyResult.Reason,
			ExecutionAllowed: false,
			PolicyAction:     string(policy.ActionRejected),
		}
	case policy.PolicyRewrite:
		target := policyResult.Target
		if target == "" {
			target = tc.Name // Default to same tool if no target specified
		}
		return model.ValidationResult{
			ToolCallID:       tc.ID,
			Status:           "rewritten",
			Confidence:       1.0,
			Reason:           policyResult.Reason,
			ExecutionAllowed: true,
			PolicyAction:     string(policy.ActionRewritten),
			Modifications:    map[string]interface{}{"name": target},
			SuggestedCorrection: &model.ToolCall{
				ID:         tc.ID,
				Name:       target,
				Parameters: tc.Parameters,
				Context:    tc.Context,
				Timestamp:  tc.Timestamp,
			},
		}
	case policy.PolicyLog:
		return model.ValidationResult{
			ToolCallID:       tc.ID,
			Status:           "approved",
			Confidence:       1.0,
			Reason:           policyResult.Reason,
			ExecutionAllowed: true,
			PolicyAction:     string(policy.ActionLogged),
		}
	case policy.PolicyAllow:
		return model.ValidationResult{
			ToolCallID:       tc.ID,
			Status:           "approved",
			Confidence:       1.0,
			Reason:           policyResult.Reason,
			ExecutionAllowed: true,
			PolicyAction:     string(policy.ActionApproved),
		}
	default:
		return model.ValidationResult{
			ToolCallID:       tc.ID,
			Status:           "approved",
			Confidence:       1.0,
			Reason:           policyResult.Reason,
			ExecutionAllowed: true,
			PolicyAction:     string(policy.ActionApproved),
		}
	}
}

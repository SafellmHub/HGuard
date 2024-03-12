package validation

import (
	"fmt"

	"github.com/fishonamos/hallucination-shield/internal/core/fuzzy"
	"github.com/fishonamos/hallucination-shield/internal/core/model"
	"github.com/fishonamos/hallucination-shield/internal/core/policy"
	"github.com/fishonamos/hallucination-shield/internal/schema"
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
			}
		}
		return model.ValidationResult{
			ToolCallID:       tc.ID,
			Status:           "rejected",
			Confidence:       1.0,
			Reason:           "Unknown tool name",
			ExecutionAllowed: false,
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
		}
	case policy.PolicyAllow, policy.PolicyLog, policy.PolicyRewrite:
		// For MVP, treat LOG and REWRITE as ALLOW (can extend later)
		return model.ValidationResult{
			ToolCallID:       tc.ID,
			Status:           "approved",
			Confidence:       1.0,
			ExecutionAllowed: true,
		}
	default:
		return model.ValidationResult{
			ToolCallID:       tc.ID,
			Status:           "approved",
			Confidence:       1.0,
			ExecutionAllowed: true,
		}
	}
}

package schema

import (
	"fmt"
	"os"

	"github.com/SafellmHub/hguard-go/pkg/internal/core/fuzzy"
	"github.com/SafellmHub/hguard-go/pkg/internal/core/model"
	"github.com/SafellmHub/hguard-go/pkg/internal/core/policy"
	"gopkg.in/yaml.v3"
)

// ParameterSchema defines the expected type and requirements for a parameter
type ParameterSchema struct {
	Type      string // e.g., "string", "number", "boolean"
	Required  bool
	Enum      []string // allowed values (optional)
	Pattern   string   // regex pattern (optional)
	MaxLength int      // for strings (optional)
}

// ToolSchema defines the schema for a tool
type ToolSchema struct {
	Name       string
	Parameters map[string]ParameterSchema
}

// In-memory registry of tool schemas
var toolSchemas = map[string]ToolSchema{}

// RegisterToolSchema adds a tool schema to the registry
func RegisterToolSchema(schema ToolSchema) {
	toolSchemas[schema.Name] = schema
}

// GetToolSchema retrieves a tool schema by name
func GetToolSchema(name string) (ToolSchema, bool) {
	schema, ok := toolSchemas[name]
	return schema, ok
}

// ToolSchemas returns a copy of all registered tool schemas
func ToolSchemas() map[string]ToolSchema {
	copy := make(map[string]ToolSchema, len(toolSchemas))
	for k, v := range toolSchemas {
		copy[k] = v
	}
	return copy
}

// ValidateParameters checks if the parameters conform to the schema
func ValidateParameters(schema ToolSchema, params map[string]interface{}) error {
	for paramName, paramSchema := range schema.Parameters {
		value, exists := params[paramName]
		if paramSchema.Required && !exists {
			return fmt.Errorf("missing required parameter: %s", paramName)
		}
		if exists {
			switch paramSchema.Type {
			case "string":
				str, ok := value.(string)
				if !ok {
					return fmt.Errorf("parameter %s should be a string", paramName)
				}
				if paramSchema.MaxLength > 0 && len(str) > paramSchema.MaxLength {
					return fmt.Errorf("parameter %s exceeds max length", paramName)
				}
				// I can use Pattern and Enum checks here
			case "number":
				_, ok1 := value.(float64)
				_, ok2 := value.(int)
				if !ok1 && !ok2 {
					return fmt.Errorf("parameter %s should be a number", paramName)
				}
			case "boolean":
				_, ok := value.(bool)
				if !ok {
					return fmt.Errorf("parameter %s should be a boolean", paramName)
				}
			}
		}
	}
	return nil
}

// LoadSchemasFromYAML loads tool schemas from a YAML file and registers them
func LoadSchemasFromYAML(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	var data struct {
		Schemas []ToolSchema `yaml:"schemas"`
	}
	if err := yaml.NewDecoder(f).Decode(&data); err != nil {
		return err
	}
	for _, s := range data.Schemas {
		RegisterToolSchema(s)
	}
	return nil
}

// ValidateAndPolicy validates a tool call and applies the policy, returning a ValidationResult
func ValidateAndPolicy(tc model.ToolCall) model.ValidationResult {
	result := model.ValidationResult{
		ToolCallID:       tc.ID,
		Status:           "approved",
		Confidence:       1.0,
		ExecutionAllowed: true,
		PolicyAction:     string(policy.PolicyAllow),
	}

	schema, ok := GetToolSchema(tc.Name)
	if !ok {
		// Check for REWRITE policy for this tool name
		ptype := policy.ApplyPolicy(tc)
		if ptype == policy.PolicyRewrite {
			// Fuzzy match to suggest correction
			known := make([]string, 0, len(toolSchemas))
			for k := range toolSchemas {
				known = append(known, k)
			}
			if suggestion, _ := fuzzy.FuzzyMatchToolName(tc.Name, known, 2); suggestion != "" {
				result.Status = "rewritten"
				result.Confidence = 1.0
				result.ExecutionAllowed = true
				result.PolicyAction = string(policy.PolicyRewrite)
				result.SuggestedCorrection = &model.ToolCall{
					ID:         tc.ID,
					Name:       suggestion,
					Parameters: tc.Parameters,
					Context:    tc.Context,
					Timestamp:  tc.Timestamp,
				}
				result.Reason = "Did you mean '" + suggestion + "'?"
				return result
			}
			result.Status = "rejected"
			result.Confidence = 0.0
			result.ExecutionAllowed = false
			result.Reason = "Unknown tool name and no close match found"
			result.PolicyAction = string(policy.PolicyReject)
			return result
		}
		result.Status = "rejected"
		result.Confidence = 0.0
		result.ExecutionAllowed = false
		result.Reason = "Unknown tool name"
		result.PolicyAction = string(policy.PolicyReject)
		return result
	}

	err := ValidateParameters(schema, tc.Parameters)
	if err != nil {
		result.Status = "rejected"
		result.Confidence = 0.0
		result.ExecutionAllowed = false
		result.Reason = err.Error()
		result.PolicyAction = string(policy.PolicyReject)
		return result
	}

	ptype := policy.ApplyPolicy(tc)
	result.PolicyAction = string(ptype)
	if ptype == policy.PolicyReject {
		result.Status = "rejected"
		result.Confidence = 1.0
		result.ExecutionAllowed = false
		result.Reason = "Policy rejected tool call"
	} else if ptype == policy.PolicyAllow {
		result.Status = "approved"
		result.Confidence = 1.0
		result.ExecutionAllowed = true
	} else if ptype == policy.PolicyLog {
		result.Status = "approved"
		result.Confidence = 1.0
		result.ExecutionAllowed = true
	} else if ptype == policy.PolicyRewrite {
		result.Status = "rewritten"
		result.Confidence = 1.0
		result.ExecutionAllowed = true
		// Optionally, set result.Modifications or result.SuggestedCorrection
	}
	return result
}

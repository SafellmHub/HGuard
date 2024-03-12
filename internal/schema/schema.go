package schema

import (
	"fmt"
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
				// Pattern and Enum checks can be added here
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

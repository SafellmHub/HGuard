# HGuard - LLM Tool Call Guardrails

HGuard is a lightweight middleware system for detecting and preventing hallucinated tool use in large language models (LLMs). It works by intercepting and validating function/tool calls made by the model, then filtering, rewriting, or blocking hallucinated or invalid ones.

## Installation

```bash
go get github.com/fishonamos/HGuard@v0.1.0
```

## Quick Start

```go
package main

import (
    "fmt"
    "time"

    "github.com/fishonamos/HGuard/pkg/hallucinationguard"
    "github.com/fishonamos/HGuard/internal/core/model"
)

func main() {
    // Create a new guard instance
    guard := hallucinationguard.New()

    // Load tool schemas and policies
    if err := guard.LoadSchemasFromFile("schemas.yaml"); err != nil {
        panic(err)
    }
    if err := guard.LoadPoliciesFromFile("policies.yaml"); err != nil {
        panic(err)
    }

    // Validate a tool call
    toolCall := model.ToolCall{
        ID: "call_001",
        Name: "weather",
        Parameters: map[string]interface{}{
            "location": "London",
            "unit": "C",
        },
        Context: model.CallContext{
            UserID: "user123",
        },
        Timestamp: time.Now(),
    }

    result := guard.ValidateToolCall(toolCall)
    fmt.Printf("Validation result: %+v\n", result)

    if result.ExecutionAllowed {
        // Execute the tool call
        fmt.Println("Tool call approved, executing...")
    } else if result.SuggestedCorrection != nil {
        // Handle typo correction
        fmt.Printf("Did you mean: %s?\n", result.SuggestedCorrection.Name)
    } else {
        // Handle rejection
        fmt.Printf("Tool call rejected: %s\n", result.Reason)
    }
}
```

## Configuration Files

### Tool Schemas (schemas.yaml)

Define your tool schemas to validate parameters and types:

```yaml
tools:
  - name: weather
    description: Get current weather for a location
    parameters:
      - name: location
        type: string
        required: true
        description: City name or coordinates
      - name: unit
        type: string
        required: false
        enum: [C, F]
        default: C
        description: Temperature unit

  - name: stock_price
    description: Get stock price for a symbol
    parameters:
      - name: symbol
        type: string
        required: true
        description: Stock ticker symbol
      - name: date
        type: string
        required: false
        format: date
        description: Historical date (YYYY-MM-DD)
```

### Policies (policies.yaml)

Define policies to control tool access and behavior:

```yaml
policies:
  - tool_name: weather
    type: ALLOW # Always allow weather queries

  - tool_name: stock_price
    type: ALLOW # Allow stock price queries

  - tool_name: transfer_money
    type: REJECT # Never allow money transfers

  - tool_name: wether # Common typo
    type: REWRITE # Auto-correct to 'weather'
```

## Advanced Usage

### 1. Custom Validation Logic

```go
// Create a custom validator
type CustomValidator struct {
    guard *hallucinationguard.Guard
}

func (v *CustomValidator) Validate(tc model.ToolCall) model.ValidationResult {
    // Get base validation from HGuard
    result := v.guard.ValidateToolCall(tc)

    // Add custom validation logic
    if tc.Name == "weather" {
        // Example: Block weather queries for certain locations
        if loc, ok := tc.Parameters["location"].(string); ok {
            if loc == "Area 51" {
                result.Status = "rejected"
                result.ExecutionAllowed = false
                result.Reason = "Location not allowed"
            }
        }
    }

    return result
}
```

### 2. Handling Typo Corrections

```go
func handleToolCall(guard *hallucinationguard.Guard, tc model.ToolCall) {
    result := guard.ValidateToolCall(tc)

    switch result.Status {
    case "approved":
        executeToolCall(tc)
    case "rewritten":
        if correction := result.SuggestedCorrection; correction != nil {
            fmt.Printf("Corrected '%s' to '%s'\n", tc.Name, correction.Name)
            executeToolCall(*correction)
        }
    case "rejected":
        fmt.Printf("Rejected: %s\n", result.Reason)
    }
}
```

### 3. Integration with OpenAI Function Calling

```go
func validateOpenAIFunctionCall(guard *hallucinationguard.Guard, functionCall openai.FunctionCall) model.ValidationResult {
    // Convert OpenAI function call to HGuard tool call
    toolCall := model.ToolCall{
        ID: functionCall.ID,
        Name: functionCall.Name,
        Parameters: functionCall.Arguments,
        Context: model.CallContext{
            UserID: "openai_user",
        },
        Timestamp: time.Now(),
    }

    return guard.ValidateToolCall(toolCall)
}
```

## Best Practices

1. **Always Validate**: Never execute tool calls without validation

   ```go
   // ❌ Bad
   executeToolCall(toolCall)

   // ✅ Good
   if result := guard.ValidateToolCall(toolCall); result.ExecutionAllowed {
       executeToolCall(toolCall)
   }
   ```

2. **Handle Corrections**: Always check for suggested corrections

   ```go
   if result.SuggestedCorrection != nil {
       // Use the corrected version
       toolCall = *result.SuggestedCorrection
   }
   ```

3. **Log Rejections**: Keep track of rejected calls for monitoring

   ```go
   if !result.ExecutionAllowed {
       log.Printf("Rejected tool call: %s (Reason: %s)\n",
           toolCall.Name, result.Reason)
   }
   ```

4. **Use Context**: Include user and session context for better validation
   ```go
   toolCall.Context = model.CallContext{
       UserID: "user123",
       SessionID: "sess456",
       IP: "192.168.1.1",
   }
   ```

## API Reference

### Guard

```go
type Guard struct {
    // ... internal fields
}

// New creates a new Guard instance
func New() *Guard

// LoadSchemasFromFile loads tool schemas from a YAML file
func (g *Guard) LoadSchemasFromFile(path string) error

// LoadPoliciesFromFile loads policies from a YAML file
func (g *Guard) LoadPoliciesFromFile(path string) error

// ValidateToolCall validates a tool call using loaded schemas and policies
func (g *Guard) ValidateToolCall(tc model.ToolCall) model.ValidationResult
```

### ValidationResult

```go
type ValidationResult struct {
    ToolCallID       string
    Status           string    // "approved", "rejected", "rewritten"
    Confidence       float64
    ExecutionAllowed bool
    Reason           string
    SuggestedCorrection *ToolCall
    PolicyAction     string
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

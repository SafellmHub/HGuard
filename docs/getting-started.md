# Getting Started with HallucinationGuard

This guide will help you get started with HallucinationGuard, a middleware system for detecting and preventing hallucinated tool use in LLMs.

## Installation

```bash
go get github.com/fishonamos/HGuard@v0.1.0
```

## Basic Usage

1. Create a new guard instance:

```go
guard := hallucinationguard.New()
```

2. Load your tool schemas and policies:

```go
if err := guard.LoadSchemasFromFile("schemas.yaml"); err != nil {
    panic(err)
}
if err := guard.LoadPoliciesFromFile("policies.yaml"); err != nil {
    panic(err)
}
```

3. Validate tool calls:

```go
result := guard.ValidateToolCall(toolCall)
if result.ExecutionAllowed {
    // Execute the tool call
} else if result.SuggestedCorrection != nil {
    // Handle typo correction
} else {
    // Handle rejection
}
```

## Configuration Files

### Tool Schemas (schemas.yaml)

Define your tool schemas to validate parameters and types:

```yaml
schemas:
  - name: weather
    description: Get current weather for a location
    parameters:
      location:
        type: string
        required: true
        pattern: "^[a-zA-Z\\s,]+$" # Only letters, spaces, and commas
        minLength: 2
        maxLength: 100
      unit:
        type: string
        required: false
        enum: ["C", "F"]
        default: "C"
```

### Policies (policies.yaml)

Define policies to control tool access and behavior:

```yaml
policies:
  - tool_name: weather
    type: ALLOW # Always allow weather queries

  - tool_name: transfer_money
    type: REJECT # Never allow money transfers
    reason: "Money transfers are not allowed for security reasons"

  - tool_name: wether # Common typo
    type: REWRITE # Auto-correct to 'weather'
    target: weather
```

## Common Use Cases

### 1. Basic Tool Validation

```go
toolCall := model.ToolCall{
    ID: "call_001",
    Name: "weather",
    Parameters: map[string]interface{}{
        "location": "London",
        "unit": "C",
    },
}

result := guard.ValidateToolCall(toolCall)
if result.ExecutionAllowed {
    fmt.Println("Tool call approved!")
} else {
    fmt.Printf("Tool call rejected: %s\n", result.Reason)
}
```

### 2. Handling Typo Corrections

```go
toolCall := model.ToolCall{
    ID: "call_001",
    Name: "wether", // Common typo
    Parameters: map[string]interface{}{
        "location": "Paris",
    },
}

result := guard.ValidateToolCall(toolCall)
if result.SuggestedCorrection != nil {
    fmt.Printf("Did you mean: %s?\n", result.SuggestedCorrection.Name)
    // Use the corrected version
    toolCall = *result.SuggestedCorrection
}
```

### 3. Rate Limiting

```go
// Rate limiting is handled automatically by the policy engine
// Just define the policy in policies.yaml:
policies:
  - tool_name: weather
    type: RATE_LIMIT
    max_calls: 10
    window: 60 # seconds
```

### 4. Context-Based Policies

```go
// Add context to your tool calls
toolCall.Context = model.CallContext{
    UserID: "user123",
    SessionID: "sess456",
    PreviousCalls: []string{"weather", "stock_price"},
}

// Define context-based policies in policies.yaml:
policies:
  - tool_name: transfer_money
    type: CONTEXT_REJECT
    condition: "previous_calls contains 'transfer_money'"
    reason: "Multiple transfer attempts detected"
```

## Best Practices

1. **Always Validate**: Never execute tool calls without validation
2. **Handle Corrections**: Always check for suggested corrections
3. **Log Rejections**: Keep track of rejected calls for monitoring
4. **Use Context**: Include user and session context for better validation
5. **Keep Schemas Updated**: Ensure your schemas match your actual tools
6. **Test Policies**: Regularly test your policies with edge cases

## Next Steps

- Check out the [examples](https://github.com/fishonamos/HGuard/tree/main/example) for more use cases
- Read the [API Reference](../README.md#api-reference) for detailed documentation
- Join our [Discord community](https://discord.gg/hallucinationguard) for support

## Troubleshooting

### Common Issues

1. **Tool Calls Not Validating**

   - Check that your schemas and policies are loaded correctly
   - Verify that tool names match exactly
   - Ensure parameters match the schema definitions

2. **Rate Limiting Not Working**

   - Verify the policy is defined correctly
   - Check that the window and max_calls values are set
   - Ensure context is being passed correctly

3. **Typo Corrections Not Working**
   - Verify the REWRITE policy is defined
   - Check that the target tool exists
   - Ensure the tool name is close enough for correction

### Getting Help

- Check the [FAQ](../docs/faq.md)
- Open an [issue](https://github.com/fishonamos/HGuard/issues)
- Join our [Discord community](https://discord.gg/hallucinationguard)

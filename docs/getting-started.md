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

3. Validate tool calls with context:

```go
toolCall := hallucinationguard.ToolCall{
    Name: "transfer_money",
    Parameters: map[string]interface{}{
        "amount": 5000,
    },
    Context: &hallucinationguard.CallContext{
        UserRole: "admin",
        UserID:   "user123",
        SessionID: "session456",
        PreviousCalls: []string{"get_balance"},
        UserPermissions: []string{"financial_access"},
        TimeOfDay: 14, // 2 PM
    },
}

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

Define policies to control tool access and behavior. HallucinationGuard now supports context-aware and conditional policies:

```yaml
policies:
  # Simple policies
  - tool_name: weather
    type: ALLOW # Always allow weather queries

  - tool_name: transfer_money
    type: REJECT # Never allow money transfers
    reason: "Money transfers are not allowed for security reasons"

  # Role-based access control
  - tool_name: admin_tool
    type: REJECT
    condition: "user.role != 'admin'"
    reason: "Only administrators can use this tool"
    priority: 10

  - tool_name: admin_tool
    type: ALLOW
    condition: "user.role == 'admin'"
    reason: "Admin access granted"
    priority: 20

  # Parameter-based restrictions
  - tool_name: transfer_money
    type: ALLOW
    condition: "user.role == 'admin' && params.amount < 1000"
    reason: "Small transfers allowed for admins"
    priority: 15

  # Session-based restrictions
  - tool_name: sensitive_operation
    type: REJECT
    condition: "'sensitive_operation' in session.previous_calls"
    reason: "Sensitive operation already performed in this session"
    priority: 5

  # Time-based restrictions
  - tool_name: maintenance_tool
    type: REJECT
    condition: "time.hour < 9 || time.hour > 17"
    reason: "Maintenance tools only available during business hours"
    priority: 5

  # Permission-based access
  - tool_name: financial_data
    type: ALLOW
    condition: "'read_financial' in user.permissions"
    reason: "User has financial data access permission"
    priority: 8

  # Complex conditions
  - tool_name: high_value_transaction
    type: ALLOW
    condition: "user.role == 'admin' && params.amount < 10000 && time.hour >= 9 && time.hour <= 17"
    reason: "High-value transaction approved under controlled conditions"
    priority: 25

  # Auto-correct common typos
  - tool_name: wether # Common typo
    type: REWRITE # Auto-correct to 'weather'
    target: weather
    reason: "Auto-corrected 'wether' to 'weather'"

  # Fallback policy
  - tool_name: "*"
    type: REJECT
    reason: "Unknown tools are not allowed"
    priority: 1
```

## Context-Aware Policies

HallucinationGuard supports rich context-aware policies that can make decisions based on:

### Available Context Fields

- **user**: User information
  - `user.id`: User ID
  - `user.role`: User role (admin, user, etc.)
  - `user.permissions`: Array of user permissions
- **session**: Session information
  - `session.id`: Session ID
  - `session.conversation_id`: Conversation ID
  - `session.previous_calls`: Array of previous tool calls in session
- **params**: Tool parameters
  - `params.amount`: Access any parameter passed to the tool
- **time**: Time information
  - `time.hour`: Current hour (0-23)
- **request**: Request information
  - `request.ip`: Client IP address
- **metadata**: Custom metadata
  - `metadata.subscription_tier`: Any custom metadata fields

### Expression Syntax

Conditions use a powerful expression syntax supporting:

- **Comparison operators**: `==`, `!=`, `<`, `<=`, `>`, `>=`
- **Logical operators**: `&&`, `||`, `!`
- **Array membership**: `'item' in array`
- **Built-in functions**: `len(array)`
- **Parentheses**: For grouping expressions

### Policy Priority

When multiple policies match the same tool, the policy with the highest priority value is applied. This allows for specific overrides of general rules.

## Common Use Cases

### 1. Basic Tool Validation

```go
toolCall := hallucinationguard.ToolCall{
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
    fmt.Printf("Tool call rejected: %s\n", result.Error)
}
```

### 2. Role-Based Access Control

```go
toolCall := hallucinationguard.ToolCall{
    Name: "admin_tool",
    Parameters: map[string]interface{}{},
    Context: &hallucinationguard.CallContext{
        UserRole: "admin",
        UserID:   "user123",
    },
}

result := guard.ValidateToolCall(toolCall)
// Will be allowed if user is admin, rejected otherwise
```

### 3. Parameter-Based Restrictions

```go
toolCall := hallucinationguard.ToolCall{
    Name: "transfer_money",
    Parameters: map[string]interface{}{
        "amount": 500,
    },
    Context: &hallucinationguard.CallContext{
        UserRole: "admin",
    },
}

result := guard.ValidateToolCall(toolCall)
// Will be allowed if user is admin and amount < 1000
```

### 4. Session-Based Restrictions

```go
toolCall := hallucinationguard.ToolCall{
    Name: "sensitive_operation",
    Parameters: map[string]interface{}{},
    Context: &hallucinationguard.CallContext{
        SessionID: "session123",
        PreviousCalls: []string{"login", "get_balance"},
    },
}

result := guard.ValidateToolCall(toolCall)
// Will be allowed if sensitive_operation not in previous calls
```

### 5. Time-Based Restrictions

```go
toolCall := hallucinationguard.ToolCall{
    Name: "maintenance_tool",
    Parameters: map[string]interface{}{},
    Context: &hallucinationguard.CallContext{
        TimeOfDay: 14, // 2 PM
    },
}

result := guard.ValidateToolCall(toolCall)
// Will be allowed during business hours (9-17)
```

### 6. Permission-Based Access

```go
toolCall := hallucinationguard.ToolCall{
    Name: "financial_data",
    Parameters: map[string]interface{}{},
    Context: &hallucinationguard.CallContext{
        UserPermissions: []string{"read_financial", "write_basic"},
    },
}

result := guard.ValidateToolCall(toolCall)
// Will be allowed if user has 'read_financial' permission
```

## Best Practices

1. **Always Validate**: Never execute tool calls without validation
2. **Use Context**: Always provide context for better policy decisions
3. **Handle Corrections**: Always check for suggested corrections
4. **Log Rejections**: Keep track of rejected calls for monitoring
5. **Use Priorities**: Use priority levels to create specific overrides
6. **Test Policies**: Regularly test your policies with edge cases
7. **Keep Schemas Updated**: Ensure your schemas match your actual tools

## Next Steps

- Check out the [examples](https://github.com/fishonamos/HGuard/tree/main/example) for more use cases
- Read the [API Reference](../README.md#api-reference) for detailed documentation
- See the [sample policies](../pkg/internal/config/sample_policies.yaml) for more examples

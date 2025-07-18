# HallucinationGuard

[![Go Report Card](https://goreportcard.com/badge/github.com/SafellmHub/hguard-go)](https://goreportcard.com/report/github.com/SafellmHub/hguard-go)
[![GoDoc](https://godoc.org/github.com/SafellmHub/hguard-go?status.svg)](https://godoc.org/github.com/SafellmHub/hguard-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

HallucinationGuard is a Go SDK for validating and enforcing guardrails on LLM tool calls. It provides schema validation, policy enforcement, and extensibility for production-grade AI integrations.

![status: experimental](https://img.shields.io/badge/status-experimental-orange)

⚠️ Experimental Notice

This package is currently experimental and still under active development.

We welcome your feedback and encourage you to report issues or suggest improvements.

## Features

- **Schema Validation**: Structured validation of tool calls against JSON schemas
- **Context-Aware Policies**: Role-based, time-based, and session-based policy enforcement
- **Conditional Logic**: Complex conditional expressions for advanced policy rules
- **Policy Priority**: Hierarchical policy system with priority-based rule resolution
- **Auto-Correction**: Automatic tool name correction for common typos
- **Thread-Safe**: Safe for concurrent use in production environments
- **Extensible**: Custom schema loaders and policy engines

## Installation

```sh
go get github.com/SafellmHub/hguard-go
```

## Usage Example

Add HallucinationGuard to your agent with context-aware policies:

```go
import (
    "context"
    "log"
    "github.com/SafellmHub/hguard-go/pkg/hallucinationguard"
)

type HGuardAgent struct {
    guard *hallucinationguard.Guard
}

func NewHGuardAgent(schemaPath, policyPath string) *HGuardAgent {
    ctx := context.Background()
    guard := hallucinationguard.New()
    if err := guard.LoadSchemasFromFile(ctx, schemaPath); err != nil {
        log.Fatalf("Schema load error: %v", err)
    }
    if err := guard.LoadPoliciesFromFile(ctx, policyPath); err != nil {
        log.Fatalf("Policy load error: %v", err)
    }
    return &HGuardAgent{guard: guard}
}

func (a *HGuardAgent) ValidateToolCall(ctx context.Context, toolCall hallucinationguard.ToolCall) hallucinationguard.ValidationResult {
    return a.guard.ValidateToolCall(ctx, toolCall)
}
```

## Context-Aware Policies

HallucinationGuard supports rich context-aware policies:

```yaml
policies:
  # Role-based access control
  - tool_name: admin_tool
    type: REJECT
    condition: "user.role != 'admin'"
    reason: "Only administrators can use this tool"
    priority: 10

  # Parameter-based restrictions
  - tool_name: transfer_money
    type: ALLOW
    condition: "user.role == 'admin' && params.amount < 1000"
    reason: "Small transfers allowed for admins"
    priority: 15

  # Time-based restrictions
  - tool_name: maintenance_tool
    type: REJECT
    condition: "time.hour < 9 || time.hour > 17"
    reason: "Maintenance tools only available during business hours"
    priority: 5

  # Session-based restrictions
  - tool_name: sensitive_operation
    type: REJECT
    condition: "'sensitive_operation' in session.previous_calls"
    reason: "Operation already performed in this session"
    priority: 8
```

## Usage with Context

```go
toolCall := hallucinationguard.ToolCall{
    Name: "transfer_money",
    Parameters: map[string]interface{}{
        "amount": 500,
    },
    Context: &hallucinationguard.CallContext{
        UserRole: "admin",
        UserID:   "user123",
        SessionID: "session456",
        PreviousCalls: []string{"get_balance"},
        UserPermissions: []string{"financial_access"},
        TimeOfDay: 14, // 2 PM
        Metadata: map[string]interface{}{
            "subscription_tier": "premium",
        },
    },
}

result := guard.ValidateToolCall(ctx, toolCall)
```

## Configuration

You can customize the Guard with functional options:

```go
guard := hallucinationguard.New(
    hallucinationguard.WithSchemaLoader(myCustomLoader),
    hallucinationguard.WithPolicyEngine(myCustomPolicyEngine),
)
```

## ValidationResult

The `ValidationResult` struct provides detailed information:

- `ExecutionAllowed` (bool): Whether the tool call is allowed.
- `Error` (string): Error message if validation failed.
- `PolicyAction` (string): Action taken by policy (ALLOW, REJECT, REWRITE, etc.).
- `SuggestedCorrection` (\*ToolCall): Suggestion for correction if available.
- `ToolCallID` (string): ID of the validated tool call.
- `Status` (string): Status of the validation (approved, rejected, rewritten).
- `Confidence` (float64): Confidence score for the validation decision.

## Policy Types

HallucinationGuard supports multiple policy types:

- **ALLOW**: Allow the tool call
- **REJECT**: Reject the tool call
- **REWRITE**: Auto-correct tool name to target
- **LOG**: Allow but log the call
- **CONTEXT_REJECT**: Reject based on context conditions

## Thread Safety

The Guard is safe for concurrent use.

## Extensibility

You can provide your own schema loader or policy engine by implementing the respective interfaces and passing them as options.

## More

- **Agent Scaffold:** See the [`scaffold/`](./scaffold/README.md) directory for a full agent scaffold and usage examples.
- **Web Demo:** See the [`webapp/`](./webapp/README.md) directory for a web demo and UI. Each has its own README for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Community

- [Discord](https://discord.gg/hallucinationguard) - Join our community
- [GitHub Issues](https://github.com/SafellmHub/hguard-go/issues) - Report bugs or request features
- [Contributing Guide] - Help improve HallucinationGuard. Create an issue and raise a PR!

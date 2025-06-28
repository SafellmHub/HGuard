# HallucinationGuard

[![Go Report Card](https://goreportcard.com/badge/github.com/SafellmHub/HGuard)](https://goreportcard.com/report/github.com/SafellmHub/HGuard)
[![GoDoc](https://godoc.org/github.com/SafellmHub/HGuard?status.svg)](https://godoc.org/github.com/SafellmHub/HGuard)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

HallucinationGuard is a Go SDK for validating and enforcing guardrails on LLM tool calls. It provides schema validation, policy enforcement, and extensibility for production-grade AI integrations.

## Features

- Structured validation of tool calls
- Policy-based allow/reject/rewrite
- Extensible schema and policy loading
- Thread-safe and context-aware
- Meaningful error messages for production

## Installation

```sh
go get github.com/SafellmHub/HGuard
```

## Usage Example

Add HallucinationGuard to your agent in just a few lines:

```go
import (
    "context"
    "log"
    "github.com/SafellmHub/HGuard/pkg/hallucinationguard"
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

// Usage in your agent:
func main() {
    ctx := context.Background()
    agent := NewHGuardAgent("schemas.yaml", "policies.yaml")

    // Tool call (from LLM or user)
    toolCall := hallucinationguard.ToolCall{
        Name: "weather",
        Parameters: map[string]interface{}{"city": "London"},
    }

    result := agent.ValidateToolCall(ctx, toolCall)
    if !result.ExecutionAllowed {
        log.Printf("Tool call rejected: %s (action: %s)", result.Error, result.PolicyAction)
        if result.SuggestedCorrection != nil {
            log.Printf("Did you mean: %v?", result.SuggestedCorrection)
        }
        return
    }
    // you can now proceed to execute the tool call
}
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

## Thread Safety

The Guard is safe for concurrent use.

## Extensibility

You can provide your own schema loader or policy engine by implementing the respective interfaces and passing them as options.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Community

- [Discord](https://discord.gg/hallucinationguard) - Join our community
- [GitHub Issues](https://github.com/SafellmHub/HGuard/issues) - Report bugs or request features
- [Contributing Guide]- Help improve HallucinationGuard. Create and issue and raise a PR!

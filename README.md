# HallucinationGuard

[![Go Report Card](https://goreportcard.com/badge/github.com/SafellmHub/hguard-go)](https://goreportcard.com/report/github.com/SafellmHub/hguard-go)
[![GoDoc](https://godoc.org/github.com/SafellmHub/hguard-go?status.svg)](https://godoc.org/github.com/SafellmHub/hguard-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

HallucinationGuard is a Go SDK for validating and enforcing guardrails on LLM tool calls. It provides schema validation, policy enforcement, and extensibility for production-grade AI integrations.

## Features

- Structured validation of tool calls
- Policy-based allow/reject/rewrite
- Extensible schema and policy loading
- Thread-safe and context-aware

## Installation

```sh
go get github.com/SafellmHub/hguard-go
```

## Usage Example

Add HallucinationGuard to your agent in just a few lines:

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

## More

- **Agent Scaffold:** See the [`scaffold/`](./scaffold/README.md) directory for a full agent scaffold and usage examples.
- **Web Demo:** See the [`webapp/`](./webapp/README.md) directory for a web demo and UI. Each has its own README for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Community

- [Discord](https://discord.gg/hallucinationguard) - Join our community
- [GitHub Issues](https://github.com/SafellmHub/hguard-go/issues) - Report bugs or request features
- [Contributing Guide] - Help improve HallucinationGuard. Create an issue and raise a PR!

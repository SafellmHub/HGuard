# HallucinationGuard

[![Go Report Card](https://goreportcard.com/badge/github.com/fishonamos/HGuard)](https://goreportcard.com/report/github.com/fishonamos/HGuard)
[![GoDoc](https://godoc.org/github.com/fishonamos/HGuard?status.svg)](https://godoc.org/github.com/fishonamos/HGuard)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

HallucinationGuard is a middleware system that validates and secures tool calls from LLMs. It acts as a security layer between your LLM and its tools, ensuring that only valid and authorized tool calls are executed.

## Installation

```bash
go get github.com/fishonamos/HGuard
```

## Usage

```go
package main

import (
    "github.com/fishonamos/HGuard"
)

func main() {
    // Initialize the guard
    guard := hallucinationguard.New()

    // Load your tool schemas and policies
    if err := guard.LoadSchemasFromFile("schemas.yaml"); err != nil {
        log.Fatal(err)
    }
    if err := guard.LoadPoliciesFromFile("policies.yaml"); err != nil {
        log.Fatal(err)
    }

    // Validate a tool call
    result := guard.ValidateToolCall(toolCall)
    if result.ExecutionAllowed {
        // Execute the tool
    }
}
```

## Features

- **Tool Validation**: Enforce strict schema validation for all tool calls
- **Policy Enforcement**: Define and enforce granular access policies
- **Typo Correction**: Automatically correct common tool name typos
- **Rate Limiting**: Prevent tool abuse with configurable rate limits
- **Context Awareness**: Make validation decisions based on call history
- **Hallucination Prevention**: Block non-existent or unauthorized tools


## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- [GitHub Issues](https://github.com/fishonamos/HGuard/issues)
- [Discord Community](https://discord.gg/hallucinationguard)

## Why HallucinationGuard?

LLMs can sometimes hallucinate tool calls - attempting to use tools that don't exist or aren't authorized. This can lead to:

- Security vulnerabilities
- System errors
- Unauthorized access
- Resource abuse

HallucinationGuard prevents these issues by:

1. Validating all tool calls against defined schemas
2. Enforcing access policies
3. Correcting common typos
4. Rate limiting tool usage
5. Tracking call context

## Community

- [Discord](https://discord.gg/hallucinationguard) - Join our community
- [GitHub Issues](https://github.com/fishonamos/HGuard/issues) - Report bugs or request features
- [Contributing Guide]- Help improve HallucinationGuard. Create and issue and raise a PR!

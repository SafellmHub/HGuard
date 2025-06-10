# HallucinationGuard

A middleware system for detecting and preventing hallucinated tool use in LLMs. HallucinationGuard acts as a security layer between your LLM and its tools, ensuring that only valid, authorized, and safe tool calls are executed.

## Features

- 🔍 **Tool Validation**: Validate tool calls against defined schemas
- 🛡️ **Policy Enforcement**: Control tool access with flexible policies
- ✨ **Typo Correction**: Auto-correct common tool name typos
- ⏱️ **Rate Limiting**: Prevent tool abuse with rate limiting
- 🔄 **Context Awareness**: Make decisions based on call history
- 🚫 **Hallucination Prevention**: Block non-existent or unauthorized tools

## Quick Start

```go
import "github.com/fishonamos/HGuard"

// Create a new guard instance
guard := hallucinationguard.New()

// Load your schemas and policies
guard.LoadSchemasFromFile("schemas.yaml")
guard.LoadPoliciesFromFile("policies.yaml")

// Validate a tool call
result := guard.ValidateToolCall(toolCall)
if result.ExecutionAllowed {
    // Execute the tool
}
```

For detailed usage instructions, see our [Getting Started Guide](docs/getting-started.md).

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

## Key Benefits

- **Security**: Prevent unauthorized tool access
- **Reliability**: Block invalid tool calls
- **User Experience**: Auto-correct common mistakes
- **Resource Control**: Prevent tool abuse
- **Flexibility**: Define your own schemas and policies

## Documentation

- [Getting Started Guide](docs/getting-started.md) - Quick start and basic usage
- [API Reference](docs/api-reference.md) - Detailed API documentation
- [Examples](example/) - Code examples and use cases
- [FAQ](docs/faq.md) - Common questions and answers

## Community

- [Discord](https://discord.gg/hallucinationguard) - Join our community
- [GitHub Issues](https://github.com/fishonamos/HGuard/issues) - Report bugs or request features
- [Contributing Guide](CONTRIBUTING.md) - Help improve HallucinationGuard

## License

HallucinationGuard is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by the need for secure LLM tool use
- Built with ❤️ by the HallucinationGuard team
- Thanks to all our contributors and users

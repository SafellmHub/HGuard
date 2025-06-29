# HallucinationGuard Agent Scaffold

This directory contains a reference agent implementation using the HallucinationGuard Go SDK. Use this scaffold as a starting point for building your own robust, production-ready LLM agent.

## Project Structure

```
scaffold/
  agent.go         # Agent logic
  tools.go         # Tool implementations
  config.go        # Configuration and API keys
  schemas.yaml     # Tool schemas
  policies.yaml    # Policy rules
  prompt.txt       # LLM prompt
```

## Getting Started

1. **Install dependencies:**
   ```sh
   go mod tidy
   ```
2. **Set environment variables (optional):**
   ```sh
   export ANTHROPIC_API_KEY=your_anthropic_key
   export MAVAPAY_API_KEY=your_mavapay_key
   export OPENWEATHER_API_KEY=your_openweather_key
   ```
## Configuration

- **schemas.yaml:** Defines available tools and parameters.
- **policies.yaml:** Defines which tools are allowed.
- **prompt.txt:** Instructs the LLM how to format tool calls.

## Extending the Agent

- Add new tools in `tools.go` and update `schemas.yaml` and `policies.yaml`.
- Edit `prompt.txt` to change agent behavior or supported tools.

## Web Demo

See [`../webapp/`](../webapp/README.md) for a web server and UI demo using this agent scaffold.

## License

MIT

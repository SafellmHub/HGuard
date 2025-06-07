# Example: Claude Agent with HallucinationGuard

This example demonstrates how to use HallucinationGuard as a middleware for tool call validation in a Claude-powered AI agent.

## Prerequisites

- Python 3.8+
- Claude API key (set as `CLAUDE_API_KEY`)
- HallucinationGuard server running (see below)

## 1. Start HallucinationGuard Server

Make sure HallucinationGuard is running with your desired policies (see `internal/config/policies.yaml`).

```sh
# From the project root
export HALLUCINATION_GUARD_API_KEY=your_guard_api_key
go run ./cmd/server/main.go
```

## 2. Install Python Dependencies

```sh
pip install anthropic requests
```

## 3. Set API Keys

```sh
export CLAUDE_API_KEY=your_claude_api_key
export HALLUCINATION_GUARD_API_KEY=your_guard_api_key
```

## 4. Run the Example Agent

```sh
python claude_agent.py
```

## What This Does

- Sends prompts to Claude to trigger tool calls (weather, money transfer, stock price, flight booking, typo scenario).
- Validates each tool call with HallucinationGuard via REST API.
- Demonstrates approval, rejection, and correction scenarios based on your policy config.

## Customization

- Edit `claude_agent.py` to add more tools, scenarios, or change prompts.
- Update `internal/config/policies.yaml` to change which tools are allowed, rejected, or rewritten.

---

For more details, see the main project README and research documentation in `docs/`.

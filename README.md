LLM Tool Guardrails is a lightweight middleware system for detecting and preventing hallucinated tool use in large language models (LLMs). It works by intercepting and validating function/tool calls made by the model, then filtering, rewriting, or blocking hallucinated or invalid ones.

## Why HallucinationGuard is Important

Modern LLMs (like GPT-3, Claude, etc.) can generate tool or API calls that are hallucinated, unsafe, or non-compliant. If executed blindly, these calls can cause security breaches, data loss, financial risk, or compliance violations. HallucinationGuard acts as an automated, real-time firewall for LLM tool use, ensuring:

- **Security:** Prevents LLMs from accessing or leaking sensitive data.
- **Reliability:** Stops system crashes or errors from malformed tool calls.
- **Compliance:** Helps organizations meet regulatory requirements (GDPR, SOC2, etc.).
- **Trust:** Increases user and developer confidence in AI-powered systems.

Without a tool like this, LLMs might call non-existent or dangerous APIs, leak or corrupt data, or cause financial/reputational damage. HallucinationGuard automates validation, rewriting, and blocking of tool calls. This makes LLM-powered applications safer and more reliable.

This project aims to improve the reliability and safety of tool-augmented LLMs by introducing programmatic checks before external tool execution. It supports rule-based and prompt-aware strategies and is compatible with OpenAI function calling, LangChain tools, and similar LLM frameworks.

## Getting Started

1. Install Go 1.21+
2. Copy `.env.example` to `.env` and configure as needed
3. Run the server:
   ```sh
   go run ./cmd/server/main.go
   ```

## REST API Usage

### Endpoint

`POST /api/v1/validate`

Validates one or more tool calls from an LLM, returning a decision for each.

#### Request Body

```json
{
  "tool_calls": [
    {
      "id": "call_001",
      "name": "get_weather",
      "parameters": {
        "location": "San Francisco",
        "units": "celsius"
      }
    }
  ],
  "context": {
    "user_id": "user123",
    "session_id": "sess456"
  }
}
```

#### Response Body (Success)

```json
{
  "status": "success",
  "results": [
    {
      "tool_call_id": "call_001",
      "status": "approved",
      "confidence": 1.0,
      "execution_allowed": true
    }
  ],
  "processing_time_ms": 1
}
```

#### Response Body (Rejected/Correction)

```json
{
  "status": "success",
  "results": [
    {
      "tool_call_id": "call_003",
      "status": "rejected",
      "confidence": 0.9,
      "reason": "Unknown tool name. Did you mean 'get_weather'?",
      "suggested_correction": {
        "id": "call_003",
        "name": "get_weather",
        "parameters": {
          "location": "San Francisco",
          "units": "celsius"
        },
        "context": {
          "user_id": "user123",
          "session_id": "sess456"
        }
      },
      "execution_allowed": false
    }
  ],
  "processing_time_ms": 1
}
```

### Example: Validate a Tool Call (with curl)

```sh
curl -X POST http://localhost:8080/api/v1/validate \
  -H "Content-Type: application/json" \
  -d '{
    "tool_calls": [
      {
        "id": "call_001",
        "name": "get_weather",
        "parameters": {"location": "London", "units": "celsius"}
      }
    ],
    "context": {"user_id": "user123", "session_id": "sess456"}
  }'
```

### Integration with OpenAI Function Calling (Python Example)

```python
import requests

tool_call = {
    "id": "call_001",
    "name": "get_weather",
    "parameters": {"location": "London", "units": "celsius"}
}
context = {"user_id": "user123", "session_id": "sess456"}

resp = requests.post(
    "http://localhost:8080/api/v1/validate",
    json={"tool_calls": [tool_call], "context": context}
)
result = resp.json()["results"][0]

if result["status"] == "approved":
    # Execute the tool call
    print("Tool call approved, executing...")
elif result["status"] == "rejected" and "suggested_correction" in result:
    print("Rejected, suggestion:", result["suggested_correction"])
else:
    print("Rejected:", result["reason"])
```

### Error Handling

- If the request is malformed, a 400 error is returned.
- If a tool call is rejected, the `reason` field explains why.
- If a typo is detected, a `suggested_correction` is provided.

### Advanced Usage

- Register your own tool schemas and policies in `cmd/server/main.go`.
- Extend the API or validation logic as needed for your use case.

## API Key Security

To secure your HallucinationGuard server, set an API key as an environment variable:

1. Generate a random key (example):
   ```sh
   openssl rand -hex 32
   # or
   head -c 32 /dev/urandom | base64
   ```
2. Set the key before starting the server:
   ```sh
   export HALLUCINATION_GUARD_API_KEY=your_generated_key
   go run ./cmd/server/main.go
   ```
3. All clients must include this key in the `Authorization` header:
   ```sh
   -H "Authorization: Bearer your_generated_key"
   ```

- If the key is not set, the server will log a warning and allow all requests (for development only).
- For production, always set an API key!

## Policy Management via Config

HallucinationGuard supports config-based policy management. You can define tool policies in a YAML file without changing code or recompiling.

### Edit Policies

1. Open `internal/config/policies.yaml` in your editor.
2. Add, remove, or change policies as needed. Example:
   ```yaml
   policies:
     - tool_name: get_weather
       type: ALLOW
     - tool_name: transfer_money
       type: REJECT
     - tool_name: get_whether
       type: REWRITE
   ```
3. Save the file and restart the server for changes to take effect.

### Policy Types

- `ALLOW`: Approve the tool call if valid.
- `REJECT`: Block the tool call.
- `REWRITE`: Automatically rewrite the tool call if a correction is available (e.g., typo fix).
- `LOG`: Approve the call but flag it for logging/monitoring.

> **Note:** For now, the server must be restarted after editing the policy file. (Hot-reload can be added in the future.)

## Using HallucinationGuard as a Go SDK

You can embed HallucinationGuard in your Go project to validate tool calls using config-driven schemas and policies.

### 1. Install

```
go get github.com/fishonamos/hallucination-shield/pkg/hallucinationguard
```

### 2. Example Usage

```go
package main

import (
    "fmt"
    "github.com/fishonamos/hallucination-shield/pkg/hallucinationguard"
    "github.com/fishonamos/hallucination-shield/internal/core/model"
    "time"
)

func main() {
    guard := hallucinationguard.New()
    err := guard.LoadSchemasFromFile("schemas.yaml")
    if err != nil {
        panic(err)
    }
    err = guard.LoadPoliciesFromFile("policies.yaml")
    if err != nil {
        panic(err)
    }
    toolCall := model.ToolCall{
        ID: "1",
        Name: "weather",
        Parameters: map[string]interface{}{"location": "London"},
        Context: model.CallContext{UserID: "user1"},
        Timestamp: time.Now(),
    }
    result := guard.ValidateToolCall(toolCall)
    fmt.Printf("Validation result: %+v\n", result)
}
```

### 3. Configuration Files

See `schemas.yaml` and `policies.yaml` for examples.

## License

See [LICENSE](LICENSE).

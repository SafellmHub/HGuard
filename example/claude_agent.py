import os
import requests
import anthropic

# Set your API keys
CLAUDE_API_KEY = os.getenv("CLAUDE_API_KEY")
HALLUCINATION_GUARD_API_KEY = os.getenv("HALLUCINATION_GUARD_API_KEY")

# 1. Get a tool call from Claude
def get_claude_tool_call(prompt):
    client = anthropic.Anthropic(api_key=CLAUDE_API_KEY)
    response = client.messages.create(
        model="claude-3-opus-20240229",
        max_tokens=512,
        messages=[{"role": "user", "content": prompt}],
        tools=[
            {
                "name": "get_weather",
                "description": "Get weather for a location",
                "input_schema": {
                    "type": "object",
                    "properties": {
                        "location": {"type": "string", "description": "City name"},
                        "units": {"type": "string", "enum": ["celsius", "fahrenheit"]}
                    },
                    "required": ["location"]
                }
            },
            {
                "name": "transfer_money",
                "description": "Transfer money between accounts",
                "input_schema": {
                    "type": "object",
                    "properties": {
                        "from": {"type": "string", "description": "Source account"},
                        "to": {"type": "string", "description": "Destination account"},
                        "amount": {"type": "number", "description": "Amount to transfer"},
                        "currency": {"type": "string", "enum": ["usd", "eur", "gbp"]}
                    },
                    "required": ["from", "to", "amount", "currency"]
                }
            },
            {
                "name": "get_stock_price",
                "description": "Get the current price of a stock symbol",
                "input_schema": {
                    "type": "object",
                    "properties": {
                        "symbol": {"type": "string", "description": "Stock symbol (e.g., AAPL)"}
                    },
                    "required": ["symbol"]
                }
            },
            {
                "name": "book_flight",
                "description": "Book a flight between two cities",
                "input_schema": {
                    "type": "object",
                    "properties": {
                        "from": {"type": "string", "description": "Departure city"},
                        "to": {"type": "string", "description": "Destination city"},
                        "date": {"type": "string", "description": "Flight date (YYYY-MM-DD)"}
                    },
                    "required": ["from", "to", "date"]
                }
            }
        ]
    )
    # Find the ToolUseBlock in the content
    tool_call = None
    for block in response.content:
        if getattr(block, "type", None) == "tool_use":
            tool_call = {
                "id": getattr(block, "id", None),
                "name": getattr(block, "name", None),
                "parameters": getattr(block, "input", None)
            }
            break
    return tool_call

# 2. Validate the tool call with HallucinationGuard
def validate_tool_call(tool_call, user_id="test_user", session_id="sess1"):
    headers = {"Content-Type": "application/json"}
    if HALLUCINATION_GUARD_API_KEY:
        headers["Authorization"] = f"Bearer {HALLUCINATION_GUARD_API_KEY}"
    payload = {
        "tool_calls": [tool_call],
        "context": {"user_id": user_id, "session_id": session_id}
    }
    resp = requests.post(
        "http://localhost:8080/api/v1/validate",
        json=payload,
        headers=headers
    )
    return resp.json()["results"][0]

# 3. Example tool execution (stub)
def execute_tool(tool_call):
    if tool_call["name"] == "get_weather":
        location = tool_call["parameters"]["location"]
        units = tool_call["parameters"].get("units", "celsius")
        return f"Weather in {location}: 20° {units}"
    elif tool_call["name"] == "transfer_money":
        from_acct = tool_call["parameters"]["from"]
        to_acct = tool_call["parameters"]["to"]
        amount = tool_call["parameters"]["amount"]
        currency = tool_call["parameters"].get("currency", "usd")
        return f"Transferred {amount} {currency} from {from_acct} to {to_acct} (simulated)"
    elif tool_call["name"] == "get_stock_price":
        symbol = tool_call["parameters"]["symbol"]
        return f"Stock price for {symbol}: $123.45 (simulated)"
    elif tool_call["name"] == "book_flight":
        from_city = tool_call["parameters"]["from"]
        to_city = tool_call["parameters"]["to"]
        date = tool_call["parameters"]["date"]
        return f"Flight booked from {from_city} to {to_city} on {date} (simulated)"
    return "Unknown tool"

# 4. Main agent loop
def agent(prompt):
    tool_call = get_claude_tool_call(prompt)
    print("Claude tool call:", tool_call)
    if not tool_call:
        print("No tool call found in Claude's response.")
        return
    validation = validate_tool_call(tool_call)
    print("Validation result:", validation)
    if validation["status"] == "approved":
        result = execute_tool(tool_call)
        print("Tool execution result:", result)
    elif validation.get("suggested_correction"):
        print("Correction suggested:", validation["suggested_correction"])
    else:
        print("Tool call rejected:", validation["reason"])

if __name__ == "__main__":
    print("--- Weather Query (should be approved) ---")
    agent("What's the weather in Paris in celsius?")
    print("\n--- Money Transfer (should be rejected) ---")
    agent("Transfer $500 from savings to checking in usd.")
    print("\n--- Stock Price Query (should be approved if allowed by policy) ---")
    agent("What is the current price of AAPL stock?")
    print("\n--- Flight Booking (should be approved if allowed by policy) ---")
    agent("Book a flight from New York to London on 2024-07-01.")
    print("\n--- Typo Tool Name (should be corrected or rejected) ---")
    agent("What is the whether in Berlin?") 
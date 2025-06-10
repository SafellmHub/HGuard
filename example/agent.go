package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/fishonamos/HGuard/pkg/hallucinationguard"
)

// ToolCall represents a tool call request from the LLM
type ToolCall struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
}

type LLMResponse struct {
	ToolCall ToolCall `json:"tool_call"`
}

type AtomaResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func callLLM(prompt string) (ToolCall, error) {
	apiKey := os.Getenv("ATOMA_API_KEY")
	// Updated system prompt with stricter parameter validation
	systemPrompt := `You are a tool-using assistant. You must respond with a valid JSON object for tool calls.
Available tools and their parameter constraints:
- weather: Get weather for a location
  parameters:
    location: string, 2-100 chars, only letters/spaces/commas
    unit: enum ["C", "F"], default "C"
    format: enum ["text", "json"], default "text"

- stock_price: Get stock price
  parameters:
    symbol: string, 1-5 uppercase letters only
    date: string, YYYY-MM-DD format, optional

- math_calculation: Calculate math expressions
  parameters:
    expression: string, only numbers and +-*/() operators
    precision: integer 0-10, default 2

- transfer_money: Transfer money (restricted)
  parameters:
    from_account: string, 10 alphanumeric chars
    to_account: string, 10 alphanumeric chars
    amount: number, 0.01-1000000
    currency: enum ["USD", "EUR", "GBP"], default "USD"

Example response format:
{
  "id": "call_001",
  "name": "weather",
  "parameters": {
    "location": "London",
    "unit": "C"
  }
}

IMPORTANT:
1. Only use the tools listed above
2. Strictly follow parameter constraints
3. For invalid requests, return a valid tool call with appropriate parameters
4. Never include explanations, only the JSON object
5. Always validate parameters before returning`

	// Add exponential backoff for rate limiting
	maxRetries := 3
	baseDelay := 1 * time.Second
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			fmt.Printf("Rate limited, retrying in %v...\n", delay)
			time.Sleep(delay)
		}

		reqBody := map[string]interface{}{
			"model": "deepseek-ai/DeepSeek-V3-0324",
			"messages": []map[string]string{
				{"role": "system", "content": systemPrompt},
				{"role": "user", "content": prompt},
			},
			"max_tokens":         256,
			"temperature":        0.3,
			"top_p":              0.7,
			"top_k":              50,
			"repetition_penalty": 1.0,
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "https://api.atoma.network/v1/chat/completions", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()
		data, _ := ioutil.ReadAll(resp.Body)

		// Check for rate limit error
		var errorResp struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(data, &errorResp); err == nil && errorResp.Error.Code == "TOO_MANY_REQUESTS" {
			lastErr = fmt.Errorf("rate limited: %s", errorResp.Error.Message)
			continue
		}

		fmt.Println("Raw LLM response:", string(data))

		var atomaResp AtomaResponse
		if err := json.Unmarshal(data, &atomaResp); err != nil {
			lastErr = err
			continue
		}
		if len(atomaResp.Choices) == 0 {
			lastErr = fmt.Errorf("no choices in response")
			continue
		}

		toolCallJson := atomaResp.Choices[0].Message.Content
		// Remove Markdown code block if present
		toolCallJson = strings.TrimSpace(toolCallJson)
		if strings.HasPrefix(toolCallJson, "```") {
			first := strings.Index(toolCallJson, "```")
			last := strings.LastIndex(toolCallJson, "```")
			if first != -1 && last != -1 && last > first {
				toolCallJson = toolCallJson[first+3 : last]
				toolCallJson = strings.TrimPrefix(toolCallJson, "json")
				toolCallJson = strings.TrimSpace(toolCallJson)
			}
		}

		var toolCall ToolCall
		if err := json.Unmarshal([]byte(toolCallJson), &toolCall); err != nil {
			lastErr = fmt.Errorf("failed to parse tool call JSON: %v\nContent: %s", err, toolCallJson)
			continue
		}

		// Validate parameters before returning
		if err := validateToolCallParameters(&toolCall); err != nil {
			lastErr = err
			continue
		}

		return toolCall, nil
	}

	return ToolCall{}, fmt.Errorf("failed after %d retries: %v", maxRetries, lastErr)
}

// Add parameter validation
func validateToolCallParameters(tc *ToolCall) error {
	switch tc.Name {
	case "weather":
		if loc, ok := tc.Parameters["location"].(string); ok {
			if len(loc) < 2 || len(loc) > 100 {
				return fmt.Errorf("location must be 2-100 chars")
			}
			if !regexp.MustCompile(`^[a-zA-Z\s,]+$`).MatchString(loc) {
				return fmt.Errorf("location can only contain letters, spaces, and commas")
			}
		}
		if unit, ok := tc.Parameters["unit"].(string); ok && unit != "" {
			if unit != "C" && unit != "F" {
				return fmt.Errorf("unit must be C or F")
			}
		}
	case "stock_price":
		if sym, ok := tc.Parameters["symbol"].(string); ok {
			if !regexp.MustCompile(`^[A-Z]{1,5}$`).MatchString(sym) {
				return fmt.Errorf("symbol must be 1-5 uppercase letters")
			}
		}
		if date, ok := tc.Parameters["date"].(string); ok && date != "" {
			if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(date) {
				return fmt.Errorf("date must be YYYY-MM-DD")
			}
		}
	case "math_calculation":
		if expr, ok := tc.Parameters["expression"].(string); ok {
			if !regexp.MustCompile(`^[0-9\s\+\-\*\/\(\)\.]+$`).MatchString(expr) {
				return fmt.Errorf("expression can only contain numbers and basic operators")
			}
		}
		if prec, ok := tc.Parameters["precision"].(float64); ok {
			if prec < 0 || prec > 10 || prec != float64(int(prec)) {
				return fmt.Errorf("precision must be integer 0-10")
			}
		}
	case "transfer_money":
		if from, ok := tc.Parameters["from_account"].(string); ok {
			if !regexp.MustCompile(`^[A-Z0-9]{10}$`).MatchString(from) {
				return fmt.Errorf("from_account must be 10 alphanumeric chars")
			}
		}
		if to, ok := tc.Parameters["to_account"].(string); ok {
			if !regexp.MustCompile(`^[A-Z0-9]{10}$`).MatchString(to) {
				return fmt.Errorf("to_account must be 10 alphanumeric chars")
			}
		}
		if amt, ok := tc.Parameters["amount"].(float64); ok {
			if amt < 0.01 || amt > 1000000 {
				return fmt.Errorf("amount must be 0.01-1000000")
			}
		}
		if curr, ok := tc.Parameters["currency"].(string); ok && curr != "" {
			if curr != "USD" && curr != "EUR" && curr != "GBP" {
				return fmt.Errorf("currency must be USD, EUR, or GBP")
			}
		}
	}
	return nil
}

// Add a test function to verify LLM responses
func testLLMResponses() {
	testPrompts := []string{
		"What's the weather in London?",
		"Get the stock price of GOOGL",
		"Calculate 2+2",
		"Transfer $100 to hacker123", // Should be rejected by policy
		"Get the wether in Paris",    // Should test typo correction
	}

	fmt.Println("\n--- Testing Atoma LLM Responses ---")
	for _, prompt := range testPrompts {
		fmt.Printf("\nTesting prompt: %s\n", prompt)
		toolCall, err := callLLM(prompt)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		fmt.Printf("LLM Response: %+v\n", toolCall)

		// Pretty print the parameters
		if len(toolCall.Parameters) > 0 {
			params, _ := json.MarshalIndent(toolCall.Parameters, "", "  ")
			fmt.Printf("Parameters: %s\n", params)
		}
	}
}

// Add a struct to hold policy state
type policyEnforcer struct {
	guard            *hallucinationguard.Guard
	weatherCallCount int
	lastWeatherCall  time.Time
	previousCalls    []string
}

func newPolicyEnforcer(guard *hallucinationguard.Guard) *policyEnforcer {
	return &policyEnforcer{
		guard:         guard,
		previousCalls: make([]string, 0),
	}
}

func (pe *policyEnforcer) validateAndExecute(toolCall ToolCall) {
	// Convert to package's ToolCall type
	result := pe.guard.ValidateToolCall(hallucinationguard.ToolCall{
		Name:       toolCall.Name,
		Parameters: toolCall.Parameters,
	})

	if !result.ExecutionAllowed {
		fmt.Printf("Tool call rejected: %s\n", result.Error)
		return
	}

	// Handle policy actions
	switch result.PolicyAction {
	case hallucinationguard.PolicyActionALLOW:
		fmt.Printf("Executing tool call: %s\n", toolCall.Name)
		// ... execute the tool ...
	case hallucinationguard.PolicyActionREJECT:
		fmt.Printf("Tool call rejected: %s\n", result.Error)
	case hallucinationguard.PolicyActionREWRITE:
		if result.SuggestedCorrection != nil {
			fmt.Printf("Correcting tool call: %s -> %s\n",
				toolCall.Name, result.SuggestedCorrection.Name)
			pe.validateAndExecute(ToolCall{
				Name:       result.SuggestedCorrection.Name,
				Parameters: result.SuggestedCorrection.Parameters,
			})
		}
	case hallucinationguard.PolicyActionRATE_LIMIT:
		if pe.weatherCallCount <= 10 {
			fmt.Printf("Executing rate-limited tool call: %s (call %d/10)\n",
				toolCall.Name, pe.weatherCallCount)
			// ... execute the tool ...
		} else {
			fmt.Printf("Rate limit exceeded for %s\n", toolCall.Name)
		}
	case hallucinationguard.PolicyActionCONTEXT_REJECT:
		fmt.Printf("Context-based rejection: %s\n", result.Error)
	default:
		fmt.Printf("Unknown policy action: %s\n", result.PolicyAction)
	}
}

func (pe *policyEnforcer) testRateLimiting() {
	fmt.Println("\nTesting rate limiting...")

	// Test weather tool rate limiting
	for i := 0; i < 12; i++ {
		call := ToolCall{
			Name: "weather",
			Parameters: map[string]interface{}{
				"location": "London",
				"unit":     "C",
			},
		}
		pe.validateAndExecute(call)
		time.Sleep(100 * time.Millisecond)
	}
}

func (pe *policyEnforcer) testContextRejection() {
	fmt.Println("\nTesting context-based rejection...")

	// Test transfer money context rejection
	call := ToolCall{
		Name: "transfer_money",
		Parameters: map[string]interface{}{
			"from":     "acc123",
			"to":       "acc456",
			"amount":   100.0,
			"currency": "USD",
		},
	}

	// First call should be allowed
	pe.validateAndExecute(call)

	// Second call should be rejected
	pe.validateAndExecute(call)
}

func (pe *policyEnforcer) testTypoCorrections() {
	fmt.Println("\nTesting typo corrections...")

	testCases := []struct {
		input    string
		expected string
	}{
		{"wether", "weather"},
		{"stok_price", "stock_price"},
		{"math_calc", "math_calculation"},
	}

	for _, tc := range testCases {
		call := ToolCall{
			Name: tc.input,
			Parameters: map[string]interface{}{
				"location": "London",
			},
		}

		result := pe.guard.ValidateToolCall(hallucinationguard.ToolCall{
			Name:       call.Name,
			Parameters: call.Parameters,
		})

		if result.PolicyAction == hallucinationguard.PolicyActionREWRITE &&
			result.SuggestedCorrection != nil {
			fmt.Printf("Typo correction: %s -> %s\n",
				call.Name, result.SuggestedCorrection.Name)
			if result.SuggestedCorrection.Name != tc.expected {
				fmt.Printf("Warning: Expected correction to %s but got %s\n",
					tc.expected, result.SuggestedCorrection.Name)
			}
		} else {
			fmt.Printf("No typo correction applied (got action: %s)\n",
				result.PolicyAction)
		}
	}
}

func main() {
	// Add LLM response testing at the start
	testLLMResponses()

	// Initialize the guard
	guard := hallucinationguard.New()

	// Load schemas and policies
	if err := guard.LoadSchemasFromFile("schemas.yaml"); err != nil {
		log.Fatalf("Failed to load schemas: %v", err)
	}
	if err := guard.LoadPoliciesFromFile("policies.yaml"); err != nil {
		log.Fatalf("Failed to load policies: %v", err)
	}

	// Create policy enforcer
	enforcer := newPolicyEnforcer(guard)

	// Run tests
	enforcer.testRateLimiting()
	enforcer.testContextRejection()
	enforcer.testTypoCorrections()
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/fishonamos/HGuard/internal/core/model"
	"github.com/fishonamos/HGuard/pkg/hallucinationguard"
)

type LLMResponse struct {
	ToolCall struct {
		ID         string                 `json:"id"`
		Name       string                 `json:"name"`
		Parameters map[string]interface{} `json:"parameters"`
	} `json:"tool_call"`
}

type AtomaResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func callLLM(prompt string) (model.ToolCall, error) {
	apiKey := "ecXtlnDvomQXCPDIv6ch3yUs2pLYEf"
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

		var toolCall model.ToolCall
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

	return model.ToolCall{}, fmt.Errorf("failed after %d retries: %v", maxRetries, lastErr)
}

// Add parameter validation
func validateToolCallParameters(tc *model.ToolCall) error {
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
		guard:            guard,
		weatherCallCount: 0,
		lastWeatherCall:  time.Now(),
		previousCalls:    make([]string, 0, 10),
	}
}

func (pe *policyEnforcer) validateAndExecute(toolCall model.ToolCall) {
	result := pe.guard.ValidateToolCall(toolCall)
	fmt.Printf("Validation result: %+v\n", result)

	// Handle rate limiting
	if toolCall.Name == "weather" {
		now := time.Now()
		if now.Sub(pe.lastWeatherCall) > 60*time.Second {
			pe.weatherCallCount = 0
			pe.lastWeatherCall = now
		}
		pe.weatherCallCount++
		if pe.weatherCallCount > 10 {
			fmt.Printf("Rate limit exceeded: %d calls in 60s\n", pe.weatherCallCount)
			return
		}
	}

	// Handle context-based policies
	if toolCall.Name == "transfer_money" {
		for _, prevCall := range pe.previousCalls {
			if prevCall == "transfer_money" {
				fmt.Printf("Context-based rejection: Multiple transfer attempts detected\n")
				return
			}
		}
	}

	// Track previous calls
	pe.previousCalls = append(pe.previousCalls, toolCall.Name)
	if len(pe.previousCalls) > 10 {
		pe.previousCalls = pe.previousCalls[1:]
	}

	// Handle policy actions
	switch string(result.PolicyAction) {
	case string(model.PolicyActionALLOW):
		if result.ExecutionAllowed {
			fmt.Printf("Executing tool call: %s\n", toolCall.Name)
			// ... execute the tool ...
		}
	case string(model.PolicyActionREJECT):
		fmt.Printf("Tool call rejected: %s\n", result.Reason)
		return
	case string(model.PolicyActionREWRITE):
		if result.SuggestedCorrection != nil {
			fmt.Printf("Correcting tool call: %s -> %s\n", toolCall.Name, result.SuggestedCorrection.Name)
			pe.validateAndExecute(*result.SuggestedCorrection)
			return
		}
	case string(model.PolicyActionRATE_LIMIT):
		if pe.weatherCallCount <= 10 {
			fmt.Printf("Executing rate-limited tool call: %s (call %d/10)\n",
				toolCall.Name, pe.weatherCallCount)
			// ... execute the tool ...
		} else {
			fmt.Printf("Rate limit exceeded for %s\n", toolCall.Name)
			return
		}
	case string(model.PolicyActionCONTEXT_REJECT):
		fmt.Printf("Context-based rejection: %s\n", result.Reason)
		return
	default:
		fmt.Printf("Unknown policy action: %s\n", result.PolicyAction)
		return
	}
}

// Add a method to test rate limiting
func (pe *policyEnforcer) testRateLimiting() {
	fmt.Println("\n--- Testing Rate Limiting ---")
	for i := 0; i < 12; i++ {
		fmt.Printf("\nRate limit test %d:\n", i+1)
		if toolCall, err := callLLM("What's the weather in London?"); err != nil {
			fmt.Printf("LLM error: %v\n", err)
		} else {
			pe.validateAndExecute(toolCall)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// Add a method to test context rejection
func (pe *policyEnforcer) testContextRejection() {
	fmt.Println("\n--- Testing Context Rejection ---")
	prompts := []string{
		"Transfer $100 to hacker123",
		"Transfer $200 to hacker123",
	}
	for _, prompt := range prompts {
		fmt.Printf("\nTesting: %s\n", prompt)
		if toolCall, err := callLLM(prompt); err != nil {
			fmt.Printf("LLM error: %v\n", err)
		} else {
			pe.validateAndExecute(toolCall)
		}
	}
}

// Add a method to test typo corrections
func (pe *policyEnforcer) testTypoCorrections() {
	fmt.Println("\n--- Testing Typo Corrections ---")
	testCases := []struct {
		prompt   string
		expected string
	}{
		{"Get the wether in Paris", "weather"},
		{"Get the wether_forcast for Tokyo", "weather_forecast"},
		{"Get the stock price of AAPL", "stock_price"},
		{"Get the wether@2.0 for Tokyo", "weather"},
	}

	for _, tc := range testCases {
		fmt.Printf("\nTesting: %s\n", tc.prompt)
		if toolCall, err := callLLM(tc.prompt); err != nil {
			fmt.Printf("LLM error: %v\n", err)
		} else {
			result := pe.guard.ValidateToolCall(toolCall)
			if string(result.PolicyAction) == string(model.PolicyActionREWRITE) && result.SuggestedCorrection != nil {
				fmt.Printf("Typo correction: %s -> %s\n", toolCall.Name, result.SuggestedCorrection.Name)
				if result.SuggestedCorrection.Name != tc.expected {
					fmt.Printf("Warning: Expected correction to %s but got %s\n",
						tc.expected, result.SuggestedCorrection.Name)
				}
				pe.validateAndExecute(*result.SuggestedCorrection)
			} else {
				fmt.Printf("No typo correction applied (got action: %s)\n", result.PolicyAction)
			}
		}
	}
}

func main() {
	// Add LLM response testing at the start
	testLLMResponses()

	guard := hallucinationguard.New()
	if err := guard.LoadSchemasFromFile("schemas.yaml"); err != nil {
		panic(err)
	}
	if err := guard.LoadPoliciesFromFile("policies.yaml"); err != nil {
		panic(err)
	}

	enforcer := newPolicyEnforcer(guard)

	// Run specific tests
	enforcer.testRateLimiting()
	enforcer.testContextRejection()
	enforcer.testTypoCorrections()

	// --- More scenarios ---
	fmt.Println("\n--- More Scenarios ---")

	stockCall := model.ToolCall{
		ID:         "call_002",
		Name:       "stock_price",
		Parameters: map[string]interface{}{"symbol": "AAPL", "date": "2024-06-08"},
		Context:    model.CallContext{UserID: "user123"},
		Timestamp:  time.Now(),
	}
	stockResult := guard.ValidateToolCall(stockCall)
	fmt.Printf("Stock price tool call result: %+v\n", stockResult)

	flightCall := model.ToolCall{
		ID:         "call_003",
		Name:       "flight_booking",
		Parameters: map[string]interface{}{"origin": "NYC", "destination": "LON", "date": "2024-07-01", "passengers": 2},
		Context:    model.CallContext{UserID: "user123"},
		Timestamp:  time.Now(),
	}
	flightResult := guard.ValidateToolCall(flightCall)
	fmt.Printf("Flight booking tool call result: %+v\n", flightResult)

	mathCall := model.ToolCall{
		ID:         "call_004",
		Name:       "math_calculation",
		Parameters: map[string]interface{}{"expression": "2+2*2"},
		Context:    model.CallContext{UserID: "user123"},
		Timestamp:  time.Now(),
	}
	mathResult := guard.ValidateToolCall(mathCall)
	fmt.Printf("Math calculation tool call result: %+v\n", mathResult)

	// Test REWRITE scenario (typo)
	wetherCall := model.ToolCall{
		ID:         "call_005",
		Name:       "wether",
		Parameters: map[string]interface{}{"location": "London", "unit": "C"},
		Context:    model.CallContext{UserID: "user123"},
		Timestamp:  time.Now(),
	}
	wetherResult := guard.ValidateToolCall(wetherCall)
	fmt.Printf("Typo tool call (wether) result: %+v\n", wetherResult)
	if wetherResult.SuggestedCorrection != nil {
		fmt.Printf("Suggested correction: %+v\n", *wetherResult.SuggestedCorrection)
	}

	// --- Hallucination Test Scenarios with Real LLM ---
	fmt.Println("\n--- Hallucination Test Scenarios with Real LLM ---")

	// 1. Test non-existent tool
	fmt.Println("\n1. Testing non-existent tool (time_travel)...")
	toolCall, err := callLLM("I want to travel back in time to January 1st, 2024")
	if err != nil {
		fmt.Printf("LLM error: %v\n", err)
	} else {
		result := guard.ValidateToolCall(toolCall)
		fmt.Printf("Time travel tool call result: %+v\n", result)
	}

	// 2. Test malformed parameters
	fmt.Println("\n2. Testing malformed parameters...")
	toolCall, err = callLLM("Get the weather for location 123 in units [C, F]")
	if err != nil {
		fmt.Printf("LLM error: %v\n", err)
	} else {
		result := guard.ValidateToolCall(toolCall)
		fmt.Printf("Malformed parameters result: %+v\n", result)
	}

	// 3. Test concurrent tool calls with real LLM
	fmt.Println("\n3. Testing concurrent tool calls with real LLM...")
	prompts := []string{
		"What's the weather in Paris?",
		"Get me the stock price of Google for March 15th, 2024",
		"Transfer $1000 to hacker123", // Should be rejected by policy
	}

	results := make(chan struct {
		prompt string
		result model.ValidationResult
		err    error
	}, len(prompts))

	for _, prompt := range prompts {
		go func(p string) {
			tc, err := callLLM(p)
			if err != nil {
				results <- struct {
					prompt string
					result model.ValidationResult
					err    error
				}{p, model.ValidationResult{}, err}
				return
			}
			result := guard.ValidateToolCall(tc)
			results <- struct {
				prompt string
				result model.ValidationResult
				err    error
			}{p, result, nil}
		}(prompt)
	}

	// Collect results
	for i := 0; i < len(prompts); i++ {
		r := <-results
		if r.err != nil {
			fmt.Printf("Error with prompt '%s': %v\n", r.prompt, r.err)
		} else {
			fmt.Printf("Result for '%s': %+v\n", r.prompt, r.result)
		}
	}

	// 4. Test complex hallucination scenarios
	fmt.Println("\n4. Testing complex hallucination scenarios...")

	// 4.1 Test special characters in tool name
	fmt.Println("\n4.1 Testing special characters in tool name...")
	toolCall, err = callLLM("Get the weather@2.0 for Tokyo")
	if err != nil {
		fmt.Printf("LLM error: %v\n", err)
	} else {
		result := guard.ValidateToolCall(toolCall)
		fmt.Printf("Special character tool name result: %+v\n", result)
	}

	// 4.2 Test nested parameter hallucination
	fmt.Println("\n4.2 Testing nested parameter hallucination...")
	toolCall, err = callLLM("Get the weather for Berlin, Germany in Kelvin")
	if err != nil {
		fmt.Printf("LLM error: %v\n", err)
	} else {
		result := guard.ValidateToolCall(toolCall)
		fmt.Printf("Nested parameter result: %+v\n", result)
	}

	// 4.3 Test multiple typos
	fmt.Println("\n4.3 Testing multiple typos...")
	toolCall, err = callLLM("Get the wether forcast for London in unt C")
	if err != nil {
		fmt.Printf("LLM error: %v\n", err)
	} else {
		result := guard.ValidateToolCall(toolCall)
		fmt.Printf("Multiple typos result: %+v\n", result)
	}

	// 4.4 Test context-based validation
	fmt.Println("\n4.4 Testing context-based validation...")
	toolCall, err = callLLM("After transferring $1,000,000, what's the weather in London?")
	if err != nil {
		fmt.Printf("LLM error: %v\n", err)
	} else {
		result := guard.ValidateToolCall(toolCall)
		fmt.Printf("Context-based validation result: %+v\n", result)
	}

	// 4.5 Test rate limiting
	fmt.Println("\n4.5 Testing rate limiting...")
	for i := 0; i < 3; i++ {
		toolCall, err = callLLM("What's the weather in London?")
		if err != nil {
			fmt.Printf("LLM error: %v\n", err)
		} else {
			result := guard.ValidateToolCall(toolCall)
			fmt.Printf("Rate limit test %d result: %+v\n", i+1, result)
		}
		time.Sleep(1 * time.Second) // Respect API rate limits
	}

	// 5. Test parameter validation rules
	fmt.Println("\n5. Testing parameter validation rules...")
	validationTests := []string{
		// Pattern validation
		"Get weather for London123",                   // Should fail: numbers in location
		"Get weather for L",                           // Should fail: too short
		"Get weather for " + strings.Repeat("A", 101), // Should fail: too long

		// Enum validation
		"Get weather in London in Kelvin", // Should fail: K not in enum

		// Format validation
		"Get stock price for GOOGLE",              // Should fail: too long for symbol
		"Get stock price for googl",               // Should fail: lowercase not allowed
		"Get stock price for GOOGL on 2024-13-45", // Should fail: invalid date
	}

	for _, prompt := range validationTests {
		fmt.Printf("\nTesting: %s\n", prompt)
		toolCall, err = callLLM(prompt)
		if err != nil {
			fmt.Printf("LLM error: %v\n", err)
			continue
		}
		result := guard.ValidateToolCall(toolCall)
		fmt.Printf("Validation result: %+v\n", result)
	}

	// 6. Test multiple typo corrections
	fmt.Println("\n6. Testing multiple typo corrections...")
	typoTests := []string{
		"Get the wether in London",         // Simple typo
		"Get the wether_forcast for Paris", // Multiple typos
		"Get the stock price of AAPL",      // Shorthand correction
		"Get the wether@2.0 for Tokyo",     // Special characters
	}

	for _, prompt := range typoTests {
		fmt.Printf("\nTesting typos: %s\n", prompt)
		toolCall, err = callLLM(prompt)
		if err != nil {
			fmt.Printf("LLM error: %v\n", err)
			continue
		}
		result := guard.ValidateToolCall(toolCall)
		fmt.Printf("Typo correction result: %+v\n", result)
		if result.SuggestedCorrection != nil {
			fmt.Printf("Suggested correction: %+v\n", *result.SuggestedCorrection)
		}
	}

	// 7. Test edge cases
	fmt.Println("\n7. Testing edge cases...")
	edgeCases := []string{
		"Get weather for ", // Empty location
		"Get weather for London in C with format XML", // Invalid format
		"Calculate 2+2 with precision 2.5",            // Non-integer precision
		"Get stock price for AAPL on tomorrow",        // Invalid date format
		"Transfer $100 from ABC1234567 to ABC1234567", // Same account
	}

	for _, prompt := range edgeCases {
		fmt.Printf("\nTesting edge case: %s\n", prompt)
		toolCall, err = callLLM(prompt)
		if err != nil {
			fmt.Printf("LLM error: %v\n", err)
			continue
		}
		result := guard.ValidateToolCall(toolCall)
		fmt.Printf("Edge case result: %+v\n", result)
	}
}

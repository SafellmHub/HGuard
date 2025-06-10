package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	apiKey := "ecXtlnDvomQXCPDIv6ch3yUs2pLYEf" // Your Atoma API key
	systemPrompt := `You are a tool-using assistant. When asked a question, respond ONLY with a JSON object describing the tool call you would make, e.g.: {"id": "call_001", "name": "weather", "parameters": {"location": "London", "unit": "C"}}.`

	reqBody := map[string]interface{}{
		"model": "deepseek-ai/DeepSeek-V3-0324",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
		"max_tokens":         128,
		"temperature":        0.7,
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
		return model.ToolCall{}, err
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)

	fmt.Println("Raw LLM response:", string(data))

	var atomaResp AtomaResponse
	if err := json.Unmarshal(data, &atomaResp); err != nil {
		return model.ToolCall{}, err
	}
	if len(atomaResp.Choices) == 0 {
		return model.ToolCall{}, fmt.Errorf("no choices in response")
	}
	toolCallJson := atomaResp.Choices[0].Message.Content
	// Remove Markdown code block if present
	toolCallJson = strings.TrimSpace(toolCallJson)
	if strings.HasPrefix(toolCallJson, "```") {
		first := strings.Index(toolCallJson, "```")
		last := strings.LastIndex(toolCallJson, "```")
		if first != -1 && last != -1 && last > first {
			toolCallJson = toolCallJson[first+3 : last]
			// Remove 'json' language tag if present
			toolCallJson = strings.TrimPrefix(toolCallJson, "json")
			toolCallJson = strings.TrimSpace(toolCallJson)
		}
	}
	var toolCall model.ToolCall
	if err := json.Unmarshal([]byte(toolCallJson), &toolCall); err != nil {
		return model.ToolCall{}, fmt.Errorf("failed to parse tool call JSON from LLM content: %v\nContent: %s", err, toolCallJson)
	}
	return toolCall, nil
}

func main() {
	guard := hallucinationguard.New()
	if err := guard.LoadSchemasFromFile("schemas.yaml"); err != nil {
		panic(err)
	}
	if err := guard.LoadPoliciesFromFile("policies.yaml"); err != nil {
		panic(err)
	}

	// Simulate a user prompt
	prompt := "What's the weather in London in Celsius?"

	// Get tool call from LLM
	toolCall, err := callLLM(prompt)
	if err != nil {
		fmt.Println("LLM error:", err)
		return
	}
	fmt.Printf("Tool call from LLM: %+v\n", toolCall)

	// Validate tool call
	result := guard.ValidateToolCall(toolCall)
	fmt.Printf("Tool call validation result: %+v\n", result)

	if result.ExecutionAllowed {
		fmt.Println("Executing tool call:", toolCall.Name)
		// ...execute the tool...
	} else {
		fmt.Println("Tool call rejected:", result.Reason)
		if result.SuggestedCorrection != nil {
			fmt.Printf("Suggested correction: %+v\n", result.SuggestedCorrection)
		}
	}

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
}

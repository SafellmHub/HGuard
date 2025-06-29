package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/SafellmHub/hguard-go/scaffold"
)

type PromptRequest struct {
	Prompt string `json:"prompt"`
}

func main() {
	config := scaffold.LoadConfig()
	promptBytes, _ := ioutil.ReadFile("example/prompt.txt")
	systemPrompt := string(promptBytes)

	hguardAgent := scaffold.NewHGuardAgent("example/schemas.yaml", "example/policies.yaml")

	http.HandleFunc("/prompt", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req PromptRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "invalid request"}`))
			return
		}
		ctx := r.Context()
		fullPrompt := fmt.Sprintf("%s\nUser: %s", systemPrompt, req.Prompt)
		response, err := scaffold.CallAnthropic(ctx, config.AnthropicAPIKey, fullPrompt)
		fmt.Println("Claude raw response:", response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error": "Anthropic API error: %v"}`, err)))
			return
		}
		var toolCall scaffold.ToolCallResponse
		if err := json.Unmarshal([]byte(response), &toolCall); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf(`{"error": "Failed to parse tool call: %v", "raw": %q}`, err, response)))
			return
		}
		result := hguardAgent.ValidateToolCall(ctx, toolCall)
		fmt.Printf("ValidationResult: allowed=%v, error=%v, confidence=%.2f\n", result.ExecutionAllowed, result.Error, result.Confidence) // Debug print
		if !result.ExecutionAllowed {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"answer":       "You are not allowed to use this feature.",
				"confidence":   result.Confidence,
				"raw_response": response,
			})
			return
		}
		answer, err := hguardAgent.ExecuteTool(ctx, toolCall)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"answer":       fmt.Sprintf("Tool error: %v", err),
				"confidence":   result.Confidence,
				"raw_response": response,
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"answer":       answer,
			"confidence":   result.Confidence,
			"raw_response": response,
		})
	})

	// Serve the frontend
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		index, _ := ioutil.ReadFile("example/webdemo/index.html")
		w.Write(index)
	})

	fmt.Println("Server running on http://localhost:8080 ...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

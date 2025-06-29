package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/SafellmHub/hguard-go/pkg/hallucinationguard"
)

const (
	anthropicAPIURL   = "https://api.anthropic.com/v1/messages"
	openWeatherAPIURL = "https://api.openweathermap.org/data/2.5/weather"
	mavapayQuoteURL   = "https://staging.api.mavapay.co/api/v1/quote"
	mavapayPriceURL   = "https://staging.api.mavapay.co/api/v1/price"
)

// ToolCallResponse is used to parse the Claude response
// expecting a JSON object with name and parameters fields
// Example: {"name": "search", "parameters": {"query": "golang"}}
type ToolCallResponse struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
}

type HGuardAgent struct {
	guard *hallucinationguard.Guard
}

func NewHGuardAgent(schemaPath, policyPath string) *HGuardAgent {
	ctx := context.Background()
	guard := hallucinationguard.New()
	if err := guard.LoadSchemasFromFile(ctx, schemaPath); err != nil {
		log.Fatalf("Schema load error: %v", err)
	}
	if err := guard.LoadPoliciesFromFile(ctx, policyPath); err != nil {
		log.Fatalf("Policy load error: %v", err)
	}
	return &HGuardAgent{guard: guard}
}

func (a *HGuardAgent) ValidateToolCall(ctx context.Context, toolCall ToolCallResponse) hallucinationguard.ValidationResult {
	return a.guard.ValidateToolCall(ctx, hallucinationguard.ToolCall{
		Name:       toolCall.Name,
		Parameters: toolCall.Parameters,
	})
}

func main() {
	ctx := context.Background()
	apiKey := getenvOrDefault("ANTHROPIC_API_KEY", "")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable not set")
	}

	hguardAgent := NewHGuardAgent("example/schemas.yaml", "example/policies.yaml")

	prompts := []struct {
		name   string
		prompt string
	}{
		{"weather", `What is the weather in London? Respond ONLY with a JSON object for the tool call, e.g. {"name": "weather", "parameters": {"city": "London"}}. Do not include any other text.`},
		{"price", `Get the current price for NGN. Respond ONLY with a JSON object for the tool call, e.g. {"name": "price", "parameters": {"currency": "NGN"}}. Do not include any other text.`},
		{"quote", `Get a quote to convert 1000 BTCSAT to NGNKOBO using LIGHTNING. Respond ONLY with a JSON object for the tool call, e.g. {"name": "quote", "parameters": {"amount": 1000, "sourceCurrency": "BTCSAT", "targetCurrency": "NGNKOBO", "paymentMethod": "LIGHTNING", "paymentCurrency": "NGNKOBO"}}. Do not include any other text.`},
		{"search", `Search for the latest Go programming news. Respond ONLY with a JSON object for the tool call, e.g. {"name": "search", "parameters": {"query": "latest Go programming news"}}. Do not include any other text.`},
	}

	for _, p := range prompts {
		fmt.Printf("\n--- Testing tool: %s ---\n", p.name)
		response, err := callAnthropic(ctx, apiKey, p.prompt)
		if err != nil {
			log.Printf("Anthropic API error: %v", err)
			continue
		}

		var toolCall ToolCallResponse
		fmt.Printf("Claude raw response: %s\n", response)
		if err := json.Unmarshal([]byte(response), &toolCall); err != nil {
			log.Printf("Failed to parse tool call from Claude response: %v\nResponse: %s", err, response)
			continue
		}
		fmt.Printf("Parsed tool call: %+v\n", toolCall)

		result := hguardAgent.ValidateToolCall(ctx, toolCall)
		if !result.ExecutionAllowed {
			log.Printf("Tool call rejected: %s (action: %s)", result.Error, result.PolicyAction)
			if result.SuggestedCorrection != nil {
				log.Printf("Did you mean: %v?", result.SuggestedCorrection)
			}
			continue
		}

		switch toolCall.Name {
		case "weather":
			city, _ := toolCall.Parameters["city"].(string)
			country, _ := toolCall.Parameters["country"].(string)
			weatherKey := getenvOrDefault("OPENWEATHER_API_KEY", "")
			weather, err := getWeather(ctx, city, country, weatherKey)
			if err != nil {
				log.Printf("Weather API error: %v", err)
				continue
			}
			fmt.Printf("Weather in %s: %s\n", city, weather)
		case "quote":
			mavapayKey := getenvOrDefault("MAVAPAY_API_KEY", "")
			quote, err := getQuote(ctx, toolCall.Parameters, mavapayKey)
			if err != nil {
				log.Printf("Quote API error: %v", err)
				continue
			}
			fmt.Printf("Quote result: %s\n", quote)
		case "price":
			mavapayKey := getenvOrDefault("MAVAPAY_API_KEY", "")
			currency, _ := toolCall.Parameters["currency"].(string)
			price, err := getPrice(ctx, currency, mavapayKey)
			if err != nil {
				log.Printf("Price API error: %v", err)
				continue
			}
			fmt.Printf("Price for %s: %s\n", currency, price)
		case "search":
			query, _ := toolCall.Parameters["query"].(string)
			fmt.Printf("Search query: %s\n", query)
		default:
			fmt.Printf("Tool '%s' is not implemented.\n", toolCall.Name)
		}
	}
}

func callAnthropic(ctx context.Context, apiKey, prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model":      "claude-3-opus-20240229",
		"max_tokens": 256,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", anthropicAPIURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return "", err
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("Anthropic API error: %s", string(b))
	}
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	// Anthropic's response format: content is an array of blocks with 'text' fields
	if contentArr, ok := result["content"].([]interface{}); ok && len(contentArr) > 0 {
		if contentBlock, ok := contentArr[0].(map[string]interface{}); ok {
			if text, ok := contentBlock["text"].(string); ok {
				return text, nil
			}
		}
	}
	return "", fmt.Errorf("unexpected response format: %v", result)
}

func getWeather(ctx context.Context, city, country, apiKey string) (string, error) {
	q := city
	if country != "" {
		q = fmt.Sprintf("%s,%s", city, country)
	}
	url := fmt.Sprintf("%s?q=%s&appid=%s&units=metric", openWeatherAPIURL, q, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenWeatherMap error: %s", string(b))
	}
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if weatherArr, ok := data["weather"].([]interface{}); ok && len(weatherArr) > 0 {
		if weather, ok := weatherArr[0].(map[string]interface{}); ok {
			return weather["description"].(string), nil
		}
	}
	return "", fmt.Errorf("unexpected weather response: %v", data)
}

func getQuote(ctx context.Context, params map[string]interface{}, apiKey string) (string, error) {
	bodyBytes, _ := json.Marshal(params)
	req, err := http.NewRequestWithContext(ctx, "POST", mavapayQuoteURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return "", err
	}
	req.Header.Set("X-API-KEY", apiKey)
	req.Header.Set("content-type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	return string(b), nil
}

func getPrice(ctx context.Context, currency, apiKey string) (string, error) {
	url := fmt.Sprintf("%s?currency=%s", mavapayPriceURL, currency)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-API-KEY", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	return string(b), nil
}

func getenvOrDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

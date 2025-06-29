package scaffold

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Config struct {
	AnthropicAPIKey   string
	OpenWeatherAPIKey string
	MavapayAPIKey     string
}

func LoadConfig() Config {
	return Config{
		AnthropicAPIKey:   getenvOrDefault("ANTHROPIC_API_KEY", ""),
		OpenWeatherAPIKey: getenvOrDefault("OPENWEATHER_API_KEY", ""),
		MavapayAPIKey:     getenvOrDefault("MAVAPAY_API_KEY", ""),
	}
}

func CallAnthropic(ctx context.Context, apiKey, prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model":      "claude-3-opus-20240229",
		"max_tokens": 256,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", strings.NewReader(string(bodyBytes)))
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
		return "", fmt.Errorf("Anthropic API error: %s", resp.Status)
	}
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if contentArr, ok := result["content"].([]interface{}); ok && len(contentArr) > 0 {
		if contentBlock, ok := contentArr[0].(map[string]interface{}); ok {
			if text, ok := contentBlock["text"].(string); ok {
				return text, nil
			}
		}
	}
	return "", fmt.Errorf("unexpected response format: %v", result)
}

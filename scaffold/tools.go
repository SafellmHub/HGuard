package scaffold

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type ToolFunc func(ctx context.Context, params map[string]interface{}) (string, error)

var ToolRegistry = map[string]ToolFunc{
	"weather":  WeatherTool,
	"addition": AdditionTool,
	"search":   SearchTool,
	"quote":    QuoteTool,
	"price":    PriceTool,
}

func WeatherTool(ctx context.Context, params map[string]interface{}) (string, error) {
	// Accept both "city" and "location" as synonyms
	city, _ := params["city"].(string)
	if city == "" {
		city, _ = params["location"].(string)
	}
	country, _ := params["country"].(string)
	apiKey := getenvOrDefault("OPENWEATHER_API_KEY", "")
	weather, err := getWeather(ctx, city, country, apiKey)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Weather in %s: %s", city, weather), nil
}

func getWeather(ctx context.Context, city, country, apiKey string) (string, error) {
	q := city
	if country != "" {
		q = fmt.Sprintf("%s,%s", city, country)
	}
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", q, apiKey)
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

func AdditionTool(ctx context.Context, params map[string]interface{}) (string, error) {
	aVal, _ := params["a"].(float64)
	bVal, _ := params["b"].(float64)
	sum := aVal + bVal
	return fmt.Sprintf("The sum of %.2f and %.2f is %.2f", aVal, bVal, sum), nil
}

func SearchTool(ctx context.Context, params map[string]interface{}) (string, error) {
	query, _ := params["query"].(string)
	return fmt.Sprintf("Search query: %s", query), nil
}

func QuoteTool(ctx context.Context, params map[string]interface{}) (string, error) {
	apiKey := getenvOrDefault("MAVAPAY_API_KEY", "")
	bodyBytes, _ := json.Marshal(params)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://staging.api.mavapay.co/api/v1/quote", strings.NewReader(string(bodyBytes)))
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

func PriceTool(ctx context.Context, params map[string]interface{}) (string, error) {
	apiKey := getenvOrDefault("MAVAPAY_API_KEY", "")
	currency, _ := params["currency"].(string)
	url := fmt.Sprintf("https://staging.api.mavapay.co/api/v1/price?currency=%s", currency)
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

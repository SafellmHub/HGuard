package scaffold

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type ToolFunc func(ctx context.Context, params map[string]interface{}) (string, error)

var ToolRegistry = map[string]ToolFunc{
	// Basic tools
	"weather":  WeatherTool,
	"addition": AdditionTool,
	"search":   SearchTool,
	"quote":    QuoteTool,
	"price":    PriceTool,

	// Business tools
	"file_operations": FileOperationsTool,
	"system_info":     SystemInfoTool,
	"user_management": UserManagementTool,
	"send_email":      SendEmailTool,
	"database_query":  DatabaseQueryTool,
	"calendar":        CalendarTool,
	"task_management": TaskManagementTool,
	"analytics":       AnalyticsTool,
	"document_gen":    DocumentGenerationTool,
	"notification":    NotificationTool,
}

// WeatherTool gets weather information for a location
func WeatherTool(ctx context.Context, params map[string]interface{}) (string, error) {
	// Accept both "city" and "location" as synonyms
	city, _ := params["city"].(string)
	if city == "" {
		city, _ = params["location"].(string)
	}
	country, _ := params["country"].(string)
	apiKey := getenvOrDefault("OPENWEATHER_API_KEY", "")

	if apiKey == "" {
		return fmt.Sprintf("Weather service unavailable - API key not configured. City: %s", city), nil
	}

	weather, err := getWeather(ctx, city, country, apiKey)
	if err != nil {
		return fmt.Sprintf("Error getting weather for %s: %v", city, err), nil
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

// AdditionTool performs mathematical addition
func AdditionTool(ctx context.Context, params map[string]interface{}) (string, error) {
	aVal, _ := params["a"].(float64)
	bVal, _ := params["b"].(float64)
	sum := aVal + bVal
	return fmt.Sprintf("The sum of %.2f and %.2f is %.2f", aVal, bVal, sum), nil
}

// SearchTool performs web search (simulated)
func SearchTool(ctx context.Context, params map[string]interface{}) (string, error) {
	query, _ := params["query"].(string)
	return fmt.Sprintf("Search results for: %s\n\nThis is a simulated search. In a real implementation, this would connect to a search API.", query), nil
}

// QuoteTool gets financial quotes
func QuoteTool(ctx context.Context, params map[string]interface{}) (string, error) {
	apiKey := getenvOrDefault("MAVAPAY_API_KEY", "")
	if apiKey == "" {
		return "Quote service unavailable - API key not configured", nil
	}

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

// PriceTool gets currency pricing
func PriceTool(ctx context.Context, params map[string]interface{}) (string, error) {
	apiKey := getenvOrDefault("MAVAPAY_API_KEY", "")
	if apiKey == "" {
		return "Price service unavailable - API key not configured", nil
	}

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

// FileOperationsTool manages file operations (admin only)
func FileOperationsTool(ctx context.Context, params map[string]interface{}) (string, error) {
	operation, _ := params["operation"].(string)
	filepath, _ := params["filepath"].(string)

	switch operation {
	case "list":
		return fmt.Sprintf("Listing files in directory: %s\n\nThis is a simulated file listing. In a real implementation, this would list actual files.", filepath), nil
	case "read":
		return fmt.Sprintf("Reading file: %s\n\nThis is a simulated file read. In a real implementation, this would read the actual file content.", filepath), nil
	case "write":
		content, _ := params["content"].(string)
		return fmt.Sprintf("Writing to file: %s\nContent: %s\n\nThis is a simulated file write. In a real implementation, this would write to the actual file.", filepath, content), nil
	case "delete":
		return fmt.Sprintf("Deleting file: %s\n\nThis is a simulated file deletion. In a real implementation, this would delete the actual file.", filepath), nil
	default:
		return fmt.Sprintf("Unknown file operation: %s", operation), nil
	}
}

// SystemInfoTool gets system information
func SystemInfoTool(ctx context.Context, params map[string]interface{}) (string, error) {
	infoType, _ := params["info_type"].(string)

	switch infoType {
	case "cpu":
		return "CPU Usage: 45%\nCPU Model: Intel Core i7-9700K\nCores: 8", nil
	case "memory":
		return "Memory Usage: 8.2GB / 16GB (51%)\nAvailable: 7.8GB", nil
	case "disk":
		return "Disk Usage: 256GB / 512GB (50%)\nFree Space: 256GB", nil
	case "network":
		return "Network Status: Connected\nBandwidth: 100 Mbps\nLatency: 12ms", nil
	default:
		return "System Status: Online\nUptime: 5 days, 14 hours\nLoad Average: 0.85", nil
	}
}

// UserManagementTool manages users (admin only)
func UserManagementTool(ctx context.Context, params map[string]interface{}) (string, error) {
	action, _ := params["action"].(string)
	username, _ := params["username"].(string)

	switch action {
	case "create":
		role, _ := params["role"].(string)
		return fmt.Sprintf("Created user: %s with role: %s\n\nThis is a simulated user creation. In a real implementation, this would create an actual user account.", username, role), nil
	case "delete":
		return fmt.Sprintf("Deleted user: %s\n\nThis is a simulated user deletion. In a real implementation, this would delete the actual user account.", username), nil
	case "list":
		return "Active Users:\n- admin (Admin)\n- john_doe (Manager)\n- jane_smith (User)\n- dev_user (Developer)\n\nThis is a simulated user listing.", nil
	case "update":
		newRole, _ := params["new_role"].(string)
		return fmt.Sprintf("Updated user: %s to role: %s\n\nThis is a simulated user update.", username, newRole), nil
	default:
		return fmt.Sprintf("Unknown user management action: %s", action), nil
	}
}

// SendEmailTool sends emails (requires permission)
func SendEmailTool(ctx context.Context, params map[string]interface{}) (string, error) {
	to, _ := params["to"].(string)
	subject, _ := params["subject"].(string)
	body, _ := params["body"].(string)

	return fmt.Sprintf("Email sent successfully!\nTo: %s\nSubject: %s\nBody: %s\n\nThis is a simulated email. In a real implementation, this would send an actual email.", to, subject, body), nil
}

// DatabaseQueryTool queries databases (requires permission)
func DatabaseQueryTool(ctx context.Context, params map[string]interface{}) (string, error) {
	query, _ := params["query"].(string)
	database, _ := params["database"].(string)

	return fmt.Sprintf("Database Query Results:\nDatabase: %s\nQuery: %s\n\nResults:\n- Record 1: Sample data\n- Record 2: More sample data\n\nThis is a simulated database query. In a real implementation, this would query an actual database.", database, query), nil
}

// CalendarTool manages calendar events
func CalendarTool(ctx context.Context, params map[string]interface{}) (string, error) {
	action, _ := params["action"].(string)

	switch action {
	case "create":
		title, _ := params["title"].(string)
		date, _ := params["date"].(string)
		return fmt.Sprintf("Created calendar event: %s on %s\n\nThis is a simulated calendar event creation.", title, date), nil
	case "list":
		return "Upcoming Events:\n- Team Meeting (Tomorrow 2:00 PM)\n- Project Review (Friday 10:00 AM)\n- Client Call (Monday 9:00 AM)\n\nThis is a simulated calendar listing.", nil
	case "update":
		eventId, _ := params["event_id"].(string)
		title, _ := params["title"].(string)
		return fmt.Sprintf("Updated event %s: %s\n\nThis is a simulated calendar event update.", eventId, title), nil
	case "delete":
		eventId, _ := params["event_id"].(string)
		return fmt.Sprintf("Deleted event: %s\n\nThis is a simulated calendar event deletion.", eventId), nil
	default:
		return fmt.Sprintf("Unknown calendar action: %s", action), nil
	}
}

// TaskManagementTool manages tasks
func TaskManagementTool(ctx context.Context, params map[string]interface{}) (string, error) {
	action, _ := params["action"].(string)

	switch action {
	case "create":
		title, _ := params["title"].(string)
		priority, _ := params["priority"].(string)
		return fmt.Sprintf("Created task: %s (Priority: %s)\n\nThis is a simulated task creation.", title, priority), nil
	case "list":
		return "Active Tasks:\n- Complete project documentation (High)\n- Review code changes (Medium)\n- Update system dependencies (Low)\n\nThis is a simulated task listing.", nil
	case "update":
		taskId, _ := params["task_id"].(string)
		status, _ := params["status"].(string)
		return fmt.Sprintf("Updated task %s status to: %s\n\nThis is a simulated task update.", taskId, status), nil
	case "delete":
		taskId, _ := params["task_id"].(string)
		return fmt.Sprintf("Deleted task: %s\n\nThis is a simulated task deletion.", taskId), nil
	default:
		return fmt.Sprintf("Unknown task management action: %s", action), nil
	}
}

// AnalyticsTool provides analytics data
func AnalyticsTool(ctx context.Context, params map[string]interface{}) (string, error) {
	reportType, _ := params["report_type"].(string)
	dateRange, _ := params["date_range"].(string)

	switch reportType {
	case "sales":
		return fmt.Sprintf("Sales Report (%s):\n- Total Revenue: $125,000\n- Orders: 1,250\n- Average Order Value: $100\n- Growth: +15%% vs previous period\n\nThis is simulated analytics data.", dateRange), nil
	case "users":
		return fmt.Sprintf("User Analytics (%s):\n- Active Users: 5,420\n- New Signups: 234\n- Retention Rate: 87%%\n- Churn Rate: 3.2%%\n\nThis is simulated analytics data.", dateRange), nil
	case "performance":
		return fmt.Sprintf("Performance Report (%s):\n- Average Response Time: 125ms\n- Error Rate: 0.02%%\n- Uptime: 99.98%%\n- Throughput: 1,250 req/sec\n\nThis is simulated analytics data.", dateRange), nil
	default:
		return fmt.Sprintf("Analytics Report (%s):\n- Overall system health: Excellent\n- Key metrics trending positive\n\nThis is simulated analytics data.", dateRange), nil
	}
}

// DocumentGenerationTool generates documents
func DocumentGenerationTool(ctx context.Context, params map[string]interface{}) (string, error) {
	docType, _ := params["document_type"].(string)
	template, _ := params["template"].(string)
	data, _ := params["data"].(map[string]interface{})

	return fmt.Sprintf("Generated %s document using template: %s\nData: %v\n\nDocument ID: DOC-%d\nThis is a simulated document generation.", docType, template, data, time.Now().Unix()), nil
}

// NotificationTool sends notifications
func NotificationTool(ctx context.Context, params map[string]interface{}) (string, error) {
	notificationType, _ := params["type"].(string)
	message, _ := params["message"].(string)
	recipients, _ := params["recipients"].([]interface{})

	return fmt.Sprintf("Sent %s notification to %d recipients\nMessage: %s\n\nThis is a simulated notification.", notificationType, len(recipients), message), nil
}

func getenvOrDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

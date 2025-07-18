package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/SafellmHub/hguard-go/scaffold"
)

func main() {
	fmt.Println("🤖 Standard AI Agent with HGuard Integration")
	fmt.Println("===============================================")
	fmt.Println()

	// Load configuration
	config := scaffold.LoadConfig()

	// Initialize the agent
	agent := scaffold.NewStandardAgent("../../scaffold/schemas.yaml", "../../scaffold/policies.yaml", config)

	// User selection
	fmt.Println("Select your role:")
	fmt.Println("1. Guest (Limited access)")
	fmt.Println("2. User (Standard access)")
	fmt.Println("3. Manager (Extended access)")
	fmt.Println("4. Developer (Technical access)")
	fmt.Println("5. Admin (Full access)")
	fmt.Print("Choose (1-5): ")

	reader := bufio.NewReader(os.Stdin)
	roleChoice, _ := reader.ReadString('\n')
	roleChoice = strings.TrimSpace(roleChoice)

	var userRole scaffold.UserRole
	var userID string

	switch roleChoice {
	case "1":
		userRole = scaffold.RoleGuest
		userID = "guest_user"
	case "2":
		userRole = scaffold.RoleUser
		userID = "standard_user"
	case "3":
		userRole = scaffold.RoleManager
		userID = "manager_user"
	case "4":
		userRole = scaffold.RoleDeveloper
		userID = "developer_user"
	case "5":
		userRole = scaffold.RoleAdmin
		userID = "admin_user"
	default:
		userRole = scaffold.RoleGuest
		userID = "guest_user"
	}

	// Create user and session contexts
	userCtx := scaffold.CreateUserContext(userID, userRole, "127.0.0.1")
	sessionCtx := scaffold.CreateSessionContext(fmt.Sprintf("session_%d", time.Now().Unix()))

	fmt.Printf("\n✅ Logged in as: %s (%s)\n", userID, userRole)
	fmt.Printf("🔐 Permissions: %s\n", strings.Join(userCtx.Permissions, ", "))
	fmt.Printf("📊 Session ID: %s\n", sessionCtx.ID)

	// Show available tools
	availableTools := agent.GetAvailableTools(userCtx)
	fmt.Printf("\n🛠️  Available Tools (%d):\n", len(availableTools))
	for _, tool := range availableTools {
		fmt.Printf("  - %s: %s\n", tool, agent.GetToolDescription(tool))
	}

	fmt.Println("\n💬 Chat with the agent (type 'exit' to quit, 'help' for commands):")
	fmt.Println("===============================================")

	ctx := context.Background()

	for {
		fmt.Print("\nYou: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading input: %v", err)
			continue
		}

		input = strings.TrimSpace(input)

		if input == "exit" || input == "quit" {
			fmt.Println("\n👋 Goodbye!")
			break
		}

		if input == "help" {
			printHelp()
			continue
		}

		if input == "stats" {
			printStats(agent, userCtx, sessionCtx)
			continue
		}

		if input == "tools" {
			printAvailableTools(agent, userCtx)
			continue
		}

		if input == "clear" {
			agent.ClearConversationHistory(sessionCtx.ID)
			fmt.Println("🗑️  Conversation history cleared")
			continue
		}

		if input == "" {
			continue
		}

		// Process the message
		fmt.Print("🤖 Agent: ")
		response, err := agent.ProcessMessage(ctx, input, userCtx, sessionCtx)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		fmt.Println(response)
	}
}

func printHelp() {
	fmt.Println("\n📋 Available Commands:")
	fmt.Println("  help   - Show this help message")
	fmt.Println("  stats  - Show user and session statistics")
	fmt.Println("  tools  - Show available tools for your role")
	fmt.Println("  clear  - Clear conversation history")
	fmt.Println("  exit   - Exit the application")
	fmt.Println("\n💡 Example queries:")
	fmt.Println("  - What's the weather in London?")
	fmt.Println("  - Add 15 and 27")
	fmt.Println("  - Search for the latest AI news")
	fmt.Println("  - Create a task: Complete project documentation")
	fmt.Println("  - Send an email to team@company.com about the meeting")
	fmt.Println("  - Show system information")
	fmt.Println("  - Generate a sales report for this month")
}

func printStats(agent *scaffold.StandardAgent, userCtx scaffold.UserContext, sessionCtx scaffold.SessionContext) {
	stats := agent.GetUserStats(userCtx, sessionCtx)
	fmt.Println("\n📊 User & Session Statistics:")
	fmt.Printf("  User ID: %s\n", stats["user_id"])
	fmt.Printf("  Role: %s\n", stats["role"])
	fmt.Printf("  Session ID: %s\n", stats["session_id"])
	fmt.Printf("  Session Start: %s\n", stats["session_start"])
	fmt.Printf("  Calls Made: %d\n", stats["calls_made"])
	fmt.Printf("  Available Tools: %d\n", stats["available_tools"])

	if permissions, ok := stats["permissions"].([]string); ok {
		fmt.Printf("  Permissions: %s\n", strings.Join(permissions, ", "))
	}
}

func printAvailableTools(agent *scaffold.StandardAgent, userCtx scaffold.UserContext) {
	tools := agent.GetAvailableTools(userCtx)
	fmt.Printf("\n🛠️  Available Tools for %s role:\n", userCtx.Role)
	for _, tool := range tools {
		fmt.Printf("  - %s: %s\n", tool, agent.GetToolDescription(tool))
	}
}

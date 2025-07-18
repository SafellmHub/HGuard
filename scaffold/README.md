# Standard AI Agent with HGuard Integration

This directory contains a comprehensive standard AI agent implementation using the HallucinationGuard Go SDK. The agent supports role-based access control, conversation flows, session management, and a wide range of business tools.

## Features

- **Role-Based Access Control**: Different user roles (Admin, Manager, Developer, User, Guest) with specific permissions
- **Conversation Management**: Maintains conversation history and context across sessions
- **Session Management**: Tracks user sessions and previous tool calls
- **Comprehensive Business Tools**: 15+ tools covering weather, calculations, file operations, system management, email, databases, calendar, tasks, analytics, and document generation
- **Context-Aware Policies**: Advanced policy engine supporting conditional logic based on user roles, parameters, time, and session state
- **Anthropic Integration**: Full conversation capabilities using Claude API
- **Security**: Built-in security through HGuard policy validation and enforcement

## Project Structure

```
scaffold/
  agent.go         # Standard AI agent with conversation and role management
  tools.go         # Comprehensive business tool implementations
  config.go        # Configuration with role management and Anthropic integration
  schemas.yaml     # Complete tool schemas with parameter validation
  policies.yaml    # Role-based and context-aware policy definitions
  prompt.txt       # Comprehensive prompt for conversation and tool usage
```

## User Roles & Permissions

### Guest (Limited Access)

- Basic tools: weather, addition, search
- Cannot access business or administrative functions

### User (Standard Access)

- All basic tools
- Calendar, task management, document generation
- Email sending (with rate limits)
- Cannot access system administration or sensitive data

### Manager (Extended Access)

- All User permissions
- Analytics and reporting
- Financial operations (quotes, pricing)
- Task deletion and management
- SMS notifications

### Developer (Technical Access)

- All User permissions
- System information access
- File operations (list, read, write)
- Database queries
- Technical tools

### Admin (Full Access)

- All permissions
- User management
- File deletion
- System administration
- Override most restrictions

## Available Tools

### Basic Tools

- **weather**: Get weather information for any location
- **addition**: Perform mathematical calculations
- **search**: Search the web for information

### Financial Tools

- **quote**: Get financial quotes for currency exchanges
- **price**: Get current currency pricing information

### System & Administration

- **file_operations**: Manage files and directories (Admin/Developer)
- **system_info**: Get system information and metrics
- **user_management**: Manage user accounts and roles (Admin only)

### Communication

- **send_email**: Send email notifications
- **notification**: Send various types of notifications (email, SMS, push, etc.)

### Data & Analytics

- **database_query**: Query databases (requires permission)
- **analytics**: Generate analytics reports and insights

### Productivity

- **calendar**: Manage calendar events and appointments
- **task_management**: Create and manage tasks
- **document_gen**: Generate documents from templates

## Getting Started

1. **Install dependencies:**

   ```bash
   go mod tidy
   ```

2. **Run the interactive demo:**

   ```bash
   cd cmd/agent_demo
   go run main.go
   ```

3. **Configuration:**
   - The agent uses the provided Anthropic API key
   - Set optional environment variables for weather and payment APIs:
     ```bash
     export OPENWEATHER_API_KEY=your_key
     export MAVAPAY_API_KEY=your_key
     ```

## Usage Examples

### Basic Usage

```go
package main

import (
    "context"
    "github.com/SafellmHub/hguard-go/scaffold"
)

func main() {
    // Load configuration
    config := scaffold.LoadConfig()

    // Initialize agent
    agent := scaffold.NewStandardAgent("schemas.yaml", "policies.yaml", config)

    // Create user context
    userCtx := scaffold.CreateUserContext("user123", scaffold.RoleManager, "192.168.1.100")
    sessionCtx := scaffold.CreateSessionContext("session456")

    // Process messages
    response, err := agent.ProcessMessage(context.Background(), "What's the weather in London?", userCtx, sessionCtx)
    if err != nil {
        panic(err)
    }

    println(response)
}
```

### Role-Based Access

```go
// Admin user - full access
adminCtx := scaffold.CreateUserContext("admin1", scaffold.RoleAdmin, "127.0.0.1")

// Manager user - business access
managerCtx := scaffold.CreateUserContext("manager1", scaffold.RoleManager, "127.0.0.1")

// Guest user - limited access
guestCtx := scaffold.CreateUserContext("guest1", scaffold.RoleGuest, "127.0.0.1")

// Get available tools for each role
adminTools := agent.GetAvailableTools(adminCtx)
managerTools := agent.GetAvailableTools(managerCtx)
guestTools := agent.GetAvailableTools(guestCtx)
```

## Policy Examples

The agent supports sophisticated policy rules:

```yaml
# Role-based access
- tool_name: user_management
  type: REJECT
  condition: "user.role != 'admin'"
  reason: "User management requires admin privileges"

# Parameter-based restrictions
- tool_name: quote
  type: REJECT
  condition: "params.amount > 10000 && user.role != 'admin'"
  reason: "Large transactions require admin approval"

# Time-based restrictions
- tool_name: database_query
  type: REJECT
  condition: "time.hour < 8 || time.hour > 18"
  reason: "Database access only during business hours"

# Session-based restrictions
- tool_name: send_email
  type: REJECT
  condition: "len(session.previous_calls) > 10"
  reason: "Email rate limit exceeded"
```

## Interactive Demo

The interactive demo (`cmd/agent_demo/main.go`) provides:

- Role selection interface
- Real-time conversation with the agent
- Tool usage examples for each role
- Session statistics and management
- Help system with available commands

## Security Features

- **Input Validation**: All tool parameters are validated against schemas
- **Role-Based Access**: Tools are restricted based on user roles
- **Rate Limiting**: Session-based limits prevent abuse
- **Context Awareness**: Policies can consider user context, time, and session state
- **Safe Execution**: Tools are simulated for safety (can be replaced with real implementations)

## Integration

This agent is designed to be easily integrated into:

- Web applications
- Chat interfaces
- API services
- MCP (Model Context Protocol) servers
- Enterprise applications

The agent provides a clean API for message processing, tool execution, and session management while maintaining security through HGuard's policy engine.

## License

MIT

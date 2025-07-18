package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/SafellmHub/hguard-go/pkg/hallucinationguard"
)

func main() {
	// Create a temporary policy file for demonstration
	policyContent := `policies:
  # Role-based access control
  - tool_name: admin_tool
    type: REJECT
    condition: "user.role != 'admin'"
    reason: "Only administrators can use this tool"
    priority: 10

  - tool_name: admin_tool
    type: ALLOW
    condition: "user.role == 'admin'"
    reason: "Admin access granted"
    priority: 20

  # Parameter-based restrictions
  - tool_name: transfer_money
    type: REJECT
    condition: "params.amount > 1000"
    reason: "Transfer amount too high"
    priority: 15

  - tool_name: transfer_money
    type: ALLOW
    condition: "user.role == 'admin' && params.amount <= 1000"
    reason: "Admin transfer approved"
    priority: 25

  # Time-based restrictions
  - tool_name: maintenance_tool
    type: REJECT
    condition: "time.hour < 9 || time.hour > 17"
    reason: "Maintenance tools only available during business hours"
    priority: 5

  # Session-based restrictions
  - tool_name: sensitive_operation
    type: REJECT
    condition: "'sensitive_operation' in session.previous_calls"
    reason: "Operation already performed in this session"
    priority: 8

  # Permission-based access
  - tool_name: financial_data
    type: ALLOW
    condition: "'read_financial' in user.permissions"
    reason: "User has financial access permission"
    priority: 12

  # Auto-correct typos
  - tool_name: wheather
    type: REWRITE
    target: weather
    reason: "Auto-corrected typo"

  # Basic tools
  - tool_name: weather
    type: ALLOW

  # Fallback
  - tool_name: "*"
    type: REJECT
    reason: "Unknown tool"
    priority: 1`

	schemaContent := `schemas:
  - name: admin_tool
    parameters:
      action:
        type: string
        required: true
  - name: transfer_money
    parameters:
      amount:
        type: number
        required: true
      recipient:
        type: string
        required: true
  - name: maintenance_tool
    parameters:
      operation:
        type: string
        required: true
  - name: sensitive_operation
    parameters:
      data:
        type: string
        required: false
  - name: financial_data
    parameters:
      query:
        type: string
        required: true
  - name: weather
    parameters:
      location:
        type: string
        required: true`

	// Write temporary files
	if err := os.WriteFile("temp_policies.yaml", []byte(policyContent), 0644); err != nil {
		log.Fatal(err)
	}
	defer os.Remove("temp_policies.yaml")

	if err := os.WriteFile("temp_schemas.yaml", []byte(schemaContent), 0644); err != nil {
		log.Fatal(err)
	}
	defer os.Remove("temp_schemas.yaml")

	// Initialize HallucinationGuard
	ctx := context.Background()
	guard := hallucinationguard.New()

	if err := guard.LoadSchemasFromFile(ctx, "temp_schemas.yaml"); err != nil {
		log.Fatalf("Failed to load schemas: %v", err)
	}

	if err := guard.LoadPoliciesFromFile(ctx, "temp_policies.yaml"); err != nil {
		log.Fatalf("Failed to load policies: %v", err)
	}

	fmt.Println("=== Context-Aware Policy Examples ===")

	// Example 1: Role-based access control
	fmt.Println("1. Role-based Access Control")

	// Admin user trying to use admin tool
	adminToolCall := hallucinationguard.ToolCall{
		Name: "admin_tool",
		Parameters: map[string]interface{}{
			"action": "delete_user",
		},
		Context: &hallucinationguard.CallContext{
			UserRole: "admin",
			UserID:   "admin123",
		},
	}

	result := guard.ValidateToolCall(ctx, adminToolCall)
	fmt.Printf("Admin user -> %s: %s\n", result.Status, result.Error)

	// Regular user trying to use admin tool
	adminToolCall.Context.UserRole = "user"
	result = guard.ValidateToolCall(ctx, adminToolCall)
	fmt.Printf("Regular user -> %s: %s\n\n", result.Status, result.Error)

	// Example 2: Parameter-based restrictions
	fmt.Println("2. Parameter-based Restrictions")

	transferCall := hallucinationguard.ToolCall{
		Name: "transfer_money",
		Parameters: map[string]interface{}{
			"amount":    500,
			"recipient": "john@example.com",
		},
		Context: &hallucinationguard.CallContext{
			UserRole: "admin",
			UserID:   "admin123",
		},
	}

	result = guard.ValidateToolCall(ctx, transferCall)
	fmt.Printf("Admin $500 transfer -> %s: %s\n", result.Status, result.Error)

	// High amount transfer
	transferCall.Parameters["amount"] = 5000
	result = guard.ValidateToolCall(ctx, transferCall)
	fmt.Printf("Admin $5000 transfer -> %s: %s\n\n", result.Status, result.Error)

	// Example 3: Time-based restrictions
	fmt.Println("3. Time-based Restrictions")

	maintenanceCall := hallucinationguard.ToolCall{
		Name: "maintenance_tool",
		Parameters: map[string]interface{}{
			"operation": "restart_server",
		},
		Context: &hallucinationguard.CallContext{
			UserRole:  "admin",
			TimeOfDay: 14, // 2 PM
		},
	}

	result = guard.ValidateToolCall(ctx, maintenanceCall)
	fmt.Printf("Maintenance at 2 PM -> %s: %s\n", result.Status, result.Error)

	// Outside business hours
	maintenanceCall.Context.TimeOfDay = 20 // 8 PM
	result = guard.ValidateToolCall(ctx, maintenanceCall)
	fmt.Printf("Maintenance at 8 PM -> %s: %s\n\n", result.Status, result.Error)

	// Example 4: Session-based restrictions
	fmt.Println("4. Session-based Restrictions")

	sensitiveCall := hallucinationguard.ToolCall{
		Name: "sensitive_operation",
		Parameters: map[string]interface{}{
			"data": "secret_info",
		},
		Context: &hallucinationguard.CallContext{
			UserRole:      "admin",
			SessionID:     "session123",
			PreviousCalls: []string{"login", "get_balance"},
		},
	}

	result = guard.ValidateToolCall(ctx, sensitiveCall)
	fmt.Printf("First sensitive operation -> %s: %s\n", result.Status, result.Error)

	// Already performed in session
	sensitiveCall.Context.PreviousCalls = []string{"login", "sensitive_operation", "get_balance"}
	result = guard.ValidateToolCall(ctx, sensitiveCall)
	fmt.Printf("Repeated sensitive operation -> %s: %s\n\n", result.Status, result.Error)

	// Example 5: Permission-based access
	fmt.Println("5. Permission-based Access")

	financialCall := hallucinationguard.ToolCall{
		Name: "financial_data",
		Parameters: map[string]interface{}{
			"query": "get_balance",
		},
		Context: &hallucinationguard.CallContext{
			UserRole:        "user",
			UserID:          "user123",
			UserPermissions: []string{"read_financial", "write_basic"},
		},
	}

	result = guard.ValidateToolCall(ctx, financialCall)
	fmt.Printf("User with financial permission -> %s: %s\n", result.Status, result.Error)

	// User without financial permission
	financialCall.Context.UserPermissions = []string{"read_basic"}
	result = guard.ValidateToolCall(ctx, financialCall)
	fmt.Printf("User without financial permission -> %s: %s\n\n", result.Status, result.Error)

	// Example 6: Auto-correction
	fmt.Println("6. Auto-correction")

	typoCall := hallucinationguard.ToolCall{
		Name: "wheather",
		Parameters: map[string]interface{}{
			"location": "London",
		},
		Context: &hallucinationguard.CallContext{
			UserRole: "user",
		},
	}

	result = guard.ValidateToolCall(ctx, typoCall)
	fmt.Printf("Typo correction -> %s: %s\n", result.Status, result.Error)
	if result.SuggestedCorrection != nil {
		fmt.Printf("Suggested correction: %s\n", result.SuggestedCorrection.Name)
	}

	fmt.Println("\n=== Context-Aware Policy Demo Complete ===")
}

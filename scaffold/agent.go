package scaffold

import (
	"context"
	"log"

	"github.com/SafellmHub/hguard-go/pkg/hallucinationguard"
)

type ToolCallResponse struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
}

type HGuardAgent struct {
	guard *hallucinationguard.Guard
	tools map[string]ToolFunc
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
	return &HGuardAgent{guard: guard, tools: ToolRegistry}
}

func (a *HGuardAgent) ValidateToolCall(ctx context.Context, toolCall ToolCallResponse) hallucinationguard.ValidationResult {
	return a.guard.ValidateToolCall(ctx, hallucinationguard.ToolCall{
		Name:       toolCall.Name,
		Parameters: toolCall.Parameters,
	})
}

func (a *HGuardAgent) ExecuteTool(ctx context.Context, toolCall ToolCallResponse) (string, error) {
	toolFunc, ok := a.tools[toolCall.Name]
	if !ok {
		return "", nil
	}
	return toolFunc(ctx, toolCall.Parameters)
}

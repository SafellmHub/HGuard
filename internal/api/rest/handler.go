package rest

import (
	"encoding/json"
	"net/http"

	"github.com/fishonamos/hallucination-shield/internal/core/model"
	"github.com/fishonamos/hallucination-shield/internal/core/validation"
)

// ValidationRequest is the input for the validation API
type ValidationRequest struct {
	ToolCalls []model.ToolCall  `json:"tool_calls"`
	Context   model.CallContext `json:"context"`
}

// ValidationResponse is the output for the validation API
type ValidationResponse struct {
	Status           string                   `json:"status"`
	Results          []model.ValidationResult `json:"results"`
	ProcessingTimeMs int                      `json:"processing_time_ms"`
}

// ValidateHandler handles POST /api/v1/validate
func ValidateHandler(w http.ResponseWriter, r *http.Request) {
	var req ValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}
	results := make([]model.ValidationResult, 0, len(req.ToolCalls))
	for _, tc := range req.ToolCalls {
		// Attach context to each tool call
		tc.Context = req.Context
		results = append(results, validation.ValidateToolCall(tc))
	}
	resp := ValidationResponse{
		Status:           "success",
		Results:          results,
		ProcessingTimeMs: 1, // Placeholder, can add timing later
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

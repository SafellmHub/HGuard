package rest

import (
	"encoding/json"
	"net/http"

	"github.com/fishonamos/hallucination-shield/internal/core/model"
	"github.com/fishonamos/hallucination-shield/internal/core/validation"
	"github.com/fishonamos/hallucination-shield/internal/logging"
	"github.com/google/uuid"
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
	RequestID        string                   `json:"request_id"`
}

type ErrorResponse struct {
	Error     string `json:"error"`
	Code      int    `json:"code"`
	RequestID string `json:"request_id"`
}

// ValidateHandler handles POST /api/v1/validate
func ValidateHandler(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.NewString()
	var req ValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logging.ErrorWithID(requestID, "Invalid request body", map[string]interface{}{"error": err.Error()})
		writeError(w, "invalid request body", http.StatusBadRequest, requestID)
		return
	}
	if len(req.ToolCalls) == 0 {
		logging.WarnWithID(requestID, "No tool_calls provided", nil)
		writeError(w, "no tool_calls provided", http.StatusBadRequest, requestID)
		return
	}
	results := make([]model.ValidationResult, 0, len(req.ToolCalls))
	for _, tc := range req.ToolCalls {
		// Attach context to each tool call
		tc.Context = req.Context
		vr := validation.ValidateToolCall(tc)
		vr.Modifications = nil // Clean up if not used
		results = append(results, vr)
		if vr.Status == "rejected" || vr.Status == "rewritten" {
			logging.WarnWithID(requestID, "Tool call not approved", map[string]interface{}{
				"tool_call_id":  vr.ToolCallID,
				"status":        vr.Status,
				"reason":        vr.Reason,
				"policy_action": vr.PolicyAction,
			})
		}
	}
	resp := ValidationResponse{
		Status:           "success",
		Results:          results,
		ProcessingTimeMs: 1, // Todo: add timing later
		RequestID:        requestID,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, msg string, code int, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:     msg,
		Code:      code,
		RequestID: requestID,
	})
}

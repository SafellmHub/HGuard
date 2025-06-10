package model

// PolicyAction represents the action to take for a policy
type PolicyAction string

const (
	PolicyActionALLOW          PolicyAction = "ALLOW"
	PolicyActionREJECT         PolicyAction = "REJECT"
	PolicyActionREWRITE        PolicyAction = "REWRITE"
	PolicyActionRATE_LIMIT     PolicyAction = "RATE_LIMIT"
	PolicyActionCONTEXT_REJECT PolicyAction = "CONTEXT_REJECT"
)

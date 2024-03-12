package main

import (
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/fishonamos/hallucination-shield/internal/api/rest"
	"github.com/fishonamos/hallucination-shield/internal/core/policy"
	"github.com/fishonamos/hallucination-shield/internal/logging"
	"github.com/fishonamos/hallucination-shield/internal/schema"
)

var apiKey = os.Getenv("HALLUCINATION_GUARD_API_KEY")

// --- Rate Limiting ---
const rateLimit = 60 // requests per minute per IP

type clientInfo struct {
	count     int
	timestamp time.Time
}

var (
	rateLimitMu  sync.Mutex
	rateLimitMap = make(map[string]*clientInfo)
)

func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		// Remove port if present
		if colon := len(ip) - 1; colon > 0 && ip[colon] == ':' {
			ip = ip[:colon]
		}
		now := time.Now()
		rateLimitMu.Lock()
		info, ok := rateLimitMap[ip]
		if !ok || now.Sub(info.timestamp) > time.Minute {
			// New window
			rateLimitMap[ip] = &clientInfo{count: 1, timestamp: now}
			rateLimitMu.Unlock()
			next(w, r)
			return
		}
		if info.count >= rateLimit {
			rateLimitMu.Unlock()
			logging.Warn("Rate limit exceeded for IP %s", ip)
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Too Many Requests: rate limit exceeded"))
			return
		}
		info.count++
		rateLimitMu.Unlock()
		next(w, r)
	}
}

func apiKeyMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if apiKey == "" {
			// No key set: allow all, but warn
			logging.Warn("No API key set! All requests are allowed. Set HALLUCINATION_GUARD_API_KEY for production security.")
			next(w, r)
			return
		}
		head := r.Header.Get("Authorization")
		if head != "Bearer "+apiKey {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized: missing or invalid API key"))
			return
		}
		next(w, r)
	}
}

func registerExampleSchemasAndPolicies() {
	schema.RegisterToolSchema(schema.ToolSchema{
		Name: "get_weather",
		Parameters: map[string]schema.ParameterSchema{
			"location": {Type: "string", Required: true, MaxLength: 100},
			"units":    {Type: "string", Required: false, Enum: []string{"celsius", "fahrenheit"}},
		},
	})

	policy.RegisterPolicy(policy.Policy{
		ToolName: "get_weather",
		Type:     policy.PolicyAllow,
	})

	schema.RegisterToolSchema(schema.ToolSchema{
		Name: "transfer_money",
		Parameters: map[string]schema.ParameterSchema{
			"from":     {Type: "string", Required: true},
			"to":       {Type: "string", Required: true},
			"amount":   {Type: "number", Required: true},
			"currency": {Type: "string", Required: true, Enum: []string{"usd", "eur", "gbp"}},
		},
	})

	policy.RegisterPolicy(policy.Policy{
		ToolName: "transfer_money",
		Type:     policy.PolicyReject, // Example: block money transfers by default
	})
}

func main() {
	logging.Info("HallucinationGuard server starting on :8080 ...")
	registerExampleSchemasAndPolicies()

	http.HandleFunc("/api/v1/validate", rateLimitMiddleware(apiKeyMiddleware(func(w http.ResponseWriter, r *http.Request) {
		logging.Info("Received validation request from %s", r.RemoteAddr)
		rest.ValidateHandler(w, r)
	})))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

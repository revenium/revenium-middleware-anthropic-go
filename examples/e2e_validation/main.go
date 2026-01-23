// E2E Validation Test - Tests ALL metadata fields
// This example validates every possible field is correctly sent and received

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/revenium/revenium-middleware-anthropic-go/revenium"
)

func main() {
	// Initialize middleware
	if err := revenium.Initialize(); err != nil {
		log.Fatalf("Failed to initialize middleware: %v", err)
	}

	client, err := revenium.GetClient()
	if err != nil {
		log.Fatalf("Failed to get client: %v", err)
	}
	defer client.Close()

	// Generate unique trace ID for this test
	traceID := fmt.Sprintf("e2e-anthropic-go-%d", time.Now().UnixMilli())
	transactionID := fmt.Sprintf("txn-anthropic-go-%d", time.Now().UnixMilli())

	fmt.Println("=== E2E Validation Test - Anthropic Go Middleware ===")
	fmt.Printf("Trace ID: %s\n", traceID)
	fmt.Printf("Transaction ID: %s\n\n", transactionID)

	// Create context with ALL supported metadata fields
	ctx := context.Background()
	metadata := map[string]interface{}{
		// Core tracking fields
		"organizationId":   "e2e-validation-org",
		"productId":        "e2e-validation-product",
		"subscriptionId":   "e2e-sub-premium",
		"agent":            "e2e-validation-agent",
		"taskType":         "e2e-validation",
		"taskId":           "task-e2e-001",
		"transactionId":    transactionID,

		// Distributed tracing fields (ALL 10)
		"traceId":               traceID,
		"traceType":             "e2e-validation-trace",
		"traceName":             "Anthropic-Go-E2E-Test",
		"environment":           "qa-e2e",
		"region":                "us-east-1",
		"operationType":         "chat-completion",
		"operationSubtype":      "e2e-test",
		"retryNumber":           0,
		"credentialAlias":       "anthropic-e2e-key",
		"parentTransactionId":   "parent-e2e-001",

		// Quality scoring and model metadata
		"responseQualityScore": 0.95,
		"modelSource":          "e2e-validation-source",
		"temperature":          0.7,
		"mediationLatency":     15,

		// Subscriber tracking (full object)
		"subscriber": map[string]interface{}{
			"id":       "e2e-user-123",
			"email":    "e2e-test@revenium.io",
			"name":     "E2E Test User",
			"tier":     "premium",
			"credential": map[string]interface{}{
				"name":  "anthropic-e2e-key",
				"value": "e2e-key-identifier",
			},
		},

		// Custom attributes
		"customField1": "custom-value-1",
		"customField2": "custom-value-2",
	}
	ctx = revenium.WithUsageMetadata(ctx, metadata)

	// Make API call
	fmt.Println("Sending request with ALL metadata fields...")
	params := anthropic.MessageNewParams{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 100,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewTextBlock("Say 'E2E validation successful' in exactly 5 words."),
			),
		},
	}

	resp, err := client.Messages().CreateMessage(ctx, params)
	if err != nil {
		log.Fatalf("Failed to create message: %v", err)
	}

	// Print response
	fmt.Printf("\nResponse: ")
	if len(resp.Content) > 0 {
		contentValue := reflect.ValueOf(resp.Content[0])
		if contentValue.Kind() == reflect.Ptr {
			contentValue = contentValue.Elem()
		}
		textField := contentValue.FieldByName("Text")
		if textField.IsValid() {
			if text, ok := textField.Interface().(string); ok {
				fmt.Printf("%s", text)
			}
		}
	}
	fmt.Println()

	// Print usage
	fmt.Println("\n=== Token Usage ===")
	fmt.Printf("  Input tokens:  %d\n", resp.Usage.InputTokens)
	fmt.Printf("  Output tokens: %d\n", resp.Usage.OutputTokens)

	// Output verification info
	fmt.Println("\n=== Verification Info ===")
	fmt.Printf("Trace ID:       %s\n", traceID)
	fmt.Printf("Transaction ID: %s\n", transactionID)
	// Use REVENIUM_UI_URL env var or default to production
	uiURL := os.Getenv("REVENIUM_UI_URL")
	if uiURL == "" {
		uiURL = "https://app.revenium.ai"
	}
	fmt.Printf("\nTrace URL: %s/traces?traceId=%s\n", uiURL, traceID)

	// Print all metadata sent for comparison
	fmt.Println("\n=== Metadata Sent (for verification) ===")
	fmt.Println("organizationId:      e2e-validation-org")
	fmt.Println("productId:           e2e-validation-product")
	fmt.Println("subscriptionId:      e2e-sub-premium")
	fmt.Println("agent:               e2e-validation-agent")
	fmt.Println("taskType:            e2e-validation")
	fmt.Println("traceType:           e2e-validation-trace")
	fmt.Println("traceName:           Anthropic-Go-E2E-Test")
	fmt.Println("environment:         qa-e2e")
	fmt.Println("region:              us-east-1")
	fmt.Println("operationType:       chat-completion")
	fmt.Println("operationSubtype:    e2e-test")
	fmt.Println("retryNumber:         0")
	fmt.Println("credentialAlias:     anthropic-e2e-key")
	fmt.Println("parentTransactionId: parent-e2e-001")
	fmt.Println("subscriber.id:       e2e-user-123")
	fmt.Println("subscriber.email:    e2e-test@revenium.io")

	// Wait for async metering to complete
	fmt.Println("\nWaiting 3 seconds for metering to complete...")
	time.Sleep(3 * time.Second)

	fmt.Println("\n=== E2E Test Complete ===")
	fmt.Println("Use the verification API to confirm data received matches this output.")
}

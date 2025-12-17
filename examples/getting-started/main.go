// Getting Started with Revenium Anthropic Middleware
//
// This is the simplest example to verify your setup is working.

package main

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/revenium/revenium-middleware-anthropic-go/revenium"
)

func main() {
	// Initialize middleware (automatically uses environment variables)
	if err := revenium.Initialize(); err != nil {
		log.Fatalf("Failed to initialize middleware: %v", err)
	}

	// Get the wrapped client
	client, err := revenium.GetClient()
	if err != nil {
		log.Fatalf("Failed to get client: %v", err)
	}

	fmt.Println("Testing Anthropic with Revenium tracking...")

	// Create context with usage metadata
	// All supported metadata fields shown below (uncomment as needed)
	ctx := context.Background()
	metadata := map[string]interface{}{
		// Required/Common fields
		"organizationId": "org-getting-started",
		"productId":      "product-getting-started",
		"taskType":       "text-generation",

		// Optional: Subscription and agent tracking
		// "subscriptionId": "sub-premium-tier",
		// "agent":          "my-agent-name",

		// Optional: Distributed tracing
		// "traceId": "trace-abc123-def456",

		// Optional: Quality scoring (0.0-1.0 scale)
		// "responseQualityScore": 0.95,

		// Optional: Subscriber details (for user attribution)
		// "subscriber": map[string]interface{}{
		// 	"id":    "user-123",
		// 	"email": "user@example.com",
		// 	"credential": map[string]interface{}{
		// 		"name":  "API Key Name",
		// 		"value": "key-identifier",
		// 	},
		// },
	}
	ctx = revenium.WithUsageMetadata(ctx, metadata)

	// Simple chat completion
	params := anthropic.MessageNewParams{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 100,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewTextBlock("Please verify you are ready to assist me."),
			),
		},
	}

	resp, err := client.Messages().CreateMessage(ctx, params)
	if err != nil {
		log.Fatalf("Failed to create message: %v", err)
	}

	// Extract and print response text
	if len(resp.Content) > 0 {
		contentValue := reflect.ValueOf(resp.Content[0])
		if contentValue.Kind() == reflect.Ptr {
			contentValue = contentValue.Elem()
		}
		textField := contentValue.FieldByName("Text")
		if textField.IsValid() {
			if text, ok := textField.Interface().(string); ok {
				fmt.Println("Response:", text)
			}
		}
	}

	fmt.Println("\nUsage:")
	fmt.Printf("  Input tokens: %d\n", resp.Usage.InputTokens)
	fmt.Printf("  Output tokens: %d\n", resp.Usage.OutputTokens)

	// Allow time for metering to complete
	fmt.Println("\nWaiting for metering to complete...")
	time.Sleep(2 * time.Second)

	fmt.Println("Tracking successful! Check your Revenium dashboard.")
}

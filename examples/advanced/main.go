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
	fmt.Println("=== Revenium Middleware - Advanced Example ===")
	fmt.Println()

	// Initialize the middleware
	if err := revenium.Initialize(); err != nil {
		log.Fatalf("Failed to initialize middleware: %v", err)
	}

	// Get the client
	client, err := revenium.GetClient()
	if err != nil {
		log.Fatalf("Failed to get client: %v", err)
	}

	// Create context with custom metadata
	ctx := context.Background()
	metadata := map[string]interface{}{
		"organizationId": "org-123",
		"productId":      "product-456",
		"subscriber": map[string]interface{}{
			"id":    "user-789",
			"email": "user@example.com",
		},
		"taskType": "chat-completion",
	}
	ctx = revenium.WithUsageMetadata(ctx, metadata)

	// Create message
	params := anthropic.MessageNewParams{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 300,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewTextBlock("Explain the benefits of using middleware for API metering"),
			),
		},
	}

	resp, err := client.Messages().CreateMessage(ctx, params)
	if err != nil {
		log.Fatalf("Failed to create message: %v", err)
	}

	// Display response
	fmt.Println("Response:")
	fmt.Println("─────────────────────────────────────────")

	if len(resp.Content) > 0 {
		contentValue := reflect.ValueOf(resp.Content[0])
		if contentValue.Kind() == reflect.Ptr {
			contentValue = contentValue.Elem()
		}

		textField := contentValue.FieldByName("Text")
		if textField.IsValid() {
			if text, ok := textField.Interface().(string); ok {
				fmt.Println(text)
			}
		}
	}

	fmt.Println("─────────────────────────────────────────")
	fmt.Println()

	fmt.Printf("Input Tokens: %d\n", resp.Usage.InputTokens)
	fmt.Printf("Output Tokens: %d\n", resp.Usage.OutputTokens)

	// Wait for metering to complete
	time.Sleep(2 * time.Second)

	fmt.Println("\nAdvanced example completed successfully!")
}

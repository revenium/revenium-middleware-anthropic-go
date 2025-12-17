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
	fmt.Println("=== Revenium Middleware - Streaming Example ===")
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
		"organizationId": "org-streaming-example",
		"productId":      "product-streaming",
		"subscriber": map[string]interface{}{
			"id":    "user-streaming",
			"email": "streaming@example.com",
		},
		"taskType": "streaming-chat",
	}
	ctx = revenium.WithUsageMetadata(ctx, metadata)

	// Create streaming message
	params := anthropic.MessageNewParams{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 200,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewTextBlock("Write a haiku about programming"),
			),
		},
	}

	// Start streaming
	streamInterface, err := client.Messages().CreateMessageStream(ctx, params)
	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}

	// Cast to streaming wrapper
	stream, ok := streamInterface.(*revenium.StreamingWrapper)
	if !ok {
		log.Fatalf("Invalid stream type")
	}

	// Display streaming response
	fmt.Println("Streaming response:")
	fmt.Println("──────────")

	for stream.Next() {
		event := stream.Current()
		if event == nil {
			continue
		}

		text := extractTextFromEvent(event)
		if text != "" {
			fmt.Print(text)
		}
	}

	// Handle errors and close
	if err := stream.Err(); err != nil {
		log.Printf("Stream error: %v", err)
	}

	if err := stream.Close(); err != nil {
		log.Printf("Close error: %v", err)
	}

	fmt.Println()
	fmt.Println()

	// Wait for metering to complete
	time.Sleep(2 * time.Second)

	fmt.Println("\nStreaming example completed successfully!")
}

// Helper function to extract text from stream events
func extractTextFromEvent(event interface{}) string {
	if event == nil {
		return ""
	}

	eventValue := reflect.ValueOf(event)
	if eventValue.Kind() == reflect.Ptr {
		eventValue = eventValue.Elem()
	}

	deltaField := eventValue.FieldByName("Delta")
	if deltaField.IsValid() && !deltaField.IsZero() {
		deltaValue := deltaField.Interface()
		if deltaValue == nil {
			return ""
		}

		deltaReflect := reflect.ValueOf(deltaValue)
		if deltaReflect.Kind() == reflect.Ptr {
			deltaReflect = deltaReflect.Elem()
		}

		textField := deltaReflect.FieldByName("Text")
		if textField.IsValid() {
			if text, ok := textField.Interface().(string); ok {
				return text
			}
		}
	}

	return ""
}

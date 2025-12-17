package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/revenium/revenium-middleware-anthropic-go/revenium"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationBasicMessage tests basic message creation end-to-end
func TestIntegrationBasicMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Only set environment variables if they don't already exist (from .env)
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping integration test: ANTHROPIC_API_KEY not set in .env")
	}
	if os.Getenv("REVENIUM_METERING_API_KEY") == "" {
		t.Skip("Skipping integration test: REVENIUM_METERING_API_KEY not set in .env")
	}

	// Initialize middleware
	err := revenium.Initialize()
	require.NoError(t, err, "Middleware initialization should succeed")

	client, err := revenium.GetClient()
	require.NoError(t, err, "Getting client should succeed")
	defer func() {
		err := client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
	}()

	t.Run("Basic message without metadata", func(t *testing.T) {
		ctx := context.Background()

		params := anthropic.MessageNewParams{
			Model:     anthropic.ModelClaude3_7SonnetLatest, // Latest Claude Sonnet 4.5
			MaxTokens: 100,
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(
					anthropic.NewTextBlock("Say hello in one word"),
				),
			},
		}

		startTime := time.Now()
		resp, err := client.Messages().CreateMessage(ctx, params)
		duration := time.Since(startTime)

		require.NoError(t, err, "Message creation should succeed")
		assert.NotNil(t, resp, "Response should not be nil")
		assert.Greater(t, resp.Usage.InputTokens, int64(0), "Input tokens should be greater than 0")
		assert.Greater(t, resp.Usage.OutputTokens, int64(0), "Output tokens should be greater than 0")
		assert.Less(t, duration, 30*time.Second, "Response should be received within 30 seconds")
	})

	t.Run("Message with comprehensive metadata", func(t *testing.T) {
		ctx := context.Background()

		metadata := map[string]interface{}{
			"organizationId": "integration-test-org",
			"productId":      "integration-test-product",
			"taskType":       "integration-test",
			"taskId":         "test-task-001",
			"agent":          "integration-test-agent",
			"subscriptionId": "test-subscription",
			"traceId":        "test-trace-123",
			"subscriber": map[string]interface{}{
				"id":    "test-user-123",
				"email": "test@integration.com",
				"credential": map[string]interface{}{
					"name":  "test-api-key",
					"value": "test-key-value",
				},
			},
			"responseQualityScore": 0.95,
		}

		ctxWithMetadata := revenium.WithUsageMetadata(ctx, metadata)

		params := anthropic.MessageNewParams{
			Model:     anthropic.ModelClaude3_7SonnetLatest, // Latest Claude Sonnet 4.5
			MaxTokens: 150,
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(
					anthropic.NewTextBlock("Explain what integration testing is in one sentence"),
				),
			},
		}

		startTime := time.Now()
		resp, err := client.Messages().CreateMessage(ctxWithMetadata, params)
		duration := time.Since(startTime)

		require.NoError(t, err, "Message creation with metadata should succeed")
		assert.NotNil(t, resp, "Response should not be nil")
		assert.Greater(t, resp.Usage.InputTokens, int64(0), "Input tokens should be greater than 0")
		assert.Greater(t, resp.Usage.OutputTokens, int64(0), "Output tokens should be greater than 0")
		assert.Less(t, duration, 30*time.Second, "Response should be received within 30 seconds")

		// Verify metadata is preserved
		extractedMetadata := revenium.GetUsageMetadata(ctxWithMetadata)
		assert.Equal(t, "integration-test-org", extractedMetadata["organizationId"])
		assert.Equal(t, "integration-test-product", extractedMetadata["productId"])
	})
}

// TestIntegrationStreamingMessage tests streaming message creation end-to-end
func TestIntegrationStreamingMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Only run if environment variables are set (from .env)
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping integration test: ANTHROPIC_API_KEY not set in .env")
	}
	if os.Getenv("REVENIUM_METERING_API_KEY") == "" {
		t.Skip("Skipping integration test: REVENIUM_METERING_API_KEY not set in .env")
	}

	// Initialize middleware
	err := revenium.Initialize()
	require.NoError(t, err, "Middleware initialization should succeed")

	client, err := revenium.GetClient()
	require.NoError(t, err, "Getting client should succeed")
	defer func() {
		err := client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
	}()

	t.Run("Basic streaming message", func(t *testing.T) {
		ctx := context.Background()

		params := anthropic.MessageNewParams{
			Model:     anthropic.ModelClaude3_7SonnetLatest, // Latest Claude Sonnet 4.5
			MaxTokens: 100,
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(
					anthropic.NewTextBlock("Count from 1 to 5"),
				),
			},
		}

		startTime := time.Now()
		streamInterface, err := client.Messages().CreateMessageStream(ctx, params)
		require.NoError(t, err, "Stream creation should succeed")

		stream, ok := streamInterface.(*revenium.StreamingWrapper)
		require.True(t, ok, "Stream should be a StreamingWrapper")

		tokenCount := 0
		var firstTokenTime time.Time

		for stream.Next() {
			event := stream.Current()
			assert.NotNil(t, event, "Stream event should not be nil")

			if tokenCount == 0 {
				firstTokenTime = time.Now()
			}
			tokenCount++
		}

		err = stream.Err()
		assert.NoError(t, err, "Stream should not have errors")

		err = stream.Close()
		assert.NoError(t, err, "Stream close should succeed")

		totalTime := time.Since(startTime)
		assert.Greater(t, tokenCount, 0, "Should receive at least one token")
		assert.Less(t, totalTime, 30*time.Second, "Streaming should complete within 30 seconds")

		if !firstTokenTime.IsZero() {
			timeToFirstToken := firstTokenTime.Sub(startTime)
			assert.Less(t, timeToFirstToken, 10*time.Second, "First token should arrive within 10 seconds")
		}
	})

	t.Run("Streaming message with metadata", func(t *testing.T) {
		ctx := context.Background()

		metadata := map[string]interface{}{
			"organizationId": "streaming-test-org",
			"productId":      "streaming-test-product",
			"taskType":       "streaming-test",
			"agent":          "streaming-test-agent",
			"subscriber": map[string]interface{}{
				"id":    "streaming-user",
				"email": "streaming@test.com",
				"credential": map[string]interface{}{
					"name":  "streaming-key",
					"value": "streaming-value",
				},
			},
		}

		ctxWithMetadata := revenium.WithUsageMetadata(ctx, metadata)

		params := anthropic.MessageNewParams{
			Model:     anthropic.ModelClaude3_7SonnetLatest, // Latest Claude Sonnet 4.5
			MaxTokens: 150,
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(
					anthropic.NewTextBlock("Write a haiku about testing"),
				),
			},
		}

		startTime := time.Now()
		streamInterface, err := client.Messages().CreateMessageStream(ctxWithMetadata, params)
		require.NoError(t, err, "Stream creation with metadata should succeed")

		stream, ok := streamInterface.(*revenium.StreamingWrapper)
		require.True(t, ok, "Stream should be a StreamingWrapper")

		tokenCount := 0
		for stream.Next() {
			event := stream.Current()
			assert.NotNil(t, event, "Stream event should not be nil")
			tokenCount++
		}

		err = stream.Err()
		assert.NoError(t, err, "Stream should not have errors")

		err = stream.Close()
		assert.NoError(t, err, "Stream close should succeed")

		totalTime := time.Since(startTime)
		assert.Greater(t, tokenCount, 0, "Should receive at least one token")
		assert.Less(t, totalTime, 30*time.Second, "Streaming should complete within 30 seconds")

		// Verify metadata is preserved
		extractedMetadata := revenium.GetUsageMetadata(ctxWithMetadata)
		assert.Equal(t, "streaming-test-org", extractedMetadata["organizationId"])
		assert.Equal(t, "streaming-test-product", extractedMetadata["productId"])
	})
}

// TestIntegrationBedrockFallback tests Bedrock fallback behavior
func TestIntegrationBedrockFallback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Only run if environment variables are set (from .env)
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping integration test: ANTHROPIC_API_KEY not set in .env")
	}
	if os.Getenv("REVENIUM_METERING_API_KEY") == "" {
		t.Skip("Skipping integration test: REVENIUM_METERING_API_KEY not set in .env")
	}

	// Set invalid AWS credentials to force fallback
	os.Setenv("AWS_ACCESS_KEY_ID", "invalid-key")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "invalid-secret")
	os.Setenv("AWS_REGION", "us-east-1")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Unsetenv("AWS_REGION")
	}()

	// Initialize middleware
	err := revenium.Initialize()
	require.NoError(t, err, "Middleware initialization should succeed")

	client, err := revenium.GetClient()
	require.NoError(t, err, "Getting client should succeed")
	defer func() {
		err := client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
	}()

	t.Run("Fallback to Anthropic with invalid AWS credentials", func(t *testing.T) {
		ctx := context.Background()

		params := anthropic.MessageNewParams{
			Model:     anthropic.ModelClaude3_7SonnetLatest, // Latest Claude Sonnet 4.5
			MaxTokens: 50,
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(
					anthropic.NewTextBlock("Say 'fallback test successful'"),
				),
			},
		}

		startTime := time.Now()
		resp, err := client.Messages().CreateMessage(ctx, params)
		duration := time.Since(startTime)

		// Should succeed via fallback to Anthropic
		require.NoError(t, err, "Message creation should succeed via fallback")
		assert.NotNil(t, resp, "Response should not be nil")
		assert.Greater(t, resp.Usage.InputTokens, int64(0), "Input tokens should be greater than 0")
		assert.Greater(t, resp.Usage.OutputTokens, int64(0), "Output tokens should be greater than 0")
		assert.Less(t, duration, 30*time.Second, "Response should be received within 30 seconds")
	})
}

// TestIntegrationErrorHandling tests error handling scenarios
func TestIntegrationErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Invalid API key", func(t *testing.T) {
		// Save original environment variables
		origAnthropicKey := os.Getenv("ANTHROPIC_API_KEY")
		origReveniumKey := os.Getenv("REVENIUM_METERING_API_KEY")

		// Set invalid API keys
		os.Setenv("REVENIUM_METERING_API_KEY", "hak_invalid_key")
		os.Setenv("ANTHROPIC_API_KEY", "sk-ant-invalid-key")
		defer func() {
			// Restore original values
			os.Setenv("ANTHROPIC_API_KEY", origAnthropicKey)
			os.Setenv("REVENIUM_METERING_API_KEY", origReveniumKey)
		}()

		err := revenium.Initialize()
		require.NoError(t, err, "Initialization should succeed even with invalid keys")

		client, err := revenium.GetClient()
		require.NoError(t, err, "Getting client should succeed")

		ctx := context.Background()
		params := anthropic.MessageNewParams{
			Model:     anthropic.ModelClaude3_7SonnetLatest, // Latest Claude Sonnet 4.5
			MaxTokens: 50,
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(
					anthropic.NewTextBlock("Test message"),
				),
			},
		}

		// This should fail due to invalid Anthropic API key
		_, err = client.Messages().CreateMessage(ctx, params)
		assert.Error(t, err, "Message creation should fail with invalid API key")

		err = client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
	})

	t.Run("Invalid model", func(t *testing.T) {
		// Use real API keys from .env for this test
		// (we're testing invalid model, not invalid API key)

		err := revenium.Initialize()
		require.NoError(t, err, "Initialization should succeed")

		client, err := revenium.GetClient()
		require.NoError(t, err, "Getting client should succeed")
		defer func() {
			err := client.Close()
			assert.NoError(t, err, "Client cleanup should succeed")
		}()

		ctx := context.Background()
		params := anthropic.MessageNewParams{
			Model:     "invalid-model-name",
			MaxTokens: 50,
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(
					anthropic.NewTextBlock("Test message"),
				),
			},
		}

		// This should fail due to invalid model
		_, err = client.Messages().CreateMessage(ctx, params)
		assert.Error(t, err, "Message creation should fail with invalid model")
	})
}

// TestIntegrationConcurrency tests concurrent usage
func TestIntegrationConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Only run if environment variables are set (from .env)
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping integration test: ANTHROPIC_API_KEY not set in .env")
	}
	if os.Getenv("REVENIUM_METERING_API_KEY") == "" {
		t.Skip("Skipping integration test: REVENIUM_METERING_API_KEY not set in .env")
	}

	// Initialize middleware
	err := revenium.Initialize()
	require.NoError(t, err, "Middleware initialization should succeed")

	client, err := revenium.GetClient()
	require.NoError(t, err, "Getting client should succeed")
	defer func() {
		err := client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
	}()

	t.Run("Concurrent message creation", func(t *testing.T) {
		const numGoroutines = 5
		done := make(chan bool, numGoroutines)
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				ctx := context.Background()

				metadata := map[string]interface{}{
					"organizationId": "concurrent-test-org",
					"taskId":         fmt.Sprintf("concurrent-task-%d", id),
				}
				ctxWithMetadata := revenium.WithUsageMetadata(ctx, metadata)

				params := anthropic.MessageNewParams{
					Model:     anthropic.ModelClaude3_7SonnetLatest, // Latest Claude Sonnet 4.5
					MaxTokens: 50,
					Messages: []anthropic.MessageParam{
						anthropic.NewUserMessage(
							anthropic.NewTextBlock(fmt.Sprintf("Concurrent test %d", id)),
						),
					},
				}

				_, err := client.Messages().CreateMessage(ctxWithMetadata, params)
				errors <- err
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Check results
		successCount := 0
		for i := 0; i < numGoroutines; i++ {
			err := <-errors
			if err == nil {
				successCount++
			} else {
				t.Logf("Goroutine error: %v", err)
			}
		}

		// At least some should succeed (allowing for rate limiting)
		assert.Greater(t, successCount, 0, "At least one concurrent request should succeed")
	})
}

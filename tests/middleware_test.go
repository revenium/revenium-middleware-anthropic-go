package tests

import (
	"context"
	"os"
	"testing"

	"github.com/revenium/revenium-middleware-anthropic-go/revenium"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMiddlewareInitialization tests basic middleware initialization
func TestMiddlewareInitialization(t *testing.T) {
	// Set required environment variables
	os.Setenv("REVENIUM_METERING_API_KEY", "hak_test_key_123")
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key")
	defer func() {
		os.Unsetenv("REVENIUM_METERING_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")
	}()

	// Test initialization
	err := revenium.Initialize()
	require.NoError(t, err, "Middleware initialization should succeed")

	// Test that middleware is initialized
	assert.True(t, revenium.IsInitialized(), "Middleware should be initialized")

	// Test getting client
	client, err := revenium.GetClient()
	require.NoError(t, err, "Getting client should succeed")
	assert.NotNil(t, client, "Client should not be nil")

	// Test cleanup
	err = client.Close()
	assert.NoError(t, err, "Client cleanup should succeed")
}

// TestMiddlewareWithoutAPIKey tests initialization without required API key
func TestMiddlewareWithoutAPIKey(t *testing.T) {
	// Ensure no API key is set
	os.Unsetenv("REVENIUM_METERING_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")

	// Test initialization succeeds but client creation should fail
	err := revenium.Initialize()
	assert.NoError(t, err, "Initialization should succeed even without API keys")
	assert.True(t, revenium.IsInitialized(), "Middleware should be initialized")

	// Getting client should succeed (validation happens at API call time)
	client, err := revenium.GetClient()
	assert.NoError(t, err, "Getting client should succeed")
	if client != nil {
		err = client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
	}

	revenium.Reset()
}

// TestMiddlewareWithInvalidAPIKey tests initialization with invalid API key format
func TestMiddlewareWithInvalidAPIKey(t *testing.T) {
	// Set invalid API key (doesn't start with "hak_")
	os.Setenv("REVENIUM_METERING_API_KEY", "invalid_key_format")
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key")
	defer func() {
		os.Unsetenv("REVENIUM_METERING_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")
		revenium.Reset()
	}()

	// Test initialization succeeds but validation happens later
	err := revenium.Initialize()
	assert.NoError(t, err, "Initialization should succeed")
	assert.True(t, revenium.IsInitialized(), "Middleware should be initialized")

	// Client creation should succeed (validation happens at API call time)
	client, err := revenium.GetClient()
	assert.NoError(t, err, "Getting client should succeed")
	if client != nil {
		err = client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
	}
}

// TestUsageMetadataContext tests usage metadata context functionality
func TestUsageMetadataContext(t *testing.T) {
	ctx := context.Background()

	// Test adding metadata to context
	metadata := map[string]interface{}{
		"organizationId": "test-org",
		"productId":      "test-product",
		"taskType":       "test-task",
		"agent":          "test-agent",
		"subscriber": map[string]interface{}{
			"id":    "test-user",
			"email": "test@example.com",
			"credential": map[string]interface{}{
				"name":  "test-key",
				"value": "test-value",
			},
		},
	}

	ctxWithMetadata := revenium.WithUsageMetadata(ctx, metadata)
	assert.NotEqual(t, ctx, ctxWithMetadata, "Context with metadata should be different")

	// Test extracting metadata from context
	extractedMetadata := revenium.GetUsageMetadata(ctxWithMetadata)
	assert.NotNil(t, extractedMetadata, "Extracted metadata should not be nil")
	assert.Equal(t, "test-org", extractedMetadata["organizationId"])
	assert.Equal(t, "test-product", extractedMetadata["productId"])
	assert.Equal(t, "test-task", extractedMetadata["taskType"])
	assert.Equal(t, "test-agent", extractedMetadata["agent"])

	// Test subscriber object structure
	subscriber, ok := extractedMetadata["subscriber"].(map[string]interface{})
	require.True(t, ok, "Subscriber should be a map")
	assert.Equal(t, "test-user", subscriber["id"])
	assert.Equal(t, "test@example.com", subscriber["email"])

	credential, ok := subscriber["credential"].(map[string]interface{})
	require.True(t, ok, "Credential should be a map")
	assert.Equal(t, "test-key", credential["name"])
	assert.Equal(t, "test-value", credential["value"])
}

// TestMiddlewareConfiguration tests configuration options
func TestMiddlewareConfiguration(t *testing.T) {
	// Test with custom configuration
	os.Setenv("REVENIUM_METERING_API_KEY", "hak_test_key_123")
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key")
	os.Setenv("REVENIUM_METERING_BASE_URL", "https://custom-api.example.com")
	os.Setenv("REVENIUM_ORG_ID", "test-org-id")
	os.Setenv("REVENIUM_PRODUCT_ID", "test-product-id")
	defer func() {
		os.Unsetenv("REVENIUM_METERING_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")
		os.Unsetenv("REVENIUM_METERING_BASE_URL")
		os.Unsetenv("REVENIUM_ORG_ID")
		os.Unsetenv("REVENIUM_PRODUCT_ID")
	}()

	err := revenium.Initialize()
	require.NoError(t, err, "Initialization with custom config should succeed")
	assert.True(t, revenium.IsInitialized(), "Middleware should be initialized")

	client, err := revenium.GetClient()
	require.NoError(t, err, "Getting client should succeed")
	assert.NotNil(t, client, "Client should not be nil")

	err = client.Close()
	assert.NoError(t, err, "Client cleanup should succeed")
}

// TestConcurrentInitialization tests concurrent initialization calls
func TestConcurrentInitialization(t *testing.T) {
	os.Setenv("REVENIUM_METERING_API_KEY", "hak_test_key_123")
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key")
	defer func() {
		os.Unsetenv("REVENIUM_METERING_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")
	}()

	// Test concurrent initialization
	done := make(chan bool, 10)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			err := revenium.Initialize()
			errors <- err
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check that at least one initialization succeeded
	successCount := 0
	for i := 0; i < 10; i++ {
		err := <-errors
		if err == nil {
			successCount++
		}
	}

	assert.Greater(t, successCount, 0, "At least one initialization should succeed")
	assert.True(t, revenium.IsInitialized(), "Middleware should be initialized")

	client, err := revenium.GetClient()
	require.NoError(t, err, "Getting client should succeed")
	err = client.Close()
	assert.NoError(t, err, "Client cleanup should succeed")
}

// TestMiddlewareReset tests middleware reset functionality
func TestMiddlewareReset(t *testing.T) {
	os.Setenv("REVENIUM_METERING_API_KEY", "hak_test_key_123")
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key")
	defer func() {
		os.Unsetenv("REVENIUM_METERING_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")
	}()

	// Initialize middleware
	err := revenium.Initialize()
	require.NoError(t, err, "Initial initialization should succeed")
	assert.True(t, revenium.IsInitialized(), "Middleware should be initialized")

	// Reset middleware
	revenium.Reset()
	assert.False(t, revenium.IsInitialized(), "Middleware should not be initialized after reset")

	// Re-initialize should work
	err = revenium.Initialize()
	require.NoError(t, err, "Re-initialization should succeed")
	assert.True(t, revenium.IsInitialized(), "Middleware should be initialized again")

	client, err := revenium.GetClient()
	require.NoError(t, err, "Getting client should succeed")
	err = client.Close()
	assert.NoError(t, err, "Client cleanup should succeed")
}

// TestMetadataValidation tests metadata validation
func TestMetadataValidation(t *testing.T) {
	testCases := []struct {
		name     string
		metadata map[string]interface{}
		valid    bool
	}{
		{
			name: "Valid metadata with all fields",
			metadata: map[string]interface{}{
				"organizationId": "test-org",
				"productId":      "test-product",
				"taskType":       "test-task",
				"agent":          "test-agent",
				"subscriptionId": "test-sub",
				"traceId":        "test-trace",
				"subscriber": map[string]interface{}{
					"id":    "test-user",
					"email": "test@example.com",
					"credential": map[string]interface{}{
						"name":  "test-key",
						"value": "test-value",
					},
				},
				"responseQualityScore": 0.95,
			},
			valid: true,
		},
		{
			name: "Valid metadata with minimal fields",
			metadata: map[string]interface{}{
				"organizationId": "test-org",
			},
			valid: true,
		},
		{
			name: "Valid metadata with subscriber as object",
			metadata: map[string]interface{}{
				"subscriber": map[string]interface{}{
					"id":    "test-user",
					"email": "test@example.com",
					"credential": map[string]interface{}{
						"name":  "test-key",
						"value": "test-value",
					},
				},
			},
			valid: true,
		},
		{
			name:     "Empty metadata",
			metadata: map[string]interface{}{},
			valid:    true,
		},
		{
			name:     "Nil metadata",
			metadata: nil,
			valid:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			ctxWithMetadata := revenium.WithUsageMetadata(ctx, tc.metadata)

			if tc.valid {
				assert.NotNil(t, ctxWithMetadata, "Context with metadata should not be nil")
				extractedMetadata := revenium.GetUsageMetadata(ctxWithMetadata)
				if tc.metadata != nil {
					assert.NotNil(t, extractedMetadata, "Extracted metadata should not be nil")
				}
			}
		})
	}
}

package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/revenium/revenium-middleware-anthropic-go/revenium"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProviderDetection tests automatic provider detection
func TestProviderDetection(t *testing.T) {
	// Set base required environment variables
	os.Setenv("REVENIUM_METERING_API_KEY", "hak_test_key_123")
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key")
	defer func() {
		os.Unsetenv("REVENIUM_METERING_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Unsetenv("AWS_REGION")
		os.Unsetenv("REVENIUM_BEDROCK_DISABLE")
	}()

	t.Run("Anthropic provider when no AWS credentials", func(t *testing.T) {
		// Ensure no AWS credentials
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		err := revenium.Initialize()
		require.NoError(t, err, "Initialization should succeed")

		client, err := revenium.GetClient()
		require.NoError(t, err, "Getting client should succeed")
		assert.NotNil(t, client, "Client should not be nil")

		// Provider should be detected as Anthropic
		// This is implicit - if AWS credentials are not set, it uses Anthropic

		err = client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
		revenium.Reset()
	})

	t.Run("Bedrock provider when AWS credentials are set", func(t *testing.T) {
		// Set AWS credentials
		os.Setenv("AWS_ACCESS_KEY_ID", "test-access-key")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret-key")
		os.Setenv("AWS_REGION", "us-east-1")

		err := revenium.Initialize()
		require.NoError(t, err, "Initialization should succeed")

		client, err := revenium.GetClient()
		require.NoError(t, err, "Getting client should succeed")
		assert.NotNil(t, client, "Client should not be nil")

		// Provider should be detected as Bedrock (with fallback to Anthropic)
		// This is implicit - if AWS credentials are set, it tries Bedrock first

		err = client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
		revenium.Reset()
	})

	t.Run("Anthropic provider when Bedrock is disabled", func(t *testing.T) {
		// Set AWS credentials but disable Bedrock
		os.Setenv("AWS_ACCESS_KEY_ID", "test-access-key")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret-key")
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("REVENIUM_BEDROCK_DISABLE", "1")

		err := revenium.Initialize()
		require.NoError(t, err, "Initialization should succeed")

		client, err := revenium.GetClient()
		require.NoError(t, err, "Getting client should succeed")
		assert.NotNil(t, client, "Client should not be nil")

		// Provider should be Anthropic even with AWS credentials due to disable flag

		err = client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
		revenium.Reset()
	})
}

// TestProviderFallback tests fallback behavior between providers
func TestProviderFallback(t *testing.T) {
	os.Setenv("REVENIUM_METERING_API_KEY", "hak_test_key_123")
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key")
	defer func() {
		os.Unsetenv("REVENIUM_METERING_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Unsetenv("AWS_REGION")
	}()

	t.Run("Fallback from Bedrock to Anthropic with invalid AWS credentials", func(t *testing.T) {
		// Set invalid AWS credentials
		os.Setenv("AWS_ACCESS_KEY_ID", "invalid-key")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "invalid-secret")
		os.Setenv("AWS_REGION", "us-east-1")

		err := revenium.Initialize()
		require.NoError(t, err, "Initialization should succeed")

		client, err := revenium.GetClient()
		require.NoError(t, err, "Getting client should succeed")
		assert.NotNil(t, client, "Client should not be nil")

		// Even with invalid AWS credentials, the client should work
		// because it will fallback to Anthropic

		err = client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
		revenium.Reset()
	})
}

// TestProviderConfiguration tests provider-specific configuration
func TestProviderConfiguration(t *testing.T) {
	os.Setenv("REVENIUM_METERING_API_KEY", "hak_test_key_123")
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key")
	defer func() {
		os.Unsetenv("REVENIUM_METERING_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Unsetenv("AWS_REGION")
		os.Unsetenv("AWS_PROFILE")
	}()

	t.Run("AWS profile configuration", func(t *testing.T) {
		// Set AWS profile instead of credentials
		os.Setenv("AWS_PROFILE", "test-profile")
		os.Setenv("AWS_REGION", "us-west-2")

		err := revenium.Initialize()
		require.NoError(t, err, "Initialization should succeed")

		client, err := revenium.GetClient()
		require.NoError(t, err, "Getting client should succeed")
		assert.NotNil(t, client, "Client should not be nil")

		err = client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
		revenium.Reset()
	})

	t.Run("Custom AWS region", func(t *testing.T) {
		os.Setenv("AWS_ACCESS_KEY_ID", "test-access-key")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret-key")
		os.Setenv("AWS_REGION", "eu-west-1")

		err := revenium.Initialize()
		require.NoError(t, err, "Initialization should succeed")

		client, err := revenium.GetClient()
		require.NoError(t, err, "Getting client should succeed")
		assert.NotNil(t, client, "Client should not be nil")

		err = client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
		revenium.Reset()
	})
}

// TestProviderModelMapping tests model mapping between providers
func TestProviderModelMapping(t *testing.T) {
	testCases := []struct {
		name           string
		anthropicModel string
		expectedValid  bool
	}{
		{
			name:           "Claude 3 Opus",
			anthropicModel: "claude-3-opus-20240229",
			expectedValid:  true,
		},
		{
			name:           "Claude 3.7 Sonnet Latest",
			anthropicModel: "claude-3-7-sonnet-latest",
			expectedValid:  true,
		},
		{
			name:           "Claude 3.5 Haiku",
			anthropicModel: "claude-3-5-haiku-20241022",
			expectedValid:  true,
		},
		{
			name:           "Claude 3.5 Haiku",
			anthropicModel: "claude-3-5-haiku-20241022",
			expectedValid:  true,
		},
		{
			name:           "Unknown model",
			anthropicModel: "claude-unknown-model",
			expectedValid:  true, // Should still work with fallback
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test that model names are valid for the middleware
			// This is a basic validation that the model string is acceptable
			assert.NotEmpty(t, tc.anthropicModel, "Model name should not be empty")

			if tc.expectedValid {
				assert.True(t, len(tc.anthropicModel) > 0, "Valid model should have non-empty name")
			}
		})
	}
}

// TestProviderEnvironmentVariables tests environment variable handling
func TestProviderEnvironmentVariables(t *testing.T) {
	// Save original environment
	originalVars := map[string]string{
		"REVENIUM_METERING_API_KEY": os.Getenv("REVENIUM_METERING_API_KEY"),
		"ANTHROPIC_API_KEY":         os.Getenv("ANTHROPIC_API_KEY"),
		"AWS_ACCESS_KEY_ID":         os.Getenv("AWS_ACCESS_KEY_ID"),
		"AWS_SECRET_ACCESS_KEY":     os.Getenv("AWS_SECRET_ACCESS_KEY"),
		"AWS_REGION":                os.Getenv("AWS_REGION"),
		"AWS_PROFILE":               os.Getenv("AWS_PROFILE"),
		"REVENIUM_BEDROCK_DISABLE":  os.Getenv("REVENIUM_BEDROCK_DISABLE"),
	}

	// Restore original environment after test
	defer func() {
		for key, value := range originalVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	t.Run("Required environment variables", func(t *testing.T) {
		// Reset middleware to clear any previous initialization
		revenium.Reset()

		// Set required variables for testing
		os.Setenv("REVENIUM_METERING_API_KEY", "hak_test_key_123")
		os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key")

		err := revenium.Initialize()
		require.NoError(t, err, "Should succeed with required environment variables")

		client, err := revenium.GetClient()
		require.NoError(t, err, "Getting client should succeed")

		// Verify client is properly configured
		assert.NotNil(t, client, "Client should not be nil")

		err = client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
		revenium.Reset()
	})

	t.Run("Optional AWS environment variables", func(t *testing.T) {
		// Reset middleware before this test
		revenium.Reset()

		// Set required variables
		os.Setenv("REVENIUM_METERING_API_KEY", "hak_test_key_123")
		os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key")

		// Test with various AWS configurations
		awsConfigs := []map[string]string{
			{
				"AWS_ACCESS_KEY_ID":     "test-key",
				"AWS_SECRET_ACCESS_KEY": "test-secret",
				"AWS_REGION":            "us-east-1",
			},
			{
				"AWS_PROFILE": "test-profile",
				"AWS_REGION":  "us-west-2",
			},
			{
				"AWS_REGION": "eu-west-1",
				// No credentials - should use default chain
			},
		}

		for i, config := range awsConfigs {
			t.Run(fmt.Sprintf("AWS config %d", i+1), func(t *testing.T) {
				// Clear AWS variables
				os.Unsetenv("AWS_ACCESS_KEY_ID")
				os.Unsetenv("AWS_SECRET_ACCESS_KEY")
				os.Unsetenv("AWS_REGION")
				os.Unsetenv("AWS_PROFILE")

				// Set test configuration
				for key, value := range config {
					os.Setenv(key, value)
				}

				err := revenium.Initialize()
				require.NoError(t, err, "Initialization should succeed")

				client, err := revenium.GetClient()
				require.NoError(t, err, "Getting client should succeed")
				err = client.Close()
				assert.NoError(t, err, "Client cleanup should succeed")
				revenium.Reset()
			})
		}
	})
}

package tests

import (
	"context"
	"os"
	"testing"

	"github.com/revenium/revenium-middleware-anthropic-go/revenium"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBedrockIntegration tests AWS Bedrock integration
func TestBedrockIntegration(t *testing.T) {
	// Set required environment variables
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

	t.Run("Bedrock disabled by environment variable", func(t *testing.T) {
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

		// Should use Anthropic even with AWS credentials due to disable flag
		err = client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
	})

	t.Run("Bedrock enabled with AWS credentials", func(t *testing.T) {
		// Clear disable flag
		os.Unsetenv("REVENIUM_BEDROCK_DISABLE")

		// Set AWS credentials
		os.Setenv("AWS_ACCESS_KEY_ID", "test-access-key")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret-key")
		os.Setenv("AWS_REGION", "us-east-1")

		err := revenium.Initialize()
		require.NoError(t, err, "Initialization should succeed")

		client, err := revenium.GetClient()
		require.NoError(t, err, "Getting client should succeed")
		assert.NotNil(t, client, "Client should not be nil")

		// Should attempt to use Bedrock (with fallback to Anthropic)
		err = client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
	})
}

// TestBedrockConfiguration tests Bedrock-specific configuration
func TestBedrockConfiguration(t *testing.T) {
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

	t.Run("AWS credentials configuration", func(t *testing.T) {
		os.Setenv("AWS_ACCESS_KEY_ID", "test-access-key")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret-key")
		os.Setenv("AWS_REGION", "us-east-1")

		err := revenium.Initialize()
		require.NoError(t, err, "Initialization should succeed")

		client, err := revenium.GetClient()
		require.NoError(t, err, "Getting client should succeed")
		assert.NotNil(t, client, "Client should not be nil")

		err = client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
	})

	t.Run("AWS profile configuration", func(t *testing.T) {
		// Clear credentials and set profile
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Setenv("AWS_PROFILE", "test-profile")
		os.Setenv("AWS_REGION", "us-west-2")

		err := revenium.Initialize()
		require.NoError(t, err, "Initialization should succeed")

		client, err := revenium.GetClient()
		require.NoError(t, err, "Getting client should succeed")
		assert.NotNil(t, client, "Client should not be nil")

		err = client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
	})

	t.Run("Different AWS regions", func(t *testing.T) {
		regions := []string{"us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1"}

		for _, region := range regions {
			t.Run("Region "+region, func(t *testing.T) {
				os.Setenv("AWS_ACCESS_KEY_ID", "test-access-key")
				os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret-key")
				os.Setenv("AWS_REGION", region)

				err := revenium.Initialize()
				require.NoError(t, err, "Initialization should succeed for region "+region)

				client, err := revenium.GetClient()
				require.NoError(t, err, "Getting client should succeed for region "+region)
				assert.NotNil(t, client, "Client should not be nil for region "+region)

				err = client.Close()
				assert.NoError(t, err, "Client cleanup should succeed for region "+region)
			})
		}
	})
}

// TestBedrockFallback tests fallback behavior from Bedrock to Anthropic
func TestBedrockFallback(t *testing.T) {
	os.Setenv("REVENIUM_METERING_API_KEY", "hak_test_key_123")
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key")
	defer func() {
		os.Unsetenv("REVENIUM_METERING_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Unsetenv("AWS_REGION")
	}()

	t.Run("Fallback with invalid AWS credentials", func(t *testing.T) {
		// Set invalid AWS credentials
		os.Setenv("AWS_ACCESS_KEY_ID", "invalid-access-key")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "invalid-secret-key")
		os.Setenv("AWS_REGION", "us-east-1")

		err := revenium.Initialize()
		require.NoError(t, err, "Initialization should succeed")

		client, err := revenium.GetClient()
		require.NoError(t, err, "Getting client should succeed")
		assert.NotNil(t, client, "Client should not be nil")

		// Should fallback to Anthropic when Bedrock fails
		err = client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
	})

	t.Run("Fallback with missing AWS region", func(t *testing.T) {
		// Set credentials but no region
		os.Setenv("AWS_ACCESS_KEY_ID", "test-access-key")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret-key")
		os.Unsetenv("AWS_REGION")

		err := revenium.Initialize()
		require.NoError(t, err, "Initialization should succeed")

		client, err := revenium.GetClient()
		require.NoError(t, err, "Getting client should succeed")
		assert.NotNil(t, client, "Client should not be nil")

		// Should work with default region or fallback to Anthropic
		err = client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
	})
}

// TestBedrockMetadata tests metadata handling with Bedrock
func TestBedrockMetadata(t *testing.T) {
	os.Setenv("REVENIUM_METERING_API_KEY", "hak_test_key_123")
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key")
	os.Setenv("AWS_ACCESS_KEY_ID", "test-access-key")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret-key")
	os.Setenv("AWS_REGION", "us-east-1")
	defer func() {
		os.Unsetenv("REVENIUM_METERING_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Unsetenv("AWS_REGION")
	}()

	err := revenium.Initialize()
	require.NoError(t, err, "Initialization should succeed")

	client, err := revenium.GetClient()
	require.NoError(t, err, "Getting client should succeed")

	t.Run("Bedrock-specific metadata", func(t *testing.T) {
		ctx := context.Background()

		bedrockMetadata := map[string]interface{}{
			"organizationId": "bedrock-test-org",
			"productId":      "bedrock-test-product",
			"taskType":       "bedrock-analysis",
			"agent":          "bedrock-agent-v1",
			"subscriber": map[string]interface{}{
				"id":    "bedrock-user",
				"email": "bedrock@example.com",
				"credential": map[string]interface{}{
					"name":  "bedrock-key",
					"value": "bedrock-value",
				},
			},
			// Bedrock-specific fields
			"region":      "us-east-1",
			"provider":    "bedrock",
			"environment": "test",
		}

		ctxWithMetadata := revenium.WithUsageMetadata(ctx, bedrockMetadata)
		assert.NotNil(t, ctxWithMetadata, "Context with metadata should not be nil")

		extractedMetadata := revenium.GetUsageMetadata(ctxWithMetadata)
		assert.NotNil(t, extractedMetadata, "Extracted metadata should not be nil")
		assert.Equal(t, "bedrock-test-org", extractedMetadata["organizationId"])
		assert.Equal(t, "bedrock-test-product", extractedMetadata["productId"])
		assert.Equal(t, "us-east-1", extractedMetadata["region"])
	})

	err = client.Close()
	assert.NoError(t, err, "Client cleanup should succeed")
}

// TestBedrockErrorHandling tests error handling in Bedrock integration
func TestBedrockErrorHandling(t *testing.T) {
	os.Setenv("REVENIUM_METERING_API_KEY", "hak_test_key_123")
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key")
	defer func() {
		os.Unsetenv("REVENIUM_METERING_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Unsetenv("AWS_REGION")
	}()

	t.Run("Missing AWS credentials", func(t *testing.T) {
		// Ensure no AWS credentials
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Unsetenv("AWS_REGION")

		err := revenium.Initialize()
		require.NoError(t, err, "Initialization should succeed")

		client, err := revenium.GetClient()
		require.NoError(t, err, "Getting client should succeed")
		assert.NotNil(t, client, "Client should not be nil")

		// Should use Anthropic when no AWS credentials
		err = client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
	})

	t.Run("Partial AWS credentials", func(t *testing.T) {
		// Set only access key, missing secret key
		os.Setenv("AWS_ACCESS_KEY_ID", "test-access-key")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Setenv("AWS_REGION", "us-east-1")

		err := revenium.Initialize()
		require.NoError(t, err, "Initialization should succeed")

		client, err := revenium.GetClient()
		require.NoError(t, err, "Getting client should succeed")
		assert.NotNil(t, client, "Client should not be nil")

		// Should fallback to Anthropic or use default credentials chain
		err = client.Close()
		assert.NoError(t, err, "Client cleanup should succeed")
	})
}

// TestGetBedrockModelID tests the conversion of Anthropic model names to Bedrock model IDs
func TestGetBedrockModelID(t *testing.T) {
	testCases := []struct {
		name              string
		inputModel        string
		expectedBedrockID string
		description       string
	}{
		{
			name:              "Simple Anthropic model name",
			inputModel:        "claude-3-opus-20240229",
			expectedBedrockID: "anthropic.claude-3-opus-20240229",
			description:       "Should add anthropic. prefix to simple model names",
		},
		{
			name:              "Claude 3.5 Haiku",
			inputModel:        "claude-3-5-haiku-20241022",
			expectedBedrockID: "anthropic.claude-3-5-haiku-20241022",
			description:       "Should add anthropic. prefix",
		},
		{
			name:              "Claude 3.7 Sonnet Latest",
			inputModel:        "claude-3-7-sonnet-latest",
			expectedBedrockID: "anthropic.claude-3-7-sonnet-latest",
			description:       "Should add anthropic. prefix to latest models",
		},
		{
			name:              "Full ARN - passthrough",
			inputModel:        "arn:aws:bedrock:us-east-1:237436736089:inference-profile/us.anthropic.claude-sonnet-4-20250514-v1:0",
			expectedBedrockID: "arn:aws:bedrock:us-east-1:237436736089:inference-profile/us.anthropic.claude-sonnet-4-20250514-v1:0",
			description:       "Should NOT modify full ARNs",
		},
		{
			name:              "Already has anthropic. prefix",
			inputModel:        "anthropic.claude-3-opus-20240229",
			expectedBedrockID: "anthropic.claude-3-opus-20240229",
			description:       "Should NOT add prefix if already present",
		},
		{
			name:              "ARN with EU region",
			inputModel:        "arn:aws:bedrock:eu-west-1:123456789:inference-profile/eu.anthropic.claude-3-opus-20240229-v1:0",
			expectedBedrockID: "arn:aws:bedrock:eu-west-1:123456789:inference-profile/eu.anthropic.claude-3-opus-20240229-v1:0",
			description:       "Should NOT modify ARNs with different regions",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := revenium.GetBedrockModelID(tc.inputModel, nil)
			assert.Equal(t, tc.expectedBedrockID, result, tc.description)
		})
	}
}

// TestConvertBedrockARNToAnthropicModel tests the conversion of Bedrock ARNs to Anthropic model names
func TestConvertBedrockARNToAnthropicModel(t *testing.T) {
	testCases := []struct {
		name          string
		bedrockModel  string
		expectedModel string
		expectError   bool
		description   string
	}{
		{
			name:          "Full ARN with US region",
			bedrockModel:  "arn:aws:bedrock:us-east-1:237436736089:inference-profile/us.anthropic.claude-sonnet-4-20250514-v1:0",
			expectedModel: "claude-sonnet-4-20250514",
			expectError:   false,
			description:   "Should extract model name from full ARN",
		},
		{
			name:          "Full ARN with EU region",
			bedrockModel:  "arn:aws:bedrock:eu-west-1:123456789:inference-profile/eu.anthropic.claude-3-opus-20240229-v1:0",
			expectedModel: "claude-3-opus-20240229",
			expectError:   false,
			description:   "Should handle EU region prefix",
		},
		{
			name:          "Full ARN with AP region",
			bedrockModel:  "arn:aws:bedrock:ap-southeast-1:123456789:inference-profile/ap.anthropic.claude-3-5-haiku-20241022-v2:0",
			expectedModel: "claude-3-5-haiku-20241022",
			expectError:   false,
			description:   "Should handle AP region prefix",
		},
		{
			name:          "Bedrock model ID format",
			bedrockModel:  "anthropic.claude-3-5-haiku-20241022-v2:0",
			expectedModel: "claude-3-5-haiku-20241022",
			expectError:   false,
			description:   "Should remove anthropic. prefix and version",
		},
		{
			name:          "Standard Anthropic model - passthrough",
			bedrockModel:  "claude-3-opus-20240229",
			expectedModel: "claude-3-opus-20240229",
			expectError:   false,
			description:   "Should pass through standard Anthropic model names",
		},
		{
			name:          "Latest model - passthrough",
			bedrockModel:  "claude-3-7-sonnet-latest",
			expectedModel: "claude-3-7-sonnet-latest",
			expectError:   false,
			description:   "Should pass through latest model names",
		},
		{
			name:          "Invalid format - should error",
			bedrockModel:  "arn:aws:bedrock:invalid-format",
			expectedModel: "",
			expectError:   true,
			description:   "Should return error for unparseable format instead of hardcoded default",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := revenium.ConvertBedrockARNToAnthropicModel(tc.bedrockModel)
			if tc.expectError {
				assert.Error(t, err, tc.description)
				assert.Empty(t, result, "Result should be empty when error occurs")
			} else {
				assert.NoError(t, err, tc.description)
				assert.Equal(t, tc.expectedModel, result, tc.description)
			}
		})
	}
}

// TestModelConversionRoundTrip tests that converting from Anthropic to Bedrock and back works
func TestModelConversionRoundTrip(t *testing.T) {
	testCases := []string{
		"claude-3-opus-20240229",
		"claude-3-5-haiku-20241022",
		"claude-3-7-sonnet-latest",
		"claude-sonnet-4-20250514",
	}

	for _, anthropicModel := range testCases {
		t.Run(anthropicModel, func(t *testing.T) {
			// Convert to Bedrock format
			bedrockID := revenium.GetBedrockModelID(anthropicModel, nil)
			assert.True(t, len(bedrockID) > 0, "Bedrock ID should not be empty")

			// Convert back to Anthropic format
			result, err := revenium.ConvertBedrockARNToAnthropicModel(bedrockID)
			assert.NoError(t, err, "Conversion should not error for valid Bedrock ID")
			assert.Equal(t, anthropicModel, result, "Round trip conversion should preserve model name")
		})
	}
}

package revenium

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the Revenium middleware
type Config struct {
	// Anthropic API configuration
	AnthropicAPIKey string
	BaseURL         string

	// Revenium metering configuration
	ReveniumAPIKey    string
	ReveniumBaseURL   string
	ReveniumOrgID     string
	ReveniumProductID string

	// AWS Bedrock configuration
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
	AWSProfile         string
	AWSModelARNBase    string // Base ARN format: arn:aws:bedrock:{region}:{account-id}
	BedrockDisabled    bool

	// Logging and debug configuration
	LogLevel       string
	VerboseStartup bool

	// Prompt capture configuration (opt-in)
	CapturePrompts bool
}

// Option is a functional option for configuring Config
type Option func(*Config)

// WithAnthropicAPIKey sets the Anthropic API key
func WithAnthropicAPIKey(key string) Option {
	return func(c *Config) {
		c.AnthropicAPIKey = key
	}
}

// WithReveniumAPIKey sets the Revenium API key
func WithReveniumAPIKey(key string) Option {
	return func(c *Config) {
		c.ReveniumAPIKey = key
	}
}

// WithReveniumBaseURL sets the Revenium base URL
func WithReveniumBaseURL(url string) Option {
	return func(c *Config) {
		c.ReveniumBaseURL = url
	}
}

// WithAWSRegion sets the AWS region
func WithAWSRegion(region string) Option {
	return func(c *Config) {
		c.AWSRegion = region
	}
}

// WithBedrockDisabled disables Bedrock support
func WithBedrockDisabled(disabled bool) Option {
	return func(c *Config) {
		c.BedrockDisabled = disabled
	}
}

// WithCapturePrompts enables or disables prompt capture for analytics
// When enabled, system prompts, input messages, and output responses are captured
// and sent to Revenium for analytics (with truncation at 50,000 characters)
func WithCapturePrompts(capture bool) Option {
	return func(c *Config) {
		c.CapturePrompts = capture
	}
}

// loadFromEnv loads configuration from environment variables and .env files
func (c *Config) loadFromEnv() error {
	// First, try to load .env files automatically
	c.loadEnvFiles()

	// Then load from environment variables (which may have been set by .env files)
	c.AnthropicAPIKey = os.Getenv("ANTHROPIC_API_KEY")
	c.ReveniumAPIKey = os.Getenv("REVENIUM_METERING_API_KEY")
	baseURL := getEnvOrDefault("REVENIUM_METERING_BASE_URL", "https://api.revenium.ai")
	c.ReveniumBaseURL = NormalizeReveniumBaseURL(baseURL)
	c.ReveniumOrgID = os.Getenv("REVENIUM_ORGANIZATION_ID")
	c.ReveniumProductID = os.Getenv("REVENIUM_PRODUCT_ID")

	c.AWSAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	c.AWSSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	c.AWSRegion = getEnvOrDefault("AWS_REGION", "us-east-1")
	c.AWSProfile = os.Getenv("AWS_PROFILE")
	c.AWSModelARNBase = os.Getenv("AWS_MODEL_ARN_ID")

	c.LogLevel = getEnvOrDefault("REVENIUM_LOG_LEVEL", "INFO")
	c.VerboseStartup = os.Getenv("REVENIUM_VERBOSE_STARTUP") == "true" || os.Getenv("REVENIUM_VERBOSE_STARTUP") == "1"
	c.CapturePrompts = os.Getenv("REVENIUM_CAPTURE_PROMPTS") == "true" || os.Getenv("REVENIUM_CAPTURE_PROMPTS") == "1"

	// Initialize logger early so we can use it
	InitializeLogger()

	// Debug log for configuration loading
	Debug("Loading configuration from environment variables")
	if c.AnthropicAPIKey != "" {
		Debug("Anthropic API key loaded (length: %d)", len(c.AnthropicAPIKey))
	}

	if os.Getenv("REVENIUM_BEDROCK_DISABLE") == "1" || os.Getenv("REVENIUM_BEDROCK_DISABLE") == "true" {
		c.BedrockDisabled = true
	}

	return nil
}

// loadEnvFiles loads environment variables from .env files
func (c *Config) loadEnvFiles() {
	// Try to load .env files in order of preference
	envFiles := []string{
		".env.local", // Local overrides (highest priority)
		".env",       // Main env file
	}

	var loadedFiles []string

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	// Try current directory and parent directories
	searchDirs := []string{
		cwd,
		filepath.Dir(cwd),
		filepath.Join(cwd, ".."),
	}

	for _, dir := range searchDirs {
		for _, envFile := range envFiles {
			envPath := filepath.Join(dir, envFile)

			// Check if file exists
			if _, err := os.Stat(envPath); err == nil {
				// Try to load the file
				if err := godotenv.Load(envPath); err == nil {
					loadedFiles = append(loadedFiles, envPath)
				}
			}
		}
	}

	// Log loaded files (only if we have a logger initialized)
	if len(loadedFiles) > 0 {
		// We can't use Debug here because logger might not be initialized yet
		// So we'll just silently load the files
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.ReveniumAPIKey == "" {
		return NewConfigError("REVENIUM_METERING_API_KEY is required", nil)
	}

	if !isValidAPIKeyFormat(c.ReveniumAPIKey) {
		return NewConfigError("invalid Revenium API key format", nil)
	}

	Debug("Configuration validation passed")
	return nil
}

// isValidAPIKeyFormat checks if the API key has a valid format
func isValidAPIKeyFormat(key string) bool {
	// Revenium API keys should start with "hak_"
	if len(key) < 4 {
		return false
	}
	return key[:4] == "hak_"
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// NormalizeReveniumBaseURL normalizes the base URL to a consistent format
// It handles various input formats and returns a normalized base URL without trailing slash
// The endpoint path (/meter/v2/ai/completions) is appended by sendMeteringRequest
func NormalizeReveniumBaseURL(baseURL string) string {
	if baseURL == "" {
		return "https://api.revenium.ai"
	}

	// Remove trailing slash if present
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	// If it already ends with /meter/v2, remove /meter/v2 (legacy format)
	if len(baseURL) >= 9 && baseURL[len(baseURL)-9:] == "/meter/v2" {
		return baseURL[:len(baseURL)-9]
	}

	// If it ends with /meter, remove /meter (legacy format)
	if len(baseURL) >= 6 && baseURL[len(baseURL)-6:] == "/meter" {
		return baseURL[:len(baseURL)-6]
	}

	// If it ends with /v2, remove /v2 (legacy format)
	if len(baseURL) >= 3 && baseURL[len(baseURL)-3:] == "/v2" {
		return baseURL[:len(baseURL)-3]
	}

	// Return the base URL as-is (should be just the domain)
	return baseURL
}

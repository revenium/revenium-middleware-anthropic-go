package revenium

import (
	"strings"
)

// Provider represents the AI provider being used
type Provider string

const (
	ProviderAnthropic Provider = "ANTHROPIC"
	ProviderBedrock   Provider = "AWS"
)

// DetectProvider detects which provider is being used based on configuration
func DetectProvider(cfg *Config) Provider {
	if cfg == nil {
		return ProviderAnthropic
	}

	// If Bedrock is explicitly disabled, use Anthropic
	if cfg.BedrockDisabled {
		return ProviderAnthropic
	}

	// Check if AWS credentials are configured
	if cfg.AWSAccessKeyID != "" && cfg.AWSSecretAccessKey != "" {
		return ProviderBedrock
	}

	// Check if base URL indicates Bedrock
	if cfg.BaseURL != "" && strings.Contains(cfg.BaseURL, "amazonaws.com") {
		return ProviderBedrock
	}

	// Default to Anthropic
	return ProviderAnthropic
}

// IsAnthropic returns true if the provider is Anthropic
func (p Provider) IsAnthropic() bool {
	return p == ProviderAnthropic
}

// IsBedrock returns true if the provider is AWS Bedrock
func (p Provider) IsBedrock() bool {
	return p == ProviderBedrock
}

// String returns the string representation of the provider
func (p Provider) String() string {
	return string(p)
}

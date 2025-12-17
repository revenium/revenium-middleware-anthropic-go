package tests

import (
	"testing"

	"github.com/revenium/revenium-middleware-anthropic-go/revenium"
)

// TestNormalizeReveniumBaseURL tests the normalizeReveniumBaseURL function
func TestNormalizeReveniumBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty_url",
			input:    "",
			expected: "https://api.revenium.ai",
		},
		{
			name:     "base_domain_only",
			input:    "https://api.revenium.ai",
			expected: "https://api.revenium.ai",
		},
		{
			name:     "base_domain_with_trailing_slash",
			input:    "https://api.revenium.ai/",
			expected: "https://api.revenium.ai",
		},
		{
			name:     "legacy_format_with_meter",
			input:    "https://api.revenium.ai/meter",
			expected: "https://api.revenium.ai",
		},
		{
			name:     "legacy_format_with_meter_and_trailing_slash",
			input:    "https://api.revenium.ai/meter/",
			expected: "https://api.revenium.ai",
		},
		{
			name:     "legacy_format_with_meter_v2",
			input:    "https://api.revenium.ai/meter/v2",
			expected: "https://api.revenium.ai",
		},
		{
			name:     "legacy_format_with_meter_v2_and_trailing_slash",
			input:    "https://api.revenium.ai/meter/v2/",
			expected: "https://api.revenium.ai",
		},
		{
			name:     "custom_url_base_domain",
			input:    "https://custom.api.com",
			expected: "https://custom.api.com",
		},
		{
			name:     "custom_url_with_trailing_slash",
			input:    "https://custom.api.com/",
			expected: "https://custom.api.com",
		},
		{
			name:     "custom_url_legacy_format",
			input:    "https://custom.api.com/meter/v2",
			expected: "https://custom.api.com",
		},
		{
			name:     "localhost_development",
			input:    "http://localhost:8080",
			expected: "http://localhost:8080",
		},
		{
			name:     "localhost_with_trailing_slash",
			input:    "http://localhost:8080/",
			expected: "http://localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := revenium.NormalizeReveniumBaseURL(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeReveniumBaseURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

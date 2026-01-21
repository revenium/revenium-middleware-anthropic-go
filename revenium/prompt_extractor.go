package revenium

import (
	"encoding/json"
	"unicode/utf8"

	"github.com/anthropics/anthropic-sdk-go"
)

// truncateUTF8Safe truncates a string to maxBytes while preserving UTF-8 validity.
// It ensures we don't cut in the middle of a multi-byte character.
func truncateUTF8Safe(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}

	// Find the last valid UTF-8 character boundary before maxBytes
	for maxBytes > 0 && maxBytes < len(s) {
		// If this byte is a continuation byte (10xxxxxx), back up
		if (s[maxBytes] & 0xC0) == 0x80 {
			maxBytes--
		} else {
			break
		}
	}

	// Validate the result is valid UTF-8
	result := s[:maxBytes]
	if !utf8.ValidString(result) {
		// If still invalid, use rune iteration as fallback
		runes := []rune(s)
		result = ""
		for _, r := range runes {
			if len(result)+utf8.RuneLen(r) > maxBytes {
				break
			}
			result += string(r)
		}
	}

	return result
}

const (
	// MaxPromptLength is the maximum length for captured prompts/responses
	// Fields exceeding this limit will be truncated
	MaxPromptLength = 50000

	// TruncationMarker is appended to truncated content
	TruncationMarker = "...[TRUNCATED]"
)

// PromptData holds extracted prompt information for metering
type PromptData struct {
	// SystemPrompt contains the system message content (if any)
	SystemPrompt string
	// InputMessages contains JSON-serialized user/assistant messages
	InputMessages string
	// OutputResponse contains the assistant's response content
	OutputResponse string
	// PromptsTruncated indicates if any field was truncated
	PromptsTruncated bool
}

// ExtractPromptsFromParams extracts system prompt and input messages from Anthropic message params
func ExtractPromptsFromParams(params anthropic.MessageNewParams) PromptData {
	data := PromptData{}

	// Extract system prompt if present
	if len(params.System) > 0 {
		systemContent := extractSystemContent(params.System)
		if systemContent != "" {
			// Apply truncation if needed
			if len(systemContent) > MaxPromptLength {
				markerLen := len(TruncationMarker)
				truncateAt := MaxPromptLength - markerLen
				systemContent = truncateUTF8Safe(systemContent, truncateAt) + TruncationMarker
				data.PromptsTruncated = true
				Debug("System prompt truncated to %d characters", MaxPromptLength)
			}
			data.SystemPrompt = systemContent
		}
	}

	// Extract user messages
	if len(params.Messages) > 0 {
		var userMessages []map[string]interface{}
		halfLimit := MaxPromptLength / 2
		markerLen := len(TruncationMarker)

		for _, msg := range params.Messages {
			role, content := extractMessageContent(msg)
			if role != "" {
				messageMap := map[string]interface{}{
					"role":    role,
					"content": content,
				}

				// Truncate individual message content if too long
				if len(content) > halfLimit {
					truncateAt := halfLimit - markerLen
					messageMap["content"] = truncateUTF8Safe(content, truncateAt) + TruncationMarker
					data.PromptsTruncated = true
				}

				userMessages = append(userMessages, messageMap)
			}
		}

		if len(userMessages) > 0 {
			jsonBytes, err := json.Marshal(userMessages)
			if err != nil {
				Warn("Failed to serialize input messages to JSON: %v", err)
			} else {
				// Note: Individual messages are already truncated above.
				// We avoid truncating the final JSON to prevent invalid JSON.
				data.InputMessages = string(jsonBytes)
			}
		}
	}

	return data
}

// extractSystemContent extracts text from Anthropic system content blocks
func extractSystemContent(system []anthropic.TextBlockParam) string {
	if len(system) == 0 {
		return ""
	}

	// Concatenate all text blocks
	var content string
	for _, block := range system {
		if block.Text != "" {
			if content != "" {
				content += "\n"
			}
			content += block.Text
		}
	}
	return content
}

// extractMessageContent extracts role and content from an Anthropic message
func extractMessageContent(msg anthropic.MessageParam) (role string, content string) {
	role = string(msg.Role)

	// Content can be a string or an array of content blocks
	// Use JSON marshaling as a reliable approach
	jsonBytes, err := json.Marshal(msg.Content)
	if err != nil {
		return role, ""
	}

	// Try to unmarshal as string first
	var stringContent string
	if err := json.Unmarshal(jsonBytes, &stringContent); err == nil {
		return role, stringContent
	}

	// Try to unmarshal as array of content blocks
	var contentBlocks []map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &contentBlocks); err == nil {
		// Extract text from content blocks
		var textParts []string
		for _, block := range contentBlocks {
			if blockType, ok := block["type"].(string); ok {
				if blockType == "text" {
					if text, ok := block["text"].(string); ok {
						textParts = append(textParts, text)
					}
				}
			}
		}
		if len(textParts) > 0 {
			return role, concatenateTextParts(textParts)
		}
		// If no text blocks, return the serialized content
		return role, string(jsonBytes)
	}

	return role, string(jsonBytes)
}

// concatenateTextParts joins text parts with newlines
func concatenateTextParts(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += "\n" + parts[i]
	}
	return result
}

// ExtractResponseContent extracts output response from Anthropic message response
func ExtractResponseContent(resp *anthropic.Message, promptsTruncated bool) PromptData {
	data := PromptData{
		PromptsTruncated: promptsTruncated,
	}

	if resp == nil || len(resp.Content) == 0 {
		return data
	}

	// Extract text from content blocks
	var textParts []string
	for _, block := range resp.Content {
		// ContentBlockUnion has Type and Text fields directly
		if block.Type == "text" && block.Text != "" {
			textParts = append(textParts, block.Text)
		}
	}

	if len(textParts) == 0 {
		return data
	}

	content := concatenateTextParts(textParts)

	// Apply truncation if needed
	if len(content) > MaxPromptLength {
		markerLen := len(TruncationMarker)
		truncateAt := MaxPromptLength - markerLen
		content = truncateUTF8Safe(content, truncateAt) + TruncationMarker
		data.PromptsTruncated = true
		Debug("Output response truncated to %d characters", MaxPromptLength)
	}

	data.OutputResponse = content
	return data
}

// ExtractStreamingResponseContent extracts output from accumulated streaming content
func ExtractStreamingResponseContent(accumulatedContent string, promptsTruncated bool) PromptData {
	data := PromptData{
		PromptsTruncated: promptsTruncated,
	}

	if accumulatedContent == "" {
		return data
	}

	content := accumulatedContent

	// Apply truncation if needed
	if len(content) > MaxPromptLength {
		markerLen := len(TruncationMarker)
		truncateAt := MaxPromptLength - markerLen
		content = truncateUTF8Safe(content, truncateAt) + TruncationMarker
		data.PromptsTruncated = true
		Debug("Streaming output response truncated to %d characters", MaxPromptLength)
	}

	data.OutputResponse = content
	return data
}

// AddPromptDataToPayload adds prompt capture fields to a metering payload
func AddPromptDataToPayload(payload map[string]interface{}, data PromptData) {
	if data.SystemPrompt != "" {
		payload["systemPrompt"] = data.SystemPrompt
	}
	if data.InputMessages != "" {
		payload["inputMessages"] = data.InputMessages
	}
	if data.OutputResponse != "" {
		payload["outputResponse"] = data.OutputResponse
	}
	if data.PromptsTruncated {
		payload["promptsTruncated"] = true
	}
}

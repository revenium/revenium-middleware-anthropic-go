// Package revenium provides vision content detection for Anthropic messages
// Detects image content in message payloads for metering

package revenium

import (
	"github.com/anthropics/anthropic-sdk-go"
)

// VisionDetectionResult contains vision detection statistics
type VisionDetectionResult struct {
	// HasVisionContent indicates whether any vision/image content was found
	HasVisionContent bool
	// ImageCount is the number of images found in messages
	ImageCount int
	// TotalImageSizeBytes is the estimated size of image data in bytes (base64 decoded)
	TotalImageSizeBytes int
	// MediaTypes contains the image media types found
	MediaTypes []string
}

// DetectVisionContent scans Anthropic message parameters for image content
func DetectVisionContent(params anthropic.MessageNewParams) VisionDetectionResult {
	result := VisionDetectionResult{
		HasVisionContent:    false,
		ImageCount:          0,
		TotalImageSizeBytes: 0,
		MediaTypes:          []string{},
	}

	// Iterate through all messages
	for _, msg := range params.Messages {
		// Check each content block for images
		for _, block := range msg.Content {
			// Check if this is an image block using the OfImage field
			if block.OfImage != nil {
				processImageBlock(block.OfImage, &result)
			}
		}
	}

	return result
}

// processImageBlock extracts information from an image block
func processImageBlock(imgBlock *anthropic.ImageBlockParam, result *VisionDetectionResult) {
	if imgBlock == nil {
		return
	}

	result.HasVisionContent = true
	result.ImageCount++

	// Check for base64 image source
	if imgBlock.Source.OfBase64 != nil {
		processBase64ImageSource(imgBlock.Source.OfBase64, result)
	}
	// URL images don't contribute to byte size calculation
	// but we still count them as vision content
}

// processBase64ImageSource processes base64 encoded image data
func processBase64ImageSource(src *anthropic.Base64ImageSourceParam, result *VisionDetectionResult) {
	if src == nil {
		return
	}

	// Track media type
	mediaType := string(src.MediaType)
	if mediaType != "" && !containsString(result.MediaTypes, mediaType) {
		result.MediaTypes = append(result.MediaTypes, mediaType)
	}

	// Calculate estimated decoded size from base64
	// Base64 encoding increases size by ~4/3, so decoded = base64_len * 3 / 4
	// Account for padding characters (=) which don't represent data
	if src.Data != "" {
		base64Length := len(src.Data)
		// Count padding characters at the end
		padding := 0
		if base64Length > 0 && src.Data[base64Length-1] == '=' {
			padding++
			if base64Length > 1 && src.Data[base64Length-2] == '=' {
				padding++
			}
		}
		estimatedBytes := ((base64Length - padding) * 3) / 4
		result.TotalImageSizeBytes += estimatedBytes
	}
}

// containsString checks if a string slice contains a value
func containsString(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// BuildVisionAttributes creates attributes map for metering payload
func BuildVisionAttributes(result VisionDetectionResult) map[string]interface{} {
	if !result.HasVisionContent {
		return nil
	}

	return map[string]interface{}{
		"vision_image_count":      result.ImageCount,
		"vision_total_size_bytes": result.TotalImageSizeBytes,
		"vision_media_types":      result.MediaTypes,
	}
}

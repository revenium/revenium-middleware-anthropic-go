package revenium

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// ReveniumAnthropic is the main middleware client that wraps the Anthropic SDK
// and adds metering capabilities
type ReveniumAnthropic struct {
	client   anthropic.Client
	config   *Config
	provider Provider
	mu       sync.RWMutex
	wg       sync.WaitGroup // WaitGroup for tracking in-flight metering goroutines
}

var (
	globalClient *ReveniumAnthropic
	globalMu     sync.RWMutex
	initialized  bool
)

// Initialize sets up the global Revenium middleware with configuration
func Initialize(opts ...Option) error {
	globalMu.Lock()
	defer globalMu.Unlock()

	if initialized {
		return nil
	}

	// Initialize logger first
	InitializeLogger()
	Info("Initializing Revenium middleware...")

	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}
	// Load from environment if not provided
	if err := cfg.loadFromEnv(); err != nil {
		Warn("Failed to load configuration from environment: %v", err)
	}

	// Validate required fields
	if cfg.ReveniumAPIKey == "" {
		return NewConfigError("REVENIUM_METERING_API_KEY is required", nil)
	}

	// Create Anthropic client
	clientOpts := []option.RequestOption{}
	if cfg.AnthropicAPIKey != "" {
		clientOpts = append(clientOpts, option.WithAPIKey(cfg.AnthropicAPIKey))
	}

	anthropicClient := anthropic.NewClient(clientOpts...)

	// Detect provider
	provider := DetectProvider(cfg)

	globalClient = &ReveniumAnthropic{
		client:   anthropicClient,
		config:   cfg,
		provider: provider,
	}

	initialized = true
	Info("Revenium middleware initialized successfully")
	return nil
}

// IsInitialized checks if the middleware is properly initialized
func IsInitialized() bool {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return initialized
}

// GetClient returns the global Revenium client
func GetClient() (*ReveniumAnthropic, error) {
	globalMu.RLock()
	defer globalMu.RUnlock()

	if !initialized {
		return nil, NewConfigError("middleware not initialized, call Initialize() first", nil)
	}

	return globalClient, nil
}

// NewReveniumAnthropic creates a new Revenium client with explicit configuration
func NewReveniumAnthropic(cfg *Config) (*ReveniumAnthropic, error) {
	if cfg == nil {
		return nil, NewConfigError("config cannot be nil", nil)
	}

	// Validate required fields
	if cfg.ReveniumAPIKey == "" {
		return nil, NewConfigError("REVENIUM_METERING_API_KEY is required", nil)
	}

	// Create Anthropic client
	clientOpts := []option.RequestOption{}
	if cfg.AnthropicAPIKey != "" {
		clientOpts = append(clientOpts, option.WithAPIKey(cfg.AnthropicAPIKey))
	}

	anthropicClient := anthropic.NewClient(clientOpts...)

	// Detect provider
	provider := DetectProvider(cfg)

	return &ReveniumAnthropic{
		client:   anthropicClient,
		config:   cfg,
		provider: provider,
	}, nil
}

// GetConfig returns the configuration
func (r *ReveniumAnthropic) GetConfig() *Config {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config
}

// GetProvider returns the detected provider
func (r *ReveniumAnthropic) GetProvider() Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.provider
}

// GetAnthropicClient returns the underlying Anthropic client
func (r *ReveniumAnthropic) GetAnthropicClient() anthropic.Client {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.client
}

// Messages returns the messages interface for creating messages
func (r *ReveniumAnthropic) Messages() *MessagesInterface {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return &MessagesInterface{
		client:   r.client,
		config:   r.config,
		provider: r.provider,
		wg:       &r.wg,
	}
}

// Flush waits for all in-flight metering goroutines to complete.
// Call this before shutdown to ensure all metering data is sent.
func (r *ReveniumAnthropic) Flush() {
	r.wg.Wait()
}

// Close closes the client and cleans up resources.
// It waits for all in-flight metering goroutines to complete before returning.
func (r *ReveniumAnthropic) Close() error {
	r.Flush() // Wait for all metering goroutines to complete
	r.mu.Lock()
	defer r.mu.Unlock()

	// Cleanup resources if needed
	return nil
}

// MessagesInterface provides methods for creating messages with metering
type MessagesInterface struct {
	client   anthropic.Client
	config   *Config
	provider Provider
	wg       *sync.WaitGroup // Shared WaitGroup from ReveniumAnthropic
}

// CreateMessage creates a message with automatic metering
func (m *MessagesInterface) CreateMessage(ctx context.Context, params anthropic.MessageNewParams) (*anthropic.Message, error) {
	// Extract metadata from context
	metadata := GetUsageMetadata(ctx)

	// Call the appropriate provider
	switch m.provider {
	case ProviderAnthropic:
		return m.createMessageAnthropic(ctx, params, metadata)
	case ProviderBedrock:
		return m.createMessageBedrock(ctx, params, metadata)
	default:
		return nil, NewProviderError("unknown provider: %v", fmt.Errorf("provider: %v", m.provider))
	}
}

// CreateMessageStream creates a streaming message with automatic metering
// Returns a stream that can be iterated over to get events
func (m *MessagesInterface) CreateMessageStream(ctx context.Context, params anthropic.MessageNewParams) (interface{}, error) {
	// Extract metadata from context
	metadata := GetUsageMetadata(ctx)

	// Call the appropriate provider
	switch m.provider {
	case ProviderAnthropic:
		return m.createMessageStreamAnthropic(ctx, params, metadata)
	case ProviderBedrock:
		return m.createMessageStreamBedrock(ctx, params, metadata)
	default:
		return nil, NewProviderError("unknown provider: %v", fmt.Errorf("provider: %v", m.provider))
	}
}

// createMessageAnthropic creates a message using Anthropic native API
func (m *MessagesInterface) createMessageAnthropic(ctx context.Context, params anthropic.MessageNewParams, metadata map[string]interface{}) (*anthropic.Message, error) {
	// Record start time for duration calculation
	startTime := time.Now()

	// Extract prompts if capture is enabled
	var promptData *PromptData
	if m.config.CapturePrompts {
		data := ExtractPromptsFromParams(params)
		promptData = &data
	}

	// Convert Bedrock ARN to Anthropic model if needed
	// This handles the case where Bedrock is disabled but user passes a Bedrock ARN
	originalModel := string(params.Model)
	convertedModel, err := ConvertBedrockARNToAnthropicModel(originalModel)
	if err != nil {
		return nil, fmt.Errorf("failed to convert Bedrock model to Anthropic format: %w", err)
	}
	if convertedModel != originalModel {
		Info("Converted Bedrock model '%s' to Anthropic model '%s'", originalModel, convertedModel)
		params.Model = anthropic.Model(convertedModel)
	}

	// Call Anthropic API
	resp, err := m.client.Messages.New(ctx, params)
	if err != nil {
		return nil, err
	}

	// Calculate duration
	duration := time.Since(startTime)

	// Extract response content if prompt capture is enabled
	if promptData != nil && m.config.CapturePrompts {
		responseData := ExtractResponseContent(resp, promptData.PromptsTruncated)
		promptData.OutputResponse = responseData.OutputResponse
		promptData.PromptsTruncated = responseData.PromptsTruncated
	}

	// Send metering data asynchronously (fire-and-forget) with WaitGroup tracking
	if m.wg != nil {
		m.wg.Add(1)
		go func() {
			defer m.wg.Done()
			m.sendMeteringDataWithPrompts(ctx, resp, metadata, false, duration, "Anthropic", startTime, &params, promptData)
		}()
	} else {
		go m.sendMeteringDataWithPrompts(ctx, resp, metadata, false, duration, "Anthropic", startTime, &params, promptData)
	}

	return resp, nil
}

// createMessageBedrock creates a message using AWS Bedrock with fallback to Anthropic
func (m *MessagesInterface) createMessageBedrock(ctx context.Context, params anthropic.MessageNewParams, metadata map[string]interface{}) (*anthropic.Message, error) {
	// Record start time for duration calculation
	startTime := time.Now()

	// Extract prompts if capture is enabled
	var promptData *PromptData
	if m.config.CapturePrompts {
		data := ExtractPromptsFromParams(params)
		promptData = &data
	}

	// Create Bedrock adapter
	bedrockAdapter, err := NewBedrockAdapter(m.config)
	if err != nil {
		Warn("Failed to create Bedrock adapter, falling back to Anthropic: %v", err)
		// Convert Bedrock model to Anthropic model for fallback
		fallbackParams := params
		convertedModel, convErr := ConvertBedrockARNToAnthropicModel(string(params.Model))
		if convErr != nil {
			return nil, fmt.Errorf("failed to convert Bedrock model for fallback: %w", convErr)
		}
		fallbackParams.Model = anthropic.Model(convertedModel)
		Info("Converted Bedrock model '%s' to Anthropic model '%s' for fallback", params.Model, fallbackParams.Model)
		return m.createMessageAnthropic(ctx, fallbackParams, metadata)
	}

	// Try Bedrock with retry logic
	retryConfig := DefaultRetryConfig()
	var resp *anthropic.Message

	err = RetryWithBackoff(ctx, retryConfig, func() error {
		var bedrockErr error
		resp, bedrockErr = bedrockAdapter.CreateMessage(ctx, params)
		return bedrockErr
	})

	if err != nil {
		Warn("Bedrock request failed after retries: %v, falling back to Anthropic", err)
		// Convert Bedrock model to Anthropic model for fallback
		fallbackParams := params
		convertedModel, convErr := ConvertBedrockARNToAnthropicModel(string(params.Model))
		if convErr != nil {
			return nil, fmt.Errorf("failed to convert Bedrock model for fallback: %w", convErr)
		}
		fallbackParams.Model = anthropic.Model(convertedModel)
		Info("Converted Bedrock model '%s' to Anthropic model '%s' for fallback", params.Model, fallbackParams.Model)
		return m.createMessageAnthropic(ctx, fallbackParams, metadata)
	}

	// Calculate duration
	duration := time.Since(startTime)

	// Extract response content if prompt capture is enabled
	if promptData != nil && m.config.CapturePrompts {
		responseData := ExtractResponseContent(resp, promptData.PromptsTruncated)
		promptData.OutputResponse = responseData.OutputResponse
		promptData.PromptsTruncated = responseData.PromptsTruncated
	}

	// Send metering data asynchronously (fire-and-forget) with WaitGroup tracking
	if m.wg != nil {
		m.wg.Add(1)
		go func() {
			defer m.wg.Done()
			m.sendMeteringDataWithPrompts(ctx, resp, metadata, false, duration, "AWS", startTime, &params, promptData)
		}()
	} else {
		go m.sendMeteringDataWithPrompts(ctx, resp, metadata, false, duration, "AWS", startTime, &params, promptData)
	}

	return resp, nil
}

// createMessageStreamAnthropic creates a streaming message using Anthropic native API
func (m *MessagesInterface) createMessageStreamAnthropic(ctx context.Context, params anthropic.MessageNewParams, metadata map[string]interface{}) (interface{}, error) {
	// Extract prompts if capture is enabled
	var promptData *PromptData
	if m.config.CapturePrompts {
		data := ExtractPromptsFromParams(params)
		promptData = &data
	}

	// Convert Bedrock ARN to Anthropic model if needed
	// This handles the case where Bedrock is disabled but user passes a Bedrock ARN
	originalModel := string(params.Model)
	convertedModel, err := ConvertBedrockARNToAnthropicModel(originalModel)
	if err != nil {
		return nil, fmt.Errorf("failed to convert Bedrock model to Anthropic format: %w", err)
	}
	if convertedModel != originalModel {
		Info("Converted Bedrock model '%s' to Anthropic model '%s' for streaming", originalModel, convertedModel)
		params.Model = anthropic.Model(convertedModel)
	}

	// Call Anthropic streaming API
	stream := m.client.Messages.NewStreaming(ctx, params)

	// Prepare metadata with model information
	streamMetadata := make(map[string]interface{})
	if metadata != nil {
		// Copy all user-provided metadata
		for k, v := range metadata {
			streamMetadata[k] = v
		}
	}

	// Add model from request if not already in metadata
	if _, ok := streamMetadata["model"]; !ok {
		streamMetadata["model"] = string(params.Model)
	}

	// Wrap stream for metering tracking
	wrapper := &StreamingWrapper{
		stream:      stream,
		config:      m.config,
		metadata:    streamMetadata,
		startTime:   time.Now(),
		messagesAPI: m,
		model:       string(params.Model),
		provider:    "Anthropic",
		params:      &params,
		promptData:  promptData,
	}

	// Estimate input tokens (this is an approximation)
	// In a real implementation, you might want to use a tokenizer
	inputTokens := estimateInputTokens(params)
	wrapper.SetInputTokens(inputTokens)

	return wrapper, nil
}

// createMessageStreamBedrock creates a streaming message using AWS Bedrock with fallback to Anthropic
func (m *MessagesInterface) createMessageStreamBedrock(ctx context.Context, params anthropic.MessageNewParams, metadata map[string]interface{}) (interface{}, error) {
	// Extract prompts if capture is enabled
	var promptData *PromptData
	if m.config.CapturePrompts {
		data := ExtractPromptsFromParams(params)
		promptData = &data
	}

	// Create Bedrock adapter
	bedrockAdapter, err := NewBedrockAdapter(m.config)
	if err != nil {
		Warn("Failed to create Bedrock adapter for streaming, falling back to Anthropic: %v", err)
		// Convert Bedrock model to Anthropic model for fallback
		fallbackParams := params
		convertedModel, convErr := ConvertBedrockARNToAnthropicModel(string(params.Model))
		if convErr != nil {
			return nil, fmt.Errorf("failed to convert Bedrock model for streaming fallback: %w", convErr)
		}
		fallbackParams.Model = anthropic.Model(convertedModel)
		Info("Converted Bedrock model '%s' to Anthropic model '%s' for streaming fallback", params.Model, fallbackParams.Model)
		return m.createMessageStreamAnthropic(ctx, fallbackParams, metadata)
	}

	// Try Bedrock streaming with retry logic
	retryConfig := DefaultRetryConfig()
	var stream interface{}

	err = RetryWithBackoff(ctx, retryConfig, func() error {
		var bedrockErr error
		stream, bedrockErr = bedrockAdapter.CreateMessageStream(ctx, params)
		return bedrockErr
	})

	if err != nil {
		Warn("Bedrock streaming request failed after retries: %v, falling back to Anthropic", err)
		// Convert Bedrock model to Anthropic model for fallback
		fallbackParams := params
		convertedModel, convErr := ConvertBedrockARNToAnthropicModel(string(params.Model))
		if convErr != nil {
			return nil, fmt.Errorf("failed to convert Bedrock model for streaming fallback: %w", convErr)
		}
		fallbackParams.Model = anthropic.Model(convertedModel)
		Info("Converted Bedrock model '%s' to Anthropic model '%s' for streaming fallback", params.Model, fallbackParams.Model)
		return m.createMessageStreamAnthropic(ctx, fallbackParams, metadata)
	}

	// Prepare metadata with model information
	streamMetadata := make(map[string]interface{})
	if metadata != nil {
		// Copy all user-provided metadata
		for k, v := range metadata {
			streamMetadata[k] = v
		}
	}

	// Add model from request if not already in metadata
	if _, ok := streamMetadata["model"]; !ok {
		streamMetadata["model"] = string(params.Model)
	}

	// Wrap Bedrock stream for metering tracking
	wrapper := &StreamingWrapper{
		stream:      stream,
		config:      m.config,
		metadata:    streamMetadata,
		startTime:   time.Now(),
		messagesAPI: m,
		model:       string(params.Model),
		provider:    "AWS",
		params:      &params,
		promptData:  promptData,
	}

	// Estimate input tokens for Bedrock
	inputTokens := estimateInputTokens(params)
	wrapper.SetInputTokens(inputTokens)

	return wrapper, nil
}

// StreamingWrapper wraps a streaming response to capture metering data
type StreamingWrapper struct {
	stream         interface{} // *anthropic.MessageStream
	config         *Config
	metadata       map[string]interface{}
	startTime      time.Time
	firstTokenTime *time.Time
	mu             sync.Mutex
	messagesAPI    *MessagesInterface // Reference to MessagesInterface for metering

	// Token counting for streaming
	inputTokens  int
	outputTokens int
	totalTokens  int
	model        string
	provider     string                        // Provider name (Anthropic or AWS)
	stopReason   string                        // Stop reason from streaming events
	params       *anthropic.MessageNewParams   // Original request params for vision detection

	// Prompt capture tracking
	promptData         *PromptData
	accumulatedContent string
}

// Next returns the next event from the stream
func (sw *StreamingWrapper) Next() bool {
	// Use reflection to call Next() on the underlying stream
	if sw.stream == nil {
		return false
	}

	// Call Next() method using reflection
	streamVal := reflect.ValueOf(sw.stream)

	// Handle pointer types
	if streamVal.Kind() == reflect.Ptr {
		nextMethod := streamVal.MethodByName("Next")
		if nextMethod.IsValid() {
			result := nextMethod.Call(nil)
			if len(result) > 0 {
				// Check if result is a bool
				if b, ok := result[0].Interface().(bool); ok {
					return b
				}
			}
		}
	}

	return false
}

// Current returns the current event
func (sw *StreamingWrapper) Current() interface{} {
	if sw.stream == nil {
		return nil
	}

	// Call Current() method using reflection
	streamVal := reflect.ValueOf(sw.stream)
	if streamVal.Kind() == reflect.Ptr {
		currentMethod := streamVal.MethodByName("Current")
		if currentMethod.IsValid() {
			result := currentMethod.Call(nil)
			if len(result) > 0 {
				event := result[0].Interface()

				// Record first token time and extract real usage data
				sw.mu.Lock()

				// Debug: log event type
				Debug("Streaming event type: %T", event)

				// Check for content events to record first token time
				if isContentEvent(event) {
					if sw.firstTokenTime == nil {
						now := time.Now()
						sw.firstTokenTime = &now
					}

					// Accumulate content for prompt capture (if enabled)
					if sw.promptData != nil {
						if text := extractTextFromContentEvent(event); text != "" {
							sw.accumulatedContent += text
						}
					}
				}

				// Check for message_delta events that contain real usage data
				if isMessageDeltaEvent(event) {
					usage := extractUsageFromEvent(event)
					if usage != nil {
						sw.inputTokens = int(usage.InputTokens)
						sw.outputTokens = int(usage.OutputTokens)
						sw.totalTokens = sw.inputTokens + sw.outputTokens
						Debug("Real token usage extracted: input=%d, output=%d, total=%d", sw.inputTokens, sw.outputTokens, sw.totalTokens)
					}

					// Extract stop_reason from message_delta event (usually in Delta field)
					stopReason := extractStopReasonFromEvent(event)
					if stopReason != "" {
						sw.stopReason = stopReason
						Debug("Stop reason extracted from streaming: %s", stopReason)
					}
				}

				sw.mu.Unlock()

				return event
			}
		}
	}
	return nil
}

// Err returns any error that occurred during streaming
func (sw *StreamingWrapper) Err() error {
	if sw.stream == nil {
		return nil
	}

	// Call Err() method using reflection
	streamVal := reflect.ValueOf(sw.stream)
	if streamVal.Kind() == reflect.Ptr {
		errMethod := streamVal.MethodByName("Err")
		if errMethod.IsValid() {
			result := errMethod.Call(nil)
			if len(result) > 0 {
				if err, ok := result[0].Interface().(error); ok {
					return err
				}
			}
		}
	}
	return nil
}

// SetInputTokens sets the input token count (should be called when stream is created)
func (sw *StreamingWrapper) SetInputTokens(tokens int) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.inputTokens = tokens
	sw.totalTokens = sw.inputTokens + sw.outputTokens
}

// SetModel sets the model name
func (sw *StreamingWrapper) SetModel(model string) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.model = model
}

// GetTokenCounts returns the current token counts
func (sw *StreamingWrapper) GetTokenCounts() (input, output, total int) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.inputTokens, sw.outputTokens, sw.totalTokens
}

// Close closes the stream and sends metering data
func (sw *StreamingWrapper) Close() error {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	var err error
	if sw.stream != nil {
		// Call Close() method using reflection
		streamVal := reflect.ValueOf(sw.stream)
		if streamVal.Kind() == reflect.Ptr {
			closeMethod := streamVal.MethodByName("Close")
			if closeMethod.IsValid() {
				result := closeMethod.Call(nil)
				if len(result) > 0 {
					if e, ok := result[0].Interface().(error); ok {
						err = e
					}
				}
			}
		}
	}

	// Calculate metrics
	duration := time.Since(sw.startTime)
	timeToFirstToken := time.Duration(0)
	if sw.firstTokenTime != nil {
		timeToFirstToken = sw.firstTokenTime.Sub(sw.startTime)
	}

	// Send metering data asynchronously (fire-and-forget) with WaitGroup tracking
	meteringFunc := func() {
		defer func() {
			if r := recover(); r != nil {
				Error("Streaming metering goroutine panic: %v", r)
			}
		}()

		// Get actual token counts and stop reason from streaming
		sw.mu.Lock()
		inputTokens := sw.inputTokens
		outputTokens := sw.outputTokens
		totalTokens := sw.totalTokens
		model := sw.model
		provider := sw.provider
		startTime := sw.startTime
		streamStopReason := sw.stopReason
		sw.mu.Unlock()

		// Build metering payload for streaming using actual token counts
		// Note: We need to use reflection to set StopReason since it's not exported
		mockResp := &anthropic.Message{
			Model: anthropic.Model(model),
			Usage: anthropic.Usage{
				InputTokens:  int64(inputTokens),
				OutputTokens: int64(outputTokens),
			},
		}

		// Set StopReason using reflection (anthropic.StopReason is a string type alias)
		if streamStopReason != "" {
			respValue := reflect.ValueOf(mockResp).Elem()
			stopReasonField := respValue.FieldByName("StopReason")
			if stopReasonField.IsValid() && stopReasonField.CanSet() {
				// Create a value of the correct type (anthropic.StopReason)
				// by converting the string to reflect.Value and setting it
				stopReasonValue := reflect.ValueOf(anthropic.StopReason(streamStopReason))
				stopReasonField.Set(stopReasonValue)
			}
		}

		// Use the same payload builder as non-streaming
		payload := buildMeteringPayload(mockResp, sw.metadata, true, duration, provider, startTime, sw.params)

		// Override streaming-specific fields with actual timing data
		payload["timeToFirstToken"] = timeToFirstToken.Milliseconds()

		// Calculate correct completion start time for streaming (when first token arrived)
		if sw.firstTokenTime != nil {
			payload["completionStartTime"] = sw.firstTokenTime.Format(time.RFC3339)
		}

		// Ensure token counts are correct (override with actual counts)
		payload["inputTokenCount"] = inputTokens
		payload["outputTokenCount"] = outputTokens
		payload["totalTokenCount"] = totalTokens
		if model != "" {
			payload["model"] = model
		}

		// Add prompt capture data if enabled
		sw.mu.Lock()
		promptData := sw.promptData
		accumulatedContent := sw.accumulatedContent
		sw.mu.Unlock()

		if promptData != nil {
			// Add input prompts
			if promptData.SystemPrompt != "" {
				payload["systemPrompt"] = promptData.SystemPrompt
			}
			if promptData.InputMessages != "" {
				payload["inputMessages"] = promptData.InputMessages
			}

			// Extract streaming response content
			responseData := ExtractStreamingResponseContent(accumulatedContent, promptData.PromptsTruncated)
			if responseData.OutputResponse != "" {
				payload["outputResponse"] = responseData.OutputResponse
			}
			if responseData.PromptsTruncated {
				payload["promptsTruncated"] = true
			}
		}

		// Send to Revenium API with retry logic
		if sw.messagesAPI != nil {
			if err := sw.messagesAPI.sendMeteringWithRetry(payload); err != nil {
				Error("Failed to send streaming metering data: %v", err)
			}
		}
	}

	// Launch goroutine with WaitGroup tracking if available
	if sw.messagesAPI != nil && sw.messagesAPI.wg != nil {
		sw.messagesAPI.wg.Add(1)
		go func() {
			defer sw.messagesAPI.wg.Done()
			meteringFunc()
		}()
	} else {
		go meteringFunc()
	}

	return err
}

// isContentEvent checks if an event is a content event
func isContentEvent(event interface{}) bool {
	if event == nil {
		return false
	}

	// Use reflection to check for Delta field (same logic as examples)
	eventValue := reflect.ValueOf(event)
	if eventValue.Kind() == reflect.Ptr {
		eventValue = eventValue.Elem()
	}

	// Check if this is a ContentBlockDeltaEvent by looking for Delta field
	deltaField := eventValue.FieldByName("Delta")
	if deltaField.IsValid() && !deltaField.IsZero() {
		// This looks like a ContentBlockDeltaEvent
		deltaValue := deltaField.Interface()
		if deltaValue == nil {
			return false
		}

		// Try to get Text field from the delta
		deltaReflect := reflect.ValueOf(deltaValue)
		if deltaReflect.Kind() == reflect.Ptr {
			deltaReflect = deltaReflect.Elem()
		}

		// Look for Text field - if it exists, this is a content event
		textField := deltaReflect.FieldByName("Text")
		if textField.IsValid() {
			if text, ok := textField.Interface().(string); ok && text != "" {
				return true
			}
		}
	}

	return false
}

// extractTextFromContentEvent extracts text from a content event for prompt capture
func extractTextFromContentEvent(event interface{}) string {
	if event == nil {
		return ""
	}

	// Use reflection to get the Delta.Text field
	eventValue := reflect.ValueOf(event)
	if eventValue.Kind() == reflect.Ptr {
		eventValue = eventValue.Elem()
	}

	// Get Delta field
	deltaField := eventValue.FieldByName("Delta")
	if !deltaField.IsValid() || deltaField.IsZero() {
		return ""
	}

	deltaValue := deltaField.Interface()
	if deltaValue == nil {
		return ""
	}

	deltaReflect := reflect.ValueOf(deltaValue)
	if deltaReflect.Kind() == reflect.Ptr {
		deltaReflect = deltaReflect.Elem()
	}

	// Get Text field from Delta
	textField := deltaReflect.FieldByName("Text")
	if textField.IsValid() {
		if text, ok := textField.Interface().(string); ok {
			return text
		}
	}

	return ""
}

// isMessageDeltaEvent checks if an event is a message_delta event containing usage data
func isMessageDeltaEvent(event interface{}) bool {
	if event == nil {
		return false
	}

	// Use reflection to check for Type field with "message_delta" value
	eventValue := reflect.ValueOf(event)
	if eventValue.Kind() == reflect.Ptr {
		eventValue = eventValue.Elem()
	}

	// Check if this has a Type field with "message_delta"
	typeField := eventValue.FieldByName("Type")
	if typeField.IsValid() {
		if typeStr, ok := typeField.Interface().(string); ok && typeStr == "message_delta" {
			return true
		}
	}

	return false
}

// extractUsageFromEvent extracts usage data from a message_delta event
func extractUsageFromEvent(event interface{}) *anthropic.MessageDeltaUsage {
	if event == nil {
		return nil
	}

	// Use reflection to get Usage field
	eventValue := reflect.ValueOf(event)
	if eventValue.Kind() == reflect.Ptr {
		eventValue = eventValue.Elem()
	}

	// Look for Usage field
	usageField := eventValue.FieldByName("Usage")
	if usageField.IsValid() && !usageField.IsZero() {
		if usage, ok := usageField.Interface().(anthropic.MessageDeltaUsage); ok {
			return &usage
		}
	}

	return nil
}

// extractStopReasonFromEvent extracts stop_reason from a message_delta event
func extractStopReasonFromEvent(event interface{}) string {
	if event == nil {
		return ""
	}

	// Use reflection to get Delta field which contains stop_reason
	eventValue := reflect.ValueOf(event)
	if eventValue.Kind() == reflect.Ptr {
		eventValue = eventValue.Elem()
	}

	// Look for Delta field (MessageDelta type)
	deltaField := eventValue.FieldByName("Delta")
	if deltaField.IsValid() && !deltaField.IsZero() {
		// Get StopReason from Delta
		stopReasonField := deltaField.FieldByName("StopReason")
		if stopReasonField.IsValid() && !stopReasonField.IsZero() {
			if stopReason, ok := stopReasonField.Interface().(string); ok && stopReason != "" {
				return stopReason
			}
		}
	}

	return ""
}

// estimateInputTokens provides a rough estimate of input tokens
// This is a simple approximation - in production you'd use a proper tokenizer
func estimateInputTokens(params anthropic.MessageNewParams) int {
	totalChars := 0

	// Count characters in all messages
	for _, msg := range params.Messages {
		// For now, use a simple approximation based on content blocks
		// This is a rough estimate since we can't easily access the actual text
		// In a real implementation, you'd use the Anthropic tokenizer
		totalChars += len(msg.Content) * 10 // Rough estimate per content block
	}

	// Rough approximation: ~4 characters per token for English text
	// This is a very rough estimate - actual tokenization varies significantly
	estimatedTokens := totalChars / 4
	if estimatedTokens < 1 {
		estimatedTokens = 10 // Default minimum estimate
	}

	return estimatedTokens
}

// sendMeteringData sends metering data in the background (fire-and-forget)
// NOTE: This function is already called with 'go' from the caller, so it should NOT launch another goroutine
func (m *MessagesInterface) sendMeteringData(ctx context.Context, resp *anthropic.Message, metadata map[string]interface{}, isStreamed bool, duration time.Duration, provider string, startTime time.Time, params *anthropic.MessageNewParams) {
	m.sendMeteringDataWithPrompts(ctx, resp, metadata, isStreamed, duration, provider, startTime, params, nil)
}

// sendMeteringDataWithPrompts sends metering data with optional prompt capture
func (m *MessagesInterface) sendMeteringDataWithPrompts(ctx context.Context, resp *anthropic.Message, metadata map[string]interface{}, isStreamed bool, duration time.Duration, provider string, startTime time.Time, params *anthropic.MessageNewParams, promptData *PromptData) {
	defer func() {
		if r := recover(); r != nil {
			Error("Metering goroutine panic: %v", r)
		}
	}()

	// Build metering payload using helper function
	payload := buildMeteringPayload(resp, metadata, isStreamed, duration, provider, startTime, params)

	// Add prompt data if available
	if promptData != nil {
		AddPromptDataToPayload(payload, *promptData)
	}

	// Send to Revenium API with retry logic
	if err := m.sendMeteringWithRetry(payload); err != nil {
		Error("Failed to send metering data: %v", err)
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().UnixNano()%1000)
}

// mapStopReasonToRevenium converts Anthropic/Bedrock stop reasons to Revenium format
// with a solid fallback to END for any unknown values
func mapStopReasonToRevenium(stopReason string) string {
	// Map Anthropic/Bedrock stop reasons to Revenium enum values
	// Based on official Anthropic API documentation
	switch stopReason {
	case "end_turn":
		return "END"
	case "max_tokens", "model_context_window_exceeded":
		return "TOKEN_LIMIT"
	case "stop_sequence":
		return "END_SEQUENCE"
	case "tool_use":
		return "END" // Tool use is a natural completion
	case "pause_turn":
		return "END" // Pause is a temporary stop, treat as normal end
	case "refusal":
		return "ERROR" // Refusal due to policy violation
	case "timeout":
		return "TIMEOUT"
	case "error":
		return "ERROR"
	case "cancelled", "canceled": // Handle both British and American spellings
		return "CANCELLED"
	default:
		// Solid fallback for any unknown stop reasons or future additions
		Debug("Unknown stop reason '%s', defaulting to END", stopReason)
		return "END"
	}
}

// normalizeProviderName converts internal provider names to Revenium-compliant format
func normalizeProviderName(provider string) string {
	switch provider {
	case "AWS":
		return "Amazon Bedrock"
	case "Anthropic":
		return "Anthropic"
	default:
		// Default to the provider as-is if unknown
		return provider
	}
}

// buildMeteringPayload builds a metering payload, matching Node.js format exactly
func buildMeteringPayload(resp *anthropic.Message, metadata map[string]interface{}, isStreamed bool, duration time.Duration, provider string, startTime time.Time, params *anthropic.MessageNewParams) map[string]interface{} {
	// Calculate actual timestamps based on request timing
	requestTimeISO := startTime.Format(time.RFC3339)
	responseTime := startTime.Add(duration)
	responseTimeISO := responseTime.Format(time.RFC3339)
	completionStartTimeISO := startTime.Format(time.RFC3339) // For non-streaming, completion starts immediately

	// Normalize provider name to match Revenium spec
	normalizedProvider := normalizeProviderName(provider)

	// Map stop reason with fallback to END
	stopReason := "END" // Default fallback
	if resp.StopReason != "" {
		stopReason = mapStopReasonToRevenium(string(resp.StopReason))
	} else {
		// Log when stop reason is empty (helps identify API changes or issues)
		Debug("Stop reason is empty, defaulting to END")
	}

	// Start with required fields only (matching Node.js buildReveniumPayload)
	payload := map[string]interface{}{
		"stopReason":              stopReason,
		"costType":                "AI",
		"isStreamed":              isStreamed,
		"operationType":           "CHAT",
		"inputTokenCount":         resp.Usage.InputTokens,
		"outputTokenCount":        resp.Usage.OutputTokens,
		"reasoningTokenCount":     int64(0), // Always 0 for Anthropic (no extended thinking)
		"cacheCreationTokenCount": int64(0), // TODO: Extract from resp.Usage if available
		"cacheReadTokenCount":     int64(0), // TODO: Extract from resp.Usage if available
		"totalTokenCount":         resp.Usage.InputTokens + resp.Usage.OutputTokens,
		"model":                   resp.Model,
		"transactionId":           generateRequestID(),
		"responseTime":            responseTimeISO,
		"requestDuration":         duration.Milliseconds(),
		"provider":                normalizedProvider,
		"requestTime":             requestTimeISO,
		"completionStartTime":     completionStartTimeISO,
		"timeToFirstToken":        int64(0), // Will be overridden for streaming
		"middlewareSource":        GetMiddlewareSource(),
	}

	// Add metadata fields if they exist (based on testing with Revenium API)
	if metadata != nil {
		// Business context fields (tested individually with Revenium API)
		if organizationId, ok := metadata["organizationId"]; ok {
			payload["organizationId"] = organizationId
		}
		if productId, ok := metadata["productId"]; ok {
			payload["productId"] = productId
		}
		if taskType, ok := metadata["taskType"]; ok {
			payload["taskType"] = taskType
		}
		if agent, ok := metadata["agent"]; ok {
			payload["agent"] = agent
		}
		if subscriptionId, ok := metadata["subscriptionId"]; ok {
			payload["subscriptionId"] = subscriptionId
		}
		if traceId, ok := metadata["traceId"]; ok {
			payload["traceId"] = traceId
		}
		if subscriber, ok := metadata["subscriber"]; ok {
			// subscriber must be an object with nested structure (not a string)
			payload["subscriber"] = subscriber
		}
		if taskId, ok := metadata["taskId"]; ok {
			payload["taskId"] = taskId
		}
		if responseQualityScore, ok := metadata["responseQualityScore"]; ok {
			payload["responseQualityScore"] = responseQualityScore
		}

		// Trace visualization fields (10 fields for distributed tracing)
		if transactionId, ok := metadata["transactionId"]; ok {
			payload["transactionId"] = transactionId
		}
		if traceType, ok := metadata["traceType"]; ok {
			payload["traceType"] = traceType
		}
		if traceName, ok := metadata["traceName"]; ok {
			payload["traceName"] = traceName
		}
		if environment, ok := metadata["environment"]; ok {
			payload["environment"] = environment
		}
		if region, ok := metadata["region"]; ok {
			payload["region"] = region
		}
		// NOTE: operationType is fixed to "CHAT" and should NOT be overridden by metadata
		// (API only accepts: CHAT, GENERATE, EMBED, CLASSIFY, SUMMARIZE, TRANSLATE, OTHER)
		// operationSubtype is auto-detected, not user-provided
		if retryNumber, ok := metadata["retryNumber"]; ok {
			payload["retryNumber"] = retryNumber
		}
		if credentialAlias, ok := metadata["credentialAlias"]; ok {
			payload["credentialAlias"] = credentialAlias
		}
		if parentTransactionId, ok := metadata["parentTransactionId"]; ok {
			payload["parentTransactionId"] = parentTransactionId
		}

		// Optional fields from Revenium spec (meter_ai_completion.md)
		if modelSource, ok := metadata["modelSource"]; ok {
			payload["modelSource"] = modelSource
		}
		if mediationLatency, ok := metadata["mediationLatency"]; ok {
			payload["mediationLatency"] = mediationLatency
		}
		if temperature, ok := metadata["temperature"]; ok {
			payload["temperature"] = temperature
		}
		if systemFingerprint, ok := metadata["systemFingerprint"]; ok {
			payload["systemFingerprint"] = systemFingerprint
		}

		// Cost override fields (typically null to let Revenium calculate)
		if inputTokenCost, ok := metadata["inputTokenCost"]; ok {
			payload["inputTokenCost"] = inputTokenCost
		}
		if outputTokenCost, ok := metadata["outputTokenCost"]; ok {
			payload["outputTokenCost"] = outputTokenCost
		}
		if cacheCreationTokenCost, ok := metadata["cacheCreationTokenCost"]; ok {
			payload["cacheCreationTokenCost"] = cacheCreationTokenCost
		}
		if cacheReadTokenCost, ok := metadata["cacheReadTokenCost"]; ok {
			payload["cacheReadTokenCost"] = cacheReadTokenCost
		}
		if totalCost, ok := metadata["totalCost"]; ok {
			payload["totalCost"] = totalCost
		}

		// Error tracking
		if errorReason, ok := metadata["errorReason"]; ok {
			payload["errorReason"] = errorReason
			payload["stopReason"] = "ERROR" // Override stop reason if error occurred
		}
	}

	// Detect vision content in request parameters
	if params != nil {
		visionResult := DetectVisionContent(*params)
		if visionResult.HasVisionContent {
			payload["hasVisionContent"] = true
			// Add vision attributes
			if attrs := BuildVisionAttributes(visionResult); attrs != nil {
				payload["attributes"] = attrs
			}
		}
	}

	return payload
}

// sendMeteringWithRetry sends metering data with exponential backoff retry
func (m *MessagesInterface) sendMeteringWithRetry(payload map[string]interface{}) error {
	const maxRetries = 3
	const initialBackoff = 100 * time.Millisecond

	var lastErr error
	backoff := initialBackoff

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
		}

		err := m.sendMeteringRequest(payload)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Don't retry on validation errors
		if isValidationError(err) {
			return err
		}
	}

	return NewMeteringError("metering failed after %d retries", fmt.Errorf("retries: %d, last error: %w", maxRetries, lastErr))
}

// sendMeteringRequest sends a single metering request to Revenium API
func (m *MessagesInterface) sendMeteringRequest(payload map[string]interface{}) error {
	if m.config == nil || m.config.ReveniumAPIKey == "" {
		return NewConfigError("metering not configured", nil)
	}

	// Build request URL
	baseURL := m.config.ReveniumBaseURL
	if baseURL == "" {
		baseURL = "https://api.revenium.ai"
	}
	// Append the endpoint path: /meter/v2/ai/completions
	url := baseURL + "/meter/v2/ai/completions"

	// Marshal payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return NewMeteringError("failed to marshal metering payload", err)
	}

	// Log the exact payload being sent
	Debug("[METERING] Sending payload to %s: %s", url, string(jsonData))

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return NewMeteringError("failed to create metering request", err)
	}

	// Set headers (matching Node.js implementation)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("x-api-key", m.config.ReveniumAPIKey)
	req.Header.Set("User-Agent", "revenium-middleware-anthropic-go/1.0")

	// Send request with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return NewNetworkError("metering request failed", err)
	}
	defer resp.Body.Close()

	// Read response body for error details
	body, _ := io.ReadAll(resp.Body)

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Log response for debugging
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			// Validation error - don't retry
			return NewValidationError(
				fmt.Sprintf("metering API returned %d: %s", resp.StatusCode, string(body)),
				nil,
			)
		}
		return NewMeteringError("metering API error", fmt.Errorf("status %d: %s", resp.StatusCode, string(body)))
	}

	return nil
}

// isValidationError checks if an error is a validation error
func isValidationError(err error) bool {
	if err == nil {
		return false
	}
	return IsValidationError(err)
}

// Reset resets the global middleware state for testing
func Reset() {
	globalMu.Lock()
	defer globalMu.Unlock()

	if globalClient != nil {
		globalClient.Close()
		globalClient = nil
	}

	initialized = false
}

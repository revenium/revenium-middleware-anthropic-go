# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.5] - 2026-01-21

### Added
- **Prompt Capture** - Opt-in feature to capture system prompts, input messages, and output responses for analytics (BACK-395)
- **Vision Content Detection** - Automatic detection of Claude Vision requests with image content (BACK-348)
- `WithCapturePrompts()` option to enable prompt capture
- `REVENIUM_CAPTURE_PROMPTS` environment variable support
- UTF-8 safe truncation for multi-byte character preservation
- Individual message truncation (before JSON serialization) to prevent invalid JSON

### Fixed
- Boolean parsing consistency for `REVENIUM_CAPTURE_PROMPTS` (accepts both "true" and "1")

## [1.0.4] - 2026-01-17

### Added
- AGENTS.md for AI agent context
- Dynamic version detection using `runtime/debug.ReadBuildInfo()`

### Fixed
- Stop reason mapping with comprehensive fallback to END
- Streaming stop reason capture from message_delta events
- Provider normalization (AWS → Amazon Bedrock for spec compliance)
- Timestamp accuracy (using actual request timing)
- Optional fields support (10+ fields from Revenium spec)
- Empty string handling for stop_reason with debug logging
- Model name consistency across all examples

### Changed
- Base URL updated from api.revenium.io to api.revenium.ai
- Dashboard URL references updated to app.revenium.ai
- Response quality score scale clarified (0.0-1.0)

## [1.0.0] - 2025-01-20

### Added

- Initial release
- Anthropic API integration
- AWS Bedrock support
- Streaming response handling
- Usage tracking and metering
- Test suite and examples

[Unreleased]: https://github.com/revenium/revenium-middleware-anthropic-go/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/revenium/revenium-middleware-anthropic-go/releases/tag/v1.0.0

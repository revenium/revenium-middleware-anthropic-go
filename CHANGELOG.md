# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

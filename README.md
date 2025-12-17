# Revenium Middleware for Anthropic (Go)

A lightweight, production-ready middleware that adds **Revenium metering and tracking** to Anthropic Claude API calls.

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue)](https://golang.org/)
[![Documentation](https://img.shields.io/badge/docs-revenium.io-blue)](https://docs.revenium.io)
[![Website](https://img.shields.io/badge/website-revenium.ai-blue)](https://www.revenium.ai)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- **Seamless Integration** - Drop-in replacement with zero code changes required
- **Automatic Metering** - Tracks all API calls with detailed usage metrics
- **Streaming Support** - Full support for streaming chat completions
- **AWS Bedrock Support** - Works with AWS Bedrock Anthropic models
- **Custom Metadata** - Add custom tracking metadata to any request
- **Production Ready** - Battle-tested and optimized for production use
- **Type Safe** - Built with Go's strong typing system

## Getting Started (5 minutes)

### Step 1: Create Your Project

```bash
mkdir my-claude-app
cd my-claude-app
go mod init my-claude-app
```

### Step 2: Install Dependencies

```bash
go get github.com/revenium/revenium-middleware-anthropic-go
go mod tidy
```

This will automatically download the middleware and all its dependencies (including the Anthropic SDK).

### Step 3: Create Environment File

Create `.env` file in your project root:

```bash
# Required - Get from https://console.anthropic.com/
ANTHROPIC_API_KEY=your_anthropic_api_key_here

# Required - Get from Revenium dashboard (https://app.revenium.ai)
REVENIUM_METERING_API_KEY=your_revenium_api_key_here

# Optional - Revenium API base URL (defaults to production)
REVENIUM_METERING_BASE_URL=https://api.revenium.ai
```

**Replace the API keys with your actual keys!**

> **Automatic .env Loading**: The middleware automatically loads `.env` files from your project directory. No need to manually export environment variables!

## Examples

This repository includes runnable examples demonstrating how to use the Revenium middleware with Anthropic Claude:

- **[Examples Guide](./examples/README.md)** - Detailed guide for running examples
- **Go Examples**: `examples/basic/`, `examples/streaming/`, `examples/advanced/`, `examples/bedrock/`

**Run examples after setup:**

```bash
# Clone this repository:
git clone https://github.com/revenium/revenium-middleware-anthropic-go.git
cd revenium-middleware-anthropic-go
go mod download
go mod tidy

# Run examples:
go run examples/basic/main.go
go run examples/streaming/main.go
go run examples/advanced/main.go
go run examples/bedrock/main.go
```

See **[Examples Guide](./examples/README.md)** for detailed setup instructions and what each example demonstrates.

## What Gets Tracked

The middleware automatically captures:

- **Token Usage**: Input and output tokens for accurate billing
- **Request Duration**: Total time for each API call
- **Model Information**: Which Claude model was used
- **Provider Info**: Anthropic API or AWS Bedrock
- **Custom Metadata**: Business context you provide
- **Error Tracking**: Failed requests and error details

## Environment Variables

### Required

```bash
ANTHROPIC_API_KEY=your_anthropic_api_key_here
REVENIUM_METERING_API_KEY=your_revenium_api_key_here
```

### Optional

```bash
# Revenium API base URL (defaults to production, middleware automatically appends /meter/v2/ai/completions)
REVENIUM_METERING_BASE_URL=https://api.revenium.ai

# Default metadata for all requests
REVENIUM_ORGANIZATION_ID=my-company
REVENIUM_PRODUCT_ID=my-app

# AWS Bedrock support (optional)
AWS_ACCESS_KEY_ID=your_aws_access_key
AWS_SECRET_ACCESS_KEY=your_aws_secret_key

# AWS Bedrock Base ARN (required for Bedrock)
AWS_MODEL_ARN_ID=arn:aws:bedrock:{region}:{account-id}

# AWS Profile (optional, only needed if using AWS profiles instead of access keys)
# AWS_PROFILE=default

# Disable Bedrock support (set to 1 to disable, 0 to enable)
REVENIUM_BEDROCK_DISABLE=1

# Debug logging
REVENIUM_LOG_LEVEL=INFO
REVENIUM_VERBOSE_STARTUP=false
```

## AWS Bedrock Configuration

To use AWS Bedrock with the middleware:

### Step 1: Configure AWS Credentials

```bash
# AWS Access Keys
AWS_ACCESS_KEY_ID=your_aws_access_key
AWS_SECRET_ACCESS_KEY=your_aws_secret_key
```

**OR** use AWS Profile (if you have `~/.aws/credentials` configured):

```bash
AWS_PROFILE=your-profile-name
```

### Step 2: Configure Base ARN

```bash
# AWS Bedrock Base ARN (ONLY the base, NOT the full ARN)
AWS_MODEL_ARN_ID=arn:aws:bedrock:{region}:{account-id}
```

**Important**:

- Only provide the **base ARN** (region + account ID)
- Do NOT include `inference-profile` or model name
- The middleware will automatically construct the full ARN from the base + model from your code

### Step 3: Enable Bedrock

```bash
# Set to 0 to enable Bedrock
REVENIUM_BEDROCK_DISABLE=0
```

See the [Bedrock Example](./examples/README.md#bedrock-example) for complete usage.

## Troubleshooting

### Metering data not appearing in Revenium dashboard

**Problem**: Your app runs successfully but no data appears in Revenium.

**Solution**: The middleware sends metering data asynchronously in the background. If your program exits too quickly, the data won't be sent. Add a delay before exit:

```go
// At the end of your main() function
time.Sleep(2 * time.Second)
```

Or use a proper shutdown handler for production apps:

```go
// For HTTP servers or long-running apps
defer func() {
    time.Sleep(2 * time.Second) // Allow metering to complete
}()
```

### "Failed to initialize" error

Check your API keys:

```bash
echo $ANTHROPIC_API_KEY
echo $REVENIUM_METERING_API_KEY
```

### "Authentication failed" error

Make sure your environment URL matches your API key:

- **Production**: `https://api.revenium.ai`

### Enable debug logging

```bash
export REVENIUM_LOG_LEVEL=DEBUG
go run main.go
```

You should see the metering payload being sent:

```
[METERING] Sending payload to https://api.revenium.ai/meter/v2/ai/completions: {...}
```

(The middleware automatically appends `/meter/v2/ai/completions` to your configured base URL)

## Requirements

- **Go**: 1.23 or higher
- **Anthropic API Key**: Get from [console.anthropic.com](https://console.anthropic.com/)
- **Revenium API Key**: Get from [app.revenium.ai](https://app.revenium.ai)

## Documentation

For more information and advanced usage:

- [Revenium Documentation](https://docs.revenium.io)
- [Anthropic Claude API Docs](https://docs.anthropic.com)
- [Anthropic Models List](https://docs.anthropic.com/en/docs/about-claude/models) - Current available models
- [AWS Bedrock Documentation](https://docs.aws.amazon.com/bedrock/)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md)

## Code of Conduct

See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)

## Security

See [SECURITY.md](SECURITY.md)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For issues, feature requests, or contributions:

- **GitHub Repository**: [revenium/revenium-middleware-anthropic-go](https://github.com/revenium/revenium-middleware-anthropic-go)
- **Issues**: [Report bugs or request features](https://github.com/revenium/revenium-middleware-anthropic-go/issues)
- **Documentation**: [docs.revenium.io](https://docs.revenium.io)
- **Contact**: Reach out to the Revenium team for additional support

---

**Built by Revenium**

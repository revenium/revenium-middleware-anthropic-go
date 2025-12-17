# Running Examples - Revenium Middleware for Anthropic (Go)

This guide shows you how to run the Go examples in this repository. Examples are provided for **Anthropic Claude** and **AWS Bedrock** integration.

## Prerequisites

- Go 1.23+ installed
- Anthropic API key
- Revenium API key
- (Optional) AWS credentials for Bedrock examples

## Setup

### 1. Clone the Repository

```bash
git clone https://github.com/revenium/revenium-middleware-anthropic-go.git
cd revenium-middleware-anthropic-go
```

### 2. Install Dependencies

```bash
go mod download
go mod tidy
```

### 3. Configure Environment Variables

Create a `.env` file in the root directory:

```bash
# Create .env file
cp .env.example .env
```

Edit `.env` and add your API keys:

#### For Anthropic Claude (basic/advanced/streaming examples):

```env
# Anthropic Configuration
ANTHROPIC_API_KEY=your_anthropic_api_key_here

# Revenium Configuration
REVENIUM_METERING_API_KEY=your_revenium_api_key_here
REVENIUM_METERING_BASE_URL=https://api.revenium.ai

# Disable Bedrock for basic examples (use Anthropic API directly)
REVENIUM_BEDROCK_DISABLE=1

# Optional: Enable debug logging
REVENIUM_LOG_LEVEL=INFO
```

#### For AWS Bedrock (bedrock example):

```env
# Same as above, plus AWS configuration:
AWS_ACCESS_KEY_ID=your_aws_access_key
AWS_SECRET_ACCESS_KEY=your_aws_secret_key

# AWS Bedrock Base ARN (ONLY the base, NOT the full ARN)
# Format: arn:aws:bedrock:{region}:{account-id}
AWS_MODEL_ARN_ID=arn:aws:bedrock:us-east-1:123456789012

# Enable Bedrock (set to 0)
REVENIUM_BEDROCK_DISABLE=0
```

**How it works:**

- The middleware takes the base ARN from `AWS_MODEL_ARN_ID`
- The model comes from your code (using Anthropic SDK constants)
- The middleware automatically constructs the full ARN and calls Bedrock

**Note:** The middleware automatically appends `/meter/v2/ai/completions` to the base URL. Both trailing slash variations are supported:

- `https://api.revenium.ai`
- `https://api.revenium.ai/`

## Running Examples

### Available Examples

```bash
go run examples/basic/main.go        # Basic chat completion
go run examples/streaming/main.go    # Streaming response
go run examples/advanced/main.go     # With custom metadata
go run examples/bedrock/main.go      # AWS Bedrock integration
```

## Example Structure

```
examples/
├── basic/
│   └── main.go              # Basic chat completion
├── streaming/
│   └── main.go              # Streaming response
├── advanced/
│   └── main.go              # With custom metadata
├── bedrock/
│   └── main.go              # AWS Bedrock integration
└── README.md                # This file
```

## Examples Included

### 1. Basic Usage

**Location**: `examples/basic/main.go`

**What it demonstrates:**

- Simple message creation
- Automatic metering
- Basic error handling

**Run it:**

```bash
go run examples/basic/main.go
```

**Expected output:**

```
=== Revenium Middleware - Basic Example ===

Response:
─────────────────────────────────────────
¡Hola! That's "hello" in Spanish.
─────────────────────────────────────────

Input Tokens: 11
Output Tokens: 18

Basic example completed successfully!
```

---

### 2. Streaming

**Location**: `examples/streaming/main.go`

**What it demonstrates:**

- Real-time streaming responses
- Token tracking for streamed content
- Time to first token measurement

**Run it:**

```bash
go run examples/streaming/main.go
```

**Expected output:**

```
=== Revenium Middleware - Streaming Example ===

Streaming response:
──────────
Code flows like water
Logic blooms in silent thought
Bugs hide in shadows


Streaming example completed successfully!
```

---

### 3. Advanced Features

**Location**: `examples/advanced/main.go`

**What it demonstrates:**

- Custom metadata tracking
- Business context in metering
- Advanced configuration

**Run it:**

```bash
go run examples/advanced/main.go
```

**Expected output:**

```
=== Revenium Middleware - Advanced Example ===

Response:
─────────────────────────────────────────
# Benefits of Using Middleware for API Metering

Middleware for API metering offers several significant advantages...
[Full response text]
─────────────────────────────────────────

Input Tokens: 18
Output Tokens: 300

Advanced example completed successfully!
```

**Note:** Custom metadata (organizationId, productId, subscriber, taskType) is automatically sent to Revenium in the background.

---

### 4. AWS Bedrock Integration

**Location**: `examples/bedrock/main.go`

**What it demonstrates:**

- AWS Bedrock provider integration
- Automatic fallback to Anthropic API if Bedrock fails
- Automatic ARN construction from base ARN + model from code

**Setup:**

Update your `.env` file with the Bedrock configuration shown above (see "For AWS Bedrock" section).

The key difference is setting `REVENIUM_BEDROCK_DISABLE=0` to enable Bedrock.

**Run it:**

```bash
go run examples/bedrock/main.go
```

**Expected output:**

```
=== Revenium Middleware - AWS Bedrock Integration Example ===

Response:
─────────────────────────────────────────
# AWS Bedrock

AWS Bedrock is a fully managed service that provides access to foundation models...
[Full response text]
─────────────────────────────────────────

Input Tokens: 19
Output Tokens: 300

Bedrock example completed successfully!
```

**Note:** If Bedrock is properly configured, it will use AWS Bedrock. If there's any issue with AWS credentials or configuration, it will automatically fall back to Anthropic API.

---

## Troubleshooting

### "Missing API Key" Error

Make sure you've created a `.env` file with your API keys:

```bash
# Check if .env exists
ls -la .env

# Verify environment variables are set
go run examples/basic/main.go  # Should load from .env automatically
```

### "Module not found" Error

Download dependencies first:

```bash
go mod download
go mod tidy
```

### AWS Bedrock "Authentication failed"

1. Check your AWS credentials in `.env`:

   - `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` are set correctly
   - OR `AWS_PROFILE` points to a valid profile in `~/.aws/credentials`

2. Verify your base ARN is correct:

   - Format: `arn:aws:bedrock:{region}:{account-id}`
   - Example: `arn:aws:bedrock:us-east-1:123456789012`
   - Do NOT include `inference-profile` or model name

3. Ensure your AWS account has Bedrock permissions

4. Set `REVENIUM_BEDROCK_DISABLE=0` to enable Bedrock

### "No metering data in Revenium dashboard"

The middleware sends data asynchronously. Add a small delay before your program exits:

```go
time.Sleep(2 * time.Second)
```

---

## Getting Your API Keys

### Anthropic API Key

1. Go to [Anthropic Console](https://console.anthropic.com/)
2. Create a new API key
3. Add to `.env` as `ANTHROPIC_API_KEY`

### Revenium API Key

1. Sign up at [Revenium](https://www.revenium.ai)
2. Create a new API key in your dashboard at [app.revenium.ai](https://app.revenium.ai)
3. Add to `.env` as `REVENIUM_METERING_API_KEY`

### AWS Bedrock Setup

1. Create an [AWS Account](https://aws.amazon.com)
2. Enable Bedrock in your region
3. Create IAM credentials with Bedrock permissions
4. Add to `.env` as shown above

---

## Support

For issues or questions:

- [GitHub Issues](https://github.com/revenium/revenium-middleware-anthropic-go/issues)
- [Documentation](https://docs.revenium.io)
- Email: support@revenium.io

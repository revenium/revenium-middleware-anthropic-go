// Deprecated: This module has moved to github.com/revenium/revenium-go-sdk/anthropic.
module github.com/revenium/revenium-middleware-anthropic-go

go 1.23.0

require (
	github.com/anthropics/anthropic-sdk-go v1.14.0
	github.com/aws/aws-sdk-go-v2 v1.39.3
	github.com/aws/aws-sdk-go-v2/config v1.31.13
	github.com/aws/aws-sdk-go-v2/credentials v1.18.17
	github.com/aws/aws-sdk-go-v2/service/bedrockruntime v1.41.1
	github.com/joho/godotenv v1.5.1
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.2 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.29.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.38.7 // indirect
	github.com/aws/smithy-go v1.23.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

retract (
	v1.0.7 // Deprecation tombstone: use revenium-go-sdk/anthropic instead.
	v1.0.6 // Consolidated into revenium-go-sdk/anthropic.
	v1.0.5 // Consolidated into revenium-go-sdk/anthropic.
	v1.0.4 // Consolidated into revenium-go-sdk/anthropic.
	v1.0.3 // Consolidated into revenium-go-sdk/anthropic.
	v1.0.2 // Consolidated into revenium-go-sdk/anthropic.
	v1.0.1 // Consolidated into revenium-go-sdk/anthropic.
)

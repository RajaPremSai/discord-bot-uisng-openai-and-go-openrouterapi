# Design Document

## Overview

This design outlines the migration from OpenAI API to OpenRouter API for the Go Discord bot. OpenRouter provides a unified interface to multiple AI models with an OpenAI-compatible API structure, making the migration relatively straightforward. The design focuses on replacing the OpenAI client with an OpenRouter client while maintaining all existing functionality.

## Architecture

### Current Architecture

- **OpenAI Client**: `github.com/sashabaranov/go-openai` package
- **Configuration**: YAML-based config with OpenAI API key and model settings
- **Request Flow**: Discord Command → Bot Handler → OpenAI Client → OpenAI API
- **Response Handling**: OpenAI Response → Bot Handler → Discord Response

### Target Architecture

- **OpenRouter Client**: Custom HTTP client or adapted OpenAI client pointing to OpenRouter endpoints
- **Configuration**: YAML-based config with OpenRouter API key and model settings
- **Request Flow**: Discord Command → Bot Handler → OpenRouter Client → OpenRouter API
- **Response Handling**: OpenRouter Response → Bot Handler → Discord Response

### Key Changes

1. **API Endpoint**: Change from `https://api.openai.com/v1/` to `https://openrouter.ai/api/v1/`
2. **Authentication**: Use OpenRouter API key with `Authorization: Bearer` header
3. **Model Names**: Update model identifiers to OpenRouter format (e.g., `openai/gpt-4`, `anthropic/claude-3-sonnet`)
4. **Request Headers**: Add required OpenRouter headers (`HTTP-Referer`, `X-Title`)

## Components and Interfaces

### 1. Configuration Component

**File**: `main.go` and `credentials.yaml`

**Changes**:

- Replace `openAI` section with `openRouter` section
- Update API key field name
- Update model list to use OpenRouter model identifiers
- Add optional site URL and app name for OpenRouter headers

**New Configuration Structure**:

```yaml
openRouter:
  apiKey: "sk-or-v1-..."
  baseURL: "https://openrouter.ai/api/v1"
  siteURL: "https://your-site.com" # Optional
  siteName: "Discord Bot" # Optional
  completionModels:
    - "openai/gpt-4"
    - "openai/gpt-3.5-turbo"
    - "anthropic/claude-3-sonnet"
```

### 2. OpenRouter Client Component

**New File**: `pkg/openrouter/client.go`

**Purpose**: Wrapper around HTTP client or adapted OpenAI client for OpenRouter API

**Key Methods**:

- `NewClient(apiKey, baseURL string) *Client`
- `CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error)`
- `CreateImage(ctx context.Context, req ImageRequest) (*ImageResponse, error)`

**Implementation Options**:

1. **Option A**: Use existing `go-openai` package with custom base URL
2. **Option B**: Create custom HTTP client with OpenRouter-specific handling

### 3. Request/Response Models

**File**: `pkg/openrouter/models.go`

**Chat Completion Models**:

```go
type ChatCompletionRequest struct {
    Model       string                    `json:"model"`
    Messages    []ChatCompletionMessage   `json:"messages"`
    Temperature *float32                  `json:"temperature,omitempty"`
    MaxTokens   *int                      `json:"max_tokens,omitempty"`
    Stream      bool                      `json:"stream,omitempty"`
}

type ChatCompletionResponse struct {
    ID      string                 `json:"id"`
    Object  string                 `json:"object"`
    Created int64                  `json:"created"`
    Model   string                 `json:"model"`
    Choices []ChatCompletionChoice `json:"choices"`
    Usage   Usage                  `json:"usage"`
}
```

**Image Generation Models**:

```go
type ImageRequest struct {
    Prompt         string `json:"prompt"`
    Model          string `json:"model"`
    N              int    `json:"n,omitempty"`
    Size           string `json:"size,omitempty"`
    ResponseFormat string `json:"response_format,omitempty"`
}

type ImageResponse struct {
    Created int64       `json:"created"`
    Data    []ImageData `json:"data"`
}
```

### 4. Updated Command Handlers

**Files**:

- `pkg/commands/gpt/handler.go`
- `pkg/commands/dalle/handler.go`

**Changes**:

- Replace `*openai.Client` parameter with `*openrouter.Client`
- Update function calls to use OpenRouter client methods
- Adapt error handling for OpenRouter error responses
- Update model validation logic

### 5. Error Handling Component

**File**: `pkg/openrouter/errors.go`

**OpenRouter Error Structure**:

```go
type ErrorResponse struct {
    Error struct {
        Code    string `json:"code"`
        Message string `json:"message"`
        Type    string `json:"type"`
    } `json:"error"`
}
```

## Data Models

### Model Mapping

- OpenAI `gpt-4` → OpenRouter `openai/gpt-4`
- OpenAI `gpt-3.5-turbo` → OpenRouter `openai/gpt-3.5-turbo`
- OpenAI `dall-e-2` → OpenRouter `openai/dall-e-2`

### Request Headers

```go
headers := map[string]string{
    "Authorization": "Bearer " + apiKey,
    "Content-Type": "application/json",
    "HTTP-Referer": siteURL,        // Optional
    "X-Title": siteName,            // Optional
}
```

### Response Compatibility

OpenRouter maintains OpenAI API compatibility, so most response structures remain the same. Key differences:

- Model names include provider prefix
- Additional metadata in usage statistics
- Different error code formats

## Error Handling

### OpenRouter-Specific Errors

1. **Authentication Errors**: Invalid API key format or expired keys
2. **Model Availability**: Requested model not available or temporarily down
3. **Rate Limiting**: Different rate limits per model and provider
4. **Credit/Balance**: Insufficient credits for API calls

### Error Mapping Strategy

```go
func mapOpenRouterError(err error) error {
    // Parse OpenRouter error response
    // Map to user-friendly Discord messages
    // Maintain logging for debugging
}
```

### Retry Logic

- Implement exponential backoff for transient errors
- Different retry strategies for different error types
- Respect rate limit headers from OpenRouter

## Testing Strategy

### Unit Tests

1. **OpenRouter Client Tests**

   - Mock HTTP responses for various scenarios
   - Test request formatting and header inclusion
   - Validate response parsing

2. **Configuration Tests**

   - Test YAML parsing with new OpenRouter config
   - Validate model name transformations
   - Test default value handling

3. **Error Handling Tests**
   - Test various OpenRouter error responses
   - Validate error message formatting
   - Test retry logic behavior

### Integration Tests

1. **API Integration Tests**

   - Test actual OpenRouter API calls (with test credentials)
   - Validate end-to-end request/response flow
   - Test different model types

2. **Discord Integration Tests**
   - Test Discord command handling with OpenRouter responses
   - Validate message formatting and embedding
   - Test thread creation and management

### Migration Testing

1. **Compatibility Tests**

   - Ensure all existing Discord commands work
   - Validate response formatting matches expectations
   - Test edge cases and error scenarios

2. **Performance Tests**
   - Compare response times between OpenAI and OpenRouter
   - Test concurrent request handling
   - Validate memory usage patterns

## Implementation Approach

### Phase 1: Core Client Implementation

- Create OpenRouter client package
- Implement basic chat completion functionality
- Add configuration parsing for OpenRouter

### Phase 2: Command Handler Updates

- Update GPT command handlers
- Implement proper error handling
- Add logging and monitoring

### Phase 3: Image Generation

- Implement OpenRouter image generation
- Update DALLE command handlers
- Test image response handling

### Phase 4: Testing and Validation

- Comprehensive testing suite
- Performance validation
- Documentation updates

## Migration Considerations

### Backward Compatibility

- Maintain existing Discord command interface
- Preserve all user-facing functionality
- Keep configuration file structure similar

### Deployment Strategy

- Feature flag for OpenRouter vs OpenAI
- Gradual rollout capability
- Rollback plan if issues arise

### Monitoring and Logging

- Enhanced logging for OpenRouter API calls
- Error rate monitoring
- Performance metrics collection

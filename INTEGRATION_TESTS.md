# OpenRouter Integration Tests

This document describes the comprehensive integration tests implemented for OpenRouter API functionality.

## Overview

The integration tests are located in `integration_openrouter_test.go` and provide comprehensive testing of the OpenRouter API integration, including:

- Real API calls to OpenRouter endpoints
- Error handling and rate limiting scenarios
- End-to-end Discord command flow simulation
- Connection testing and model validation
- Logging and metrics verification

## Test Categories

### 1. Chat Completion Tests (`TestOpenRouterChatCompletionIntegration`)

Tests chat completion functionality with real OpenRouter API calls:

- **BasicChatCompletion**: Simple chat completion with user message
- **ChatCompletionWithSystemMessage**: Chat with system message context
- **ChatCompletionWithMultipleMessages**: Conversation context handling

**Requirements Covered**: 1.1, 1.2, 1.3, 2.1

### 2. Image Generation Tests (`TestOpenRouterImageGenerationIntegration`)

Tests image generation functionality with real OpenRouter API calls:

- **BasicImageGeneration**: Single image generation
- **MultipleImageGeneration**: Multiple images in one request
- **LargerImageGeneration**: Different image sizes

**Requirements Covered**: 3.1, 3.2

### 3. Error Scenarios Tests (`TestOpenRouterErrorScenariosIntegration`)

Tests error handling and rate limiting scenarios:

- **InvalidAPIKey**: Authentication error handling
- **InvalidModel**: Model not found error handling
- **EmptyPrompt**: Input validation error handling
- **ContextTimeout**: Timeout error handling

**Requirements Covered**: 5.1, 5.2, 5.3, 5.4, 5.5

### 4. Retry Logic Tests (`TestOpenRouterRetryLogicIntegration`)

Tests retry logic with simulated scenarios:

- **RetryWithValidRequest**: Successful retry after transient error
- **RetryExhaustion**: Retry exhaustion with persistent errors

**Requirements Covered**: 5.5

### 5. Connection and Models Tests (`TestOpenRouterConnectionAndModels`)

Tests basic connectivity and model information:

- **PingConnection**: API connectivity test
- **ListModels**: Available models retrieval
- **GetSpecificModel**: Individual model information

**Requirements Covered**: 1.1, 6.3

### 6. End-to-End Flow Tests (`TestEndToEndDiscordCommandFlow`)

Tests complete Discord command simulation:

- **SimulateChatCommand**: Complete chat command flow
- **SimulateImageCommand**: Complete image command flow
- **SimulateChatWithContext**: Chat with context file simulation

**Requirements Covered**: 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3

### 7. Logging Tests (`TestOpenRouterLoggingIntegration`)

Tests logging functionality during API calls:

- **LoggingDuringChatCompletion**: Request/response logging
- **LoggingDuringError**: Error logging verification

**Requirements Covered**: 6.3

## Running the Tests

### Prerequisites

1. **OpenRouter API Key**: Required for most tests

   - Set `OPENROUTER_API_KEY` environment variable, OR
   - Configure `credentials.yaml` with valid OpenRouter API key

2. **Go Dependencies**: Ensure all dependencies are installed
   ```bash
   go mod tidy
   ```

### Running All Integration Tests

```bash
# Run all OpenRouter integration tests
go test -v -run TestOpenRouter -timeout 60s

# Run with short flag to skip tests requiring API key
go test -v -run TestOpenRouter -short -timeout 60s
```

### Running Specific Test Categories

```bash
# Chat completion tests only
go test -v -run TestOpenRouterChatCompletionIntegration -timeout 60s

# Image generation tests only
go test -v -run TestOpenRouterImageGenerationIntegration -timeout 60s

# Error scenarios tests only
go test -v -run TestOpenRouterErrorScenariosIntegration -timeout 60s

# Connection tests only
go test -v -run TestOpenRouterConnectionAndModels -timeout 60s
```

## Test Behavior

### With Valid API Key

When a valid OpenRouter API key is provided:

- All tests will execute with real API calls
- Tests verify actual API responses and behavior
- Network timeouts and rate limits may affect test duration
- Tests may consume OpenRouter credits

### Without API Key

When no API key is provided:

- Most tests will be **skipped** (expected behavior)
- Tests that don't require real API calls will still run (e.g., InvalidAPIKey test)
- No API credits will be consumed
- Tests complete quickly

## Test Configuration

### Default Test Models

- **Chat Model**: `openai/gpt-3.5-turbo`
- **Image Model**: `openai/dall-e-2`

### Timeouts

- **Individual Test Timeout**: 30 seconds
- **Overall Test Suite Timeout**: 60 seconds

### Test Client Configuration

Tests use a properly configured OpenRouter client with:

- Custom site URL and name for testing
- Comprehensive logging enabled
- Proper error handling and retry logic

## Expected Outcomes

### Successful Test Run (with API key)

```
=== RUN   TestOpenRouterChatCompletionIntegration
=== RUN   TestOpenRouterChatCompletionIntegration/BasicChatCompletion
--- PASS: TestOpenRouterChatCompletionIntegration/BasicChatCompletion (2.34s)
=== RUN   TestOpenRouterChatCompletionIntegration/ChatCompletionWithSystemMessage
--- PASS: TestOpenRouterChatCompletionIntegration/ChatCompletionWithSystemMessage (1.87s)
...
PASS
```

### Test Run Without API Key

```
=== RUN   TestOpenRouterChatCompletionIntegration
--- SKIP: TestOpenRouterChatCompletionIntegration (0.00s)
=== RUN   TestOpenRouterImageGenerationIntegration
--- SKIP: TestOpenRouterImageGenerationIntegration (0.00s)
...
PASS
```

## Troubleshooting

### Common Issues

1. **Tests Skipped**: Normal behavior when no API key is configured
2. **Timeout Errors**: Increase timeout or check network connectivity
3. **Authentication Errors**: Verify API key format and validity
4. **Rate Limit Errors**: Wait and retry, or use different API key

### Debug Mode

Enable debug logging by setting environment variable:

```bash
export OPENROUTER_LOG_LEVEL=DEBUG
go test -v -run TestOpenRouter -timeout 60s
```

## Integration with CI/CD

These tests can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions step
- name: Run OpenRouter Integration Tests
  run: |
    if [ -n "$OPENROUTER_API_KEY" ]; then
      go test -v -run TestOpenRouter -timeout 60s
    else
      echo "Skipping integration tests - no API key provided"
      go test -v -run TestOpenRouter -short -timeout 60s
    fi
  env:
    OPENROUTER_API_KEY: ${{ secrets.OPENROUTER_API_KEY }}
```

## Maintenance

### Adding New Tests

1. Follow existing test patterns in `integration_openrouter_test.go`
2. Use `createTestClient(t)` helper for consistent client setup
3. Handle API key absence gracefully with `t.Skip()`
4. Include proper timeout handling
5. Verify against specific requirements

### Updating Test Models

Update the constants at the top of the test file:

```go
const (
    testModel      = "openai/gpt-3.5-turbo"  // Update as needed
    testImageModel = "openai/dall-e-2"       // Update as needed
)
```

## Requirements Coverage

This integration test suite covers all requirements specified in task 15:

- ✅ **Write integration tests for chat completion with real OpenRouter API calls**
- ✅ **Create integration tests for image generation with OpenRouter API**
- ✅ **Test error scenarios and rate limiting with OpenRouter API**
- ✅ **Validate end-to-end Discord command flow with OpenRouter responses**
- ✅ **Requirements 1.1, 1.2, 1.3, 2.1, 3.1 coverage verified**

The tests provide comprehensive validation of the OpenRouter API migration functionality while being robust enough to handle various deployment scenarios.

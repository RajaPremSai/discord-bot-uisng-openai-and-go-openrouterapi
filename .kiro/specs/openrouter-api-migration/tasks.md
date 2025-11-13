# Implementation Plan

- [x] 1. Create OpenRouter client package structure

  - Create `pkg/openrouter/` directory with core client files
  - Define OpenRouter-specific data models and interfaces
  - Set up package structure for maintainable code organization
  - _Requirements: 6.1, 6.2_

-

- [x] 2. Implement OpenRouter data models and request/response structures

  - Create `pkg/openrouter/models.go` with chat completion and image generation models
  - Define error response structures for OpenRouter API
  - Implement JSON marshaling/unmarshaling for all data types
  - Write unit tests for model serialization and deserialization
  - _Requirements: 1.2, 1.3, 5.1_

- [x] 3. Create OpenRouter HTTP client with proper authentication

  - Implement `pkg/openrouter/client.go` with HTTP client wrapper
  - Add OpenRouter-specific headers (Authorization, HTTP-Referer, X-Title)
  - Implement base URL configuration and request building
  - Write unit tests for client initialization and header setting
  - _Requirements: 1.1, 1.2, 4.3_

- [x] 4. Implement chat completion functionality for OpenRouter

  - Add `CreateChatCompletion` method to OpenRouter client
  - Handle request formatting and response parsing for chat completions
  - Implement proper error handling for OpenRouter chat API responses
  - Write unit tests with mocked HTTP responses for chat completion
  - _Requirements: 2.1, 2.2, 5.1, 5.2_

- [x] 5. Implement image generation functionality for OpenRouter

  - Add `CreateImage` method to OpenRouter client
  - Handle request formatting and response parsing for image generation
  - Implement proper error handling for OpenRouter image API responses
  - Write unit tests with mocked HTTP responses for image generation
  - _Requirements: 3.1, 3.2, 5.1_

- [x] 6. Update configuration system for OpenRouter

  - Modify `Config` struct in `main.go` to support OpenRouter configuration
  - Update `credentials.yaml` structure to use OpenRouter API key and settings
  - Add validation for OpenRouter configuration parameters
  - Write unit tests for configuration parsing and validation
  - _Requirements: 4.1, 4.2, 4.3, 4.4_

- [x] 7. Create OpenRouter error handling and mapping utilities

  - Implement `pkg/openrouter/errors.go` with error parsing and mapping
  - Create user-friendly error messages for common OpenRouter error types
  - Add retry logic for transient errors with exponential backoff
  - Write unit tests for error handling and retry mechanisms
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

- [x] 8. Update GPT command handlers to use OpenRouter client

  - Modify `pkg/commands/gpt/handler.go` to use OpenRouter client instead of OpenAI
  - Update `chatGPTHandler` function to work with OpenRouter API calls
  - Adapt model name handling for OpenRouter model format (e.g., "openai/gpt-4")
  - Update error handling to use OpenRouter error responses
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

- [x] 9. Update GPT command configuration and model selection

  - Modify `pkg/commands/gpt/command.go` to work with OpenRouter models
  - Update model validation and selection logic for OpenRouter model names
  - Adapt temperature and parameter handling for OpenRouter API
  - Write unit tests for command option parsing with OpenRouter models
  - _Requirements: 2.3, 4.4_

- [x] 10. Update DALLE command handlers to use OpenRouter client


  - Modify `pkg/commands/dalle/handler.go` to use OpenRouter client instead of OpenAI
  - Update `imageHandler` function to work with OpenRouter image generation API
  - Adapt image parameter handling for OpenRouter image models
  - Update error handling for OpenRouter image generation responses
  - _Requirements: 3.1, 3.2, 3.3_

- [ ] 11. Update DALLE command configuration for OpenRouter

  - Modify `pkg/commands/dalle/command.go` to work with OpenRouter image models
  - Update image size and parameter options for OpenRouter compatibility
  - Adapt model selection for OpenRouter image generation models
  - Write unit tests for image command option parsing
  - _Requirements: 3.1, 3.2_

- [ ] 12. Update main application initialization for OpenRouter

  - Modify `main.go` to initialize OpenRouter client instead of OpenAI client
  - Update client initialization with OpenRouter configuration parameters
  - Adapt command registration to use OpenRouter client
  - Add proper logging for OpenRouter client initialization
  - _Requirements: 1.1, 4.1, 6.3_

- [ ] 13. Update message handling and caching for OpenRouter compatibility

  - Modify `pkg/commands/gpt/message_handler.go` to work with OpenRouter responses
  - Update message caching structures to handle OpenRouter model names
  - Adapt token counting and message truncation for OpenRouter models
  - Write unit tests for message handling with OpenRouter data structures
  - _Requirements: 2.4, 2.5, 6.2_

- [ ] 14. Add comprehensive logging for OpenRouter API interactions

  - Implement detailed logging for OpenRouter API requests and responses
  - Add performance metrics logging for API call duration and success rates
  - Include error logging with OpenRouter-specific error details
  - Write unit tests for logging functionality
  - _Requirements: 6.3, 5.5_

- [ ] 15. Create integration tests for OpenRouter API functionality

  - Write integration tests for chat completion with real OpenRouter API calls
  - Create integration tests for image generation with OpenRouter API
  - Test error scenarios and rate limiting with OpenRouter API
  - Validate end-to-end Discord command flow with OpenRouter responses
  - _Requirements: 1.1, 1.2, 1.3, 2.1, 3.1_

- [ ] 16. Update go.mod dependencies and remove OpenAI package

  - Remove `github.com/sashabaranov/go-openai` dependency from go.mod
  - Add any new dependencies required for OpenRouter HTTP client
  - Update import statements throughout codebase to remove OpenAI references
  - Run `go mod tidy` to clean up unused dependencies
  - _Requirements: 6.4_

- [ ] 17. Create comprehensive unit test suite for OpenRouter migration

  - Write unit tests for all OpenRouter client methods with mocked responses
  - Create tests for configuration parsing and validation
  - Add tests for error handling and retry logic
  - Test model name mapping and parameter conversion
  - _Requirements: 1.2, 1.3, 4.2, 5.1_

- [ ] 18. Update documentation and configuration examples
  - Update README.md with OpenRouter configuration instructions
  - Create example `credentials.yaml` file with OpenRouter settings
  - Document model name mappings and available OpenRouter models
  - Add troubleshooting guide for common OpenRouter API issues
  - _Requirements: 4.1, 4.2, 6.1_

package openrouter

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestChatCompletionRequest_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		request  ChatCompletionRequest
		expected string
	}{
		{
			name: "basic request",
			request: ChatCompletionRequest{
				Model: "openai/gpt-4",
				Messages: []ChatCompletionMessage{
					{Role: "user", Content: "Hello"},
				},
			},
			expected: `{"model":"openai/gpt-4","messages":[{"role":"user","content":"Hello"}],"stream":false}`,
		},
		{
			name: "request with optional fields",
			request: ChatCompletionRequest{
				Model: "openai/gpt-3.5-turbo",
				Messages: []ChatCompletionMessage{
					{Role: "system", Content: "You are a helpful assistant"},
					{Role: "user", Content: "Hello", Name: "user1"},
				},
				Temperature:      floatPtr(0.7),
				MaxTokens:        intPtr(100),
				TopP:             floatPtr(0.9),
				FrequencyPenalty: floatPtr(0.1),
				PresencePenalty:  floatPtr(0.2),
				Stream:           true,
				Stop:             []string{"END"},
				User:             "test-user",
			},
			expected: `{"model":"openai/gpt-3.5-turbo","messages":[{"role":"system","content":"You are a helpful assistant"},{"role":"user","content":"Hello","name":"user1"}],"temperature":0.7,"max_tokens":100,"top_p":0.9,"frequency_penalty":0.1,"presence_penalty":0.2,"stream":true,"stop":["END"],"user":"test-user"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}

			// Parse both JSON strings to compare structure
			var expected, actual map[string]interface{}
			if err := json.Unmarshal([]byte(tt.expected), &expected); err != nil {
				t.Fatalf("Expected JSON parse error: %v", err)
			}
			if err := json.Unmarshal(data, &actual); err != nil {
				t.Fatalf("Actual JSON parse error: %v", err)
			}

			if !reflect.DeepEqual(expected, actual) {
				t.Errorf("Expected %s, got %s", tt.expected, string(data))
			}
		})
	}
}

func TestChatCompletionRequest_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected ChatCompletionRequest
	}{
		{
			name:     "basic request",
			jsonData: `{"model":"openai/gpt-4","messages":[{"role":"user","content":"Hello"}],"stream":false}`,
			expected: ChatCompletionRequest{
				Model: "openai/gpt-4",
				Messages: []ChatCompletionMessage{
					{Role: "user", Content: "Hello"},
				},
				Stream: false,
			},
		},
		{
			name:     "request with optional fields",
			jsonData: `{"model":"openai/gpt-3.5-turbo","messages":[{"role":"system","content":"You are a helpful assistant"},{"role":"user","content":"Hello","name":"user1"}],"temperature":0.7,"max_tokens":100,"top_p":0.9,"frequency_penalty":0.1,"presence_penalty":0.2,"stream":true,"stop":["END"],"user":"test-user"}`,
			expected: ChatCompletionRequest{
				Model: "openai/gpt-3.5-turbo",
				Messages: []ChatCompletionMessage{
					{Role: "system", Content: "You are a helpful assistant"},
					{Role: "user", Content: "Hello", Name: "user1"},
				},
				Temperature:      floatPtr(0.7),
				MaxTokens:        intPtr(100),
				TopP:             floatPtr(0.9),
				FrequencyPenalty: floatPtr(0.1),
				PresencePenalty:  floatPtr(0.2),
				Stream:           true,
				Stop:             []string{"END"},
				User:             "test-user",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var request ChatCompletionRequest
			err := json.Unmarshal([]byte(tt.jsonData), &request)
			if err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}

			if !reflect.DeepEqual(tt.expected, request) {
				t.Errorf("Expected %+v, got %+v", tt.expected, request)
			}
		})
	}
}

func TestChatCompletionResponse_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"id": "chatcmpl-123",
		"object": "chat.completion",
		"created": 1677652288,
		"model": "openai/gpt-4",
		"choices": [
			{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "Hello! How can I help you today?"
				},
				"finish_reason": "stop"
			}
		],
		"usage": {
			"prompt_tokens": 9,
			"completion_tokens": 12,
			"total_tokens": 21,
			"prompt_cost": 0.00027,
			"completion_cost": 0.00036,
			"total_cost": 0.00063
		}
	}`

	expected := ChatCompletionResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion",
		Created: 1677652288,
		Model:   "openai/gpt-4",
		Choices: []ChatCompletionChoice{
			{
				Index: 0,
				Message: ChatCompletionMessage{
					Role:    "assistant",
					Content: "Hello! How can I help you today?",
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     9,
			CompletionTokens: 12,
			TotalTokens:      21,
			PromptCost:       0.00027,
			CompletionCost:   0.00036,
			TotalCost:        0.00063,
		},
	}

	var response ChatCompletionResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(expected, response) {
		t.Errorf("Expected %+v, got %+v", expected, response)
	}
}

func TestImageRequest_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		request  ImageRequest
		expected string
	}{
		{
			name: "basic image request",
			request: ImageRequest{
				Prompt: "A beautiful sunset",
				Model:  "openai/dall-e-2",
			},
			expected: `{"prompt":"A beautiful sunset","model":"openai/dall-e-2"}`,
		},
		{
			name: "image request with all fields",
			request: ImageRequest{
				Prompt:         "A cat wearing a hat",
				Model:          "openai/dall-e-3",
				N:              2,
				Size:           "1024x1024",
				ResponseFormat: "url",
				User:           "test-user",
				Quality:        "hd",
				Style:          "vivid",
			},
			expected: `{"prompt":"A cat wearing a hat","model":"openai/dall-e-3","n":2,"size":"1024x1024","response_format":"url","user":"test-user","quality":"hd","style":"vivid"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}

			// Parse both JSON strings to compare structure
			var expected, actual map[string]interface{}
			if err := json.Unmarshal([]byte(tt.expected), &expected); err != nil {
				t.Fatalf("Expected JSON parse error: %v", err)
			}
			if err := json.Unmarshal(data, &actual); err != nil {
				t.Fatalf("Actual JSON parse error: %v", err)
			}

			if !reflect.DeepEqual(expected, actual) {
				t.Errorf("Expected %s, got %s", tt.expected, string(data))
			}
		})
	}
}

func TestImageRequest_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected ImageRequest
	}{
		{
			name:     "basic image request",
			jsonData: `{"prompt":"A beautiful sunset","model":"openai/dall-e-2"}`,
			expected: ImageRequest{
				Prompt: "A beautiful sunset",
				Model:  "openai/dall-e-2",
			},
		},
		{
			name:     "image request with all fields",
			jsonData: `{"prompt":"A cat wearing a hat","model":"openai/dall-e-3","n":2,"size":"1024x1024","response_format":"url","user":"test-user","quality":"hd","style":"vivid"}`,
			expected: ImageRequest{
				Prompt:         "A cat wearing a hat",
				Model:          "openai/dall-e-3",
				N:              2,
				Size:           "1024x1024",
				ResponseFormat: "url",
				User:           "test-user",
				Quality:        "hd",
				Style:          "vivid",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var request ImageRequest
			err := json.Unmarshal([]byte(tt.jsonData), &request)
			if err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}

			if !reflect.DeepEqual(tt.expected, request) {
				t.Errorf("Expected %+v, got %+v", tt.expected, request)
			}
		})
	}
}

func TestImageResponse_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"created": 1677652288,
		"data": [
			{
				"url": "https://example.com/image1.png",
				"revised_prompt": "A beautiful sunset over the ocean"
			},
			{
				"b64_json": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="
			}
		]
	}`

	expected := ImageResponse{
		Created: 1677652288,
		Data: []ImageData{
			{
				URL:           "https://example.com/image1.png",
				RevisedPrompt: "A beautiful sunset over the ocean",
			},
			{
				B64JSON: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==",
			},
		},
	}

	var response ImageResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(expected, response) {
		t.Errorf("Expected %+v, got %+v", expected, response)
	}
}

func TestErrorResponse_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"error": {
			"code": "invalid_request_error",
			"message": "Invalid model specified",
			"type": "invalid_request_error",
			"param": "model"
		}
	}`

	expected := ErrorResponse{
		ErrorDetail: ErrorDetail{
			Code:    "invalid_request_error",
			Message: "Invalid model specified",
			Type:    "invalid_request_error",
			Param:   "model",
		},
	}

	var response ErrorResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(expected, response) {
		t.Errorf("Expected %+v, got %+v", expected, response)
	}
}

func TestErrorResponse_Error(t *testing.T) {
	errorResp := ErrorResponse{
		ErrorDetail: ErrorDetail{
			Code:    "invalid_request_error",
			Message: "Invalid model specified",
			Type:    "invalid_request_error",
		},
	}

	expected := "OpenRouter API error: Invalid model specified (code: invalid_request_error, type: invalid_request_error)"
	actual := errorResp.Error()

	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

func TestModelsResponse_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"object": "list",
		"data": [
			{
				"id": "openai/gpt-4",
				"object": "model",
				"created": 1677610602,
				"owned_by": "openai",
				"root": "gpt-4",
				"parent": null
			},
			{
				"id": "anthropic/claude-3-sonnet",
				"object": "model",
				"created": 1677610602,
				"owned_by": "anthropic"
			}
		]
	}`

	expected := ModelsResponse{
		Object: "list",
		Data: []Model{
			{
				ID:      "openai/gpt-4",
				Object:  "model",
				Created: 1677610602,
				OwnedBy: "openai",
				Root:    "gpt-4",
			},
			{
				ID:      "anthropic/claude-3-sonnet",
				Object:  "model",
				Created: 1677610602,
				OwnedBy: "anthropic",
			},
		},
	}

	var response ModelsResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(expected, response) {
		t.Errorf("Expected %+v, got %+v", expected, response)
	}
}

func TestStreamResponse_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"id": "chatcmpl-123",
		"object": "chat.completion.chunk",
		"created": 1677652288,
		"model": "openai/gpt-4",
		"choices": [
			{
				"index": 0,
				"delta": {
					"role": "assistant",
					"content": "Hello"
				},
				"finish_reason": null
			}
		]
	}`

	expected := StreamResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion.chunk",
		Created: 1677652288,
		Model:   "openai/gpt-4",
		Choices: []StreamChoice{
			{
				Index: 0,
				Delta: ChatCompletionMessage{
					Role:    "assistant",
					Content: "Hello",
				},
				FinishReason: "",
			},
		},
	}

	var response StreamResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(expected, response) {
		t.Errorf("Expected %+v, got %+v", expected, response)
	}
}

func TestChatCompletionRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request ChatCompletionRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: ChatCompletionRequest{
				Model: "openai/gpt-4",
				Messages: []ChatCompletionMessage{
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing model",
			request: ChatCompletionRequest{
				Messages: []ChatCompletionMessage{
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr: true,
			errMsg:  "model is required",
		},
		{
			name: "no messages",
			request: ChatCompletionRequest{
				Model:    "openai/gpt-4",
				Messages: []ChatCompletionMessage{},
			},
			wantErr: true,
			errMsg:  "at least one message is required",
		},
		{
			name: "message missing role",
			request: ChatCompletionRequest{
				Model: "openai/gpt-4",
				Messages: []ChatCompletionMessage{
					{Content: "Hello"},
				},
			},
			wantErr: true,
			errMsg:  "message 0: role is required",
		},
		{
			name: "message missing content",
			request: ChatCompletionRequest{
				Model: "openai/gpt-4",
				Messages: []ChatCompletionMessage{
					{Role: "user"},
				},
			},
			wantErr: true,
			errMsg:  "message 0: content is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errMsg {
					t.Errorf("Expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestImageRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request ImageRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: ImageRequest{
				Prompt: "A beautiful sunset",
				Model:  "openai/dall-e-2",
				N:      1,
			},
			wantErr: false,
		},
		{
			name: "missing prompt",
			request: ImageRequest{
				Model: "openai/dall-e-2",
				N:     1,
			},
			wantErr: true,
			errMsg:  "prompt is required",
		},
		{
			name: "missing model",
			request: ImageRequest{
				Prompt: "A beautiful sunset",
				N:      1,
			},
			wantErr: true,
			errMsg:  "model is required",
		},
		{
			name: "negative n",
			request: ImageRequest{
				Prompt: "A beautiful sunset",
				Model:  "openai/dall-e-2",
				N:      -1,
			},
			wantErr: true,
			errMsg:  "n must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errMsg {
					t.Errorf("Expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// Helper functions for creating pointers
func floatPtr(f float32) *float32 {
	return &f
}

func intPtr(i int) *int {
	return &i
}
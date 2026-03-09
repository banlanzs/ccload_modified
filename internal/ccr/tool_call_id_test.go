package ccr

import (
	"testing"
)

func TestGenerateToolCallID(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		args     map[string]interface{}
		wantSame bool // 是否期望与另一个测试用例生成相同的 ID
	}{
		{
			name:     "simple_function_no_args",
			funcName: "get_weather",
			args:     nil,
			wantSame: false,
		},
		{
			name:     "simple_function_with_args",
			funcName: "get_weather",
			args:     map[string]interface{}{"location": "Beijing"},
			wantSame: false,
		},
		{
			name:     "same_function_same_args_should_generate_same_id",
			funcName: "get_weather",
			args:     map[string]interface{}{"location": "Beijing"},
			wantSame: true, // 应该与上一个测试用例生成相同的 ID
		},
		{
			name:     "same_function_different_args",
			funcName: "get_weather",
			args:     map[string]interface{}{"location": "Shanghai"},
			wantSame: false,
		},
		{
			name:     "different_function_same_args",
			funcName: "get_temperature",
			args:     map[string]interface{}{"location": "Beijing"},
			wantSame: false,
		},
	}

	var prevID string
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := GenerateToolCallID(tt.funcName, tt.args)

			// 验证 ID 格式
			if len(id) == 0 {
				t.Errorf("GenerateToolCallID() returned empty ID")
			}
			if id[:5] != "call_" {
				t.Errorf("GenerateToolCallID() ID should start with 'call_', got: %s", id)
			}
			if len(id) != 29 { // "call_" + 24 hex chars
				t.Errorf("GenerateToolCallID() ID length should be 29, got: %d (%s)", len(id), id)
			}

			// 验证确定性
			id2 := GenerateToolCallID(tt.funcName, tt.args)
			if id != id2 {
				t.Errorf("GenerateToolCallID() should be deterministic, got different IDs: %s vs %s", id, id2)
			}

			// 验证是否应该与前一个 ID 相同
			if i > 0 {
				if tt.wantSame {
					if id != prevID {
						t.Errorf("Expected same ID as previous test, got: %s, want: %s", id, prevID)
					}
				} else {
					if id == prevID {
						t.Errorf("Expected different ID from previous test, but got same: %s", id)
					}
				}
			}

			prevID = id
			t.Logf("Generated ID: %s for function: %s", id, tt.funcName)
		})
	}
}

func TestExtractToolCallID(t *testing.T) {
	tests := []struct {
		name     string
		toolCall *CanonicalToolCall
		wantID   string
	}{
		{
			name:     "nil_tool_call",
			toolCall: nil,
			wantID:   "",
		},
		{
			name: "existing_id",
			toolCall: &CanonicalToolCall{
				ID:   "call_existing123",
				Name: "test_func",
				Args: map[string]interface{}{"key": "value"},
			},
			wantID: "call_existing123",
		},
		{
			name: "no_id_should_generate",
			toolCall: &CanonicalToolCall{
				ID:   "",
				Name: "test_func",
				Args: map[string]interface{}{"key": "value"},
			},
			wantID: "", // 会生成，但我们不知道具体值，只验证非空
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID := ExtractToolCallID(tt.toolCall)

			if tt.wantID != "" {
				if gotID != tt.wantID {
					t.Errorf("ExtractToolCallID() = %v, want %v", gotID, tt.wantID)
				}
			} else if tt.toolCall != nil && tt.toolCall.ID == "" {
				// 应该生成 ID
				if gotID == "" {
					t.Errorf("ExtractToolCallID() should generate ID, got empty")
				}
				if gotID[:5] != "call_" {
					t.Errorf("ExtractToolCallID() generated ID should start with 'call_', got: %s", gotID)
				}
			}
		})
	}
}

func TestGeminiToolCallRoundTrip(t *testing.T) {
	// 测试 OpenAI -> Gemini -> OpenAI 的工具调用 ID 一致性
	openaiPayload := []byte(`{
		"model": "gpt-4",
		"messages": [
			{
				"role": "user",
				"content": "What's the weather in Beijing?"
			},
			{
				"role": "assistant",
				"tool_calls": [
					{
						"id": "call_abc123",
						"type": "function",
						"function": {
							"name": "get_weather",
							"arguments": "{\"location\":\"Beijing\"}"
						}
					}
				]
			}
		],
		"tools": [
			{
				"type": "function",
				"function": {
					"name": "get_weather",
					"description": "Get weather",
					"parameters": {
						"type": "object",
						"properties": {
							"location": {"type": "string"}
						}
					}
				}
			}
		]
	}`)

	// Step 1: OpenAI -> Canonical
	openaiConv := NewOpenAIConverter()
	canonical, err := openaiConv.ToCanonical(openaiPayload)
	if err != nil {
		t.Fatalf("OpenAI ToCanonical failed: %v", err)
	}

	// 验证 Canonical 中有 tool_call
	if len(canonical.Messages) < 2 {
		t.Fatalf("Expected at least 2 messages, got %d", len(canonical.Messages))
	}
	assistantMsg := canonical.Messages[1]
	if len(assistantMsg.Content) == 0 {
		t.Fatalf("Expected assistant message to have content")
	}

	var toolCallPart *CanonicalPart
	for i := range assistantMsg.Content {
		if assistantMsg.Content[i].Type == "tool_call" {
			toolCallPart = &assistantMsg.Content[i]
			break
		}
	}
	if toolCallPart == nil || toolCallPart.ToolCall == nil {
		t.Fatalf("Expected tool_call in assistant message")
	}

	originalID := toolCallPart.ToolCall.ID
	t.Logf("Original OpenAI tool call ID: %s", originalID)

	// Step 2: Canonical -> Gemini
	geminiConv := NewGeminiConverter()
	geminiPayload, err := geminiConv.FromCanonical(canonical)
	if err != nil {
		t.Fatalf("Gemini FromCanonical failed: %v", err)
	}
	t.Logf("Gemini payload: %s", string(geminiPayload))

	// Step 3: Gemini -> Canonical
	canonical2, err := geminiConv.ToCanonical(geminiPayload)
	if err != nil {
		t.Fatalf("Gemini ToCanonical failed: %v", err)
	}

	// 验证 tool_call ID 被重新生成
	if len(canonical2.Messages) < 2 {
		t.Fatalf("Expected at least 2 messages after round trip, got %d", len(canonical2.Messages))
	}
	assistantMsg2 := canonical2.Messages[1]
	if len(assistantMsg2.Content) == 0 {
		t.Fatalf("Expected assistant message to have content after round trip")
	}

	var toolCallPart2 *CanonicalPart
	for i := range assistantMsg2.Content {
		if assistantMsg2.Content[i].Type == "tool_call" {
			toolCallPart2 = &assistantMsg2.Content[i]
			break
		}
	}
	if toolCallPart2 == nil || toolCallPart2.ToolCall == nil {
		t.Fatalf("Expected tool_call in assistant message after round trip")
	}

	regeneratedID := toolCallPart2.ToolCall.ID
	t.Logf("Regenerated tool call ID: %s", regeneratedID)

	// 验证 ID 已被生成（不为空）
	if regeneratedID == "" {
		t.Errorf("Tool call ID should be generated, got empty")
	}

	// 验证 ID 格式
	if regeneratedID[:5] != "call_" {
		t.Errorf("Generated ID should start with 'call_', got: %s", regeneratedID)
	}

	// 验证 name 和 args 保持一致
	if toolCallPart2.ToolCall.Name != toolCallPart.ToolCall.Name {
		t.Errorf("Tool call name changed: %s -> %s", toolCallPart.ToolCall.Name, toolCallPart2.ToolCall.Name)
	}

	// Step 4: Canonical -> OpenAI (验证可以转回 OpenAI 格式)
	openaiPayload2, err := openaiConv.FromCanonical(canonical2)
	if err != nil {
		t.Fatalf("OpenAI FromCanonical failed: %v", err)
	}
	t.Logf("Final OpenAI payload: %s", string(openaiPayload2))
}

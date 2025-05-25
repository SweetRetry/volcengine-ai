package service

import (
	"fmt"
	"reflect"
	"testing"
)

// compareMapWithTypeConversion 比较两个map，考虑JSON序列化的类型转换
func compareMapWithTypeConversion(original, deserialized map[string]interface{}) bool {
	if len(original) != len(deserialized) {
		return false
	}

	for key, originalValue := range original {
		deserializedValue, exists := deserialized[key]
		if !exists {
			return false
		}

		if !compareValues(originalValue, deserializedValue) {
			return false
		}
	}

	return true
}

// compareValues 比较两个值，考虑JSON序列化的类型转换
func compareValues(original, deserialized interface{}) bool {
	// 处理数字类型转换：int -> float64
	if originalInt, ok := original.(int); ok {
		if deserializedFloat, ok := deserialized.(float64); ok {
			return float64(originalInt) == deserializedFloat
		}
	}

	// 处理嵌套map - 必须在类型相同检查之前
	if originalMap, ok := original.(map[string]interface{}); ok {
		if deserializedMap, ok := deserialized.(map[string]interface{}); ok {
			return compareMapWithTypeConversion(originalMap, deserializedMap)
		}
	}

	// 处理slice
	if originalSlice, ok := original.([]interface{}); ok {
		if deserializedSlice, ok := deserialized.([]interface{}); ok {
			if len(originalSlice) != len(deserializedSlice) {
				return false
			}
			for i, originalItem := range originalSlice {
				if !compareValues(originalItem, deserializedSlice[i]) {
					return false
				}
			}
			return true
		}
	}

	// 如果类型相同，直接比较
	if reflect.TypeOf(original) == reflect.TypeOf(deserialized) {
		return reflect.DeepEqual(original, deserialized)
	}

	// 其他情况直接比较
	return reflect.DeepEqual(original, deserialized)
}

func TestTaskService_SerializeInput(t *testing.T) {
	ts := &TaskService{}

	tests := []struct {
		name     string
		input    map[string]interface{}
		expected string
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: "{}",
		},
		{
			name:     "empty input",
			input:    map[string]interface{}{},
			expected: "{}",
		},
		{
			name: "simple input",
			input: map[string]interface{}{
				"prompt": "test prompt",
				"model":  "test-model",
			},
			expected: `{"model":"test-model","prompt":"test prompt"}`,
		},
		{
			name: "complex input",
			input: map[string]interface{}{
				"prompt":  "一只可爱的小猫",
				"model":   "doubao-seedream-3.0-t2i",
				"size":    "1024x1024",
				"quality": "standard",
				"options": map[string]interface{}{
					"style": "anime",
					"n":     1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ts.serializeInput(tt.input)

			if tt.name == "complex input" {
				// 对于复杂输入，我们验证能否正确反序列化
				deserialized, err := ts.deserializeInput(result)
				if err != nil {
					t.Errorf("反序列化失败: %v", err)
				}

				// 验证关键字段
				if deserialized["prompt"] != tt.input["prompt"] {
					t.Errorf("prompt不匹配: %v vs %v", deserialized["prompt"], tt.input["prompt"])
				}
				if deserialized["model"] != tt.input["model"] {
					t.Errorf("model不匹配: %v vs %v", deserialized["model"], tt.input["model"])
				}

				// 验证嵌套的options
				originalOptions := tt.input["options"].(map[string]interface{})
				deserializedOptions := deserialized["options"].(map[string]interface{})

				if deserializedOptions["style"] != originalOptions["style"] {
					t.Errorf("options.style不匹配: %v vs %v", deserializedOptions["style"], originalOptions["style"])
				}

				// 数字类型会被转换为float64，所以需要特殊处理
				if deserializedOptions["n"] != float64(originalOptions["n"].(int)) {
					t.Errorf("options.n不匹配: %v vs %v", deserializedOptions["n"], float64(originalOptions["n"].(int)))
				}
			} else {
				if result != tt.expected {
					t.Errorf("serializeInput() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestTaskService_DeserializeInput(t *testing.T) {
	ts := &TaskService{}

	tests := []struct {
		name      string
		inputStr  string
		expected  map[string]interface{}
		expectErr bool
	}{
		{
			name:     "empty string",
			inputStr: "",
			expected: map[string]interface{}{},
		},
		{
			name:     "empty json",
			inputStr: "{}",
			expected: map[string]interface{}{},
		},
		{
			name:     "simple json",
			inputStr: `{"prompt":"test","model":"test-model"}`,
			expected: map[string]interface{}{
				"prompt": "test",
				"model":  "test-model",
			},
		},
		{
			name:     "complex json",
			inputStr: `{"prompt":"一只可爱的小猫","options":{"style":"anime","n":1}}`,
			expected: map[string]interface{}{
				"prompt": "一只可爱的小猫",
				"options": map[string]interface{}{
					"style": "anime",
					"n":     float64(1), // JSON数字会被解析为float64
				},
			},
		},
		{
			name:      "invalid json",
			inputStr:  `{"invalid": json}`,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ts.deserializeInput(tt.inputStr)

			if tt.expectErr {
				if err == nil {
					t.Errorf("期望出错但没有出错")
				}
				return
			}

			if err != nil {
				t.Errorf("意外的错误: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("deserializeInput() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTaskService_SerializeDeserializeRoundTrip(t *testing.T) {
	ts := &TaskService{}

	testCases := []map[string]interface{}{
		{
			"prompt": "生成一张图片",
			"model":  "doubao-seedream-3.0-t2i",
			"size":   "1024x1024",
		},
		{
			"prompt":  "未来科技城市",
			"quality": "hd",
			"style":   "cyberpunk",
			"options": map[string]interface{}{
				"n":      2,
				"format": "png",
				"seed":   12345,
			},
		},
		{
			"text":     "Hello, world!",
			"language": "zh-CN",
			"metadata": map[string]interface{}{
				"user_id":   "user123",
				"timestamp": "2023-12-01T10:00:00Z",
				"tags":      []interface{}{"test", "translation"},
			},
		},
	}

	for i, original := range testCases {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			// 序列化
			serialized := ts.serializeInput(original)

			// 反序列化
			deserialized, err := ts.deserializeInput(serialized)
			if err != nil {
				t.Errorf("反序列化失败: %v", err)
				return
			}

			// 验证往返一致性
			if !compareValues(original, deserialized) {
				t.Errorf("往返测试失败:\n原始: %+v\n结果: %+v\n序列化: %s",
					original, deserialized, serialized)
			}
		})
	}
}

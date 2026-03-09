package ccr

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/bytedance/sonic"
)

// GenerateToolCallID 为工具调用生成稳定的 ID
// 用于 Gemini 等不提供原生 ID 的 API
//
// 设计原则：
// - 确定性：相同的 name + args 总是生成相同的 ID
// - 唯一性：不同的 name 或 args 生成不同的 ID
// - 兼容性：生成的 ID 格式与 OpenAI/Anthropic 兼容
func GenerateToolCallID(name string, args map[string]interface{}) string {
	// 构建稳定的输入字符串
	input := fmt.Sprintf("tool:%s", name)

	// 如果有参数，将其序列化并加入哈希
	if len(args) > 0 {
		// 使用 sonic.Marshal 确保稳定的 JSON 序列化
		argsBytes, err := sonic.Marshal(args)
		if err == nil {
			input = fmt.Sprintf("%s:args:%s", input, string(argsBytes))
		}
	}

	// 使用 SHA256 生成哈希
	hash := sha256.Sum256([]byte(input))

	// 取前 12 字节（24 个十六进制字符），与 OpenAI 的 ID 长度相近
	// 格式：call_<hash>
	return fmt.Sprintf("call_%s", hex.EncodeToString(hash[:12]))
}

// ExtractToolCallID 从 CanonicalToolCall 中提取或生成 ID
// 如果已有 ID 则直接返回，否则根据 name + args 生成
func ExtractToolCallID(tc *CanonicalToolCall) string {
	if tc == nil {
		return ""
	}

	// 如果已有 ID，直接返回
	if tc.ID != "" {
		return tc.ID
	}

	// 否则生成稳定的 ID
	return GenerateToolCallID(tc.Name, tc.Args)
}

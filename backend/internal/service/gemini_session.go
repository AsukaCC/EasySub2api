package service

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/AsukaCC/EasySub2api/internal/pkg/antigravity"
)

// BuildGeminiDigestChain 根据 Gemini 请求生成摘要链
// 格式: s:<hash>-u:<hash>-m:<hash>-u:<hash>-...
// s = systemInstruction, u = user, m = model
func BuildGeminiDigestChain(req *antigravity.GeminiRequest) string {
	if req == nil {
		return ""
	}

	var parts []string

	// 1. system instruction
	if req.SystemInstruction != nil && len(req.SystemInstruction.Parts) > 0 {
		partsData, _ := json.Marshal(req.SystemInstruction.Parts)
		parts = append(parts, "s:"+shortHash(partsData))
	}

	// 2. contents
	for _, c := range req.Contents {
		prefix := "u" // user
		if c.Role == "model" {
			prefix = "m"
		}
		partsData, _ := json.Marshal(c.Parts)
		parts = append(parts, prefix+":"+shortHash(partsData))
	}

	return strings.Join(parts, "-")
}

// GenerateGeminiPrefixHash 生成前缀 hash（用于分区隔离）
// 组合: userID + apiKeyID + ip + userAgent + platform + model
// 返回 16 字符的 Base64 编码的 SHA256 前缀
func GenerateGeminiPrefixHash(userID, apiKeyID string, ip, userAgent, platform, model string) string {
	// 组合所有标识符
	normalizedUserAgent := NormalizeSessionUserAgent(userAgent)
	combined := userID + ":" +
		apiKeyID + ":" +
		ip + ":" +
		normalizedUserAgent + ":" +
		platform + ":" +
		model

	hash := sha256.Sum256([]byte(combined))
	// 取前 12 字节，Base64 编码后正好 16 字符
	return base64.RawURLEncoding.EncodeToString(hash[:12])
}

// ParseGeminiSessionValue 解析 Gemini 会话缓存值
// 格式: {uuid}:{accountID}
func ParseGeminiSessionValue(value string) (sessionUUID string, accountID string, ok bool) {
	if value == "" {
		return "", "", false
	}

	// 找到最后一个 ":" 的位置（因为 uuid 可能包含 ":"）
	i := strings.LastIndex(value, ":")
	if i <= 0 || i >= len(value)-1 {
		return "", "", false
	}

	sessionUUID = value[:i]
	accountID = strings.TrimSpace(value[i+1:])
	if accountID == "" {
		return "", "", false
	}

	return sessionUUID, accountID, true
}

// FormatGeminiSessionValue 格式化 Gemini 会话缓存值
// 格式: {uuid}:{accountID}
func FormatGeminiSessionValue(sessionUUID string, accountID string) string {
	return sessionUUID + ":" + accountID
}

// geminiDigestSessionKeyPrefix Gemini 摘要 fallback 会话 key 前缀
const geminiDigestSessionKeyPrefix = "gemini:digest:"

// GenerateGeminiDigestSessionKey 生成 Gemini 摘要 fallback 的 sessionKey
// 组合 prefixHash 前 8 位 + uuid 前 8 位，确保不同会话产生不同的 sessionKey
// 用于在 SelectAccountWithLoadAwareness 中保持粘性会话
func GenerateGeminiDigestSessionKey(prefixHash, uuid string) string {
	prefix := prefixHash
	if len(prefixHash) >= 8 {
		prefix = prefixHash[:8]
	}
	uuidPart := uuid
	if len(uuid) >= 8 {
		uuidPart = uuid[:8]
	}
	return geminiDigestSessionKeyPrefix + prefix + ":" + uuidPart
}

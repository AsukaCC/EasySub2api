package service

import "regexp"

const (
	errorCodeForbidden       = "forbidden"
	errorCodeUnauthenticated = "unauthenticated"
)

var (
	sensitiveQueryParamRegex = regexp.MustCompile(`(?i)([?&](?:key|client_secret|access_token|refresh_token)=)[^&"\s]+`)
	// urlUserInfoRegex 匹配 scheme://user:pass@host 形态中的 userinfo：代理 / 上游 URL
	// 被拼进 net/http、x/net/proxy 的错误文案时会连同凭据一起暴露。
	urlUserInfoRegex = regexp.MustCompile(`(?i)([a-z][a-z0-9+.-]*://)[^/@\s"']+@`)
	// bearerTokenRegex 匹配错误文案中回显的 Bearer 令牌。
	bearerTokenRegex = regexp.MustCompile(`(?i)\bbearer\s+[A-Za-z0-9._~+/=-]{8,}`)
)

// sanitizeUpstreamErrorMessage 清洗可能进入客户端响应 / ops 记录的上游或传输层错误文案：
// 遮盖敏感查询参数、URL 内嵌凭据（代理 user:pass@）与回显的 Bearer 令牌。
// 只做值级脱敏、不改变错误语义，调用方仍可据此判断失败类别。
func sanitizeUpstreamErrorMessage(message string) string {
	if message == "" {
		return ""
	}
	message = sensitiveQueryParamRegex.ReplaceAllString(message, `$1***`)
	message = urlUserInfoRegex.ReplaceAllString(message, `$1***@`)
	message = bearerTokenRegex.ReplaceAllString(message, `Bearer ***`)
	return message
}

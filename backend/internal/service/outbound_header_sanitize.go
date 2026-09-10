package service

import (
	"net/http"
	"strings"
)

// internalOpenAIOutboundHeaders 是网关 / 平台自身的实现细节头（小写）。
// 它们描述的是 EasySub2api 的部署形态（代理链、路由来源、运行环境），与上游协议无关，
// 任何情况下都不得进入 OpenAI / ChatGPT 上游请求。
var internalOpenAIOutboundHeaders = map[string]struct{}{
	"x-easysub2api":    {},
	"x-sub2api":        {},
	"x-gateway":        {},
	"x-relay":          {},
	"x-provider":       {},
	"x-powered-by":     {},
	"x-server":         {},
	"x-request-source": {},
	"via":              {},
	"forwarded":        {},
}

// platformBrandMarkers 平台品牌字样（小写）。出现在头名或头值中即视为平台身份泄漏。
var platformBrandMarkers = [...]string{"easysub2api", "sub2api"}

// outboundValueCheckExemptHeaders 只按名字判定、不检查值的头：凭据与上游铸造的不透明 blob
// 是随机串，偶然命中品牌字样不代表泄漏，误删会直接导致上游 401 或回合状态丢失。
var outboundValueCheckExemptHeaders = map[string]struct{}{
	"authorization":       {},
	"proxy-authorization": {},
	"x-api-key":           {},
	"chatgpt-account-id":  {},
	"x-codex-turn-state":  {},
}

// containsPlatformBrandMarker 报告文本（任意大小写）是否含平台品牌字样。
func containsPlatformBrandMarker(text string) bool {
	lower := strings.ToLower(text)
	for _, marker := range platformBrandMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

// isInternalOutboundHeader 报告一个出站头是否属于平台内部标识：
// 名字命中 denylist，或名字 / 值中出现平台品牌字样。
// 供 sanitizeOpenAIOutboundHeaders（出站终态清理）与账号级 Header Override 校验共用，
// 保证「保存时拒绝」与「出站时剥离」使用同一套判定。
func isInternalOutboundHeader(name string, values ...string) bool {
	lower := strings.ToLower(strings.TrimSpace(name))
	if _, blocked := internalOpenAIOutboundHeaders[lower]; blocked {
		return true
	}
	if containsPlatformBrandMarker(lower) {
		return true
	}
	if _, exempt := outboundValueCheckExemptHeaders[lower]; exempt {
		return false
	}
	for _, value := range values {
		if containsPlatformBrandMarker(value) {
			return true
		}
	}
	return false
}

// sanitizeOpenAIOutboundHeaders 是 OpenAI / ChatGPT 出站请求的终态清理阶段：
// 在账号级覆写、身份收口之后调用，剥离平台品牌与基础设施头。
// Codex 身份（User-Agent / originator / version）由 enforceCodexIdentityHeaders* 另行收口，
// 本函数不触碰协议头。
func sanitizeOpenAIOutboundHeaders(h http.Header) {
	if h == nil {
		return
	}
	for name, values := range h {
		if isInternalOutboundHeader(name, values...) {
			delete(h, name)
		}
	}
}

// sanitizeOutboundHeaderMap 是 sanitizeOpenAIOutboundHeaders 的 map[string]string 形态，
// 供使用 req 等以 map 传头的出站客户端（如额度探针）复用同一套判定。
func sanitizeOutboundHeaderMap(h map[string]string) {
	for name, value := range h {
		if isInternalOutboundHeader(name, value) {
			delete(h, name)
		}
	}
}

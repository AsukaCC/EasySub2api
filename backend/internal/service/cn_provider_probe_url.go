package service

// CN 供应商探测端点的出站 URL 安全策略校验（配额/余额探测共用）。
//
// 背景（review B4）：这两条探测路径会把账号 API key 发往 base_url 衍生端点，
// 此前完全绕过 security.url_allowlist——在加固部署里构成任意外发与内网探测
// 面（本项目此前发生过账号测试 SSRF 生产事件）。与网关转发
// （validateUpstreamBaseURL）、Grok 探测（grokOperatorPolicyValidator）一致，
// 探测发起前必须过同一套运营者策略。

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AsukaCC/EasySub2api/internal/config"
	"github.com/AsukaCC/EasySub2api/internal/util/urlvalidator"
)

// cnValidateProbeURL 按全局出站 URL 安全策略校验探测端点，返回规范化 URL。
// 白名单开启时强制 UpstreamHosts；关闭时仅关闭主机名匹配，但协议和私网
// 地址保护仍跟随全局策略；cfg 为 nil 时退化为安全的 HTTPS 校验。
func cnValidateProbeURL(cfg *config.Config, raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", errors.New("probe url is required")
	}
	var enabled bool
	var allowedHosts []string
	var allowPrivate, allowInsecureHTTP bool
	if cfg != nil {
		enabled = cfg.Security.URLAllowlist.Enabled
		allowedHosts = cfg.Security.URLAllowlist.UpstreamHosts
		allowPrivate = cfg.Security.URLAllowlist.AllowPrivateHosts
		allowInsecureHTTP = cfg.Security.URLAllowlist.AllowInsecureHTTP
	}
	normalized, err := urlvalidator.ValidateConfiguredURL(trimmed, enabled, allowedHosts, allowPrivate, allowInsecureHTTP)
	if err != nil {
		return "", fmt.Errorf("probe target rejected by URL security policy: %w", err)
	}
	return normalized, nil
}

package service

import (
	"context"
	"strings"
	"time"
)

// Codex 出站版本来源标识（诊断用）。
const (
	CodexVersionSourceManualOverride = "manual_override"
	CodexVersionSourceAutoSync       = "auto_sync"
	CodexVersionSourceBuiltinDefault = "builtin_default"
)

// CodexOutboundDiagnostics 描述当前进程对 ChatGPT Codex 上游声明的出站身份与连接 Profile。
// 只包含运维排障所需的非敏感摘要：不含 token、Cookie、Authorization、账号凭据或完整请求头。
type CodexOutboundDiagnostics struct {
	// 版本链
	EffectiveVersion        string `json:"effective_version"`
	VersionSource           string `json:"version_source"`
	ManualOverrideVersion   string `json:"manual_override_version"`
	ManualOverrideEnabled   bool   `json:"manual_override_enabled"`
	SyncedVersion           string `json:"synced_version"`
	AutoSyncEnabled         bool   `json:"auto_sync_enabled"`
	BuiltinDefaultVersion   string `json:"builtin_default_version"`
	MinimumSupportedVersion string `json:"minimum_supported_version"`
	PrereleaseAllowed       bool   `json:"prerelease_allowed"`
	// 已配置但被门禁拒绝的版本（预发布 / 非法 / 低于门槛），供面板提示原因。
	RejectedManualOverrideVersion string `json:"rejected_manual_override_version,omitempty"`
	RejectedSyncedVersion         string `json:"rejected_synced_version,omitempty"`

	// 身份三元组（OAuth 推理面）
	UserAgent                  string `json:"user_agent"`
	Originator                 string `json:"originator"`
	IdentityEnforcementEnabled bool   `json:"identity_enforcement_enabled"`
	// ResponsesBetaHeader 为空表示 HTTP 推理面不发送 OpenAI-Beta（当前策略）。
	ResponsesBetaHeader string `json:"responses_beta_header"`
	LiveAlphaHeader     string `json:"live_alpha_header"`

	// 连接层 Profile（进程级配置快照；账号级代理 / TLS 模板在账号详情查看）
	ProxyDirectFallbackAllowed bool   `json:"proxy_direct_fallback_allowed"`
	OpenAIHTTP2Enabled         bool   `json:"openai_http2_enabled"`
	OpenAIHTTP2ProxyFallback   bool   `json:"openai_http2_proxy_fallback_to_http1"`
	TLSFingerprintEnabled      bool   `json:"tls_fingerprint_enabled"`
	TLSFingerprintProfileCount int    `json:"tls_fingerprint_profile_count"`
	ProtocolMode               string `json:"protocol_mode"`

	GeneratedAt time.Time `json:"generated_at"`
}

// GetCodexOutboundDiagnostics 组装 Codex 出站诊断快照。
// 版本链取值与 GetOpenAICodexClientVersion 同一套准入规则（AcceptCodexClientVersion），
// 保证诊断展示的 effective_version 就是热路径实际出站的版本。
func (s *SettingService) GetCodexOutboundDiagnostics(ctx context.Context) CodexOutboundDiagnostics {
	diag := CodexOutboundDiagnostics{
		BuiltinDefaultVersion:      codexCLIVersion,
		MinimumSupportedVersion:    codexUpstreamMinVersion,
		PrereleaseAllowed:          CodexPrereleaseVersionAllowed(),
		IdentityEnforcementEnabled: codexIdentityEnforcement.Load(),
		ResponsesBetaHeader:        "",
		LiveAlphaHeader:            "quicksilver=v2",
		AutoSyncEnabled:            true,
		GeneratedAt:                time.Now(),
	}

	if s != nil && s.settingRepo != nil {
		if ctx == nil {
			ctx = context.Background()
		}
		dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), openAICodexClientVersionDBTimeout)
		defer cancel()
		values, err := s.settingRepo.GetMultiple(dbCtx, []string{
			SettingKeyOpenAICodexClientVersion,
			SettingKeyOpenAICodexClientVersionSynced,
			SettingKeyOpenAICodexVersionAutoSyncEnabled,
		})
		if err == nil {
			rawOverride := strings.TrimSpace(values[SettingKeyOpenAICodexClientVersion])
			rawSynced := strings.TrimSpace(values[SettingKeyOpenAICodexClientVersionSynced])
			diag.ManualOverrideVersion = rawOverride
			diag.SyncedVersion = rawSynced
			if autoSync := strings.TrimSpace(values[SettingKeyOpenAICodexVersionAutoSyncEnabled]); autoSync != "" {
				diag.AutoSyncEnabled = autoSync == "true"
			}
			if rawOverride != "" && AcceptCodexClientVersion(rawOverride) == "" {
				diag.RejectedManualOverrideVersion = rawOverride
			}
			if rawSynced != "" && AcceptCodexClientVersion(rawSynced) == "" {
				diag.RejectedSyncedVersion = rawSynced
			}
		}
	}

	// 版本来源按与热路径相同的优先级判定。
	switch {
	case diag.ManualOverrideVersion != "" && AcceptCodexClientVersion(diag.ManualOverrideVersion) != "":
		diag.VersionSource = CodexVersionSourceManualOverride
		diag.ManualOverrideEnabled = true
	case diag.SyncedVersion != "" && AcceptCodexClientVersion(diag.SyncedVersion) != "":
		diag.VersionSource = CodexVersionSourceAutoSync
	default:
		diag.VersionSource = CodexVersionSourceBuiltinDefault
	}

	// 身份三元组走与推理相同的解析链（面板 UA 指纹 + 生效版本 + 编译期兜底）。
	identity := resolveCodexOutboundIdentity("")
	diag.EffectiveVersion = identity.version
	diag.UserAgent = identity.userAgent
	diag.Originator = identity.originator

	if s != nil && s.cfg != nil {
		diag.ProxyDirectFallbackAllowed = s.cfg.Security.ProxyFallback.AllowDirectOnError
		diag.OpenAIHTTP2Enabled = s.cfg.Gateway.OpenAIHTTP2.Enabled
		diag.OpenAIHTTP2ProxyFallback = s.cfg.Gateway.OpenAIHTTP2.AllowProxyFallbackToHTTP1
		diag.TLSFingerprintEnabled = s.cfg.Gateway.TLSFingerprint.Enabled
		diag.TLSFingerprintProfileCount = len(s.cfg.Gateway.TLSFingerprint.Profiles)
	}
	switch {
	case diag.TLSFingerprintEnabled:
		// utls 拨号路径固定 HTTP/1.1（ForceAttemptHTTP2=false），与 H2 开关互斥。
		diag.ProtocolMode = "tls_fingerprint_http1"
	case diag.OpenAIHTTP2Enabled:
		diag.ProtocolMode = "http2"
	default:
		diag.ProtocolMode = "http1"
	}
	return diag
}

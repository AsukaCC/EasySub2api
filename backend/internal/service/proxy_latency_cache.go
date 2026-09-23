package service

import (
	"context"
	"time"
)

type ProxyLatencyInfo struct {
	Success          bool      `json:"success"`
	LatencyMs        *int64    `json:"latency_ms,omitempty"`
	Message          string    `json:"message,omitempty"`
	IPAddress        string    `json:"ip_address,omitempty"`
	Timezone         string    `json:"timezone,omitempty"`
	Country          string    `json:"country,omitempty"`
	CountryCode      string    `json:"country_code,omitempty"`
	Region           string    `json:"region,omitempty"`
	City             string    `json:"city,omitempty"`
	QualityStatus    string    `json:"quality_status,omitempty"`
	QualityScore     *int      `json:"quality_score,omitempty"`
	QualityGrade     string    `json:"quality_grade,omitempty"`
	QualitySummary   string    `json:"quality_summary,omitempty"`
	QualityCheckedAt *int64    `json:"quality_checked_at,omitempty"`
	QualityCFRay     string    `json:"quality_cf_ray,omitempty"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ProxyLatencyCache interface {
	GetProxyLatencies(ctx context.Context, proxyIDs []string) (map[string]*ProxyLatencyInfo, error)
	SetProxyLatency(ctx context.Context, proxyID string, info *ProxyLatencyInfo) error
}

// ProxyTimezoneResolver resolves the IANA timezone observed through a proxy's
// exit IP. It is deliberately separate from ProxyLatencyCache so existing
// cache implementations and test doubles do not need to grow another method.
type ProxyTimezoneResolver interface {
	GetProxyTimezone(ctx context.Context, proxyID string) (string, error)
}

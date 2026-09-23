package service

import (
	"context"
	"net/http"
	"strings"
	"time"
)

// proxyTimezoneHeader is the SDK header used to describe the client timezone
// to Anthropic-compatible upstreams. It must describe the actual egress IP,
// never the gateway host or the downstream client.
const proxyTimezoneHeader = "X-Stainless-Timezone"

// applyProxyTimezoneHeader makes the timezone header deterministic for every
// Anthropic request. A missing/failed proxy probe intentionally removes the
// header instead of falling back to time.Local or the application timezone.
func applyProxyTimezoneHeader(ctx context.Context, account *Account, headers http.Header, resolver ProxyTimezoneResolver) {
	if headers == nil {
		return
	}
	deleteHeaderAllForms(headers, proxyTimezoneHeader)
	if account == nil || account.ProxyID == nil || strings.TrimSpace(*account.ProxyID) == "" || resolver == nil {
		return
	}

	name, err := resolver.GetProxyTimezone(ctx, strings.TrimSpace(*account.ProxyID))
	name = strings.TrimSpace(name)
	if err != nil || name == "" {
		return
	}
	if _, err := time.LoadLocation(name); err != nil {
		return
	}
	setHeaderRaw(headers, proxyTimezoneHeader, name)
}

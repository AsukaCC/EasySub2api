package service

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type proxyTimezoneResolverStub struct {
	timezone string
	err      error
}

func (s proxyTimezoneResolverStub) GetProxyTimezone(context.Context, string) (string, error) {
	return s.timezone, s.err
}

func TestApplyProxyTimezoneHeaderUsesProxyTimezone(t *testing.T) {
	proxyID := "proxy-1"
	account := &Account{ProxyID: &proxyID}
	headers := make(http.Header)
	headers.Set("X-Stainless-Timezone", "Asia/Shanghai")

	applyProxyTimezoneHeader(context.Background(), account, headers, proxyTimezoneResolverStub{timezone: "America/New_York"})

	require.Equal(t, "America/New_York", headers.Get("X-Stainless-Timezone"))
}

func TestApplyProxyTimezoneHeaderDoesNotFallbackToServerTimezone(t *testing.T) {
	proxyID := "proxy-1"
	account := &Account{ProxyID: &proxyID}
	headers := make(http.Header)
	headers.Set("X-Stainless-Timezone", "America/Los_Angeles")

	applyProxyTimezoneHeader(context.Background(), account, headers, proxyTimezoneResolverStub{})

	require.Empty(t, headers.Get("X-Stainless-Timezone"))
}

func TestApplyProxyTimezoneHeaderRejectsInvalidOrFailedResolution(t *testing.T) {
	proxyID := "proxy-1"
	account := &Account{ProxyID: &proxyID}

	for name, resolver := range map[string]proxyTimezoneResolverStub{
		"invalid timezone": {timezone: "Mars/Olympus"},
		"resolver failure": {timezone: "America/New_York", err: errors.New("cache unavailable")},
	} {
		t.Run(name, func(t *testing.T) {
			headers := make(http.Header)
			headers.Set("X-Stainless-Timezone", "UTC")

			applyProxyTimezoneHeader(context.Background(), account, headers, resolver)

			require.Empty(t, headers.Get("X-Stainless-Timezone"))
		})
	}
}

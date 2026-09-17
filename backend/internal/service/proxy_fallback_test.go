//go:build unit

package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func mkProxy(id string, mode string, backup *string, expiresInDays *int, now time.Time) Proxy {
	p := Proxy{ID: id, FallbackMode: mode, BackupProxyID: backup}
	if expiresInDays != nil {
		t := now.AddDate(0, 0, *expiresInDays)
		p.ExpiresAt = &t
	}
	return p
}
func strptr(v string) *string { return &v }
func di(v int) *int           { return &v }

func TestResolveFallbackTarget(t *testing.T) {
	now := time.Now()
	t.Run("none keeps original", func(t *testing.T) {
		a := mkProxy("1", FallbackModeNone, nil, di(-1), now)
		by := map[string]Proxy{"1": a}
		target, change := ResolveProxyFallbackTarget(a, by, now)
		require.False(t, change)
		require.Nil(t, target)
	})
	t.Run("direct -> nil target, change", func(t *testing.T) {
		a := mkProxy("1", FallbackModeDirect, nil, di(-1), now)
		by := map[string]Proxy{"1": a}
		target, change := ResolveProxyFallbackTarget(a, by, now)
		require.True(t, change)
		require.Nil(t, target)
	})
	t.Run("proxy -> healthy backup", func(t *testing.T) {
		b := mkProxy("2", FallbackModeNone, nil, di(30), now)
		a := mkProxy("1", FallbackModeProxy, strptr("2"), di(-1), now)
		by := map[string]Proxy{"1": a, "2": b}
		target, change := ResolveProxyFallbackTarget(a, by, now)
		require.True(t, change)
		require.NotNil(t, target)
		require.Equal(t, "2", *target)
	})
	t.Run("chain A->B(expired)->C(healthy)", func(t *testing.T) {
		c := mkProxy("3", FallbackModeNone, nil, di(30), now)
		b := mkProxy("2", FallbackModeProxy, strptr("3"), di(-1), now)
		a := mkProxy("1", FallbackModeProxy, strptr("2"), di(-1), now)
		by := map[string]Proxy{"1": a, "2": b, "3": c}
		target, change := ResolveProxyFallbackTarget(a, by, now)
		require.True(t, change)
		require.Equal(t, "3", *target)
	})
	t.Run("cycle A->B->A keeps original", func(t *testing.T) {
		b := mkProxy("2", FallbackModeProxy, strptr("1"), di(-1), now)
		a := mkProxy("1", FallbackModeProxy, strptr("2"), di(-1), now)
		by := map[string]Proxy{"1": a, "2": b}
		target, change := ResolveProxyFallbackTarget(a, by, now)
		require.False(t, change)
		require.Nil(t, target)
	})
	t.Run("chain tail direct fallback", func(t *testing.T) {
		b := mkProxy("2", FallbackModeDirect, nil, di(-1), now)
		a := mkProxy("1", FallbackModeProxy, strptr("2"), di(-1), now)
		by := map[string]Proxy{"1": a, "2": b}
		target, change := ResolveProxyFallbackTarget(a, by, now)
		require.True(t, change)
		require.Nil(t, target)
	})
}

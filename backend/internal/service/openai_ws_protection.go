package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/AsukaCC/EasySub2api/internal/pkg/tlsfingerprint"
	"net/http"
	"net/url"
	"time"
)

type openAIWSTLSContextKey struct{}
type openAIWSTLSConfig struct {
	accountID string
	profile   *tlsfingerprint.Profile
}

func withOpenAIWSTLSProfile(ctx context.Context, id string, profile *tlsfingerprint.Profile) context.Context {
	return context.WithValue(ctx, openAIWSTLSContextKey{}, openAIWSTLSConfig{id, profile.Clone()})
}

func (d *coderOpenAIWSClientDialer) fingerprintHTTPClient(cfg openAIWSTLSConfig, target, proxy string) (*http.Client, error) {
	key := fmt.Sprintf("tls:%x", sha256.Sum256([]byte(cfg.accountID+"\x00"+target+"\x00"+proxy+"\x00"+cfg.profile.CacheKey())))
	d.proxyMu.Lock()
	defer d.proxyMu.Unlock()
	now := time.Now().UnixNano()
	if entry := d.proxyClients[key]; entry != nil {
		entry.lastUsedUnixNano = now
		return entry.client, nil
	}
	var parsed *url.URL
	if proxy != "" {
		var err error
		parsed, err = url.Parse(proxy)
		if err != nil {
			return nil, err
		}
	}
	transport, err := tlsfingerprint.NewHTTPTransport(cfg.profile, parsed, tlsfingerprint.TransportOptions{RootCAs: d.tlsRootCAs})
	if err != nil {
		return nil, err
	}
	d.cleanupProxyClientsLocked(now)
	client := &http.Client{Transport: transport}
	if d.proxyClients == nil {
		d.proxyClients = make(map[string]*openAIWSProxyClientEntry)
	}
	d.proxyClients[key] = &openAIWSProxyClientEntry{client: client, lastUsedUnixNano: now}
	return client, nil
}

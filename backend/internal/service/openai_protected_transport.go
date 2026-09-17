package service

import (
	"github.com/AsukaCC/EasySub2api/internal/pkg/tlsfingerprint"
	"net/http"
)

func (s *OpenAIGatewayService) doProtectedOpenAIRequest(req *http.Request, proxyURL string, account *Account) (*http.Response, error) {
	profile, err := resolveMode1TLSProfile(account)
	if err != nil {
		return nil, err
	}
	if s.cfg != nil && !s.cfg.Gateway.TLSFingerprint.Enabled {
		profile = nil
	}
	return s.httpUpstream.DoWithTLS(req, proxyURL, account.ID, account.Mode1EffectiveConcurrency(), profile)
}

func protectedTLSProfile(account *Account, enabled bool) (*tlsfingerprint.Profile, error) {
	profile, err := resolveMode1TLSProfile(account)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, nil
	}
	return profile, nil
}

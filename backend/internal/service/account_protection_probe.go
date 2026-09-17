package service

import (
	"bytes"
	"github.com/AsukaCC/EasySub2api/internal/pkg/tlsfingerprint"
	"io"
	"net/http"
)

func (s *AccountTestService) doAccountTestWithProtection(req *http.Request, proxy string, account *Account, profile *tlsfingerprint.Profile) (*http.Response, error) {
	if req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		raw, err := io.ReadAll(body)
		_ = body.Close()
		if err != nil {
			return nil, err
		}
		if err := checkActivePayloadModels(account, raw); err != nil {
			return nil, err
		}
	}
	if err := ValidateAccountProtectionConfiguration(account); err != nil {
		return nil, err
	}
	if account.IsOpenAIOAuthLike() {
		gateway := &OpenAIGatewayService{accountRepo: s.accountRepo, cfg: s.cfg}
		source, err := gateway.prepareCodexAccountIdentitySource(req.Context(), nil, account)
		if err != nil {
			return nil, err
		}
		ids := resolveCodexFingerprintIDsFromRequest(source, req.Header)
		if req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, err
			}
			raw, readErr := io.ReadAll(body)
			_ = body.Close()
			if readErr != nil {
				return nil, readErr
			}
			next, _, err := applyCodexFingerprintClientMetadataRaw(raw, ids)
			if err != nil {
				return nil, err
			}
			next, _, err = applyCodexAccountIdentityClientMetadataRaw(next, source, "")
			if err != nil {
				return nil, err
			}
			if err := checkAccountRequestIntegrity(nil, account, raw, next); err != nil {
				return nil, err
			}
			req.Body = io.NopCloser(bytes.NewReader(next))
			req.ContentLength = int64(len(next))
			req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(next)), nil }
		}
		applyCodexFingerprintHeaders(req.Header, ids)
		applyCodexAccountIdentityHeaders(req.Header, source, "")
	}
	return s.httpUpstream.DoWithTLS(req, proxy, account.ID, account.Mode1EffectiveConcurrency(), profile)
}

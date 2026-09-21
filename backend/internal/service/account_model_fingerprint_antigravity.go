package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/AsukaCC/EasySub2api/internal/pkg/antigravity"
)

// Fingerprinting makes one request, without the connectivity probe's retries or
// optional credits-overage fallback.
func (s *AntigravityGatewayService) testModelFingerprint(ctx context.Context, account *Account, model, projectID, token string) (*TestConnectionResult, error) {
	payload := map[string]any{
		"systemInstruction": map[string]any{"parts": []map[string]any{{"text": antigravity.GetDefaultIdentityPatch()}}},
	}
	applyModelFingerprintPayload(ctx, payload, "gemini", false)
	payload["sessionId"] = fingerprintProbe(ctx).conversation.id
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	body, err = s.wrapV1InternalRequest(projectID, model, body)
	if err != nil {
		return nil, err
	}
	req, err := antigravity.NewAPIRequestWithURL(ctx, resolveAntigravityForwardBaseURL(account), "streamGenerateContent", token, body)
	if err != nil {
		return nil, err
	}
	proxy := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxy = account.Proxy.URL()
	}
	resp, err := s.httpUpstream.Do(req, proxy, account.ID, account.Mode1EffectiveConcurrency())
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("fingerprint upstream status %d", resp.StatusCode)
	}
	const limit = 256 << 10
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if len(data) > limit {
		return nil, fmt.Errorf("fingerprint response too large")
	}
	return &TestConnectionResult{Text: extractTextFromSSEResponse(data), MappedModel: model}, nil
}

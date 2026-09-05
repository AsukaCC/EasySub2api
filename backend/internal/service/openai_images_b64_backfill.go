package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/logger"
	"github.com/AsukaCC/EasySub2api/internal/util/urlvalidator"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const AccountExtraImagesURLToB64JSON = "images_url_to_b64_json"

const openAIImageURLDownloadTimeout = 60 * time.Second

func ImagesURLToB64JSONEnabled(account *Account) bool {
	return account != nil && account.getExtraBool(AccountExtraImagesURLToB64JSON)
}

func (s *OpenAIGatewayService) backfillOpenAIImagesB64JSON(
	ctx context.Context,
	account *Account,
	parsed *OpenAIImagesRequest,
	body []byte,
) []byte {
	if !ImagesURLToB64JSONEnabled(account) || (parsed != nil && parsed.ResponseFormat == "url") || len(body) == 0 || !gjson.ValidBytes(body) {
		return body
	}
	items := gjson.GetBytes(body, "data")
	if !items.IsArray() {
		return body
	}
	for index, item := range items.Array() {
		if !item.IsObject() || strings.TrimSpace(item.Get("b64_json").String()) != "" {
			continue
		}
		rawURL := strings.TrimSpace(item.Get("url").String())
		if rawURL == "" {
			continue
		}
		encoded, err := s.fetchOpenAIImageURLBase64(ctx, account, rawURL)
		if err != nil {
			logger.LegacyPrintf("service.openai_gateway", "[OpenAI] Images b64_json backfill skipped account_id=%s index=%d err=%s", account.ID, index, sanitizeUpstreamErrorMessage(err.Error()))
			continue
		}
		updated, err := sjson.SetBytes(body, fmt.Sprintf("data.%d.b64_json", index), encoded)
		if err == nil {
			body = updated
		}
	}
	return body
}

func (s *OpenAIGatewayService) fetchOpenAIImageURLBase64(ctx context.Context, account *Account, rawURL string) (string, error) {
	if strings.HasPrefix(strings.ToLower(rawURL), "data:") {
		if encoded := normalizeOpenAIImageBase64(rawURL); encoded != "" {
			return encoded, nil
		}
		return "", errors.New("data url payload is not valid base64")
	}
	if s == nil || s.httpUpstream == nil {
		return "", errors.New("http upstream is not configured")
	}
	downloadURL, err := s.validateOutboundURL(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid image url: %w", err)
	}
	if err := rejectPrivateImageHost(downloadURL); err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(WithHTTPUpstreamPublicHostsOnly(ctx), openAIImageURLDownloadTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("build image download request: %w", err)
	}
	req.Header.Set("Accept", "image/*,*/*;q=0.8")
	proxyURL := ""
	if account != nil && account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	accountID := ""
	concurrency := 0
	if account != nil {
		accountID, concurrency = account.ID, account.Concurrency
	}
	resp, err := s.httpUpstream.Do(req, proxyURL, accountID, concurrency)
	if err != nil {
		return "", fmt.Errorf("download image: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("download image: unexpected status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, openAIImageMaxDownloadBytes+1))
	if err != nil {
		return "", fmt.Errorf("read image body: %w", err)
	}
	if int64(len(data)) > openAIImageMaxDownloadBytes {
		return "", fmt.Errorf("downloaded image exceeds %d bytes", openAIImageMaxDownloadBytes)
	}
	if len(data) == 0 || !isBackfillImageContent(data) {
		return "", errors.New("download image: content is not an allowed image format")
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func rejectPrivateImageHost(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid image url: %w", err)
	}
	if urlvalidator.IsBlockedHost(parsed.Hostname()) {
		return fmt.Errorf("image url host is not allowed: %s", parsed.Hostname())
	}
	return nil
}

var openAIImageBackfillContentTypes = map[string]struct{}{
	"image/png": {}, "image/jpeg": {}, "image/webp": {}, "image/gif": {},
}

func isBackfillImageContent(data []byte) bool {
	_, ok := openAIImageBackfillContentTypes[detectedImageContentType(data)]
	return ok
}

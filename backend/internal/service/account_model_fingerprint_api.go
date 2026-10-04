package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/claude"
	apperrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	"github.com/gin-gonic/gin"
)

type ModelFingerprintOptions struct {
	Model           string `json:"model_id"`
	Protocol        string `json:"protocol"`
	ReasoningEffort string `json:"reasoning_effort"`
	Source          string `json:"-"`
}

type ModelFingerprintModel struct {
	ID              string   `json:"id"`
	DisplayName     string   `json:"display_name"`
	ReasoningLevels []string `json:"reasoning_levels"`
}

func fingerprintAPIAccount(a *Account) bool {
	return a != nil && !a.IsSyntheticUITest() && a.Type == AccountTypeAPIKey && a.GetCredential("api_key") != "" &&
		(a.IsOpenAI() || a.IsAnthropic() || a.IsCNProvider() || a.IsOpenCodeGo() || a.IsGrok())
}

func (s *AccountTestService) GetFingerprintModels(ctx context.Context, id string) ([]ModelFingerprintModel, error) {
	a, err := s.accountRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.fingerprintModels(ctx, a)
}

func (s *AccountTestService) fingerprintModels(ctx context.Context, a *Account) ([]ModelFingerprintModel, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	if !fingerprintAPIAccount(a) {
		return nil, apperrors.BadRequest("UNSUPPORTED_FINGERPRINT_ACCOUNT", "Select an API key account with an OpenAI or Anthropic compatible endpoint")
	}
	// Discovery is read-only: testing must not rewrite the account's allowlist or metadata.
	catalogAccount := a
	if a.IsCNProvider() && strings.TrimSpace(a.GetCredential("base_url")) != "" {
		// Custom relay credentials must stay on their configured host during discovery.
		catalogAccount = DirectModelTestAccount(a)
		catalogAccount.Credentials["api_protocol"] = APIProtocolChatCompletions
	}
	ids, body, err := s.fetchUpstreamModelList(ctx, catalogAccount)
	if err != nil {
		return nil, apperrors.New(http.StatusBadGateway, "FINGERPRINT_MODELS_UNAVAILABLE", "Could not fetch the upstream model list")
	}
	_, metadata, _ := extractUpstreamModelCatalog(body, a.IsGrok())
	models := make([]ModelFingerprintModel, 0, len(ids))
	for _, id := range ids {
		if !IsModelFingerprintTextModel(id) {
			continue
		}
		meta := metadata[id]
		levels := meta.SupportedReasoningLevels
		if len(levels) == 0 && (meta.Reasoning == nil || *meta.Reasoning) {
			if saved, ok := a.GetUpstreamModelMetadata(id); ok {
				levels = saved.SupportedReasoningLevels
			}
		}
		if len(levels) == 0 && (meta.Reasoning == nil || *meta.Reasoning) {
			if strings.HasPrefix(id, "claude-") {
				levels = claude.EffortLevelsForModel(id)
			}
			if modelFingerprintReasoningModel(id) {
				for _, item := range configuredCodexGPTReasoningLevels(id) {
					levels = append(levels, item.Effort)
				}
			}
			if grokSupportsReasoningEffort(id) {
				levels = []string{"low", "high"}
			}
		}
		name := meta.DisplayName
		if name == "" {
			name = id
		}
		models = append(models, ModelFingerprintModel{ID: id, DisplayName: name, ReasoningLevels: levels})
	}
	return models, nil
}

func (s *AccountTestService) validateFingerprintOptions(ctx context.Context, a *Account, option *ModelFingerprintOptions) error {
	option.Model = strings.TrimSpace(option.Model)
	if option.Model == "" || len(option.Model) > 100 || strings.ContainsAny(option.Model, "\r\n*") || !IsModelFingerprintTextModel(option.Model) {
		return apperrors.BadRequest("INVALID_MODEL", "Select an upstream text model")
	}
	if err := CheckActiveModel(option.Model); err != nil {
		return err
	}
	if option.Protocol == "" {
		option.Protocol = "auto"
	}
	if !slices.Contains([]string{"auto", "chat", "anthropic"}, option.Protocol) {
		return apperrors.BadRequest("INVALID_PROTOCOL", "Invalid test protocol")
	}
	models, err := s.fingerprintModels(ctx, a)
	if err != nil {
		return err
	}
	for _, model := range models {
		if model.ID != option.Model {
			continue
		}
		if option.ReasoningEffort != "" && !slices.Contains(model.ReasoningLevels, option.ReasoningEffort) {
			return apperrors.BadRequest("INVALID_REASONING_EFFORT", "Reasoning effort is not supported by this model")
		}
		if option.Source == "" {
			option.Source = "manual"
		}
		return nil
	}
	return apperrors.BadRequest("INVALID_MODEL", "Select a model currently advertised by this upstream")
}

func (s *AccountTestService) testFingerprintAPI(c *gin.Context, account *Account) error {
	ctx := c.Request.Context()
	p := fingerprintProbe(ctx)
	if !fingerprintAPIAccount(account) {
		p.fatal = true
		return fmt.Errorf("unsupported fingerprint account")
	}
	base := strings.TrimSpace(account.GetCredential("base_url"))
	if base == "" {
		if account.IsGrok() {
			base = "https://api.x.ai"
		} else if account.IsAnthropic() {
			base = "https://api.anthropic.com"
		} else {
			base = account.GetOpenAIFormatBaseURL()
			if p.protocol == "anthropic" && (account.IsCNProvider() || account.IsOpenCodeGo()) {
				copy := DirectModelTestAccount(account)
				copy.Credentials["api_protocol"] = APIProtocolAnthropic
				base = copy.GetAnthropicProtocolBaseURL()
			}
		}
	}
	base, err := s.validateUpstreamBaseURL(base)
	if err != nil {
		p.fatal = true
		return fmt.Errorf("invalid fingerprint base URL")
	}
	endpoint := buildOpenAIChatCompletionsURL(base)
	if p.protocol == "anthropic" {
		endpoint = buildOpenAIEndpointURL(base, "/v1/messages")
	}
	payload := map[string]any{"model": p.model, "stream": false, "messages": []map[string]string{{"role": "user", "content": p.prompt}}}
	limit := 4096
	if p.effort != "" && p.effort != "low" && p.effort != "minimal" && p.effort != "none" {
		limit = 16384
	}
	if p.protocol == "chat" && modelFingerprintReasoningModel(p.model) {
		payload["max_completion_tokens"] = limit
	} else {
		payload["max_tokens"] = limit
	}
	if p.effort != "" {
		if p.protocol == "chat" {
			payload["reasoning_effort"] = p.effort
		} else {
			payload["output_config"] = map[string]string{"effort": p.effort}
			if len(claude.EffortLevelsForModel(p.model)) > 0 && !strings.Contains(p.model, "opus-4-5") {
				payload["thinking"] = map[string]string{"type": "adaptive"}
			}
		}
		p.usage.ReasoningEffort, p.usage.RequestedReasoningEffort = &p.effort, &p.effort
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if p.protocol == "anthropic" {
		req.Header.Set("anthropic-version", "2023-06-01")
		setAnthropicAPIKeyAuthHeader(req.Header, account, account.GetCredential("api_key"), base)
	} else {
		req.Header.Set("Authorization", "Bearer "+account.GetCredential("api_key"))
	}
	account.ApplyHeaderOverrides(req.Header)
	proxy := ""
	if account.Proxy != nil {
		proxy = account.Proxy.URL()
	}
	p.upstreamModel = p.model
	resp, err := s.doAccountTestWithProtection(req, proxy, account, s.tlsFPProfileService.ResolveTLSProfile(account))
	if err != nil {
		return fmt.Errorf("fingerprint upstream unavailable")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		p.fatal = resp.StatusCode == 401 || resp.StatusCode == 403 || resp.StatusCode == 429
		p.formatRejected = slices.Contains([]int{400, 404, 405, 422}, resp.StatusCode)
		return fmt.Errorf("fingerprint upstream HTTP %d", resp.StatusCode)
	}
	body, err = io.ReadAll(io.LimitReader(resp.Body, (1<<20)+1))
	if err != nil || len(body) > 1<<20 {
		return fmt.Errorf("invalid fingerprint response")
	}
	var data map[string]any
	if json.Unmarshal(body, &data) != nil {
		return fmt.Errorf("invalid fingerprint JSON")
	}
	collectModelFingerprintUsage(ctx, data)
	var output strings.Builder
	if p.protocol == "anthropic" {
		if stop, _ := data["stop_reason"].(string); stop == "max_tokens" || stop == "refusal" {
			return fmt.Errorf("incomplete fingerprint response")
		}
		blocks, _ := data["content"].([]any)
		for _, block := range blocks {
			if item, ok := block.(map[string]any); ok && item["type"] == "text" {
				text, _ := item["text"].(string)
				output.WriteString(text)
			}
		}
	} else {
		choices, _ := data["choices"].([]any)
		if len(choices) == 0 {
			return fmt.Errorf("missing fingerprint choices")
		}
		choice, _ := choices[0].(map[string]any)
		if choice["finish_reason"] == "length" || choice["finish_reason"] == "content_filter" {
			return fmt.Errorf("incomplete fingerprint response")
		}
		message, _ := choice["message"].(map[string]any)
		if text, ok := message["content"].(string); ok {
			output.WriteString(text)
		} else {
			blocks, _ := message["content"].([]any)
			for _, block := range blocks {
				if item, ok := block.(map[string]any); ok && item["type"] == "text" {
					text, _ := item["text"].(string)
					output.WriteString(text)
				}
			}
		}
	}
	p.event(TestEvent{Type: "content", Text: output.String()})
	p.firstTokenMs = nil // A buffered response does not expose time to first token.
	p.complete = true
	return nil
}

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/claude"
	apperrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	"github.com/AsukaCC/EasySub2api/internal/pkg/modeltrace"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const ModelFingerprintExtraKey = "model_fingerprint"
const ModelFingerprintRetention = 2 * time.Hour
const modelFingerprintMaxTokens = 1536

type ModelFingerprintSnapshot struct {
	ID           string             `json:"id"`
	Model        string             `json:"model"`
	Status       string             `json:"status"`
	SamplingMode string             `json:"sampling_mode"`
	Completed    int                `json:"completed"`
	Valid        int                `json:"valid"`
	Total        int                `json:"total"`
	StartedAt    time.Time          `json:"started_at"`
	ExpiresAt    time.Time          `json:"expires_at"`
	FinishedAt   *time.Time         `json:"finished_at,omitempty"`
	Error        string             `json:"error,omitempty"`
	Result       *modeltrace.Result `json:"result,omitempty"`
}

// Separate capability keeps the general account repository contract unchanged.
type modelFingerprintStore interface {
	ClaimModelFingerprint(context.Context, string, *ModelFingerprintSnapshot) (bool, error)
	SaveModelFingerprint(context.Context, string, *ModelFingerprintSnapshot) error
}

func (s *AccountTestService) GetModelFingerprint(ctx context.Context, id string) (*ModelFingerprintSnapshot, error) {
	a, err := s.accountRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ParseModelFingerprintSnapshot(a.Extra[ModelFingerprintExtraKey], time.Now())
}

// The worker lease (expires_at) is separate from the completed result's lifetime.
func ParseModelFingerprintSnapshot(raw any, now time.Time) (*ModelFingerprintSnapshot, error) {
	if raw == nil {
		return nil, nil
	}
	body, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var snapshot ModelFingerprintSnapshot
	if err := json.Unmarshal(body, &snapshot); err != nil {
		return nil, err
	}
	finished := snapshot.StartedAt
	if !snapshot.ExpiresAt.IsZero() {
		finished = snapshot.ExpiresAt
	}
	if snapshot.FinishedAt != nil {
		finished = *snapshot.FinishedAt
	}
	if !now.Before(finished.Add(ModelFingerprintRetention)) {
		return nil, nil
	}
	if snapshot.Status == "running" && !now.Before(snapshot.ExpiresAt) {
		snapshot.Status = "failed"
		snapshot.Error = "interrupted"
		snapshot.FinishedAt = &finished
	}
	return &snapshot, nil
}

// DirectModelTestAccount isolates test routing from the account's saved allowlist.
func DirectModelTestAccount(account *Account) *Account {
	copy := *account
	copy.Credentials = maps.Clone(account.Credentials)
	delete(copy.Credentials, "model_mapping")
	copy.modelMappingCacheReady = false
	copy.modelMappingCache = nil
	return &copy
}

func (s *AccountTestService) StartModelFingerprint(ctx context.Context, id, model, userID string) (*ModelFingerprintSnapshot, error) {
	model = strings.TrimSpace(model)
	a, err := s.accountRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if model == "" || len(model) > 100 || strings.ContainsAny(model, "\r\n*") {
		return nil, apperrors.BadRequest("INVALID_MODEL", "Select a text model")
	}
	if err := CheckActiveModel(model); err != nil {
		return nil, err
	}
	if !IsModelFingerprintTextModel(model) {
		return nil, apperrors.BadRequest("INVALID_MODEL", "Model fingerprinting requires an available text model")
	}
	if err := ValidateAccountProtectionConfiguration(a); err != nil {
		return nil, err
	}
	if userID == "" || s.modelFingerprintUsage == nil {
		return nil, errors.New("model fingerprint usage logging unavailable")
	}
	store, ok := s.accountRepo.(modelFingerprintStore)
	if !ok {
		return nil, errors.New("model fingerprint store unavailable")
	}
	// Bound work per server; the database claim also prevents duplicates across replicas.
	s.modelFingerprintMu.Lock()
	if s.modelFingerprintActive >= 4 {
		s.modelFingerprintMu.Unlock()
		return nil, apperrors.New(429, "MODEL_FINGERPRINT_BUSY", "Model fingerprint workers are busy")
	}
	s.modelFingerprintActive++
	s.modelFingerprintMu.Unlock()
	release := func() { s.modelFingerprintMu.Lock(); s.modelFingerprintActive--; s.modelFingerprintMu.Unlock() }
	now := time.Now().UTC()
	snapshot := &ModelFingerprintSnapshot{ID: uuid.NewString(), Model: model, Status: "running", SamplingMode: "single_conversation", Total: modeltrace.QueryCount, StartedAt: now, ExpiresAt: now.Add(5 * time.Minute)}
	claimed, err := store.ClaimModelFingerprint(ctx, id, snapshot)
	if err != nil {
		release()
		return nil, err
	}
	if !claimed {
		release()
		return nil, apperrors.New(409, "MODEL_FINGERPRINT_RUNNING", "An account fingerprint test is already running")
	}
	initial := *snapshot
	go func() { defer release(); s.runModelFingerprint(id, userID, snapshot, store) }()
	return &initial, nil
}

func IsModelFingerprintTextModel(model string) bool {
	m := strings.ToLower(model)
	for _, excluded := range []string{"image", "imagine", "video", "audio", "voice", "tts", "stt", "whisper", "realtime", "embedding", "rerank", "moderation", "dall-e", "sora", "veo", "imagen"} {
		if strings.Contains(m, excluded) {
			return false
		}
	}
	return strings.TrimSpace(model) != ""
}

func (s *AccountTestService) runModelFingerprint(id, userID string, snapshot *ModelFingerprintSnapshot, store modelFingerprintStore) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	defer func() {
		if recover() != nil {
			snapshot.Status = "failed"
			snapshot.Error = "internal_error"
		}
		finished := time.Now().UTC()
		snapshot.FinishedAt = &finished
		persistCtx, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		if err := store.SaveModelFingerprint(persistCtx, id, snapshot); err != nil {
			slog.Warn("model fingerprint result could not be saved", "account_id", id, "job_id", snapshot.ID)
		}
	}()
	outputs := make([][]int, 0, modeltrace.QueryCount)
	conversation := &modelFingerprintConversation{id: snapshot.ID}
	for turn, challenge := range modeltrace.Challenges() {
		probeCtx, stop := context.WithTimeout(ctx, 70*time.Second)
		probe := &modelFingerprintProbe{model: snapshot.Model, prompt: challenge.Prompt, expected: challenge.Expected, cancel: stop, conversation: conversation, startedAt: time.Now()}
		probeCtx = context.WithValue(probeCtx, modelFingerprintContextKey{}, probe)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/model-fingerprint", nil).WithContext(probeCtx)
		err := s.TestAccountConnection(c, id, snapshot.Model, challenge.Prompt, AccountTestModeDefault)
		timedOut := probeCtx.Err() != nil && !probe.enough
		stop()
		snapshot.Completed++
		if err := s.recordModelFingerprintUsage(userID, id, snapshot, turn, probe); err != nil {
			snapshot.Status = "failed"
			snapshot.Error = "save_failed"
			slog.Warn("model fingerprint usage could not be saved", "account_id", id, "job_id", snapshot.ID)
			return
		}
		if !probe.enough && (err != nil || probe.failed || timedOut || !probe.complete) {
			snapshot.Status = "failed"
			snapshot.Error = "upstream_failed"
			if timedOut {
				snapshot.Error = "timeout"
			}
			return
		}
		numbers := modeltrace.ParseNumbers(probe.text.String())
		conversation.turns = append(conversation.turns, modelFingerprintTurn{role: "user", text: challenge.Prompt}, modelFingerprintTurn{role: "assistant", text: probe.text.String()})
		if len(numbers) >= modeltrace.MinimumNumbers(challenge.Expected) {
			outputs = append(outputs, numbers)
			snapshot.Valid++
		}
		if err := store.SaveModelFingerprint(ctx, id, snapshot); err != nil {
			snapshot.Status = "failed"
			snapshot.Error = "save_failed"
			return
		}
	}
	if len(outputs) != modeltrace.QueryCount {
		snapshot.Status = "failed"
		snapshot.Error = "insufficient_samples"
		return
	}
	result, err := modeltrace.Analyze(outputs)
	if err != nil {
		snapshot.Status = "failed"
		snapshot.Error = "analysis_failed"
		return
	}
	snapshot.Result = result
	snapshot.Status = "completed"
}

type modelFingerprintContextKey struct{}
type modelFingerprintTurn struct{ role, text string }
type modelFingerprintConversation struct {
	id           string
	claudeUserID string
	turns        []modelFingerprintTurn
}
type modelFingerprintProbe struct {
	conversation  *modelFingerprintConversation
	model         string
	prompt        string
	expected      int
	text          strings.Builder
	cancel        context.CancelFunc
	enough        bool
	complete      bool
	failed        bool
	startedAt     time.Time
	firstTokenMs  *int
	upstreamModel string
	usage         UsageLog
}

func fingerprintProbe(ctx context.Context) *modelFingerprintProbe {
	p, _ := ctx.Value(modelFingerprintContextKey{}).(*modelFingerprintProbe)
	return p
}

func (p *modelFingerprintProbe) event(event TestEvent) {
	switch event.Type {
	case "test_start":
		p.upstreamModel = event.Model
	case "content":
		if p.firstTokenMs == nil && event.Text != "" {
			ms := int(time.Since(p.startedAt).Milliseconds())
			p.firstTokenMs = &ms
		}
		if p.enough || p.failed {
			return
		}
		if p.text.Len()+len(event.Text) > 16<<10 {
			p.failed = true
			p.cancel()
			return
		}
		_, _ = p.text.WriteString(event.Text)
		text := p.text.String()
		// Wait for a delimiter so a digit split over SSE chunks is not counted early.
		if len(text) > 0 && (text[len(text)-1] < '0' || text[len(text)-1] > '9') && len(modeltrace.ParseNumbers(text)) >= p.expected {
			p.enough = true
			p.cancel()
		}
	case "test_complete":
		p.complete = event.Success
	case "error":
		p.failed = true
	}
}

type modelFingerprintUsageWriter interface {
	Create(context.Context, *UsageLog) (bool, error)
}

func (s *AccountTestService) recordModelFingerprintUsage(userID, accountID string, snapshot *ModelFingerprintSnapshot, turn int, probe *modelFingerprintProbe) error {
	log := probe.usage
	log.UserID, log.AccountID = userID, accountID
	log.RequestID = fmt.Sprintf("fingerprint:%s:%d", snapshot.ID, turn+1)
	log.Model, log.RequestedModel = snapshot.Model, snapshot.Model
	log.RequestType, log.Stream = RequestTypeTest, true
	log.SessionID = &snapshot.ID
	log.CreatedAt = probe.startedAt
	ms := int(time.Since(probe.startedAt).Milliseconds())
	log.DurationMs, log.FirstTokenMs = &ms, probe.firstTokenMs
	if probe.upstreamModel != "" {
		log.UpstreamModel = &probe.upstreamModel
	}
	endpoint := "/admin/accounts/:id/model-fingerprint"
	log.InboundEndpoint = &endpoint
	// Persist even when sampling canceled upstream early. Tests never debit a wallet/key.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := s.modelFingerprintUsage.Create(ctx, &log)
	return err
}

// Only fingerprint contexts alter payloads. Native identity/authentication stays in
// the existing test transport, while prompts and token limits are protocol specific.
func applyModelFingerprintPayload(ctx context.Context, payload map[string]any, protocol string, codex bool) {
	p := fingerprintProbe(ctx)
	if p == nil {
		return
	}
	turns := append(append([]modelFingerprintTurn(nil), p.conversation.turns...), modelFingerprintTurn{role: "user", text: p.prompt})
	switch protocol {
	case "anthropic", "chat":
		messages := make([]map[string]any, 0, len(turns))
		for _, turn := range turns {
			messages = append(messages, map[string]any{"role": turn.role, "content": turn.text})
		}
		payload["messages"] = messages
		if metadata, ok := payload["metadata"].(map[string]string); ok {
			if p.conversation.claudeUserID == "" {
				p.conversation.claudeUserID = metadata["user_id"]
			}
			metadata["user_id"] = p.conversation.claudeUserID
		}
		payload["max_tokens"] = modelFingerprintMaxTokens
		delete(payload, "temperature")
		if protocol == "chat" {
			payload["stream_options"] = map[string]any{"include_usage": true}
			model, _ := payload["model"].(string)
			if modelFingerprintReasoningModel(model) {
				delete(payload, "max_tokens")
				payload["max_completion_tokens"] = modelFingerprintMaxTokens
			}
		}
	case "responses":
		input := make([]map[string]any, 0, len(turns))
		for _, turn := range turns {
			kind := "input_text"
			if turn.role == "assistant" {
				kind = "output_text"
			}
			input = append(input, map[string]any{"role": turn.role, "content": []map[string]any{{"type": kind, "text": turn.text}}})
		}
		payload["input"] = input
		delete(payload, "tools")
		if !codex {
			payload["max_output_tokens"] = modelFingerprintMaxTokens
		}
	case "gemini":
		contents := make([]map[string]any, 0, len(turns))
		for _, turn := range turns {
			role := turn.role
			if role == "assistant" {
				role = "model"
			}
			contents = append(contents, map[string]any{"role": role, "parts": []map[string]any{{"text": turn.text}}})
		}
		payload["contents"] = contents
		payload["generationConfig"] = map[string]any{"maxOutputTokens": modelFingerprintMaxTokens}
	}
	applyModelFingerprintEffort(p, payload, protocol)
}

func applyModelFingerprintEffort(probe *modelFingerprintProbe, payload map[string]any, protocol string) {
	model, _ := payload["model"].(string)
	if model == "" {
		// Gemini and Bedrock carry the model in the URL instead of the payload.
		model = probe.model
	}
	model = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(model), "models/"))
	effort := "low"
	switch protocol {
	case "chat":
		if !modelFingerprintReasoningModel(model) && !grokSupportsReasoningEffort(model) && payload["reasoning_effort"] == nil {
			return
		}
		payload["reasoning_effort"] = effort
	case "responses":
		if !modelFingerprintReasoningModel(model) && !grokSupportsReasoningEffort(model) && payload["reasoning"] == nil {
			return
		}
		reasoning, _ := payload["reasoning"].(map[string]any)
		if reasoning == nil {
			reasoning = map[string]any{}
		}
		reasoning["effort"] = effort
		payload["reasoning"] = reasoning
	case "anthropic":
		if len(claude.EffortLevelsForModel(model)) == 0 {
			return
		}
		config, _ := payload["output_config"].(map[string]any)
		if config == nil {
			config = map[string]any{}
		}
		config["effort"] = effort
		payload["output_config"] = config
		delete(payload, "thinking")
	case "gemini":
		config, _ := payload["generationConfig"].(map[string]any)
		switch {
		case strings.HasPrefix(model, "gemini-3"):
			config["thinkingConfig"] = map[string]any{"thinkingLevel": effort, "includeThoughts": false}
		case strings.HasPrefix(model, "gemini-2.5"), strings.HasPrefix(model, "claude-"):
			// Budget-based APIs express the low tier with at most 1024 tokens.
			config["thinkingConfig"] = map[string]any{"thinkingBudget": geminiThinkingBudgetLowMax, "includeThoughts": false}
		default:
			return
		}
	default:
		return
	}
	probe.usage.ReasoningEffort = &effort
	probe.usage.RequestedReasoningEffort = &effort
}

func modelFingerprintReasoningModel(model string) bool {
	for _, prefix := range []string{"gpt-5", "gpt-6", "o1", "o3", "o4"} {
		if strings.HasPrefix(model, prefix) {
			return true
		}
	}
	return false
}

func applyModelFingerprintHeaders(req *http.Request) {
	if p := fingerprintProbe(req.Context()); p != nil {
		req.Header.Set("Session_ID", p.conversation.id)
		req.Header.Set("Conversation_ID", p.conversation.id)
	}
}

func modelFingerprintPayloadBytes(ctx context.Context, body []byte, protocol string) []byte {
	if fingerprintProbe(ctx) == nil {
		return body
	}
	var payload map[string]any
	if json.Unmarshal(body, &payload) != nil {
		return body
	}
	applyModelFingerprintPayload(ctx, payload, protocol, false)
	result, err := json.Marshal(payload)
	if err != nil {
		return body
	}
	return result
}

func modelFingerprintSafeError(ctx context.Context, message string) string {
	if fingerprintProbe(ctx) != nil {
		return "Model fingerprint upstream request failed"
	}
	return message
}

//go:build unit

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/config"
	apperrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	"github.com/AsukaCC/EasySub2api/internal/pkg/tlsfingerprint"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type fingerprintTestRepo struct {
	AccountRepository
	mu        sync.Mutex
	snapshots map[string]*ModelFingerprintSnapshot
	done      chan *ModelFingerprintSnapshot
}

func (r *fingerprintTestRepo) GetByID(_ context.Context, id string) (*Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return &Account{ID: id, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive,
		Credentials: map[string]any{"api_key": "test-only", "base_url": "https://example.com"},
		Extra:       map[string]any{"openai_responses_mode": "force_chat_completions", ModelFingerprintExtraKey: r.snapshots[id]}}, nil
}
func (r *fingerprintTestRepo) ClaimModelFingerprint(_ context.Context, id string, snapshot *ModelFingerprintSnapshot) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if previous := r.snapshots[id]; previous != nil && previous.Status == "running" && time.Now().Before(previous.ExpiresAt) {
		return false, nil
	}
	copy := *snapshot
	r.snapshots[id] = &copy
	return true, nil
}
func (r *fingerprintTestRepo) SaveModelFingerprint(_ context.Context, id string, snapshot *ModelFingerprintSnapshot) error {
	r.mu.Lock()
	copy := *snapshot
	r.snapshots[id] = &copy
	r.mu.Unlock()
	if snapshot.Status != "running" {
		r.done <- &copy
	}
	return nil
}

type fingerprintTestTransport struct {
	HTTPUpstream
	mu       sync.Mutex
	requests map[string][][]byte
	sessions map[string][]string
	entered  chan string
	release  chan struct{}
	invalid  bool
}

func (u *fingerprintTestTransport) DoWithTLS(req *http.Request, _ string, accountID string, _ int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	u.mu.Lock()
	u.requests[accountID] = append(u.requests[accountID], body)
	u.sessions[accountID] = append(u.sessions[accountID], req.Header.Get("Session_ID"))
	first := len(u.requests[accountID]) == 1
	u.mu.Unlock()
	if first && u.entered != nil {
		u.entered <- accountID
		select {
		case <-u.release:
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
	}
	text := strings.Repeat("17,42,138,251,", 60)
	if u.invalid {
		text = "I cannot provide numbers"
	}
	delta, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"delta": map[string]string{"content": text}}}})
	return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(fmt.Sprintf("data: %s\n\ndata: [DONE]\n\n", delta)))}, nil
}

func newFingerprintTestService() (*AccountTestService, *fingerprintTestRepo, *fingerprintTestTransport) {
	repo := &fingerprintTestRepo{snapshots: make(map[string]*ModelFingerprintSnapshot), done: make(chan *ModelFingerprintSnapshot, 8)}
	upstream := &fingerprintTestTransport{requests: make(map[string][][]byte), sessions: make(map[string][]string)}
	return &AccountTestService{accountRepo: repo, httpUpstream: upstream, cfg: &config.Config{}}, repo, upstream
}

func waitFingerprint(t *testing.T, done <-chan *ModelFingerprintSnapshot) *ModelFingerprintSnapshot {
	t.Helper()
	select {
	case result := <-done:
		return result
	case <-time.After(5 * time.Second):
		t.Fatal("test worker did not finish")
		return nil
	}
}

func TestModelFingerprintSingleConversationAndAccountIsolation(t *testing.T) {
	svc, repo, upstream := newFingerprintTestService()
	upstream.entered = make(chan string, 2)
	upstream.release = make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	first, err := svc.StartModelFingerprint(ctx, "one", "gpt-6-astra")
	require.NoError(t, err)
	cancel() // Closing the browser request must not cancel the worker.
	select {
	case <-upstream.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("worker not started")
	}
	_, err = svc.StartModelFingerprint(context.Background(), "one", "gpt-6-astra")
	require.Equal(t, 409, apperrors.Code(err))
	second, err := svc.StartModelFingerprint(context.Background(), "two", "gpt-6-astra")
	require.NoError(t, err)
	select {
	case <-upstream.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("other account cannot run concurrently")
	}
	upstream.mu.Lock()
	require.Len(t, upstream.requests["one"], 1)
	require.Len(t, upstream.requests["two"], 1)
	upstream.mu.Unlock()
	close(upstream.release)
	for range 2 {
		result := waitFingerprint(t, repo.done)
		require.Equal(t, "completed", result.Status)
		require.Equal(t, 3, result.Valid)
		require.NotNil(t, result.Result)
	}
	upstream.mu.Lock()
	defer upstream.mu.Unlock()
	for _, item := range []struct{ id, session string }{{"one", first.ID}, {"two", second.ID}} {
		require.Len(t, upstream.requests[item.id], 3)
		for i, body := range upstream.requests[item.id] {
			require.Len(t, gjson.GetBytes(body, "messages").Array(), 2*i+1)
			require.Equal(t, item.session, upstream.sessions[item.id][i])
			require.Equal(t, int64(modelFingerprintMaxTokens), gjson.GetBytes(body, "max_completion_tokens").Int())
			require.Equal(t, "low", gjson.GetBytes(body, "reasoning_effort").String())
			if i > 0 {
				require.Equal(t, "assistant", gjson.GetBytes(body, "messages.1.role").String())
				require.Contains(t, gjson.GetBytes(body, "messages.1.content").String(), "17,42")
			}
		}
	}
}

func TestModelFingerprintInvalidSamplesAndMedia(t *testing.T) {
	svc, repo, upstream := newFingerprintTestService()
	upstream.invalid = true
	_, err := svc.StartModelFingerprint(context.Background(), "one", "gpt-image-2")
	require.Equal(t, 400, apperrors.Code(err))
	_, err = svc.StartModelFingerprint(context.Background(), "one", "gpt-6-astra")
	require.NoError(t, err)
	result := waitFingerprint(t, repo.done)
	require.Equal(t, "insufficient_samples", result.Error)
	require.Nil(t, result.Result)
	require.Equal(t, 3, result.Completed)
	stale := &ModelFingerprintSnapshot{Status: "running", ExpiresAt: time.Now().Add(-time.Minute)}
	repo.mu.Lock()
	repo.snapshots["stale"] = stale
	repo.mu.Unlock()
	status, err := svc.GetModelFingerprint(context.Background(), "stale")
	require.NoError(t, err)
	require.Equal(t, "interrupted", status.Error)
}

func TestModelFingerprintPayloadProtocols(t *testing.T) {
	conversation := &modelFingerprintConversation{id: "session", turns: []modelFingerprintTurn{{role: "user", text: "first"}, {role: "assistant", text: "17,42"}}}
	ctx := context.WithValue(context.Background(), modelFingerprintContextKey{}, &modelFingerprintProbe{conversation: conversation, prompt: "next"})
	for _, tc := range []struct{ protocol, key, role string }{{"anthropic", "messages", "assistant"}, {"chat", "messages", "assistant"}, {"responses", "input", "assistant"}, {"gemini", "contents", "model"}} {
		payload := map[string]any{"model": "gpt-6-astra"}
		applyModelFingerprintPayload(ctx, payload, tc.protocol, false)
		turns := payload[tc.key].([]map[string]any)
		require.Len(t, turns, 3)
		require.Equal(t, tc.role, turns[1]["role"])
	}
	codex := map[string]any{}
	applyModelFingerprintPayload(ctx, codex, "responses", true)
	require.NotContains(t, codex, "max_output_tokens")
	unchanged := map[string]any{"input": "hi"}
	applyModelFingerprintPayload(context.Background(), unchanged, "responses", false)
	require.Equal(t, map[string]any{"input": "hi"}, unchanged)
}

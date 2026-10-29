//go:build unit

package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/config"
	"github.com/AsukaCC/EasySub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type geminiSequenceUpstreamResponse struct {
	status int
	body   string
}
type geminiSequenceUpstreamStub struct {
	responses []geminiSequenceUpstreamResponse
	calls     int
}

func (s *geminiSequenceUpstreamStub) Do(*http.Request, string, string, int) (*http.Response, error) {
	r := s.responses[min(s.calls, len(s.responses)-1)]
	s.calls++
	return &http.Response{StatusCode: r.status, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(r.body))}, nil
}
func (s *geminiSequenceUpstreamStub) DoWithTLS(req *http.Request, proxy string, account string, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return s.Do(req, proxy, account, concurrency)
}
func newGeminiSingleCheckService(responses ...geminiSequenceUpstreamResponse) (*GeminiMessagesCompatService, *geminiSequenceUpstreamStub, *rateLimitAccountRepoStub) {
	repo := &rateLimitAccountRepoStub{}
	upstream := &geminiSequenceUpstreamStub{responses: responses}
	return &GeminiMessagesCompatService{accountRepo: repo, httpUpstream: upstream, rateLimitService: NewRateLimitService(repo, &config.Config{}, nil), cfg: &config.Config{}}, upstream, repo
}

var geminiSingleCheckForwarders = []struct {
	name    string
	forward func(*testing.T, *GeminiMessagesCompatService, *Account) (error, *gin.Context)
}{
	{"native", func(t *testing.T, s *GeminiMessagesCompatService, a *Account) (error, *gin.Context) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		body := []byte(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`)
		c.Request = httptest.NewRequest("POST", "/v1beta/models/gemini-3.8-flash:generateContent", bytes.NewReader(body))
		_, err := s.ForwardNative(context.Background(), c, a, "gemini-3.8-flash", "generateContent", false, body)
		return err, c
	}},
	{"claude_compat", func(t *testing.T, s *GeminiMessagesCompatService, a *Account) (error, *gin.Context) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		body := []byte(`{"model":"gemini-3.8-flash","max_tokens":16,"messages":[{"role":"user","content":"hi"}]}`)
		c.Request = httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
		_, err := s.Forward(context.Background(), c, a, body)
		return err, c
	}},
	{"chat_completions_compat", func(t *testing.T, s *GeminiMessagesCompatService, a *Account) (error, *gin.Context) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		body := []byte(`{"model":"gemini-3.8-flash","messages":[{"role":"user","content":"hi"}]}`)
		c.Request = httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(body))
		_, err := s.ForwardAsChatCompletions(context.Background(), c, a, body)
		return err, c
	}},
}

func TestRelease0219GeminiDeclaredModels(t *testing.T) {
	gid := "gemini-group"
	for _, tc := range []struct {
		name     string
		accounts []Account
		want     []string
	}{
		{"all mapped", []Account{{Platform: PlatformGemini, Credentials: map[string]any{"model_mapping": map[string]any{"custom-gemini": "gemini-3.8-flash", "gemini-3.8-*": "gemini-3.8-flash"}}}}, []string{"custom-gemini", "gemini-3.8-*"}},
		{"unrestricted account retains discovery", []Account{{Platform: PlatformGemini, Credentials: map[string]any{"model_mapping": map[string]any{"custom": "gemini-3.8-flash"}}}, {Platform: PlatformGemini}}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockAccountRepoForGemini{listByGroupFunc: func(context.Context, string, []string) ([]Account, error) { return tc.accounts, nil }}
			s := &GeminiMessagesCompatService{accountRepo: repo, cfg: &config.Config{}}
			got, err := s.DeclaredModelIDs(context.Background(), &gid)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
	require.True(t, DeclaredModelsCover([]string{"gemini-3.8-*"}, "gemini-3.8-flash"))
	require.False(t, DeclaredModelsCover([]string{"gemini-3.8-*"}, "other"))
}

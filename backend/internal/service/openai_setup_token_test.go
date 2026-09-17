package service

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestOpenAISetupTokenCompatibilityPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, kind := range []string{AccountTypeOAuth, AccountTypeSetupToken} {
		for _, endpoint := range []string{"/v1/chat/completions", "/v1/messages"} {
			t.Run(kind+endpoint, func(t *testing.T) {
				body := []byte(`{"model":"gpt-5.2","max_tokens":128,"messages":[{"role":"user","content":"hello"}],"stream":false}`)
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = httptest.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
				upstream := &httpUpstreamRecorder{resp: openAICompatSSECompletedResponse("resp_setup", "gpt-5.2")}
				svc := &OpenAIGatewayService{httpUpstream: upstream, cfg: &config.Config{}}
				account := &Account{ID: "setup-test", Platform: PlatformOpenAI, Type: kind,
					Credentials: map[string]any{"access_token": "test-token", "chatgpt_account_id": "test-account"}}
				var err error
				if endpoint == "/v1/messages" {
					_, err = svc.ForwardAsAnthropic(context.Background(), c, account, body, "", "gpt-5.2")
				} else {
					_, err = svc.ForwardAsChatCompletions(context.Background(), c, account, body, "", "gpt-5.2")
				}
				require.NoError(t, err)
				require.Equal(t, chatgptCodexURL, upstream.lastReq.URL.String())
				require.Equal(t, "Bearer test-token", upstream.lastReq.Header.Get("Authorization"))
				require.Equal(t, "test-account", upstream.lastReq.Header.Get("ChatGPT-Account-Id"))
				require.Empty(t, upstream.lastReq.Header.Get("originator"))
				require.Equal(t, codexCLIUserAgent, upstream.lastReq.Header.Get("User-Agent"))
				require.True(t, gjson.GetBytes(upstream.lastBody, "instructions").Exists())
				require.False(t, gjson.GetBytes(upstream.lastBody, "max_output_tokens").Exists())
				require.False(t, gjson.GetBytes(upstream.lastBody, "store").Bool())
			})
		}
	}
}

func TestOpenAISetupTokenPassthrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, kind := range []string{AccountTypeOAuth, AccountTypeSetupToken, AccountTypeAPIKey} {
		t.Run(kind, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
			c.Request.Header.Set("session_id", "client-session")
			account := &Account{ID: "setup-test", Platform: PlatformOpenAI, Type: kind,
				Credentials: map[string]any{"chatgpt_account_id": "test-account"}}
			svc := &OpenAIGatewayService{}
			req, err := svc.buildUpstreamRequestOpenAIPassthrough(context.Background(), c, account,
				[]byte(`{"model":"gpt-5.2","instructions":"test","input":[],"stream":true}`), "test-token")
			require.NoError(t, err)
			if kind == AccountTypeAPIKey {
				require.Equal(t, openaiPlatformAPIURL, req.URL.String())
				require.Empty(t, req.Header.Get("ChatGPT-Account-Id"))
				return
			}
			require.Equal(t, chatgptCodexURL, req.URL.String())
			require.Equal(t, "chatgpt.com", req.Host)
			require.Equal(t, "test-account", req.Header.Get("ChatGPT-Account-Id"))
			require.NotEmpty(t, req.Header.Get("originator"))
			require.NotEmpty(t, req.Header.Get("session_id"))
			require.NotEqual(t, "client-session", req.Header.Get("session_id"))
		})
	}
}

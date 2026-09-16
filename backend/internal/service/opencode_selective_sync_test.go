package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/config"
	"github.com/AsukaCC/EasySub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type openCodeSyncTransport struct {
	HTTPUpstream
	request *http.Request
	body    []byte
	status  int
}

func (u *openCodeSyncTransport) Do(req *http.Request, _ string, _ string, _ int) (*http.Response, error) {
	u.request = req
	u.body, _ = io.ReadAll(req.Body)
	status := u.status
	if status == 0 {
		status = 200
	}
	payload := `{"error":{"message":"quota exhausted","type":"insufficient_quota"}}`
	if status == 200 {
		switch {
		case strings.HasSuffix(req.URL.Path, "/messages"):
			payload = "event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"msg_fixture\",\"type\":\"message\",\"role\":\"assistant\",\"model\":\"fixture\",\"content\":[],\"usage\":{\"input_tokens\":5,\"output_tokens\":0}}}\n\nevent: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\nevent: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"OK\"}}\n\nevent: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\nevent: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":2}}\n\nevent: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"
		case strings.HasSuffix(req.URL.Path, "/chat/completions"):
			payload = "data: {\"id\":\"chat_fixture\",\"object\":\"chat.completion.chunk\",\"model\":\"fixture\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"content\":\"OK\"},\"finish_reason\":null}]}\n\ndata: {\"id\":\"chat_fixture\",\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":5,\"completion_tokens\":2,\"total_tokens\":7}}\n\ndata: [DONE]\n\n"
		default:
			payload = "data: {\"type\":\"response.created\",\"response\":{\"id\":\"resp_fixture\",\"status\":\"in_progress\"}}\n\ndata: {\"type\":\"response.output_text.delta\",\"delta\":\"OK\",\"output_index\":0,\"content_index\":0,\"item_id\":\"msg_fixture\"}\n\ndata: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_fixture\",\"object\":\"response\",\"status\":\"completed\",\"output\":[{\"type\":\"message\",\"role\":\"assistant\",\"content\":[{\"type\":\"output_text\",\"text\":\"OK\"}]}],\"usage\":{\"input_tokens\":5,\"output_tokens\":2,\"total_tokens\":7}}}\n\n"
		}
	}
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": []string{"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(payload)), Request: req}, nil
}
func (u *openCodeSyncTransport) DoWithTLS(req *http.Request, proxy, id string, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.Do(req, proxy, id, concurrency)
}

func TestOpenCodeThreeProtocolRoutingAndMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, inbound := range []string{"responses", "chat", "messages"} {
		for _, tc := range []struct{ model, path string }{{"gpt-fixture", "/zen/go/v1/responses"}, {"minimax-fixture", "/zen/go/v1/messages"}, {"glm-fixture", "/zen/go/v1/chat/completions"}} {
			t.Run(inbound+"/"+tc.model, func(t *testing.T) {
				transport := &openCodeSyncTransport{}
				svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: transport}
				account := &Account{ID: "fixture-account", Platform: PlatformOpenCodeGo, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "fixture-key", "account_mode": "go", "model_mapping": map[string]any{"alias": tc.model}}}
				body := []byte(`{"model":"alias","stream":true,"prompt_cache_key":"fixture-session","input":"hi","messages":[{"role":"user","content":"hi"}],"max_tokens":100}`)
				recorder := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/"+inbound, bytes.NewReader(body))
				var result *OpenAIForwardResult
				var err error
				switch inbound {
				case "responses":
					result, err = svc.Forward(context.Background(), c, account, body)
				case "chat":
					result, err = svc.ForwardAsChatCompletions(context.Background(), c, account, body, "", "")
				case "messages":
					result, err = svc.ForwardAsAnthropic(context.Background(), c, account, body, "", "")
				}
				require.NoError(t, err)
				require.NotNil(t, result)
				require.NotNil(t, transport.request)
				require.Equal(t, tc.path, transport.request.URL.Path)
				require.Equal(t, tc.model, gjson.GetBytes(transport.body, "model").String())
				require.Equal(t, "fixture-session", transport.request.Header.Get(openCodeSessionHeader))
				require.Contains(t, recorder.Body.String(), "OK")
			})
		}
	}
}

func TestOpenCodeCredentialAndAccountIsolation(t *testing.T) {
	for _, mode := range []string{"zen", "go"} {
		account := &Account{ID: "fixture", Platform: PlatformOpenCodeGo, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"account_mode": mode, "api_key": "fixture"}}
		require.NoError(t, ValidateOpenCodeAccount(account))
		require.True(t, account.IsSchedulable())
		for _, field := range []string{"api_key", "account_mode", "api_protocol"} {
			t.Run(fmt.Sprintf("%s/%s", mode, field), func(t *testing.T) {
				original, present := account.Credentials[field]
				account.Credentials[field] = ""
				require.Error(t, ValidateOpenCodeAccount(account))
				require.False(t, account.IsSchedulable())
				if present {
					account.Credentials[field] = original
				} else {
					delete(account.Credentials, field)
				}
			})
		}
	}
	require.Equal(t, PlatformOpenAI, NormalizeOpenAICompatiblePlatform(PlatformOpenAI))
	require.Equal(t, PlatformOpenCodeGo, NormalizeOpenAICompatiblePlatform(PlatformOpenCodeGo))
}

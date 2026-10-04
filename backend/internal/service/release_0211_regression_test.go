//go:build unit

package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/pkg/apicompat"
	"github.com/AsukaCC/EasySub2api/internal/pkg/openai"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestGPT61MessagesEffortRoundTrip(t *testing.T) {
	for _, effort := range []string{"low", "medium", "high", "xhigh", "max", "ultra"} {
		for _, prefix := range []string{"gpt-6.1-sol-", "openai/gpt61-sol-"} {
			t.Run(prefix+effort, func(t *testing.T) {
				model := prefix + effort
				req := &apicompat.AnthropicRequest{Model: model}
				require.Equal(t, "gpt-6.1-sol", NormalizeOpenAICompatRequestedModel(model))
				applyOpenAICompatModelNormalization(req)
				require.Equal(t, "gpt-6.1-sol", req.Model)
				converted, err := apicompat.AnthropicToResponses(req)
				require.NoError(t, err)
				require.Equal(t, effort, converted.Reasoning.Effort)
				require.Equal(t, effort, openAICompatAnthropicReasoningEffort(req, req.Model, converted.Reasoning.Effort))
				req.OutputConfig.Effort = "high"
				req.Model = model
				applyOpenAICompatModelNormalization(req)
				require.Equal(t, "high", req.OutputConfig.Effort, "explicit effort wins over suffix")
			})
		}
	}
	for _, effort := range []string{"none", "minimal"} {
		req := &apicompat.AnthropicRequest{Model: "gpt-6.1-sol-" + effort}
		applyOpenAICompatModelNormalization(req)
		_, err := apicompat.AnthropicToResponses(req)
		require.Error(t, err)
	}
}

func TestGPT61CompatCacheIdentity(t *testing.T) {
	for _, model := range []string{"gpt-6.1-sol", "openai/gpt61-sol", "gpt-6.1-sol-max"} {
		require.True(t, shouldAutoInjectPromptCacheKeyForCompat(model), model)
	}
	require.False(t, shouldAutoInjectPromptCacheKeyForCompat("gpt-4o"))
}

func TestGPT61OfficialDescriptorPreservesMetadataAndOverrides(t *testing.T) {
	d := newConfiguredCodexModelDescriptor("gpt-6.1-sol")
	raw, err := json.Marshal(d)
	require.NoError(t, err)
	var actual, official map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &actual))
	require.NoError(t, json.Unmarshal(openai.CodexGPT61SolMetadata, &official))
	for key, expected := range official {
		t.Run(key, func(t *testing.T) { require.JSONEq(t, string(expected), string(actual[key])) })
	}
	require.Equal(t, d.ModelMessages.InstructionsTemplate, openai.CodexBaseInstructionsForModel("gpt-6.1-sol"))

	d.Slug = "public-sol"
	d.DisplayName = "Public Sol"
	applyUpstreamModelMetadataToCodexDescriptor(&d, codexModelMetadataOverride{UpstreamModelMetadata: UpstreamModelMetadata{
		ID: "gpt-6.1-sol", CodexToolCapabilities: accountCodexToolCapabilities(&Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}, "gpt-6.1-sol"),
	}})
	raw, err = json.Marshal(d)
	require.NoError(t, err)
	require.Equal(t, "public-sol", gjson.GetBytes(raw, "slug").String())
	require.Equal(t, "Public Sol", gjson.GetBytes(raw, "display_name").String())
	require.False(t, gjson.GetBytes(raw, "use_responses_lite").Bool())
	require.Equal(t, "xhigh", gjson.GetBytes(raw, "multi_agent_reasoning_effort").String())
	require.True(t, gjson.GetBytes(raw, "model_messages.tools").Exists())
}

func TestGPT61ChatOnlyGuardsAfterModelMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, protocol := range []string{"messages", "responses", "chat", "legacy-functions"} {
		for _, model := range []string{"gpt-6.1-sol", "public-sol", "gpt-5.6-sol"} {
			for _, reason := range []string{"tools", "disabled", "text"} {
				t.Run(protocol+"/"+model+"/"+reason, func(t *testing.T) {
					body := map[string]any{"model": model, "stream": false}
					switch protocol {
					case "messages":
						body["messages"] = []any{map[string]any{"role": "user", "content": "hello"}}
						body["max_tokens"] = 16
						if reason == "tools" {
							body["tools"] = []any{map[string]any{"name": "lookup", "input_schema": map[string]any{"type": "object"}}}
						} else if reason == "disabled" {
							body["output_config"] = map[string]any{"effort": "none"}
						}
					case "responses":
						body["input"] = "hello"
						if reason == "tools" {
							body["tools"] = []any{map[string]any{"type": "function", "name": "lookup", "parameters": map[string]any{"type": "object"}}}
						} else if reason == "disabled" {
							body["reasoning"] = map[string]any{"effort": "none"}
						}
					default:
						body["messages"] = []any{map[string]any{"role": "user", "content": "hello"}}
						if reason == "tools" {
							function := map[string]any{"name": "lookup", "parameters": map[string]any{"type": "object"}}
							if protocol == "legacy-functions" {
								body["functions"] = []any{function}
							} else {
								body["tools"] = []any{map[string]any{"type": "function", "function": function}}
							}
						} else if reason == "disabled" {
							body["reasoning_effort"] = "none"
						}
					}
					encoded, err := json.Marshal(body)
					require.NoError(t, err)
					w := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(w)
					c.Request = httptest.NewRequest(http.MethodPost, "/v1/"+protocol, strings.NewReader(string(encoded)))
					upstream := &httpUpstreamRecorder{resp: &http.Response{
						StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}},
						Body: io.NopCloser(strings.NewReader(`{"id":"test","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":3,"completion_tokens":2}}`)),
					}}
					svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}
					account := forceChatMessagesFallbackAccount()
					mapped := model
					if model == "public-sol" {
						mapped = "gpt-6.1-sol"
					}
					account.Credentials["model_mapping"] = map[string]any{model: mapped}
					switch protocol {
					case "messages":
						_, err = svc.ForwardAsAnthropic(context.Background(), c, account, encoded, "", "")
					case "responses":
						_, err = svc.Forward(context.Background(), c, account, encoded)
					default:
						_, err = svc.ForwardAsChatCompletions(context.Background(), c, account, encoded, "", "")
					}
					if model != "gpt-5.6-sol" && reason != "text" {
						require.Error(t, err)
						require.Equal(t, http.StatusBadRequest, w.Code)
						require.Nil(t, upstream.lastReq, "invalid requests must not reach upstream")
					} else {
						require.NoError(t, err)
						require.NotNil(t, upstream.lastReq)
					}
				})
			}
		}
	}
}

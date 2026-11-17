package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestImageToolForwardingPreservesExplicitToolsWithoutLite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, accountType := range []string{AccountTypeAPIKey, AccountTypeOAuth, AccountTypeSetupToken} {
		for _, passthrough := range []bool{false, true} {
			for _, lite := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/passthrough=%t/lite=%t", accountType, passthrough, lite), func(t *testing.T) {
					upstream := &httpUpstreamRecorder{resp: &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{"Content-Type": []string{"application/json"}},
						Body:       io.NopCloser(strings.NewReader(`{"id":"resp_image","model":"gpt-6.1-sol","output":[],"usage":{"input_tokens":1,"output_tokens":1}}`)),
					}}
					svc := newOpenAIImageGenerationControlTestService(upstream)
					c, _ := newOpenAIImageGenerationControlTestContext(true, "codex_cli_rs/0.169.0")
					if lite {
						c.Request.Header.Set(responsesLiteHeader, "true")
					}
					account := newOpenAIImageGenerationControlTestAccount()
					account.Type = accountType
					account.Credentials["access_token"] = "test-token"
					account.Credentials["chatgpt_account_id"] = "test-account"
					account.Extra = map[string]any{"openai_passthrough": passthrough, featureKeyCodexImageGenerationExplicitToolPolicy: codexImageGenerationExplicitToolPolicyStrip}
					body := []byte(`{"model":"gpt-6.1-sol","input":"draw a square","stream":false,"tools":[{"type":"image_generation","model":"gpt-image-2","quality":"low"}],"tool_choice":{"type":"image_generation"}}`)
					_, err := svc.Forward(context.Background(), c, account, body)
					require.NoError(t, err)
					require.NotNil(t, upstream.lastReq)
					require.Empty(t, upstream.lastReq.Header.Get(responsesLiteHeader))
					require.Equal(t, "image_generation", gjson.GetBytes(upstream.lastBody, "tool_choice.type").String())
					require.Equal(t, "gpt-image-2", gjson.GetBytes(upstream.lastBody, `tools.#(type=="image_generation").model`).String())
					require.Equal(t, "low", gjson.GetBytes(upstream.lastBody, `tools.#(type=="image_generation").quality`).String())
				})
			}
		}
	}
}

func TestImageToolForwardingDoesNotInjectDespiteLegacyBridge(t *testing.T) {
	upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"id":"resp_text","model":"gpt-6.1-sol","output":[],"usage":{"input_tokens":1,"output_tokens":1}}`))}}
	svc := newOpenAIImageGenerationControlTestService(upstream)
	svc.cfg.Gateway.CodexImageGenerationBridgeEnabled = true
	c, _ := newOpenAIImageGenerationControlTestContext(true, "codex_cli_rs/0.169.0")
	account := newOpenAIImageGenerationControlTestAccount()
	account.Extra = map[string]any{featureKeyCodexImageGenerationBridge: true}
	_, err := svc.Forward(context.Background(), c, account, []byte(`{"model":"gpt-6.1-sol","input":"draw a square","stream":false}`))
	require.NoError(t, err)
	require.False(t, hasNativeImageGenerationTool(upstream.lastBody))
	require.False(t, gjson.GetBytes(upstream.lastBody, "tool_choice").Exists())
	require.NotContains(t, gjson.GetBytes(upstream.lastBody, "instructions").String(), "image_generation")
}

func TestImageToolWebSocketPayloadUsesFullResponses(t *testing.T) {
	body := []byte(`{"type":"response.create","model":"gpt-6.1-sol","client_metadata":{"ws_request_header_x_openai_internal_codex_responses_lite":"true","keep":"value"},"tools":[{"type":"image_generation","quality":"high"},{"type":"function","name":"shell"}],"tool_choice":{"type":"image_generation"}}`)
	updated, changed, err := normalizeOpenAIResponsesLiteToolsPayload(body)
	require.NoError(t, err)
	require.True(t, changed)
	require.False(t, isOpenAIResponsesLiteWebSocketPayload(updated))
	require.Equal(t, "value", gjson.GetBytes(updated, "client_metadata.keep").String())
	require.Equal(t, gjson.GetBytes(body, "tools").Raw, gjson.GetBytes(updated, "tools").Raw)
	require.Equal(t, "image_generation", gjson.GetBytes(updated, "tool_choice.type").String())
}

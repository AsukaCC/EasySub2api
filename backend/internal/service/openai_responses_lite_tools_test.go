//go:build unit

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
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestNormalizeOpenAIResponsesLiteTools_MovesNamespacesAndKeepsSupportedTools(t *testing.T) {
	reqBody := map[string]any{
		"model": "gpt-5.6-terra",
		"tools": []any{
			map[string]any{"type": "function", "name": "shell"},
			map[string]any{"type": "custom", "name": "exec"},
			map[string]any{"type": "tool_search", "execution": "client"},
			map[string]any{"type": "namespace", "name": "collaboration", "tools": []any{
				map[string]any{"type": "function", "name": "spawn_agent"},
			}},
		},
		"input": []any{
			map[string]any{"type": "message", "role": "user", "content": "hello"},
			map[string]any{"type": "additional_tools", "role": "developer", "tools": []any{
				map[string]any{"type": "namespace", "name": "image_gen"},
				map[string]any{"type": "namespace", "name": "collaboration", "tools": []any{
					map[string]any{"type": "function", "name": "spawn_agent"},
				}},
			}},
		},
		"tool_choice": map[string]any{"type": "namespace", "name": "collaboration"},
	}

	changed, err := normalizeOpenAIResponsesLiteTools(reqBody)

	require.NoError(t, err)
	require.True(t, changed)
	tools := reqBody["tools"].([]any)
	require.Len(t, tools, 3)
	require.Equal(t, "function", tools[0].(map[string]any)["type"])
	require.Equal(t, "custom", tools[1].(map[string]any)["type"])
	require.Equal(t, "tool_search", tools[2].(map[string]any)["type"])
	input := reqBody["input"].([]any)
	require.Len(t, input, 2)
	additional := input[1].(map[string]any)["tools"].([]any)
	require.Len(t, additional, 2)
	require.Equal(t, "image_gen", additional[0].(map[string]any)["name"])
	require.Equal(t, "collaboration", additional[1].(map[string]any)["name"], "existing namespace must not be duplicated")
	require.Equal(t, map[string]any{"type": "namespace", "name": "collaboration"}, reqBody["tool_choice"])
}

func TestNormalizeOpenAIResponsesLiteTools_RejectsConflictingAdditionalTool(t *testing.T) {
	reqBody := map[string]any{
		"tools": []any{map[string]any{
			"type":  "namespace",
			"name":  "collaboration",
			"tools": []any{map[string]any{"type": "function", "name": "spawn_agent"}},
		}},
		"input": []any{map[string]any{
			"type": "additional_tools",
			"tools": []any{map[string]any{
				"type":  "namespace",
				"name":  "collaboration",
				"tools": []any{map[string]any{"type": "function", "name": "send_message"}},
			}},
		}},
	}

	changed, err := normalizeOpenAIResponsesLiteTools(reqBody)

	require.ErrorContains(t, err, `conflicts with migrated tool type "namespace" name "collaboration"`)
	require.False(t, changed)
	require.Len(t, reqBody["tools"], 1, "conflicts must not partially remove top-level tools")
}

func TestNormalizeOpenAIResponsesLiteTools_DeduplicatesAcrossAdditionalToolItems(t *testing.T) {
	namespace := map[string]any{
		"type":  "namespace",
		"name":  "collaboration",
		"tools": []any{map[string]any{"type": "function", "name": "spawn_agent"}},
	}
	reqBody := map[string]any{
		"tools": []any{namespace},
		"input": []any{
			map[string]any{
				"type":  "additional_tools",
				"tools": []any{map[string]any{"type": "custom", "name": "exec"}},
			},
			map[string]any{
				"type":  "additional_tools",
				"tools": []any{namespace},
			},
		},
	}

	changed, err := normalizeOpenAIResponsesLiteTools(reqBody)

	require.NoError(t, err)
	require.True(t, changed)
	require.NotContains(t, reqBody, "tools")
	input := reqBody["input"].([]any)
	require.Len(t, input[0].(map[string]any)["tools"], 1)
	require.Len(t, input[1].(map[string]any)["tools"], 1)
}

func TestNormalizeOpenAIResponsesLiteTools_ConvertsStringInput(t *testing.T) {
	reqBody := map[string]any{
		"input": "hello",
		"tools": []any{map[string]any{
			"type": "namespace",
			"name": "collaboration",
		}},
	}

	changed, err := normalizeOpenAIResponsesLiteTools(reqBody)

	require.NoError(t, err)
	require.True(t, changed)
	require.NotContains(t, reqBody, "tools")
	input := reqBody["input"].([]any)
	require.Len(t, input, 2)
	require.Equal(t, "message", input[0].(map[string]any)["type"])
	require.Equal(t, "hello", input[0].(map[string]any)["content"])
	require.Equal(t, "additional_tools", input[1].(map[string]any)["type"])
}

func TestNormalizeOpenAIResponsesLiteTools_KeepsSupportedTopLevelTools(t *testing.T) {
	reqBody := map[string]any{
		"parallel_tool_calls": false,
		"reasoning":           map[string]any{"context": "all_turns"},
		"tools": []any{
			map[string]any{"type": "function", "name": "shell"},
			map[string]any{"type": "custom", "name": "exec"},
			map[string]any{"type": "tool_search", "execution": "client"},
			"custom shorthand",
		},
	}

	changed, err := normalizeOpenAIResponsesLiteTools(reqBody)

	require.NoError(t, err)
	require.False(t, changed)
	require.Len(t, reqBody["tools"], 4)
}

func TestNormalizeOpenAIResponsesLiteTools_DisablesParallelToolCalls(t *testing.T) {
	for _, tools := range []string{"", `,"tools":[]`, `,"tools":[{"type":"function","name":"shell"}]`, `,"tools":[{"type":"namespace","name":"helpers"}]`} {
		for _, parallel := range []string{"", `,"parallel_tool_calls":true`, `,"parallel_tool_calls":false`, `,"parallel_tool_calls":null`, `,"parallel_tool_calls":"false"`} {
			body := []byte(`{"type":"response.create","input":"hello","reasoning":{"context":"all_turns","effort":"high"}` + tools + parallel + `}`)
			out, _, err := normalizeOpenAIResponsesLiteToolsPayload(body)
			require.NoError(t, err)
			require.Equal(t, gjson.False, gjson.GetBytes(out, "parallel_tool_calls").Type)
			require.Equal(t, "all_turns", gjson.GetBytes(out, "reasoning.context").String())
			require.Equal(t, "high", gjson.GetBytes(out, "reasoning.effort").String())
			require.Equal(t, "response.create", gjson.GetBytes(out, "type").String())
			replayed, changed, err := normalizeOpenAIResponsesLiteToolsPayload(out)
			require.NoError(t, err)
			require.False(t, changed)
			require.Equal(t, out, replayed)
		}
	}
}

func TestNormalizeOpenAIResponsesLiteTools_EnsuresReasoningContext(t *testing.T) {
	tests := []struct {
		name      string
		reasoning any
	}{
		{name: "missing"},
		{name: "missing context", reasoning: map[string]any{"effort": "high"}},
		{name: "wrong context", reasoning: map[string]any{"effort": "medium", "context": "current_turn"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := map[string]any{"input": "hello"}
			if tt.reasoning != nil {
				reqBody["reasoning"] = tt.reasoning
			}

			changed, err := normalizeOpenAIResponsesLiteTools(reqBody)

			require.NoError(t, err)
			require.True(t, changed)
			reasoning := reqBody["reasoning"].(map[string]any)
			require.Equal(t, "all_turns", reasoning["context"])
			if tt.name != "missing" {
				require.Equal(t, tt.reasoning.(map[string]any)["effort"], reasoning["effort"])
			}
		})
	}
}

func TestNormalizeOpenAIResponsesLiteTools_RejectsNonObjectReasoning(t *testing.T) {
	reqBody := map[string]any{"reasoning": "high"}

	changed, err := normalizeOpenAIResponsesLiteTools(reqBody)

	require.ErrorContains(t, err, "reasoning to be an object")
	require.False(t, changed)
	require.Equal(t, "high", reqBody["reasoning"])
}

func TestNormalizeOpenAIResponsesLiteTools_RejectsUnsupportedTools(t *testing.T) {
	tests := []struct {
		name string
		tool map[string]any
		want string
	}{
		{name: "hosted web search", tool: map[string]any{"type": "web_search"}, want: `top-level tool type "web_search"`},
		{name: "hosted image generation", tool: map[string]any{"type": "image_generation"}, want: `top-level tool type "image_generation"`},
		{name: "hosted tool search", tool: map[string]any{"type": "tool_search", "execution": "server"}, want: "execution=client"},
		{name: "default hosted tool search", tool: map[string]any{"type": "tool_search"}, want: "execution=client"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := map[string]any{"tools": []any{tt.tool}}
			changed, err := normalizeOpenAIResponsesLiteTools(reqBody)
			require.ErrorContains(t, err, tt.want)
			require.False(t, changed)
			require.Len(t, reqBody["tools"], 1, "validation errors must not partially mutate tools")
		})
	}
}

func TestNormalizeOpenAIResponsesLiteToolsPayload_PreservesResponseCreateShape(t *testing.T) {
	body := []byte(`{
		"type":"response.create",
		"model":"gpt-5.6-terra",
		"client_metadata":{"ws_request_header_x_openai_internal_codex_responses_lite":"true"},
		"input":[{"type":"message","role":"user","content":"hello"}],
		"tools":[{"type":"namespace","name":"collaboration","tools":[{"type":"function","name":"spawn_agent"}]}],
		"tool_choice":{"type":"namespace","name":"collaboration"}
	}`)

	updated, changed, err := normalizeOpenAIResponsesLiteToolsPayload(body)

	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, "response.create", gjson.GetBytes(updated, "type").String())
	require.False(t, gjson.GetBytes(updated, "tools").Exists())
	require.Equal(t, "collaboration", gjson.GetBytes(updated, `input.#(type=="additional_tools").tools.0.name`).String())
	require.Equal(t, "namespace", gjson.GetBytes(updated, "tool_choice.type").String())
}

func TestNormalizeOpenAIResponsesLiteToolsPayload_RejectsTrailingJSON(t *testing.T) {
	body := []byte(`{"model":"gpt-6.2-sol","tools":[]} {"tools":[{"type":"namespace","name":"functions"}]}`)
	out, changed, err := normalizeOpenAIResponsesLiteToolsPayload(body)
	require.ErrorContains(t, err, "invalid JSON")
	require.False(t, changed)
	require.Equal(t, body, out)
}

func TestApplyCodexOAuthTransform_PreservesLiteNamespaceToolChoice(t *testing.T) {
	reqBody := map[string]any{
		"model": "gpt-5.6-terra",
		"input": []any{map[string]any{
			"type": "additional_tools",
			"tools": []any{map[string]any{
				"type": "namespace",
				"name": "collaboration",
			}},
		}},
		"tool_choice": map[string]any{"type": "namespace", "name": "collaboration"},
	}

	applyCodexOAuthTransform(reqBody, true, false)

	require.Equal(t, map[string]any{"type": "namespace", "name": "collaboration"}, reqBody["tool_choice"])
}

func TestOpenAIGatewayServiceForward_NormalizesResponsesLiteToolsForOAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, passthrough := range []bool{false, true} {
		name := "managed"
		if passthrough {
			name = "passthrough"
		}
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(nil))
			c.Request.Header.Set("User-Agent", "codex_cli_rs/0.144.1")
			c.Request.Header.Set(responsesLiteHeader, "true")
			upstream := &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
				Body: io.NopCloser(strings.NewReader(
					"data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_lite\",\"usage\":{\"input_tokens\":1,\"output_tokens\":1}}}\n\n" +
						"data: [DONE]\n\n",
				)),
			}}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
			account := &Account{
				ID: "501", Name: "responses-lite", Platform: PlatformOpenAI, Type: AccountTypeOAuth,
				Concurrency: 1, Status: StatusActive, Schedulable: true, RateMultiplier: f64p(1),
				Credentials: map[string]any{"access_token": "oauth-token", "chatgpt_account_id": "chatgpt-account"},
				Extra:       map[string]any{"openai_passthrough": passthrough},
			}
			body := []byte(`{
				"model":"gpt-5.6-terra","stream":true,"instructions":"test",
				"parallel_tool_calls":true,
				"reasoning":{"effort":"high","context":"current_turn"},
				"tools":[
					{"type":"function","name":"shell","parameters":{"type":"object"}},
					{"type":"custom","name":"exec"},
					{"type":"tool_search","execution":"client"},
					{"type":"namespace","name":"collaboration","tools":[{"type":"function","name":"spawn_agent","parameters":{"type":"object"}}]}
				],
				"input":[{"type":"message","role":"user","content":"hello"}],
				"tool_choice":{"type":"namespace","name":"collaboration"}
			}`)

			result, err := svc.Forward(context.Background(), c, account, body)

			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, "true", upstream.lastReq.Header.Get(responsesLiteHeader))
			require.Equal(t, "high", gjson.GetBytes(upstream.lastBody, "reasoning.effort").String())
			require.Equal(t, "all_turns", gjson.GetBytes(upstream.lastBody, "reasoning.context").String())
			require.Equal(t, gjson.False, gjson.GetBytes(upstream.lastBody, "parallel_tool_calls").Type)
			require.False(t, gjson.GetBytes(upstream.lastBody, `tools.#(type=="namespace")`).Exists())
			require.Equal(t, "shell", gjson.GetBytes(upstream.lastBody, `tools.#(type=="function").name`).String())
			require.Equal(t, "exec", gjson.GetBytes(upstream.lastBody, `tools.#(type=="custom").name`).String())
			require.True(t, gjson.GetBytes(upstream.lastBody, `tools.#(type=="tool_search")`).Exists())
			require.Equal(t, "collaboration", gjson.GetBytes(upstream.lastBody, `input.#(type=="additional_tools").tools.0.name`).String())
			require.Equal(t, "namespace", gjson.GetBytes(upstream.lastBody, "tool_choice.type").String())
			require.Equal(t, "collaboration", gjson.GetBytes(upstream.lastBody, "tool_choice.name").String())
		})
	}
}

func TestResponsesLiteSimulatedHeaderMatchesFinalBody(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
		account := &Account{ID: "1", Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"chatgpt_account_id": "account-1", "access_token": "token"}}
		body := []byte(`{"model":"gpt-5.6-terra","reasoning":{"effort":"high","context":"all_turns"},"parallel_tool_calls":true,"input":"test"}`)
		svc := &OpenAIGatewayService{cfg: &config.Config{}}
		var req *http.Request
		var err error
		if passthrough {
			req, err = svc.buildUpstreamRequestOpenAIPassthrough(context.Background(), c, account, body, "token")
		} else {
			req, err = svc.buildUpstreamRequest(context.Background(), c, account, body, "token", false, "", true)
		}
		require.NoError(t, err)
		require.Equal(t, "true", req.Header.Get(responsesLiteHeader))
		out, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		require.Equal(t, "all_turns", gjson.GetBytes(out, "reasoning.context").String())
		require.Equal(t, gjson.False, gjson.GetBytes(out, "parallel_tool_calls").Type)
		require.Equal(t, "high", gjson.GetBytes(out, "reasoning.effort").String())
		replay, err := req.GetBody()
		require.NoError(t, err)
		replayed, err := io.ReadAll(replay)
		require.NoError(t, err)
		require.Equal(t, out, replayed)
		require.EqualValues(t, len(replayed), req.ContentLength)
	}
}

func TestResponsesLiteSimulatedHeaderNormalizesNamespaceTools(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, passthrough := range []bool{false, true} {
		for _, clientLite := range []bool{false, true} {
			t.Run(fmt.Sprintf("passthrough=%t/client_lite=%t", passthrough, clientLite), func(t *testing.T) {
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
				if clientLite {
					c.Request.Header.Set(responsesLiteHeader, "true")
				}
				account := &Account{ID: "1", Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"chatgpt_account_id": "account-1", "access_token": "token"}}
				body := []byte(`{
					"model":"gpt-6.2-sol","input":"test","parallel_tool_calls":false,
					"reasoning":{"effort":"high","context":"all_turns"},
					"tools":[
						{"type":"function","name":"shell","parameters":{"type":"object","properties":{"id":{"type":"integer","maximum":9007199254740993}}}},
						{"type":"custom","name":"exec"},
						{"type":"tool_search","execution":"client"},
						{"type":"namespace","name":"functions","tools":[{"type":"function","name":"lookup","parameters":{"type":"object"}}]}
					],
					"tool_choice":{"type":"namespace","name":"functions"}
				}`)
				svc := &OpenAIGatewayService{cfg: &config.Config{}}
				var req *http.Request
				var err error
				if passthrough {
					req, err = svc.buildUpstreamRequestOpenAIPassthrough(context.Background(), c, account, body, "token")
				} else {
					req, err = svc.buildUpstreamRequest(context.Background(), c, account, body, "token", false, "", true)
				}
				require.NoError(t, err)
				require.Equal(t, "true", req.Header.Get(responsesLiteHeader))
				out, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				require.False(t, gjson.GetBytes(out, `tools.#(type=="namespace")`).Exists())
				require.Equal(t, "lookup", gjson.GetBytes(out, `input.#(type=="additional_tools").tools.0.tools.0.name`).String())
				require.Equal(t, "functions", gjson.GetBytes(out, "tool_choice.name").String())
				require.Equal(t, "client", gjson.GetBytes(out, `tools.#(type=="tool_search").execution`).String())
				require.Equal(t, "9007199254740993", gjson.GetBytes(out, `tools.#(name=="shell").parameters.properties.id.maximum`).Raw)
				replay, err := req.GetBody()
				require.NoError(t, err)
				defer replay.Close()
				replayed, err := io.ReadAll(replay)
				require.NoError(t, err)
				require.Equal(t, out, replayed)
				require.EqualValues(t, len(out), req.ContentLength)
			})
		}
	}
}

func TestResponsesLiteFinalBodyPreservesHeaders(t *testing.T) {
	for _, lite := range []string{"", "false", "true"} {
		t.Run("lite="+lite, func(t *testing.T) {
			body := `{"model":"gpt-6.2-sol","input":"test","tools":[{"type":"namespace","name":"functions"}]}`
			req, err := http.NewRequest(http.MethodPost, "https://example.com/responses", strings.NewReader(body))
			require.NoError(t, err)
			req.Header.Set("User-Agent", "codex_cli_rs/0.169.0")
			req.Header.Set("Authorization", "Bearer test")
			if lite != "" {
				req.Header.Set(responsesLiteHeader, lite)
			}
			headers := req.Header.Clone()
			require.NoError(t, finalizeOpenAIResponsesLiteRequest(req))
			require.Equal(t, headers, req.Header)
			out, err := io.ReadAll(req.Body)
			require.NoError(t, err)
			if lite == "true" {
				require.False(t, gjson.GetBytes(out, `tools.#(type=="namespace")`).Exists())
			} else {
				require.Equal(t, body, string(out))
			}
		})
	}
}

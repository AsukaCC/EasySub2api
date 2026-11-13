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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func forwardingTestAccount() *Account {
	return &Account{
		ID: "forwarding-account", Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
		Status: StatusActive, Schedulable: true,
		Credentials: map[string]any{
			"api_key": "test-key", "base_url": "http://upstream.example",
			"model_mapping":    map[string]any{"gpt-5.6-sol": "gpt-5.6-sol"},
			"model_forwarding": map[string]any{"gpt-5.6-luna": "gpt-5.6-terra"},
		},
	}
}

func TestAccountModelForwarding_AdmissionAndIsolation(t *testing.T) {
	a := forwardingTestAccount()
	require.True(t, a.IsModelSupported("gpt-5.6-luna"))
	require.True(t, a.IsModelSupported("gpt-5.6-sol"))
	require.False(t, a.IsModelSupported("gpt-6-astra"))
	require.Equal(t, "gpt-5.6-terra", a.GetMappedModel("gpt-5.6-luna"))
	other := forwardingTestAccount()
	other.ID = "other-account"
	delete(other.Credentials, "model_forwarding")
	require.False(t, other.IsModelSupported("gpt-5.6-luna"))
	svc := &OpenAIGatewayService{accountRepo: stubOpenAIAccountRepo{accounts: []Account{*other, *a}}}
	selected, err := svc.SelectAccountForModelWithExclusions(context.Background(), nil, "", "gpt-5.6-luna", nil)
	require.NoError(t, err)
	require.Equal(t, a.ID, selected.ID)
}

func TestAccountModelForwarding_PrecedenceAndSingleHop(t *testing.T) {
	a := forwardingTestAccount()
	a.Credentials["model_forwarding"] = map[string]any{
		"gpt-*": "gpt-6-sol", "gpt-5.6-*": "gpt-5.6-sol",
		"gpt-5.6-luna": "gpt-5.6-terra", "gpt-5.6-terra": "gpt-6-astra",
	}
	a.Credentials["model_mapping"] = map[string]any{"gpt-5.6-luna": "gpt-6-luna"}
	require.Equal(t, "gpt-5.6-terra", a.GetMappedModel("gpt-5.6-luna"))
	require.Equal(t, "gpt-5.6-sol", a.GetMappedModel("gpt-5.6-custom"))
	require.Equal(t, "gpt-6-sol", a.GetMappedModel("gpt-6-custom"))
	a.Credentials["model_forwarding"] = map[string]any{"gpt-5.6-luna": "gpt-5.4"}
	require.False(t, a.IsModelSupported("gpt-5.6-luna"))
}

func TestAccountModelForwarding_DynamicRules(t *testing.T) {
	a := forwardingTestAccount()
	require.Equal(t, "gpt-5.6-terra", a.GetMappedModel("gpt-5.6-luna"))
	a.Credentials["model_forwarding"] = map[string]any{"gpt-5.6-luna": "gpt-6.1-sol", "custom-client-model": "gpt-6-luna"}
	require.Equal(t, "gpt-6.1-sol", a.GetMappedModel("gpt-5.6-luna"))
	require.True(t, a.IsModelSupported("custom-client-model"))
	require.Equal(t, "gpt-6-luna", a.GetMappedModel("custom-client-model"))
	require.Contains(t, openAIConfiguredCodexModelIDs([]Account{*a}), "custom-client-model")
	a.Credentials["model_forwarding"] = map[string]any{}
	require.False(t, a.IsModelSupported("gpt-5.6-luna"))
	require.False(t, a.IsModelSupported("custom-client-model"))
	require.True(t, a.IsModelSupported("gpt-5.6-sol"))
}

func TestAccountModelForwarding_HTTPAndWebSocket(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, passthrough := range []bool{false, true} {
		t.Run(map[bool]string{false: "normal", true: "passthrough"}[passthrough], func(t *testing.T) {
			a := forwardingTestAccount()
			a.Extra = map[string]any{"openai_passthrough": passthrough}
			body := []byte(`{"model":"gpt-5.6-luna","input":"hello","stream":false}`)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
			upstream := &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}},
				Body: io.NopCloser(strings.NewReader(`{"id":"resp-forward","model":"gpt-5.6-terra","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"ok"}]}],"usage":{"input_tokens":10,"output_tokens":2}}`)),
			}}
			svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}
			result, err := svc.Forward(context.Background(), c, a, body)
			require.NoError(t, err)
			require.Equal(t, "gpt-5.6-terra", gjson.GetBytes(upstream.lastBody, "model").String())
			require.Equal(t, "gpt-5.6-terra", result.UpstreamModel)
			require.Equal(t, "gpt-5.6-terra", resolveOpenAIAccountUpstreamModelForRequest(a, "gpt-5.6-luna", false))
			require.Equal(t, "gpt-5.6-terra", openAIWSPassthroughPolicyModelForFrame(a, body))
		})
	}
}

func TestAccountModelForwarding_BillsFinalModelPoints(t *testing.T) {
	for _, source := range []string{BillingModelSourceRequested, BillingModelSourceChannelMapped, BillingModelSourceUpstream, BillingModelSourceResponse} {
		t.Run(source, func(t *testing.T) {
			usageRepo := &openAIRecordUsageLogRepoStub{inserted: true}
			userRepo := &openAIRecordUsageUserRepoStub{}
			svc := newOpenAIRecordUsageServiceForTest(usageRepo, userRepo, &openAIRecordUsageSubRepoStub{}, nil)
			svc.resolver = NewModelPricingResolver(nil, svc.billingService)
			cheap, expensive := 1.0, 7.0
			group := &Group{ID: "group", Platform: PlatformOpenAI, RateMultiplier: 1, ModelPricing: []ChannelModelPricing{
				{Platform: PlatformOpenAI, Models: []string{"gpt-5.6-luna"}, BillingMode: BillingModePerRequest, PerRequestPrice: &cheap},
				{Platform: PlatformOpenAI, Models: []string{"gpt-5.6-terra"}, BillingMode: BillingModePerRequest, PerRequestPrice: &expensive},
			}}
			err := svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
				Result: &OpenAIForwardResult{
					RequestID: "forwarded-" + source, Model: "gpt-5.6-luna", UpstreamModel: "gpt-5.6-terra",
					UpstreamResponseModel: "gpt-5.6-luna", Usage: OpenAIUsage{InputTokens: 10, OutputTokens: 2}, Duration: time.Second,
				},
				APIKey: &APIKey{ID: "key", GroupID: &group.ID, Group: group}, User: &User{ID: "user"}, Account: forwardingTestAccount(),
				ChannelUsageFields: ChannelUsageFields{OriginalModel: "gpt-5.6-luna", ChannelMappedModel: "gpt-5.6-luna", BillingModelSource: source},
			})
			require.NoError(t, err)
			require.NotNil(t, usageRepo.lastLog)
			require.InDelta(t, 7, usageRepo.lastLog.ActualCost, 1e-9)
			require.InDelta(t, 7, userRepo.lastAmount, 1e-9)
			require.Equal(t, "gpt-5.6-luna", usageRepo.lastLog.RequestedModel)
			require.Equal(t, "gpt-5.6-terra", *usageRepo.lastLog.UpstreamModel)
		})
	}
}

func TestAccountModelForwarding_Validation(t *testing.T) {
	for _, raw := range []any{"invalid", map[string]any{"a": 1}, map[string]any{"a*b": "gpt-5.6-terra"}, map[string]any{"a": "gpt-*"}, map[string]any{"a": ""}, map[string]any{"a": "gpt-5.4"}} {
		require.Error(t, NormalizeOpenAIModelForwardingCredentials(PlatformOpenAI, map[string]any{"model_forwarding": raw}))
	}
	credentials := map[string]any{"model_forwarding": map[string]any{" gpt-5.6-luna ": " gpt-5.6-terra "}}
	require.NoError(t, NormalizeOpenAIModelForwardingCredentials(PlatformOpenAI, credentials))
	require.Equal(t, map[string]any{"gpt-5.6-luna": "gpt-5.6-terra"}, credentials["model_forwarding"])
	require.NoError(t, NormalizeOpenAIModelForwardingCredentials(PlatformOpenAI, map[string]any{"model_forwarding": map[string]any{}}))
}

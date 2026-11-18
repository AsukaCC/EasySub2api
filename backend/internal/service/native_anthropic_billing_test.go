//go:build unit

package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/apicompat"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestNativeAnthropicCacheBilling(t *testing.T) {
	const model = "gpt-6-luna"
	const groupID = "00000000-0000-0000-0000-000000000777"
	for _, route := range []string{"chat-buffered", "chat-stream", "responses-buffered", "responses-stream", "messages-json", "messages-stream"} {
		t.Run(route, func(t *testing.T) {
			logs := &openAIRecordUsageLogRepoStub{inserted: true}
			users := &openAIRecordUsageUserRepoStub{}
			s := newOpenAIRecordUsageServiceForTest(logs, users, &openAIRecordUsageSubRepoStub{}, nil)
			group := &Group{ID: groupID, Platform: PlatformOpenAI, RateMultiplier: 0.5}
			account := &Account{ID: "00000000-0000-0000-0000-000000000001", Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
			pricing := ChannelModelPricing{Platform: PlatformOpenAI, Models: []string{model}, BillingMode: BillingModeToken,
				InputPrice: float64Ptr(1e-6), OutputPrice: float64Ptr(2e-6), CacheReadPrice: float64Ptr(0.5e-6),
				CacheWritePrice: float64Ptr(3e-6), CacheWrite1hPrice: float64Ptr(7e-6)}
			ch := Channel{ID: account.ID, Status: StatusActive, GroupIDs: []string{groupID}, ModelPricing: []ChannelModelPricing{pricing}, ApplyPricingToAccountStats: true}
			s.channelService = &ChannelService{}
			s.channelService.cache.Store(populateChannelCache([]Channel{ch}, map[string]string{groupID: PlatformOpenAI}))
			s.resolver = NewModelPricingResolver(s.channelService, s.billingService)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
			body := strings.Replace(miniAnthropicSSEStream(), `"input_tokens":10`, `"input_tokens":10,"cache_read_input_tokens":20,"cache_creation_input_tokens":100,"cache_creation":{"ephemeral_5m_input_tokens":5}`, 1)
			body = strings.Replace(body, `"output_tokens":5`, `"output_tokens":5,"cache_creation":{"ephemeral_1h_input_tokens":40}`, 1)
			if route == "messages-json" {
				body = `{"usage":{"input_tokens":10,"output_tokens":5,"cache_read_input_tokens":20,"cache_creation_input_tokens":100,"cache_creation":{"ephemeral_5m_input_tokens":5,"ephemeral_1h_input_tokens":40}}}`
			}
			resp := &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}
			var result *OpenAIForwardResult
			var err error
			switch route {
			case "chat-buffered":
				result, err = s.handleCCBufferedFromNativeAnthropic(resp, c, model, model, model, nil, time.Now())
			case "chat-stream":
				result, err = s.handleCCStreamingFromNativeAnthropic(resp, c, model, model, model, nil, time.Now(), true)
			case "responses-buffered":
				result, err = s.handleResponsesBufferedFromNativeAnthropic(resp, c, model, model, model, nil, time.Now(), apicompat.ResponsesClientToolMapping{})
			case "responses-stream":
				result, err = s.handleResponsesStreamingFromNativeAnthropic(resp, c, model, model, model, nil, time.Now(), apicompat.ResponsesClientToolMapping{})
			case "messages-json":
				result, err = s.handleNativeAnthropicBufferedResponse(context.Background(), resp, c, account, model, model, model, time.Now())
			case "messages-stream":
				result, err = s.handleNativeAnthropicStreamingResponse(context.Background(), resp, c, account, model, model, model, time.Now())
			}
			require.NoError(t, err)
			require.Equal(t, 130, result.Usage.InputTokens)
			err = s.RecordUsage(context.Background(), &OpenAIRecordUsageInput{Result: result, Account: account,
				User: &User{ID: "00000000-0000-0000-0000-000000000002"}, APIKey: &APIKey{ID: account.ID, GroupID: &group.ID, Group: group},
				ChannelUsageFields: ChannelUsageFields{OriginalModel: model, BillingModelSource: BillingModelSourceRequested}})
			require.NoError(t, err)
			log := logs.lastLog
			require.NotNil(t, log)
			require.Equal(t, 10, log.InputTokens)
			require.Equal(t, 100, log.CacheCreationTokens)
			require.Equal(t, 5, log.CacheCreation5mTokens) // observed, not the inferred billable count
			require.Equal(t, 40, log.CacheCreation1hTokens)
			require.Zero(t, log.CacheCreation12hTokens)
			const writes = (60*3 + 40*7) * 1e-6
			const total = writes + (10+5*2+20*0.5)*1e-6
			require.InDelta(t, writes, log.CacheCreationCost, 1e-12)
			require.NotNil(t, log.AccountStatsCost)
			require.InDelta(t, total, *log.AccountStatsCost, 1e-12)
			require.InDelta(t, total*0.5, users.lastAmount, 1e-12)
		})
	}
}

func TestNativeAnthropicUsageWithoutWriteTotal(t *testing.T) {
	var usage ClaudeUsage
	parseSSEUsagePassthrough(`{"type":"message_start","message":{"usage":{"input_tokens":10,"output_tokens":4,"cache_creation":{"ephemeral_5m_input_tokens":3}}}}`, &usage)
	parseSSEUsagePassthrough(`{"type":"message_delta","usage":{"cache_creation":{"ephemeral_1h_input_tokens":6}}}`, &usage)
	require.Equal(t, 9, usage.CacheCreationInputTokens)
	require.Equal(t, 19, usage.toOpenAIUsage().InputTokens)
	require.Equal(t, 4, usage.OutputTokens)
	parseSSEUsagePassthrough(`{"type":"message_delta","usage":{"cache_creation_input_tokens":7}}`, &usage)
	parseSSEUsagePassthrough(`{"type":"message_delta","usage":{"cache_creation":{"ephemeral_5m_input_tokens":8}}}`, &usage)
	require.Equal(t, 7, usage.CacheCreationInputTokens, "an explicit total remains authoritative")
}

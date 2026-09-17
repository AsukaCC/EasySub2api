package handler

import (
	"context"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRetiredGatewayModelResponse(t *testing.T) {
	c, w := gin.CreateTestContext(httptest.NewRecorder())
	_ = w
	require.True(t, rejectRetiredGatewayModel(c, "openai/gpt5.5-pro"))
	require.Equal(t, http.StatusBadRequest, c.Writer.Status())
	require.True(t, c.IsAborted())
}

func TestRetiredModelsNotExposedByModelListWriters(t *testing.T) {
	for _, writer := range []func(*gin.Context, []string){
		writeOpenAIModelsList,
		func(c *gin.Context, ids []string) { writeModelsList(c, service.PlatformAnthropic, ids) },
	} {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		writer(c, []string{"gpt5.4-mini", "gpt-5.5", "gpt-5.6-sol"})
		require.NotContains(t, rec.Body.String(), "gpt5.4")
		require.NotContains(t, rec.Body.String(), "gpt-5.5")
		require.Contains(t, rec.Body.String(), "gpt-5.6-sol")
	}
}

func TestRetiredCompositeOriginalModelIsRejected(t *testing.T) {
	ctx := service.WithCompositeRouteDecision(context.Background(), service.CompositeRouteDecision{
		Matched: true, PublicModel: "gpt-5.5", UpstreamModel: "gpt-5.6-sol", TargetPlatform: service.PlatformOpenAI,
	})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil).WithContext(ctx)
	require.True(t, rejectRetiredGatewayModel(c, "gpt-5.6-sol"))
	require.Equal(t, http.StatusBadRequest, c.Writer.Status())
}

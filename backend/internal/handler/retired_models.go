package handler

import (
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

func rejectRetiredGatewayModel(c *gin.Context, model string, writers ...func(*gin.Context, int, string, string)) bool {
	err := service.CheckActiveModel(model)
	if c.Request != nil {
		err = service.CheckActiveRequestModel(c.Request.Context(), model)
	}
	if err == nil {
		return false
	}
	service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonLocalPolicyDenied)
	if len(writers) > 0 {
		c.Abort()
		writers[0](c, http.StatusBadRequest, "invalid_request_error", "GPT-5.4 and GPT-5.5 model families have been retired")
		return true
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": gin.H{
		"type": "invalid_request_error", "code": "model_retired",
		"message": "GPT-5.4 and GPT-5.5 model families have been retired",
	}})
	return true
}

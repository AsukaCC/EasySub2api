package service

import (
	"github.com/gin-gonic/gin"
	"strings"
)

func gatewaySSEStarted(c *gin.Context) bool {
	return c.Writer.Written() && strings.HasPrefix(strings.ToLower(c.Writer.Header().Get("Content-Type")), "text/event-stream")
}

// A concurrency or user-message queue may have already flushed a keepalive.
// In that case the HTTP status is fixed, and a JSON body would be lost by SSE clients.
func writeGatewayErrorResponse(c *gin.Context, status int, errType, message string) {
	payload := gin.H{
		"type": "error",
		"error": gin.H{
			"type":    errType,
			"message": message,
		},
	}
	if gatewaySSEStarted(c) {
		MarkOpsStreamError(c, errType, message, status)
		c.SSEvent("error", payload)
		c.Writer.Flush()
		return
	}
	c.JSON(status, payload)
}

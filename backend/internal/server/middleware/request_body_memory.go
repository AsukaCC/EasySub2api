package middleware

import (
	"errors"
	"net/http"
	"strings"

	pkghttputil "github.com/AsukaCC/EasySub2api/internal/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// RejectRequestBodyReadError handles early ingress termination before the route
// handler runs. Cancelled clients have no response body and an uncommitted
// response is marked 499, matching the gateway handler's cancellation path.
// A read deadline on HTTP/1 also cancels the request context; timeouts are left
// to AbortWithRequestBodyReadFailure (408).
func RejectRequestBodyReadError(c *gin.Context, component string, err error) bool {
	if err != nil && c != nil && c.Request != nil && c.Request.Context().Err() != nil && !pkghttputil.IsRequestBodyReadTimeout(err) {
		logRequestBodyReadFailure(c, component, pkghttputil.DescribeRequestBodyReadFailure(c.Request, err))
		if !c.Writer.Written() {
			c.Status(499)
		}
		c.Abort()
		return true
	}
	return RejectRequestBodyMemoryOverload(c, err)
}

// RejectRequestBodyMemoryOverload keeps temporary capacity errors distinct
// from invalid JSON and the per-request 413 size limit.
func RejectRequestBodyMemoryOverload(c *gin.Context, err error) bool {
	if !errors.Is(err, pkghttputil.ErrRequestBodyMemoryBudget) {
		return false
	}
	c.Header("Retry-After", "1")
	requestBodyErrorWriter(c)(c, http.StatusServiceUnavailable, "Request body buffering capacity is exhausted; retry later")
	c.Abort()
	return true
}

func requestBodyErrorWriter(c *gin.Context) GatewayErrorWriter {
	path := ""
	if c.Request != nil && c.Request.URL != nil {
		path = c.Request.URL.Path
	}
	switch {
	case strings.HasPrefix(path, "/v1beta") || strings.HasPrefix(path, "/antigravity/v1beta"):
		return GoogleErrorWriter
	case strings.Contains(path, "/messages"):
		return AnthropicErrorWriter
	default:
		return func(c *gin.Context, status int, message string) {
			c.AbortWithStatusJSON(status, gin.H{"error": gin.H{"type": "server_error", "message": message}})
		}
	}
}

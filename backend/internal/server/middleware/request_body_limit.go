package middleware

import (
	pkghttputil "github.com/AsukaCC/EasySub2api/internal/pkg/httputil"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequestBodyLimit 使用 MaxBytesReader 限制请求体大小。
func RequestBodyLimit(maxBytes int64, budgets ...*pkghttputil.BodyMemoryBudget) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(budgets) > 0 {
			ctx, release := pkghttputil.WithRequestBodyMemoryBudget(c.Request.Context(), budgets[0])
			defer release()
			c.Request = c.Request.WithContext(ctx)
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

// TextRequestBodyLimit also caps the decoded representation, without installing
// a second memory budget or copying a body already read by composite routing.
func TextRequestBodyLimit(maxBytes int64, budget *pkghttputil.BodyMemoryBudget) gin.HandlerFunc {
	wireLimit := RequestBodyLimit(maxBytes, budget)
	return func(c *gin.Context) {
		c.Request = c.Request.WithContext(pkghttputil.WithDecodedBodyLimit(c.Request.Context(), maxBytes))
		wireLimit(c)
	}
}

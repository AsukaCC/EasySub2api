package service

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func prepareCodexHTTPIdentityBody(c *gin.Context, account *Account, body []byte) ([]byte, error) {
	if account == nil || !account.IsOpenAIOAuthLike() || isOpenAIResponsesCompactPath(c) {
		return body, nil
	}
	ids := stagedCodexFingerprintIDs(c, account)
	if ids == nil {
		var headers http.Header
		if c != nil && c.Request != nil {
			headers = c.Request.Header
		}
		ids = resolveCodexFingerprintIDsFromRequest(codexAccountIdentitySource(c, account), headers)
		stageCodexFingerprintIDs(c, ids)
	}
	next, _, err := applyCodexFingerprintClientMetadataRaw(body, ids)
	return next, err
}

package service

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func protectOpenAIWSIdentity(c *gin.Context, account *Account, body []byte, reuseStaged ...bool) ([]byte, error) {
	var headers http.Header
	if c != nil && c.Request != nil {
		headers = c.Request.Header
	}
	var ids *codexFingerprintIDs
	if len(reuseStaged) > 0 && reuseStaged[0] {
		ids = stagedCodexFingerprintIDs(c, account)
	}
	if ids == nil {
		ids = resolveCodexFingerprintIDsFromRequest(codexAccountIdentitySource(c, account), headers)
	}
	stageCodexFingerprintIDs(c, ids)
	next, _, err := applyCodexFingerprintClientMetadataRaw(body, ids)
	if err != nil {
		return nil, err
	}
	next, _, err = applyCodexAccountIdentityClientMetadataRaw(next, codexAccountIdentitySource(c, account), "")
	return next, err
}

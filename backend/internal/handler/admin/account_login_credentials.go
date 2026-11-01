package admin

import (
	"github.com/AsukaCC/EasySub2api/internal/pkg/response"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// GetLoginCredentials is an admin-only, audited read, separate from account lists.
func (h *AccountHandler) GetLoginCredentials(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	accountID, err := parseEntityID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	account, err := h.adminService.GetAccount(c.Request.Context(), accountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	result := make(map[string]string, 3)
	for _, key := range []string{service.AccountLoginUsernameKey, service.AccountLoginPasswordKey, service.AccountLoginTOTPKey} {
		result[key] = account.GetCredential(key)
	}
	response.Success(c, result)
}

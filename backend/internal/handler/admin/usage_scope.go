package admin

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	usageRoleScopeAll     = "all"
	usageRoleScopeRegular = "regular"
	usageRoleScopeAdmin   = "admin"
)

func parseUsageRoleScope(c *gin.Context, fallback string, forced string) (string, error) {
	if forced != "" {
		return forced, nil
	}
	scope := strings.ToLower(strings.TrimSpace(c.Query("scope")))
	if scope == "" {
		scope = fallback
	}
	switch scope {
	case usageRoleScopeAll, usageRoleScopeRegular, usageRoleScopeAdmin:
		return scope, nil
	default:
		return "", errors.New("invalid scope, use all, regular, or admin")
	}
}

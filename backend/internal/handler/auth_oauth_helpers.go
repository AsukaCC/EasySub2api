package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	"github.com/AsukaCC/EasySub2api/internal/pkg/response"
	servermiddleware "github.com/AsukaCC/EasySub2api/internal/server/middleware"
	"github.com/AsukaCC/EasySub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

const (
	oauthBindAccessTokenCookiePath = "/api/v1/auth/oauth"
	oauthBindAccessTokenCookieName = "oauth_bind_access_token"
	oauthCookieMaxAgeSec           = 10 * 60
	oauthDefaultRedirectTo         = "/dashboard"
	oauthMaxRedirectLen            = 2048
	oauthMaxFragmentValueLen       = 512

	oauthIntentLogin           = "login"
	oauthIntentBindCurrentUser = "bind_current_user"
)

func redirectOAuthError(c *gin.Context, frontendCallback string, code string, message string, description string) {
	fragment := url.Values{}
	fragment.Set("error", truncateFragmentValue(code))
	if strings.TrimSpace(message) != "" {
		fragment.Set("error_message", truncateFragmentValue(message))
	}
	if strings.TrimSpace(description) != "" {
		fragment.Set("error_description", truncateFragmentValue(description))
	}
	redirectWithFragment(c, frontendCallback, fragment)
}

func redirectWithFragment(c *gin.Context, frontendCallback string, fragment url.Values) {
	u, err := url.Parse(frontendCallback)
	if err != nil {
		c.Redirect(http.StatusFound, oauthDefaultRedirectTo)
		return
	}
	if u.Scheme != "" && !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		c.Redirect(http.StatusFound, oauthDefaultRedirectTo)
		return
	}
	u.Fragment = fragment.Encode()
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.Redirect(http.StatusFound, u.String())
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func parseOAuthProviderError(body string) (providerErr string, providerDesc string) {
	body = strings.TrimSpace(body)
	if body == "" {
		return "", ""
	}

	providerErr = firstNonEmpty(
		getGJSON(body, "error"),
		getGJSON(body, "code"),
		getGJSON(body, "error.code"),
	)
	providerDesc = firstNonEmpty(
		getGJSON(body, "error_description"),
		getGJSON(body, "error.message"),
		getGJSON(body, "message"),
		getGJSON(body, "detail"),
	)
	if providerErr != "" || providerDesc != "" {
		return providerErr, providerDesc
	}

	values, err := url.ParseQuery(body)
	if err != nil {
		return "", ""
	}
	providerErr = firstNonEmpty(values.Get("error"), values.Get("code"))
	providerDesc = firstNonEmpty(values.Get("error_description"), values.Get("error_message"), values.Get("message"))
	return providerErr, providerDesc
}

func getGJSON(body string, path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	result := gjson.Get(body, path)
	if !result.Exists() {
		return ""
	}
	return result.String()
}

func truncateLogValue(value string, maxLen int) string {
	value = strings.TrimSpace(value)
	if value == "" || maxLen <= 0 {
		return ""
	}
	if len(value) <= maxLen {
		return value
	}
	value = value[:maxLen]
	for !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return value
}

func singleLine(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.Join(strings.Fields(value), " ")
}

func sanitizeFrontendRedirectPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || len(path) > oauthMaxRedirectLen {
		return ""
	}
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return ""
	}
	if strings.Contains(path, "://") || strings.ContainsAny(path, "\r\n") {
		return ""
	}
	return path
}

func isRequestHTTPS(c *gin.Context) bool {
	if c.Request.TLS != nil {
		return true
	}
	proto := strings.ToLower(strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")))
	return proto == "https"
}

func encodeCookieValue(value string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(value))
}

func decodeCookieValue(value string) (string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func readCookieDecoded(c *gin.Context, name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	return decodeCookieValue(cookie.Value)
}

func clearOAuthBindAccessTokenCookie(c *gin.Context, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     oauthBindAccessTokenCookieName,
		Value:    "",
		Path:     oauthBindAccessTokenCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func setOAuthBindAccessTokenCookie(c *gin.Context, token string, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     oauthBindAccessTokenCookieName,
		Value:    url.QueryEscape(strings.TrimSpace(token)),
		Path:     oauthBindAccessTokenCookiePath,
		MaxAge:   oauthCookieMaxAgeSec,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func truncateFragmentValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) > oauthMaxFragmentValueLen {
		value = value[:oauthMaxFragmentValueLen]
		for !utf8.ValidString(value) {
			value = value[:len(value)-1]
		}
	}
	return value
}

func buildBearerAuthorization(tokenType, accessToken string) (string, error) {
	tokenType = strings.TrimSpace(tokenType)
	if tokenType == "" {
		tokenType = "Bearer"
	}
	if !strings.EqualFold(tokenType, "Bearer") {
		return "", fmt.Errorf("unsupported token_type: %s", tokenType)
	}

	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return "", errors.New("missing access_token")
	}
	if strings.ContainsAny(accessToken, " \t\r\n") {
		return "", errors.New("access_token contains whitespace")
	}
	return "Bearer " + accessToken, nil
}

func normalizeOAuthIntent(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", oauthIntentLogin:
		return oauthIntentLogin
	case "bind", oauthIntentBindCurrentUser:
		return oauthIntentBindCurrentUser
	default:
		return oauthIntentLogin
	}
}

func (h *AuthHandler) buildOAuthBindUserCookieFromContext(c *gin.Context) (string, error) {
	userID, err := h.resolveOAuthBindTargetUserID(c)
	if err != nil || userID == nil || *userID == "" {
		return "", infraerrors.Unauthorized("UNAUTHORIZED", "authentication required")
	}
	return buildOAuthBindUserCookieValue(*userID, h.oauthBindCookieSecret())
}

func (h *AuthHandler) PrepareOAuthBindAccessTokenCookie(c *gin.Context) {
	const bearerPrefix = "Bearer "

	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if !strings.HasPrefix(strings.ToLower(authHeader), strings.ToLower(bearerPrefix)) {
		response.ErrorFrom(c, infraerrors.Unauthorized("UNAUTHORIZED", "authentication required"))
		return
	}

	token := strings.TrimSpace(authHeader[len(bearerPrefix):])
	if token == "" {
		response.ErrorFrom(c, infraerrors.Unauthorized("UNAUTHORIZED", "authentication required"))
		return
	}

	setOAuthBindAccessTokenCookie(c, token, isRequestHTTPS(c))
	c.Status(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
}

func (h *AuthHandler) resolveOAuthBindTargetUserID(c *gin.Context) (*string, error) {
	if subject, ok := servermiddleware.GetAuthSubjectFromContext(c); ok && subject.UserID != "" {
		return &subject.UserID, nil
	}
	if h == nil || h.authService == nil || h.userService == nil {
		return nil, service.ErrInvalidToken
	}

	cookie, err := c.Request.Cookie(oauthBindAccessTokenCookieName)
	clearOAuthBindAccessTokenCookie(c, isRequestHTTPS(c))
	if err != nil {
		return nil, err
	}

	tokenString, err := url.QueryUnescape(strings.TrimSpace(cookie.Value))
	if err != nil {
		return nil, err
	}
	if tokenString == "" {
		return nil, service.ErrInvalidToken
	}

	claims, err := h.authService.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	user, err := h.userService.GetByID(c.Request.Context(), claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsActive() || claims.TokenVersion != user.TokenVersion {
		return nil, service.ErrInvalidToken
	}
	return &user.ID, nil
}

func (h *AuthHandler) readOAuthBindUserIDFromCookie(c *gin.Context, cookieName string) (string, error) {
	value, err := readCookieDecoded(c, cookieName)
	if err != nil {
		return "", err
	}
	return parseOAuthBindUserCookieValue(value, h.oauthBindCookieSecret())
}

func (h *AuthHandler) oauthBindCookieSecret() string {
	if h == nil || h.cfg == nil {
		return ""
	}
	return strings.TrimSpace(h.cfg.JWT.Secret)
}

func buildOAuthBindUserCookieValue(userID string, secret string) (string, error) {
	secret = strings.TrimSpace(secret)
	if userID == "" || secret == "" {
		return "", errors.New("invalid oauth bind cookie input")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(userID))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return userID + "." + signature, nil
}

func parseOAuthBindUserCookieValue(value string, secret string) (string, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "", errors.New("missing oauth bind cookie secret")
	}
	payload, signature, ok := strings.Cut(strings.TrimSpace(value), ".")
	if !ok || payload == "" || signature == "" {
		return "", errors.New("invalid oauth bind cookie")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	expectedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return "", errors.New("invalid oauth bind cookie signature")
	}
	userID, err := parseEntityID(payload)
	if err != nil || userID == "" {
		return "", errors.New("invalid oauth bind cookie user")
	}
	return userID, nil
}

package service

import "regexp"

const (
	errorCodeForbidden       = "forbidden"
	errorCodeUnauthenticated = "unauthenticated"
)

var sensitiveQueryParamRegex = regexp.MustCompile(`(?i)([?&](?:key|client_secret|access_token|refresh_token)=)[^&"\s]+`)

func sanitizeUpstreamErrorMessage(message string) string {
	if message == "" {
		return ""
	}
	return sensitiveQueryParamRegex.ReplaceAllString(message, `$1***`)
}

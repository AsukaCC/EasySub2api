package service

import "strings"

func isCloudflareBotBlockResponse(body []byte) bool {
	return strings.Contains(strings.ToLower(string(body)), "error code: 1010")
}

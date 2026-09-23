package service

import (
	"github.com/tidwall/gjson"
	"strings"
)

// Keyword enforcement must inspect reminder text even when classifier input
// omits client metadata. Only the latest user turn is selected.
func extractContentModerationKeywordText(protocol string, body []byte) string {
	if protocol != ContentModerationProtocolAnthropicMessages {
		return ExtractContentModerationText(protocol, body)
	}
	array := gjson.GetBytes(body, "messages").Array()
	index := len(array) - 1
	for index >= 0 && strings.EqualFold(strings.TrimSpace(array[index].Get("role").String()), "system") {
		index--
	}
	if index < 0 || !strings.EqualFold(strings.TrimSpace(array[index].Get("role").String()), "user") {
		return ""
	}
	var parts, images []string
	collectContentValue(array[index].Get("content"), &parts, &images)
	return normalizeContentModerationText(strings.Join(parts, "\n"))
}

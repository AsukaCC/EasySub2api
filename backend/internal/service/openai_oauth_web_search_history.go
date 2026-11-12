package service

import (
	"github.com/tidwall/gjson"
	"strings"
)

func gjsonToolsContainWebSearch(tools gjson.Result) bool {
	if !tools.IsArray() {
		return false
	}
	for _, tool := range tools.Array() {
		if isOpenAIWebSearchToolType(tool.Get("type").String()) {
			return true
		}
	}
	return false
}

func openAIToolsContainWebSearch(rawTools any) bool {
	tools, ok := rawTools.([]any)
	if !ok {
		return false
	}
	for _, rawTool := range tools {
		tool, ok := rawTool.(map[string]any)
		if ok && isOpenAIWebSearchToolType(firstNonEmptyString(tool["type"])) {
			return true
		}
	}
	return false
}

func isOpenAIWebSearchToolType(toolType string) bool {
	return strings.HasPrefix(strings.TrimSpace(toolType), "web_search")
}

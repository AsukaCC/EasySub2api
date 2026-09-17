package service

import "github.com/AsukaCC/EasySub2api/internal/pkg/claude"

func floorClaudeCLIUserAgentVersion(ua string) (string, bool) {
	if extractProduct(ua) != claudeCLIUserAgentProduct || !isNewerVersion(claudeCLIUserAgentProduct+"/"+claude.CLIVersion(), ua) {
		return ua, false
	}
	next := claudeCLIUAVersionPrefixRegex.ReplaceAllString(ua, "${1}/"+claude.CLIVersion())
	return next, next != ua
}

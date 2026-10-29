package service

import (
	"context"
	"fmt"
	"sort"
)

func (s *GeminiMessagesCompatService) DeclaredModelIDs(ctx context.Context, groupID *string) ([]string, error) {
	accounts, err := s.listSchedulableAccountsOnce(ctx, groupID, PlatformGemini, true)
	if err != nil {
		return nil, fmt.Errorf("query accounts failed: %w", err)
	}
	seen := make(map[string]struct{})
	for i := range accounts {
		if accounts[i].Platform != PlatformGemini {
			continue
		}
		mapping := accounts[i].GetModelMapping()
		if len(mapping) == 0 {
			return nil, nil
		}
		for model := range mapping {
			seen[model] = struct{}{}
		}
	}
	models := make([]string, 0, len(seen))
	for model := range seen {
		models = append(models, model)
	}
	sort.Strings(models)
	return models, nil
}

// DeclaredModelsCover 判断声明的模型中是否有条目覆盖 model：键末尾的 * 按前缀匹配，
// 模型别名按请求路径查映射时的规则归一后再比较。
func DeclaredModelsCover(declared []string, model string) bool {
	normalized := normalizeRequestedModelForLookup(PlatformGemini, model)
	for _, pattern := range declared {
		if matchWildcard(pattern, model) || matchWildcard(pattern, normalized) {
			return true
		}
	}
	return false
}

package service

// OpenAITokenCacheKey 生成 OpenAI OAuth 账号的缓存键
// 格式: "openai:account:{account_id}"
func OpenAITokenCacheKey(account *Account) string {
	return "openai:account:" + account.ID
}

// ClaudeTokenCacheKey 生成 Claude (Anthropic) OAuth 账号的缓存键
// 格式: "claude:account:{account_id}"
func ClaudeTokenCacheKey(account *Account) string {
	return "claude:account:" + account.ID
}

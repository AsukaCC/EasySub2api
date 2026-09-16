export type OpenCodeMode = 'zen' | 'go'
export type OpenCodeProtocol = 'adaptive' | 'chat_completions' | 'responses' | 'anthropic'
export interface OpenCodeRule { pattern: string; protocol: Exclude<OpenCodeProtocol, 'adaptive'> }
export interface OpenCodeSettings {
  account_mode: OpenCodeMode
  api_protocol: OpenCodeProtocol
  protocol_rules?: OpenCodeRule[]
}

export function openCodeBaseUrl(mode: OpenCodeMode): string {
  return mode === 'zen' ? 'https://opencode.ai/zen/v1' : 'https://opencode.ai/zen/go/v1'
}

export function readOpenCodeSettings(credentials: Record<string, unknown> = {}): OpenCodeSettings {
  const protocol = credentials.api_protocol
  return {
    account_mode: credentials.account_mode === 'zen' ? 'zen' : 'go',
    api_protocol: protocol === 'chat_completions' || protocol === 'responses' || protocol === 'anthropic' ? protocol : 'adaptive',
    protocol_rules: Array.isArray(credentials.protocol_rules)
      ? credentials.protocol_rules.filter((rule): rule is OpenCodeRule => !!rule && typeof rule === 'object' && typeof rule.pattern === 'string' && ['chat_completions', 'responses', 'anthropic'].includes(rule.protocol)).map(rule => ({ ...rule }))
      : undefined
  }
}

export function defaultOpenCodeRules(mode: OpenCodeMode): OpenCodeRule[] {
  return [
    { pattern: 'grok-*', protocol: 'responses' },
    { pattern: 'gpt-*', protocol: 'responses' },
    { pattern: 'muse-spark-*', protocol: 'responses' },
    { pattern: mode === 'go' ? 'minimax-*' : 'claude-*', protocol: 'anthropic' },
    { pattern: 'qwen*', protocol: 'anthropic' }
  ]
}

export function applyOpenCodeSettings(target: Record<string, unknown>, settings: OpenCodeSettings): void {
  target.account_mode = settings.account_mode
  target.api_protocol = settings.api_protocol
  if (settings.protocol_rules) target.protocol_rules = settings.protocol_rules.map(rule => ({ ...rule }))
  else delete target.protocol_rules
}

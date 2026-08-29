import type { AccountPlatform } from '@/types'

/**
 * Platforms that can be used by a production account.
 * Keep this list limited to account platforms; composite is a group route,
 * and historical/removed platform values must not be offered for new data.
 */
export const ACCOUNT_PLATFORMS: readonly AccountPlatform[] = [
  'anthropic',
  'openai',
  'gemini',
  'antigravity',
  'grok',
  'kimi',
  'zhipu',
  'deepseek'
] as const

export type AccountPlatformOption = {
  value: AccountPlatform
  label: string
}

export function isAccountPlatform(value: unknown): value is AccountPlatform {
  return typeof value === 'string' && ACCOUNT_PLATFORMS.includes(value as AccountPlatform)
}

export function accountPlatformOptions(t: (key: string) => string): AccountPlatformOption[] {
  return ACCOUNT_PLATFORMS.map((platform) => ({
    value: platform,
    label: t(`admin.accounts.platforms.${platform}`)
  }))
}

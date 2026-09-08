import type { GroupPlatform } from '@/types'

export const OPENAI_CC_SWITCH_CODEX_MODEL = 'gpt-5.5'
export const GROK_CC_SWITCH_MODEL = 'grok-4.5'

export type CcSwitchClientType = 'claude'
export type CcSwitchApp = 'claude' | 'codex' | 'grokbuild'

export interface CcSwitchImportConfig {
  app: CcSwitchApp
  endpoint: string
  model?: string
}

export interface CcSwitchImportTarget {
  app: CcSwitchApp
  platform: GroupPlatform
}

export interface CcSwitchImportDeeplinkInput {
  baseUrl: string
  platform?: GroupPlatform | null
  app?: CcSwitchApp
  clientType: CcSwitchClientType
  providerName: string
  apiKey: string
  usageScript: string
  codexWebsocketEnabled?: boolean
}

interface CcSwitchCodexImportPayload {
  auth: {
    OPENAI_API_KEY: string
  }
  config: string
}

const CLAUDE_IMPORT_PLATFORMS = new Set<GroupPlatform>(['anthropic', 'gemini', 'antigravity', 'composite'])
const CODEX_IMPORT_PLATFORMS = new Set<GroupPlatform>(['openai', 'kimi', 'zhipu', 'deepseek'])

function encodeBase64Utf8(value: string): string {
  const bytes = new TextEncoder().encode(value)
  let binary = ''
  for (const byte of bytes) {
    binary += String.fromCharCode(byte)
  }
  return btoa(binary)
}

function asTomlString(value: string): string {
  return JSON.stringify(value)
}

function buildCodexImportPayload(
  input: CcSwitchImportDeeplinkInput,
  config: CcSwitchImportConfig
): CcSwitchCodexImportPayload {
  const websocketEnabled = input.codexWebsocketEnabled === true
  const codexConfig = [
    'model_provider = "custom"',
    `model = ${asTomlString(config.model || OPENAI_CC_SWITCH_CODEX_MODEL)}`,
    'model_reasoning_effort = "high"',
    'disable_response_storage = true',
    '',
    '[model_providers.custom]',
    `name = ${asTomlString(input.providerName)}`,
    `base_url = ${asTomlString(config.endpoint)}`,
    'wire_api = "responses"',
    'requires_openai_auth = true',
    `supports_websockets = ${websocketEnabled}`,
    '',
    '[features]',
    `responses_websockets_v2 = ${websocketEnabled}`,
    ''
  ].join('\n')

  return {
    auth: {
      OPENAI_API_KEY: input.apiKey
    },
    config: codexConfig
  }
}

function withV1Endpoint(baseUrl: string): string {
  const normalizedBaseUrl = baseUrl.replace(/\/+$/, '')
  return normalizedBaseUrl.endsWith('/v1') ? normalizedBaseUrl : `${normalizedBaseUrl}/v1`
}

export function ccsImportTargetsFromGroups(
  groups: Array<{ platform?: GroupPlatform | null }>
): CcSwitchImportTarget[] {
  const platforms = [...new Set(
    groups
      .map((group) => group.platform)
      .filter((platform): platform is GroupPlatform => Boolean(platform))
  )]
  const targets: CcSwitchImportTarget[] = []
  const claudePlatform = platforms.find((platform) => CLAUDE_IMPORT_PLATFORMS.has(platform))
  if (claudePlatform) {
    targets.push({ app: 'claude', platform: claudePlatform })
  }
  const codexPlatform = platforms.find((platform) => platform === 'openai')
    || platforms.find((platform) => CODEX_IMPORT_PLATFORMS.has(platform))
  const hasGrok = platforms.includes('grok')
  if (codexPlatform) {
    targets.push({ app: 'codex', platform: codexPlatform })
  } else if (hasGrok) {
    targets.push({ app: 'codex', platform: 'grok' })
  }
  if (hasGrok) {
    targets.push({ app: 'grokbuild', platform: 'grok' })
  }
  if (targets.length === 0) {
    targets.push({ app: 'claude', platform: 'anthropic' })
  }
  return targets
}

export function resolveCcSwitchImportConfig(
  platform: GroupPlatform | undefined | null,
  _clientType: CcSwitchClientType,
  baseUrl: string,
  appOverride?: CcSwitchApp
): CcSwitchImportConfig {
  const inferred = inferCcSwitchImportConfig(platform, baseUrl)
  if (appOverride === 'codex') {
    return {
      app: 'codex',
      endpoint: baseUrl,
      model: platform === 'grok' ? GROK_CC_SWITCH_MODEL : (inferred.model || OPENAI_CC_SWITCH_CODEX_MODEL)
    }
  }
  if (appOverride === 'grokbuild') {
    return {
      app: 'grokbuild',
      endpoint: withV1Endpoint(baseUrl),
      model: GROK_CC_SWITCH_MODEL
    }
  }
  if (appOverride === 'claude') {
    return {
      app: 'claude',
      endpoint: baseUrl
    }
  }
  return inferred
}

function inferCcSwitchImportConfig(
  platform: GroupPlatform | undefined | null,
  baseUrl: string
): CcSwitchImportConfig {
  switch (platform || 'anthropic') {
    case 'openai':
    case 'kimi':
    case 'zhipu':
    case 'deepseek':
      return {
        app: 'codex',
        endpoint: baseUrl,
        model: OPENAI_CC_SWITCH_CODEX_MODEL
      }
    case 'grok':
      return {
        app: 'grokbuild',
        endpoint: withV1Endpoint(baseUrl),
        model: GROK_CC_SWITCH_MODEL
      }
    default:
      return {
        app: 'claude',
        endpoint: baseUrl
      }
  }
}

export function buildCcSwitchImportDeeplink(input: CcSwitchImportDeeplinkInput): string {
  const config = resolveCcSwitchImportConfig(input.platform, input.clientType, input.baseUrl, input.app)
  const entries: [string, string][] = [
    ['resource', 'provider'],
    ['app', config.app],
    ['name', input.providerName],
    ['homepage', input.baseUrl],
    ['endpoint', config.endpoint],
    ['apiKey', input.apiKey],
    ['configFormat', 'json'],
    ['usageEnabled', 'true'],
    ['usageScript', encodeBase64Utf8(input.usageScript)],
    ['usageAutoInterval', '30']
  ]

  if (config.model) {
    entries.splice(2, 0, ['model', config.model])
  }

  if (config.app === 'codex') {
    const codexPayload = buildCodexImportPayload(input, config)
    entries.push(['config', encodeBase64Utf8(JSON.stringify(codexPayload))])
  }

  return `ccswitch://v1/import?${new URLSearchParams(entries).toString()}`
}

import type { GroupPlatform } from '@/types'

export const OPENAI_CC_SWITCH_CODEX_MODEL = 'gpt-5.5'
export const GROK_CC_SWITCH_MODEL = 'grok-4.5'

export type CcSwitchClientType = 'claude'

export interface CcSwitchImportConfig {
  app: string
  endpoint: string
  model?: string
}

export interface CcSwitchImportDeeplinkInput {
  baseUrl: string
  platform?: GroupPlatform | null
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

export function resolveCcSwitchImportConfig(
  platform: GroupPlatform | undefined | null,
  _clientType: CcSwitchClientType,
  baseUrl: string
): CcSwitchImportConfig {
  switch (platform || 'anthropic') {
    case 'openai':
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
  const config = resolveCcSwitchImportConfig(input.platform, input.clientType, input.baseUrl)
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

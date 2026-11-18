import { describe, expect, it } from 'vitest'
import {
  GROK_CC_SWITCH_MODEL,
  OPENAI_CC_SWITCH_CODEX_MODEL,
  SUB2API_CC_SWITCH_USAGE_SCRIPT,
  buildCcSwitchImportDeeplink,
  ccsImportTargetsFromGroups
} from '@/utils/ccswitchImport'
import type { GroupPlatform } from '@/types'

function paramsFromDeeplink(deeplink: string): URLSearchParams {
  const query = deeplink.split('?')[1] || ''
  return new URLSearchParams(query)
}

function decodeBase64Utf8(value: string): string {
  const bytes = Uint8Array.from(atob(value), character => character.charCodeAt(0))
  return new TextDecoder().decode(bytes)
}

describe('ccswitchImport utils', () => {
  const evaluateUsageScript = (code: string, baseUrl = 'https://api.example.com/v1') => {
    const substituted = code.replaceAll('{{baseUrl}}', baseUrl).replaceAll('{{apiKey}}', 'sk-test')
    return new Function(`return ${substituted}`)() as {
      request: { url: string; method: string; headers: Record<string, string> }
      extractor: (response: Record<string, unknown>) => Record<string, unknown>
    }
  }

  it.each(['claude', 'codex', 'gemini', 'grokbuild', 'opencode', 'openclaw', 'hermes'] as const)(
    'defaults %s exports to an enabled Sub2API usage script', (app) => {
      const params = paramsFromDeeplink(buildCcSwitchImportDeeplink({
        baseUrl: 'https://api.example.com', providerName: 'EasySub2api', apiKey: 'sk-test', clientType: 'claude', app,
      }))
      expect(params.get('usageEnabled')).toBe('true')
      expect(params.get('usageAutoInterval')).toBe('30')
      const script = decodeBase64Utf8(params.get('usageScript') || '')
      expect(script).toBe(SUB2API_CC_SWITCH_USAGE_SCRIPT)
      expect(evaluateUsageScript(script, params.get('endpoint')!).request).toEqual({
        url: 'https://api.example.com/v1/usage', method: 'GET',
        headers: { Authorization: 'Bearer sk-test', 'User-Agent': 'cc-switch/1.0' },
      })
    },
  )

  it.each(['', '/', '/v1', '/v1/'])('normalizes the usage URL suffix %s and preserves deployment paths', (suffix) => {
    const script = evaluateUsageScript(SUB2API_CC_SWITCH_USAGE_SCRIPT, `https://api.example.com/gateway${suffix}`)
    expect(script.request.url).toBe('https://api.example.com/gateway/v1/usage')
  })

  it.each([
    [{ isValid: true, planName: '钱包余额', remaining: 42, unit: 'points' }, { isValid: true, planName: '钱包余额', remaining: 42, unit: 'points' }],
    [{ isValid: true, remaining: 0, quota: { limit: 100, used: 100, unit: 'points' } }, { isValid: true, remaining: 0, total: 100, used: 100, unit: 'points' }],
    [{ isValid: true, planName: 'Unlimited', remaining: -1, unit: 'points' }, { isValid: true, planName: 'Unlimited', unit: 'points' }],
    [{ isValid: false, status: 'inactive' }, { isValid: false, invalidMessage: 'inactive', unit: 'points' }],
    [{ isValid: true, remaining: 1.25, unit: 'USD' }, { isValid: true, remaining: 1.25, unit: 'USD' }],
  ])('extracts Sub2API wallet, quota, subscription and invalid-key responses', (response, expected) => {
    expect(evaluateUsageScript(SUB2API_CC_SWITCH_USAGE_SCRIPT).extractor(response)).toEqual(expected)
  })

  it('defaults OpenAI CC Switch imports to the current Codex model', () => {
    expect(OPENAI_CC_SWITCH_CODEX_MODEL).toBe('gpt-6.2-sol')
  })

  it('defaults Grok Build imports to the current Grok model', () => {
    expect(GROK_CC_SWITCH_MODEL).toBe('grok-4.5')
  })

  const baseInput = {
    baseUrl: 'https://api.example.com',
    providerName: 'EasySub2api',
    apiKey: 'sk-test',
    usageScript: 'return true'
  }

  it('adds the Codex model parameter for OpenAI imports', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'openai',
        clientType: 'claude'
      })
    )

    expect(params.get('resource')).toBe('provider')
    expect(params.get('app')).toBe('codex')
    expect(params.get('endpoint')).toBe(baseInput.baseUrl + '/v1')
    expect(params.get('model')).toBe(OPENAI_CC_SWITCH_CODEX_MODEL)
    expect(atob(params.get('usageScript') || '')).toBe(baseInput.usageScript)
    const payload = JSON.parse(decodeBase64Utf8(params.get('config') || ''))
    expect(payload.auth.OPENAI_API_KEY).toBe(baseInput.apiKey)
    expect(payload.config).toContain(`model = "${OPENAI_CC_SWITCH_CODEX_MODEL}"`)
    expect(payload.config).not.toContain('default_subagent_model')
    expect(payload.config).toContain('supports_websockets = false')
    expect(payload.config).toContain('image_generation = true')
    expect(payload.config).toContain('untrusted_as_tool_output = false')
    expect(payload.config).toContain('experimental_bearer_token = "sk-test"')
    // CC Switch v4.0.34 preserves provider auth headers through its Codex link projection.
    expect(payload.config).toContain('requires_openai_auth = false')
    expect(payload.config).toContain('http_headers = { "x-openai-actor-authorization" = "local-image-extension" }')
    expect(payload.config).toContain('web_search = "live"')
    expect(payload.config).not.toMatch(/responses[-_]lite/i)
  })

  it('adds WebSocket support to the Codex configuration when enabled by admin settings', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        providerName: '测试站点',
        platform: 'openai',
        clientType: 'claude',
        codexWebsocketEnabled: true
      })
    )

    const payload = JSON.parse(decodeBase64Utf8(params.get('config') || ''))
    expect(payload.config).toContain('name = "测试站点"')
    expect(payload.config).toContain('supports_websockets = true')
    expect(payload.config).not.toContain('responses_websockets_v2')
    expect(payload.config).not.toContain('default_subagent_model')
  })

  it.each([
    'https://api.example.com',
    'https://api.example.com/',
    'https://api.example.com/v1',
    'https://api.example.com/v1/'
  ])('imports Grok Build with one /v1 suffix for base URL %s', (baseUrl) => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        baseUrl,
        platform: 'grok',
        clientType: 'claude'
      })
    )

    expect(params.get('app')).toBe('grokbuild')
    expect(params.get('endpoint')).toBe('https://api.example.com/v1')
    expect(params.get('model')).toBe(GROK_CC_SWITCH_MODEL)
  })

  it('lists native and OpenAI-compatible CC-Switch targets from bound groups', () => {
    expect(ccsImportTargetsFromGroups([
      { platform: 'anthropic' },
      { platform: 'openai' },
      { platform: 'grok' }
    ])).toEqual([
      { app: 'claude', platform: 'anthropic' },
      { app: 'codex', platform: 'openai' },
      { app: 'grokbuild', platform: 'grok' },
      { app: 'opencode', platform: 'openai' },
      { app: 'openclaw', platform: 'openai' },
      { app: 'hermes', platform: 'openai' }
    ])
  })

  it('adds a Codex target for Grok-only keys so Codex can load Grok models', () => {
    expect(ccsImportTargetsFromGroups([{ platform: 'grok' }])).toEqual([
      { app: 'codex', platform: 'grok' },
      { app: 'grokbuild', platform: 'grok' },
      { app: 'opencode', platform: 'grok' },
      { app: 'openclaw', platform: 'grok' },
      { app: 'hermes', platform: 'grok' }
    ])
  })

  it('adds Gemini CLI for Gemini-compatible groups', () => {
    expect(ccsImportTargetsFromGroups([{ platform: 'gemini' }])).toEqual([
      { app: 'claude', platform: 'gemini' },
      { app: 'gemini', platform: 'gemini' }
    ])
  })

  it('defaults to Claude when a key has no groups', () => {
    expect(ccsImportTargetsFromGroups([])).toEqual([
      { app: 'claude', platform: 'anthropic' }
    ])
  })

  it('imports Grok into Codex with the Grok model and Responses wire API', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'grok',
        app: 'codex',
        clientType: 'claude'
      })
    )

    expect(params.get('app')).toBe('codex')
    expect(params.get('endpoint')).toBe(baseInput.baseUrl + '/v1')
    expect(params.get('model')).toBe(GROK_CC_SWITCH_MODEL)
    const payload = JSON.parse(decodeBase64Utf8(params.get('config') || ''))
    expect(payload.config).toContain('wire_api = "responses"')
    expect(payload.config).toContain(`model = "${GROK_CC_SWITCH_MODEL}"`)
    expect(payload.config).not.toContain('default_subagent_model')
  })

  it('exports edited models consistently in the deep link and TOML', () => {
    const params = paramsFromDeeplink(buildCcSwitchImportDeeplink({
      ...baseInput, platform: 'openai', clientType: 'claude',
      codexModel: ' gpt-6-astra ', codexSubagentModel: 'gpt-6.1-terra', codexWebsocketEnabled: true
    }))
    const payload = JSON.parse(decodeBase64Utf8(params.get('config') || ''))
    expect(params.get('model')).toBe('gpt-6-astra')
    expect(payload.config).toContain('model = "gpt-6-astra"')
    expect(payload.config).toContain('default_subagent_model = "gpt-6.1-terra"')
    expect(payload.config).toContain('supports_websockets = true')
  })

  it.each([
    { platform: 'anthropic' as GroupPlatform, clientType: 'claude' as const, app: 'claude' }
  ])('does not add a model parameter for $platform imports', ({ platform, clientType, app }) => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform,
        clientType
      })
    )

    expect(params.get('app')).toBe(app)
    expect(params.get('endpoint')).toBe(baseInput.baseUrl)
    expect(params.has('model')).toBe(false)
  })
})

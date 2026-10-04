import { afterEach, describe, expect, it, vi } from 'vitest'
import { buildCodexModelCatalogUrl, fetchCodexModelsManifest } from '../codex'

afterEach(() => vi.unstubAllGlobals())

describe('Codex catalog', () => {
  it('normalizes the API prefix and requests the Codex manifest using the selected key', async () => {
    expect(buildCodexModelCatalogUrl('https://example.com/v1/')).toBe('https://example.com/v1/models')
    expect(buildCodexModelCatalogUrl('https://example.com/')).toBe('https://example.com/v1/models')
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ models: [{ slug: 'gpt-6.1-sol' }] })))
    vi.stubGlobal('fetch', fetchMock)
    const result = await fetchCodexModelsManifest('https://example.com', 'synthetic-key')
    expect(result.modelCount).toBe(1)
    expect(result.content).toContain('gpt-6.1-sol')
    expect(fetchMock).toHaveBeenCalledWith(expect.stringContaining('/v1/models?client_version='), expect.objectContaining({
      headers: { Accept: 'application/json', Authorization: 'Bearer synthetic-key' }, cache: 'no-store'
    }))
  })

  it('rejects an ordinary OpenAI model list and does not expose upstream error bodies', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ data: [] }))))
    await expect(fetchCodexModelsManifest('', 'synthetic-key')).rejects.toThrow('valid manifest')
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('private upstream diagnostic', { status: 403 })))
    await expect(fetchCodexModelsManifest('', 'synthetic-key')).rejects.toThrow('status 403')
  })
})

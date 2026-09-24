import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import OpenCodeAccountFields from '../OpenCodeAccountFields.vue'
import Select from '@/components/common/Select.vue'
import { applyOpenCodeSettings, defaultOpenCodeRules, openCodeBaseUrl, readOpenCodeSettings } from '../openCodeCredentials'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))

describe('OpenCode account settings', () => {
  it('keeps automatic rules absent and preserves unrelated credentials', () => {
    const target: Record<string, unknown> = { api_key: 'fixture', model_mapping: { alias: 'glm-fixture' }, protocol_rules: [] }
    applyOpenCodeSettings(target, readOpenCodeSettings({ account_mode: 'zen' }))
    expect(target).toEqual({ api_key: 'fixture', model_mapping: { alias: 'glm-fixture' }, account_mode: 'zen', api_protocol: 'adaptive' })
    expect(defaultOpenCodeRules('zen').some(rule => rule.pattern === 'claude-*')).toBe(true)
    expect(defaultOpenCodeRules('go').some(rule => rule.pattern === 'minimax-*')).toBe(true)
  })
  it('distinguishes an empty rule list from built-in defaults', () => {
    const settings = readOpenCodeSettings({ protocol_rules: [] })
    expect(settings.protocol_rules).toEqual([])
    const target = {}
    applyOpenCodeSettings(target, settings)
    expect(target).toHaveProperty('protocol_rules', [])
  })
  it('switches the official endpoint with the plan and preserves a custom endpoint', async () => {
    const wrapper = mount(OpenCodeAccountFields, { props: { modelValue: readOpenCodeSettings(), baseUrl: openCodeBaseUrl('go') } })
    await wrapper.findAllComponents(Select)[0]!.setValue('zen')
    expect(wrapper.emitted('update:baseUrl')?.[0]).toEqual([openCodeBaseUrl('zen')])
    await wrapper.setProps({ baseUrl: 'https://relay.example/v1' })
    await wrapper.findAllComponents(Select)[0]!.setValue('zen')
    expect(wrapper.emitted('update:baseUrl')).toHaveLength(1)
  })
  it('edits rules without mutating the supplied settings', async () => {
    const settings = readOpenCodeSettings({ protocol_rules: [{ pattern: 'alias*', protocol: 'responses' }] })
    const wrapper = mount(OpenCodeAccountFields, { props: { modelValue: settings, baseUrl: openCodeBaseUrl('go') } })
    await wrapper.find('input').setValue('other*')
    expect(settings.protocol_rules?.[0]?.pattern).toBe('alias*')
    expect(wrapper.emitted('update:modelValue')?.[0]?.[0]).toMatchObject({ protocol_rules: [{ pattern: 'other*', protocol: 'responses' }] })
  })
})

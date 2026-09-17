import { describe, expect, it } from 'vitest'
import { createI18n } from 'vue-i18n'
import en from '../locales/en'
import zh from '../locales/zh'

describe('runtime-only locale compilation', () => {
  it('renders both complete language packs without eval', () => {
    const i18n = createI18n({ legacy: false, locale: 'en', messages: {
      en,
      zh
    } })
    expect(i18n.global.t('admin.accounts.protection.title')).toBe('Account protection')
    expect(i18n.global.t('admin.accounts.protection.batchResult', { success: 2, failed: 1 })).toBe('Enabled: 2; failed: 1')
    i18n.global.locale.value = 'zh'
    expect(i18n.global.t('admin.accounts.protection.title')).toBe('账号保护')
  })
})

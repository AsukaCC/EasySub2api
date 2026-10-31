import { describe, expect, it } from 'vitest'
import { ensureRouteMessages, i18n, initI18n } from '../index'
import en from '../locales/en'
import zh from '../locales/zh'

describe('i18n route namespaces', () => {
  it('nests admin locale modules under the admin key', async () => {
    i18n.global.locale.value = 'en'
    await initI18n()
    await ensureRouteMessages('/admin/dashboard', 'en')
    expect(i18n.global.t('admin.dashboard.title')).toBe('Admin Dashboard')
  })

  it.each(['en', 'zh'] as const)('loads all route namespaces for %s', async (locale) => {
    i18n.global.locale.value = locale
    await ensureRouteMessages('/admin/settings', locale)
    await ensureRouteMessages('/channel-monitor', locale)
    expect(i18n.global.getLocaleMessage(locale)).toEqual(locale === 'en' ? en : zh)
    expect(i18n.global.t('adminAccessKeys.title')).toBe((locale === 'en' ? en : zh).adminAccessKeys.title)
  })
})

import { describe, expect, it } from 'vitest'
import { ensureRouteMessages, i18n, initI18n } from '../index'

describe('i18n route namespaces', () => {
  it('nests admin locale modules under the admin key', async () => {
    i18n.global.locale.value = 'en'
    await initI18n()
    await ensureRouteMessages('/admin/dashboard', 'en')
    expect(i18n.global.t('admin.dashboard.title')).toBe('Admin Dashboard')
  })
})

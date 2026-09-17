import { describe, expect, it } from 'vitest'
import featureManagementSource from '../FeatureManagementView.vue?raw'

describe('FeatureManagementView site billing mode control', () => {
  it('loads, saves, and refreshes the unified site billing mode', () => {
    expect(featureManagementSource).toContain('resolveSiteBillingMode(settings)')
    expect(featureManagementSource).toContain('site_billing_mode: next')
    expect(featureManagementSource).toContain('appStore.fetchPublicSettings(true)')
    expect(featureManagementSource).toContain('adminSettingsStore.fetch(true)')
  })
})

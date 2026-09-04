import { describe, expect, it } from 'vitest'
import featureManagementSource from '../FeatureManagementView.vue?raw'
import opsFeatureSettingsSource from '../OpsFeatureSettingsView.vue?raw'
import riskControlSource from '../RiskControlView.vue?raw'
import settingsSource from '../SettingsView.vue?raw'
import systemUpdatesSource from '../SystemUpdatesView.vue?raw'
import userLevelsSource from '../UserLevelsView.vue?raw'
import opsDashboardHeaderSource from '../ops/components/OpsDashboardHeader.vue?raw'
import promptAuditSource from '../../../features/prompt-audit/PromptAuditView.vue?raw'

describe('admin page titles', () => {
  it.each([
    ['user levels', userLevelsSource],
    ['feature management', featureManagementSource],
    ['ops feature settings', opsFeatureSettingsSource],
    ['system updates', systemUpdatesSource],
    ['risk control', riskControlSource],
    ['settings', settingsSource],
    ['prompt audit', promptAuditSource]
  ])('does not render a duplicate top-level title on %s', (_, source) => {
    expect(source).not.toMatch(/<h1\b/)
  })

  it('keeps the operations title for fullscreen mode only', () => {
    expect(opsDashboardHeaderSource).toMatch(
      /<h1\s+v-if="props\.fullscreen"\s+class="views-admin-ops-components-ops-dashboard-header__heading"/
    )
    expect(opsDashboardHeaderSource.match(/<h1\b/g)).toHaveLength(1)
  })
})

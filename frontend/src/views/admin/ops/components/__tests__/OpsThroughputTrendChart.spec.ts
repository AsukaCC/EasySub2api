import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import OpsThroughputTrendChart from '../OpsThroughputTrendChart.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

vi.mock('@/components/charts/d3/D3LineChart.vue', () => ({
  default: {
    name: 'D3LineChart',
    props: ['data', 'options'],
    template: '<div class="line-chart" />',
  },
}))

describe('OpsThroughputTrendChart', () => {
  it('renders the responsive header toolbar controls', () => {
    const wrapper = mount(OpsThroughputTrendChart, {
      props: {
        points: [],
        loading: false,
        timeRange: '1h',
      },
      global: {
        stubs: {
          EmptyState: true,
          HelpTooltip: true,
        },
      },
    })

    const header = wrapper.get('[data-testid="throughput-chart-header"]')
    expect(header.classes()).toContain('views-admin-ops-components-ops-throughput-trend-chart__panel-2')

    const toolbar = wrapper.get('[data-testid="throughput-chart-toolbar"]')
    expect(toolbar.classes()).toContain('views-admin-ops-components-ops-throughput-trend-chart__panel-3')
    expect(toolbar.findAll('button')).toHaveLength(3)
    toolbar.findAll('button').forEach((button) => {
      expect(button.classes()).toContain('views-admin-ops-components-ops-throughput-trend-chart__action')
    })
  })
})

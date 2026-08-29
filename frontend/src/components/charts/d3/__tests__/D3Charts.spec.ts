import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { describe, expect, it } from 'vitest'
import D3BarChart from '../D3BarChart.vue'
import D3DonutChart from '../D3DonutChart.vue'
import D3LineChart from '../D3LineChart.vue'
import type { D3LineChartHandle } from '../chartTypes'

const data = {
  labels: ['Mon', 'Tue', 'Wed'],
  datasets: [
    {
      label: 'Requests',
      data: [12, 18, 15],
      borderColor: '#2563eb',
      backgroundColor: '#2563eb20',
      fill: true,
      tension: 0.3,
    },
    {
      label: 'Errors',
      data: [2, 4, 1],
      borderColor: '#dc2626',
      backgroundColor: '#dc2626',
    },
  ],
}

describe('D3 chart primitives', () => {
  it('renders line and area paths and exposes chart controls', async () => {
    const wrapper = mount(D3LineChart, { props: { data } })
    await nextTick()

    const paths = wrapper.findAll('.d3-line-chart__svg path')
    expect(paths.length).toBeGreaterThanOrEqual(3)
    expect(paths.every((path) => Boolean(path.attributes('d')))).toBe(true)

    const handle = wrapper.vm as unknown as D3LineChartHandle
    expect(handle.toDataUrl()).toMatch(/^data:image\/svg\+xml/)
    expect(() => handle.resetZoom()).not.toThrow()
  })

  it('renders one donut segment for every value', async () => {
    const wrapper = mount(D3DonutChart, {
      props: {
        data: {
          labels: ['Input', 'Output', 'Cache'],
          datasets: [{ data: [50, 30, 20], backgroundColor: ['#2563eb', '#059669', '#d97706'] }],
        },
        options: { plugins: { legend: { display: false } }, cutout: '60%' },
      },
    })
    await nextTick()

    const segments = wrapper.findAll('.d3-donut-chart__arc')
    expect(segments).toHaveLength(3)
    expect(segments.every((segment) => Boolean(segment.attributes('d')))).toBe(true)
  })

  it('renders grouped bars for every dataset value', async () => {
    const wrapper = mount(D3BarChart, {
      props: {
        data,
        options: { plugins: { legend: { display: false } } },
      },
    })
    await nextTick()

    expect(wrapper.findAll('.d3-bar-chart__bar')).toHaveLength(6)
  })
})

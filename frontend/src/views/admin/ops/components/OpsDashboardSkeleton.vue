<script setup lang="ts">
import { usePageLoading } from '@/composables/usePageLoading'
usePageLoading(() => true)
interface Props {
  fullscreen?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  fullscreen: false
})
</script>

<template>
  <div class="views-admin-ops-components-ops-dashboard-skeleton__panel">
    <!-- Header (matches OpsDashboardHeader + overview blocks) -->
    <div :class="['views-admin-ops-components-ops-dashboard-skeleton__panel-29', props.fullscreen ? 'views-admin-ops-components-ops-dashboard-skeleton__panel-30' : 'views-admin-ops-components-ops-dashboard-skeleton__panel-31 card-body']">
      <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-2">
        <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-3">
          <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-4 loading-shimmer"></div>
          <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-5 loading-shimmer"></div>
        </div>
        <div v-if="!props.fullscreen" class="views-admin-ops-components-ops-dashboard-skeleton__panel-6">
          <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-7 loading-shimmer"></div>
          <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-8 loading-shimmer"></div>
          <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-9 loading-shimmer"></div>
          <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-10 loading-shimmer"></div>
          <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-11 loading-shimmer"></div>
          <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-11 loading-shimmer"></div>
          <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-10 loading-shimmer"></div>
        </div>
      </div>

      <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-12">
        <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-13">
          <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-14">
            <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-15 loading-shimmer"></div>
            <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-16">
              <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-17 loading-shimmer"></div>
              <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-18">
                <div v-for="i in 4" :key="i" class="views-admin-ops-components-ops-dashboard-skeleton__panel-19 loading-shimmer"></div>
              </div>
            </div>
          </div>
        </div>

        <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-20">
          <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-21">
            <div v-for="i in 6" :key="i" class="views-admin-ops-components-ops-dashboard-skeleton__panel-22 loading-shimmer"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Row: Concurrency + Throughput (matches OpsDashboard.vue) -->
    <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-23">
      <div :class="['views-admin-ops-components-ops-dashboard-skeleton__panel-32', props.fullscreen ? 'views-admin-ops-components-ops-dashboard-skeleton__panel-30' : 'views-admin-ops-components-ops-dashboard-skeleton__panel-31 card-body']">
        <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-24 loading-shimmer"></div>
        <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-25 loading-shimmer"></div>
      </div>
      <div :class="['views-admin-ops-components-ops-dashboard-skeleton__panel-32', props.fullscreen ? 'views-admin-ops-components-ops-dashboard-skeleton__panel-30' : 'views-admin-ops-components-ops-dashboard-skeleton__panel-31 card-body']">
        <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-24 loading-shimmer"></div>
        <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-25 loading-shimmer"></div>
      </div>
      <div :class="['views-admin-ops-components-ops-dashboard-skeleton__panel-33', props.fullscreen ? 'views-admin-ops-components-ops-dashboard-skeleton__panel-30' : 'views-admin-ops-components-ops-dashboard-skeleton__panel-31 card-body']">
        <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-26 loading-shimmer"></div>
        <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-25 loading-shimmer"></div>
      </div>
    </div>

    <!-- Row: Visual Analysis (baseline 3-up grid) -->
    <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-27">
      <div
        v-for="i in 3"
        :key="i"
        :class="['views-admin-ops-components-ops-dashboard-skeleton__panel-29', props.fullscreen ? 'views-admin-ops-components-ops-dashboard-skeleton__panel-30' : 'views-admin-ops-components-ops-dashboard-skeleton__panel-31 card-body']"
      >
        <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-24 loading-shimmer"></div>
        <div class="views-admin-ops-components-ops-dashboard-skeleton__panel-28 loading-shimmer"></div>
      </div>
    </div>

  </div>
</template>

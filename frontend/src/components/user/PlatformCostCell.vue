<template>
  <div v-if="usage" class="components-user-platform-cost-cell__panel">
    <div class="components-user-platform-cost-cell__panel-2">
      <span class="components-user-platform-cost-cell__text">{{ t('admin.users.today') }}:</span>
      <span class="components-user-platform-cost-cell__text-2">{{ formatPoints(usage.today_actual_cost) }}</span>
    </div>
    <div class="components-user-platform-cost-cell__panel-3">
      <span class="components-user-platform-cost-cell__text">{{ t('admin.users.total') }}:</span>
      <span class="components-user-platform-cost-cell__text-2">{{ formatPoints(usage.total_actual_cost) }}</span>
    </div>
  </div>
  <span v-else class="components-user-platform-cost-cell__text-3">—</span>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { PlatformUsage } from '@/api/admin/dashboard'
import { formatPoints } from '@/utils/format'

defineProps<{
  usage?: PlatformUsage
}>()

const { t } = useI18n()
</script>

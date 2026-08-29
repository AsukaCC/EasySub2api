<template>
  <!-- 用量页"用户排行"tab 内容：无卡片外观，依赖父级统一卡片；筛选/时间范围复用页面级筛选栏 -->
  <div>
    <!-- Toolbar -->
    <div class="components-admin-usage-user-token-ranking__panel">
      <p class="components-admin-usage-user-token-ranking__description">{{ t('admin.usage.tokenRanking.subtitle') }}</p>
      <div class="components-admin-usage-user-token-ranking__panel-2">
        <span v-if="!loading && items.length > 0" class="components-admin-usage-user-token-ranking__description">
          {{ t('admin.usage.tokenRanking.userCount', { count: items.length }) }}
        </span>
        <div class="components-admin-usage-user-token-ranking__panel-3">
          <Select v-model="limit" :options="limitOptions" @change="load" />
        </div>
      </div>
    </div>

    <!-- Table -->
    <div class="components-admin-usage-user-token-ranking__panel-4">
      <table class="components-admin-usage-user-token-ranking__table">
        <thead class="components-admin-usage-user-token-ranking__header">
          <tr>
            <th class="components-admin-usage-user-token-ranking__heading">#</th>
            <th class="components-admin-usage-user-token-ranking__heading-2">
              {{ t('admin.usage.tokenRanking.columns.user') }}
            </th>
            <th
              v-for="col in sortableColumns"
              :key="col.key"
              class="components-admin-usage-user-token-ranking__heading-3"
              :class="sortBy === col.key ? 'components-admin-usage-user-token-ranking__heading-4' : 'components-admin-usage-user-token-ranking__heading-5'"
              @click="setSort(col.key)"
            >
              {{ t(col.label) }}
              <span v-if="sortBy === col.key" aria-hidden="true">↓</span>
            </th>
          </tr>
        </thead>
        <tbody class="components-admin-usage-user-token-ranking__body">
          <tr v-if="loading">
            <td :colspan="sortableColumns.length + 2" class="components-admin-usage-user-token-ranking__cell">
              <LoadingSpinner />
            </td>
          </tr>
          <tr v-else-if="items.length === 0">
            <td :colspan="sortableColumns.length + 2" class="components-admin-usage-user-token-ranking__cell-2">
              {{ t('admin.dashboard.noDataAvailable') }}
            </td>
          </tr>
          <tr
            v-for="(item, index) in items"
            v-else
            :key="item.user_id"
            class="components-admin-usage-user-token-ranking__row"
            :title="t('admin.usage.tokenRanking.rowHint')"
            @click="$emit('select-user', item.user_id, item.email)"
          >
            <td class="components-admin-usage-user-token-ranking__cell-3">
              <span
                v-if="index < 3"
                class="components-admin-usage-user-token-ranking__3"
                :class="RANK_BADGE_CLASSES[index]"
              >{{ index + 1 }}</span>
              <span v-else class="components-admin-usage-user-token-ranking__text">{{ index + 1 }}</span>
            </td>
            <td class="components-admin-usage-user-token-ranking__cell-4" :title="item.email">
              {{ item.email || `User #${item.user_id}` }}
              <span class="components-admin-usage-user-token-ranking__text-2">#{{ item.user_id }}</span>
            </td>
            <td class="components-admin-usage-user-token-ranking__cell-5">{{ item.requests.toLocaleString() }}</td>
            <td class="components-admin-usage-user-token-ranking__cell-5">{{ fmtTokens(item.input_tokens) }}</td>
            <td class="components-admin-usage-user-token-ranking__cell-5">{{ fmtTokens(item.output_tokens) }}</td>
            <td class="components-admin-usage-user-token-ranking__cell-5">{{ fmtTokens(item.cache_tokens) }}</td>
            <td class="components-admin-usage-user-token-ranking__cell-6">{{ fmtTokens(item.total_tokens) }}</td>
            <td class="components-admin-usage-user-token-ranking__cell-7">{{ formatPoints(item.actual_cost) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getUserBreakdown, type UserBreakdownParams } from '@/api/admin/dashboard'
import { formatCompactNumber, formatPoints } from '@/utils/format'
import type { UserBreakdownItem } from '@/types'
import Select from '@/components/common/Select.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

const props = defineProps<{
  startDate: string
  endDate: string
  filters: Record<string, unknown>
  model?: string
}>()

defineEmits<{ (e: 'select-user', userId: string, email: string): void }>()

const { t } = useI18n()

type SortKey = NonNullable<UserBreakdownParams['sort_by']>
const sortableColumns: { key: SortKey; label: string }[] = [
  { key: 'requests', label: 'admin.usage.tokenRanking.columns.requests' },
  { key: 'input_tokens', label: 'admin.usage.tokenRanking.columns.inputTokens' },
  { key: 'output_tokens', label: 'admin.usage.tokenRanking.columns.outputTokens' },
  { key: 'cache_tokens', label: 'admin.usage.tokenRanking.columns.cacheTokens' },
  { key: 'total_tokens', label: 'admin.usage.tokenRanking.columns.totalTokens' },
  { key: 'actual_cost', label: 'admin.usage.tokenRanking.columns.cost' },
]

const limitOptions = [
  { value: 20, label: 'Top 20' },
  { value: 50, label: 'Top 50' },
  { value: 100, label: 'Top 100' },
  { value: 200, label: 'Top 200' },
]

// 前三名金/银/铜徽章
const RANK_BADGE_CLASSES = [
  'components-admin-usage-user-token-ranking__state',
  'components-admin-usage-user-token-ranking__state-2',
  'components-admin-usage-user-token-ranking__state-3',
]

const items = ref<UserBreakdownItem[]>([])
const loading = ref(false)
const sortBy = ref<SortKey>('total_tokens')
const limit = ref(50)
let reqSeq = 0

const fmtTokens = (v: number) => formatCompactNumber(v)
const setSort = (key: SortKey) => {
  if (sortBy.value === key) return
  sortBy.value = key
  load()
}

const load = async () => {
  const seq = ++reqSeq
  loading.value = true
  try {
    const params: UserBreakdownParams = {
      ...props.filters,
      start_date: props.startDate,
      end_date: props.endDate,
      sort_by: sortBy.value,
      limit: limit.value,
    }
    if (props.model) params.model = props.model
    const res = await getUserBreakdown(params)
    if (seq !== reqSeq) return
    items.value = res.users || []
  } catch {
    if (seq !== reqSeq) return
    items.value = []
  } finally {
    if (seq === reqSeq) loading.value = false
  }
}

// Reload when the shared filters / date range / model change.
watch(
  () => [props.startDate, props.endDate, props.model, JSON.stringify(props.filters)],
  () => load(),
  { immediate: true }
)

defineExpose({ reload: load })
</script>

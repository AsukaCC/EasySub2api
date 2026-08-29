<template>
  <div
    class="components-common-pagination__panel"
  >
    <div class="components-common-pagination__panel-2">
      <!-- Mobile pagination -->
      <button
        @click="goToPage(page - 1)"
        :disabled="page === 1"
        class="components-common-pagination__action"
      >
        {{ t('pagination.previous') }}
      </button>
      <span class="components-common-pagination__text">
        {{ t('pagination.pageOf', { page, total: totalPages }) }}
      </span>
      <button
        @click="goToPage(page + 1)"
        :disabled="page === totalPages"
        class="components-common-pagination__action-2"
      >
        {{ t('pagination.next') }}
      </button>
    </div>

    <div class="components-common-pagination__panel-3">
      <!-- Desktop pagination info -->
      <div class="components-common-pagination__panel-4">
        <p class="components-common-pagination__text">
          {{ t('pagination.showing') }}
          <span class="components-common-pagination__text-2">{{ fromItem }}</span>
          {{ t('pagination.to') }}
          <span class="components-common-pagination__text-2">{{ toItem }}</span>
          {{ t('pagination.of') }}
          <span class="components-common-pagination__text-2">{{ total }}</span>
          {{ t('pagination.results') }}
        </p>

        <!-- Page size selector -->
        <div v-if="showPageSizeSelector" class="components-common-pagination__panel-5">
          <span class="components-common-pagination__text"
            >{{ t('pagination.perPage') }}:</span
          >
          <div class="components-common-pagination__panel-6 page-size-select">
            <Select
              :model-value="pageSize"
              :options="pageSizeSelectOptions"
              @update:model-value="handlePageSizeChange"
            />
          </div>
        </div>

        <div v-if="showJump" class="components-common-pagination__panel-5">
          <span class="components-common-pagination__text">{{ t('pagination.jumpTo') }}</span>
          <input
            v-model="jumpPage"
            type="number"
            min="1"
            :max="totalPages"
            class="components-common-pagination__field input"
            :placeholder="t('pagination.jumpPlaceholder')"
            @keyup.enter="submitJump"
          />
          <button type="button" class="btn btn-ghost btn-sm" @click="submitJump">
            {{ t('pagination.jumpAction') }}
          </button>
        </div>
      </div>

      <!-- Desktop pagination buttons -->
      <nav
        class="components-common-pagination__navigation"
        aria-label="Pagination"
      >
        <!-- Previous button -->
        <button
          @click="goToPage(page - 1)"
          :disabled="page === 1"
          class="components-common-pagination__action-3"
          :aria-label="t('pagination.previous')"
        >
          <Icon name="chevronLeft" size="md" />
        </button>

        <!-- Page numbers -->
        <button
          v-for="(pageNum, index) in visiblePages"
          :key="`${pageNum}-${index}`"
          @click="typeof pageNum === 'number' && goToPage(pageNum)"
          :disabled="typeof pageNum !== 'number'"
          :class="[
            'components-common-pagination__action-5',
            pageNum === page
              ? 'components-common-pagination__action-6'
              : 'components-common-pagination__action-7',
            typeof pageNum !== 'number' && 'components-common-pagination__action-8'
          ]"
          :aria-label="
            typeof pageNum === 'number' ? t('pagination.goToPage', { page: pageNum }) : undefined
          "
          :aria-current="pageNum === page ? 'page' : undefined"
        >
          {{ pageNum }}
        </button>

        <!-- Next button -->
        <button
          @click="goToPage(page + 1)"
          :disabled="page === totalPages"
          class="components-common-pagination__action-4"
          :aria-label="t('pagination.next')"
        >
          <Icon name="chevronRight" size="md" />
        </button>
      </nav>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import Select from './Select.vue'
import { getConfiguredTablePageSizeOptions, normalizeTablePageSize } from '@/utils/tablePreferences'
import { setPersistedPageSize } from '@/composables/usePersistedPageSize'

const { t } = useI18n()

interface Props {
  total: number
  page: number
  pageSize: number
  pageSizeOptions?: number[]
  showPageSizeSelector?: boolean
  showJump?: boolean
}

interface Emits {
  (e: 'update:page', page: number): void
  (e: 'update:pageSize', pageSize: number): void
}

const props = withDefaults(defineProps<Props>(), {
  pageSizeOptions: () => getConfiguredTablePageSizeOptions(),
  showPageSizeSelector: true,
  showJump: false
})

const emit = defineEmits<Emits>()

const totalPages = computed(() => Math.ceil(props.total / props.pageSize))

const fromItem = computed(() => {
  if (props.total === 0) return 0
  return (props.page - 1) * props.pageSize + 1
})

const toItem = computed(() => {
  const to = props.page * props.pageSize
  return to > props.total ? props.total : to
})

const pageSizeSelectOptions = computed(() => {
  const options = Array.from(
    new Set([
      ...getConfiguredTablePageSizeOptions(),
      normalizeTablePageSize(props.pageSize)
    ])
  ).sort((a, b) => a - b)

  return options.map((size) => ({
    value: size,
    label: String(size)
  }))
})

const jumpPage = ref('')

const visiblePages = computed(() => {
  const pages: (number | string)[] = []
  const maxVisible = 7
  const total = totalPages.value

  if (total <= maxVisible) {
    // Show all pages if total is small
    for (let i = 1; i <= total; i++) {
      pages.push(i)
    }
  } else {
    // Always show first page
    pages.push(1)

    const start = Math.max(2, props.page - 2)
    const end = Math.min(total - 1, props.page + 2)

    // Add ellipsis before if needed
    if (start > 2) {
      pages.push('...')
    }

    // Add middle pages
    for (let i = start; i <= end; i++) {
      pages.push(i)
    }

    // Add ellipsis after if needed
    if (end < total - 1) {
      pages.push('...')
    }

    // Always show last page
    pages.push(total)
  }

  return pages
})

const goToPage = (newPage: number) => {
  if (newPage >= 1 && newPage <= totalPages.value && newPage !== props.page) {
    emit('update:page', newPage)
  }
}

const handlePageSizeChange = (value: string | number | boolean | null) => {
  if (value === null || typeof value === 'boolean') return
  const newPageSize = normalizeTablePageSize(typeof value === 'string' ? parseInt(value, 10) : value)
  setPersistedPageSize(newPageSize)
  emit('update:pageSize', newPageSize)
}

const submitJump = () => {
  const value = jumpPage.value.trim()
  if (!value) return
  const pageNum = Number.parseInt(value, 10)
  if (Number.isNaN(pageNum)) return
  const nextPage = Math.min(Math.max(pageNum, 1), totalPages.value)
  jumpPage.value = ''
  goToPage(nextPage)
}
</script>

<style scoped>
.page-size-select :deep(.select-trigger) {
  padding: 0.375rem 0.75rem;
  font-size: var(--font-size-sm);
}

/* ============================================================
   对齐项目风格:覆盖迁移别名样式
   容器透明化(去白底/上边框),页码从连体按钮组改为分离圆角按钮,
   当前页用主色浅底 + 主色描边(与顶部导航激活态同语言)。
   ============================================================ */
/* 悬浮磨砂玻璃条:与顶部导航壳同一材质(玻璃描边 + backdrop 模糊 + 内高光) */
.components-common-pagination__panel,
.dark .components-common-pagination__panel {
  flex-wrap: wrap;
  gap: 0.5rem 1rem;
  padding: 0.5rem 1rem;
  border: 1px solid var(--glass-border);
  border-top: 1px solid var(--glass-border);
  border-radius: var(--radius-xl);
  background: var(--glass-layer-content-bg);
  -webkit-backdrop-filter: blur(var(--glass-layer-content-blur)) saturate(var(--glass-saturate, 1.8));
  backdrop-filter: blur(var(--glass-layer-content-blur)) saturate(var(--glass-saturate, 1.8));
  box-shadow:
    var(--glass-shadow),
    0 1px 0 var(--glass-highlight) inset;
  color: var(--color-text-tertiary);
  font-variant-numeric: tabular-nums;
}

.components-common-pagination__navigation {
  gap: 0.25rem;
  border-radius: 0;
  box-shadow: none;
  font-variant-numeric: tabular-nums;
}

/* 前后箭头 + 页码:透明底圆角按钮 */
.components-common-pagination__action-3,
.components-common-pagination__action-4,
.components-common-pagination__action-5,
.dark .components-common-pagination__action-3,
.dark .components-common-pagination__action-4,
.dark .components-common-pagination__action-5 {
  justify-content: center;
  min-width: 2.25rem;
  padding: 0.375rem 0.625rem;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--color-text-secondary);
  transition: color 160ms ease, background-color 160ms ease, border-color 160ms ease;
}

.components-common-pagination__action-3:hover:not(:disabled),
.components-common-pagination__action-4:hover:not(:disabled),
.components-common-pagination__action-7:hover:not(:disabled),
.dark .components-common-pagination__action-3:hover:not(:disabled),
.dark .components-common-pagination__action-4:hover:not(:disabled),
.dark .components-common-pagination__action-7:hover:not(:disabled) {
  color: var(--color-text-primary);
  border-color: var(--glass-border);
  background-color: var(--glass-bg-interactive-hover);
  -webkit-backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate));
  box-shadow: 0 1px 0 var(--glass-highlight) inset;
}

.components-common-pagination__action-3:disabled,
.components-common-pagination__action-4:disabled {
  opacity: 0.4;
}

/* 普通页码 */
.components-common-pagination__action-7,
.dark .components-common-pagination__action-7 {
  border-color: transparent;
  background: transparent;
  color: var(--color-text-secondary);
}

/* 当前页:主色浅底 + 主色描边 + 拟态玻璃高光 */
.components-common-pagination__action-6,
.dark .components-common-pagination__action-6 {
  border-color: var(--glass-border-active);
  background-color: var(--color-primary-subtle);
  color: var(--color-text-brand);
  font-weight: var(--font-weight-semibold);
  box-shadow:
    0 2px 8px rgba(10, 132, 255, 0.15),
    0 1px 0 var(--glass-highlight-hover) inset;
}

.components-common-pagination__action-6:hover,
.dark .components-common-pagination__action-6:hover {
  border-color: var(--glass-border-hover);
  background-color: var(--color-primary-subtle);
  color: var(--color-text-brand);
}

/* 省略号:纯文本,无交互反馈 */
.components-common-pagination__action-8,
.dark .components-common-pagination__action-8 {
  border-color: transparent;
  background: transparent;
  color: var(--color-text-tertiary);
  opacity: 0.8;
  cursor: default;
}

/* 移动端上一页/下一页:玻璃次级按钮 */
.components-common-pagination__action,
.components-common-pagination__action-2,
.dark .components-common-pagination__action,
.dark .components-common-pagination__action-2 {
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-md);
  background: var(--glass-bg);
  color: var(--color-text-secondary);
  transition: background-color 160ms ease, color 160ms ease;
}

.components-common-pagination__action:hover:not(:disabled),
.components-common-pagination__action-2:hover:not(:disabled),
.dark .components-common-pagination__action:hover:not(:disabled),
.dark .components-common-pagination__action-2:hover:not(:disabled) {
  border-color: var(--glass-border-hover);
  background: var(--glass-layer-inset-bg);
  -webkit-backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
  backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
  color: var(--color-text-primary);
}

.components-common-pagination__action:disabled,
.components-common-pagination__action-2:disabled {
  opacity: 0.4;
}
</style>

<template>
  <div class="components-model-plaza-model-plaza-content__panel">
    <!-- 页面标题在独立和后台内嵌形态下都保留。 -->
    <div>
      <h1 class="components-model-plaza-model-plaza-content__heading">{{ t('modelPlaza.title') }}</h1>
      <p class="components-model-plaza-model-plaza-content__description">{{ t('modelPlaza.description') }}</p>
    </div>

    <!-- 全局价格说明(管理员配置,Markdown) -->
    <div
      v-if="descriptionHtml"
      class="components-model-plaza-model-plaza-content__panel-2 plaza-description"
      v-html="descriptionHtml"
    ></div>

    <!-- 未登录提示 -->
    <p
      v-if="!isAuthenticated"
      class="components-model-plaza-model-plaza-content__description-2"
    >
      <Icon name="infoCircle" size="xs" class="components-model-plaza-model-plaza-content__icon" />
      {{ t('modelPlaza.anonymousHint') }}
    </p>

    <!-- 加载/错误/空 -->
    <div v-if="loading" class="components-model-plaza-model-plaza-content__panel-3">
      <div class="components-model-plaza-model-plaza-content__panel-4"></div>
    </div>
    <div
      v-else-if="error"
      class="components-model-plaza-model-plaza-content__panel-5"
    >
      {{ t('modelPlaza.loadFailed') }}
    </div>
    <template v-else>
      <!-- 筛选区:平台 → 分组 → 倍率 -->
      <PlazaFilterBar
        :platforms="platforms"
        :groups="groupOptions"
        :rates="rates"
        :platform="selectedPlatform"
        :group-id="selectedGroupId"
        :rate="selectedRate"
        :search="searchQuery"
        @update:platform="selectedPlatform = $event"
        @update:group-id="selectedGroupId = $event"
        @update:rate="selectedRate = $event"
        @update:search="searchQuery = $event"
      />

      <!-- 分组分节的模型清单(默认按生效倍率升序) -->
      <div v-if="filteredGroups.length > 0" class="components-model-plaza-model-plaza-content__panel">
        <PlazaGroupSection v-for="g in filteredGroups" :key="g.id" :group="g" />
      </div>
      <div
        v-else
        class="components-model-plaza-model-plaza-content__panel-6"
      >
        {{ searchActive ? t('modelPlaza.noSearchResult') : t('modelPlaza.empty') }}
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import Icon from '@/components/icons/Icon.vue'
import PlazaFilterBar from './PlazaFilterBar.vue'
import PlazaGroupSection from './PlazaGroupSection.vue'
import type { ModelPlazaGroup, ModelPlazaResponse } from '@/api/modelPlaza'
import { useAuthStore } from '@/stores/auth'

const props = defineProps<{
  response: ModelPlazaResponse | null
  loading: boolean
  error?: boolean
  /** 后台内嵌形态(AppLayout 内)。 */
  embedded?: boolean
}>()

const { t } = useI18n()
const authStore = useAuthStore()
const isAuthenticated = computed(() => authStore.isAuthenticated)

const selectedPlatform = ref<string>('all')
const selectedGroupId = ref<string | 'all'>('all')
const selectedRate = ref<number | 'all'>('all')
const searchQuery = ref('')

const searchActive = computed(() => searchQuery.value.trim() !== '')

const descriptionHtml = computed(() => {
  const md = props.response?.description?.trim()
  if (!md) return ''
  return DOMPurify.sanitize(marked.parse(md) as string)
})

/** 生效倍率 = 用户专属倍率 ?? 分组默认倍率。 */
function effectiveRate(g: ModelPlazaGroup): number {
  return g.user_rate_multiplier ?? g.rate_multiplier
}

const platforms = computed(() =>
  [...new Set((props.response?.groups ?? []).map((g) => g.platform).filter(Boolean))].sort()
)

const groupOptions = computed(() =>
  (props.response?.groups ?? []).map((g) => ({
    id: g.id,
    name: g.name,
    platform: g.platform,
    rate: effectiveRate(g)
  }))
)

/** 全量生效倍率;当前组合下不可用的项由 FilterBar 置灰而非隐藏。 */
const rates = computed(() =>
  [...new Set((props.response?.groups ?? []).map(effectiveRate))].sort((a, b) => a - b)
)

/** 数据刷新后选中的倍率可能不复存在,重置为全部。 */
watch(rates, (list) => {
  if (selectedRate.value !== 'all' && !list.includes(selectedRate.value)) {
    selectedRate.value = 'all'
  }
})

const filteredGroups = computed(() => {
  let groups = props.response?.groups ?? []
  if (selectedPlatform.value !== 'all') {
    groups = groups.filter((g) => g.platform === selectedPlatform.value)
  }
  if (selectedGroupId.value !== 'all') {
    groups = groups.filter((g) => g.id === selectedGroupId.value)
  }
  if (selectedRate.value !== 'all') {
    groups = groups.filter((g) => effectiveRate(g) === selectedRate.value)
  }
  // 模型名搜索:分组内只留命中的模型,整组无命中则隐藏该分组。
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    groups = groups
      .map((g) => ({ ...g, models: g.models.filter((m) => m.name.toLowerCase().includes(q)) }))
      .filter((g) => g.models.length > 0)
  }
  // 专属倍率会改变生效值,不能只依赖后端按默认倍率的排序。
  return [...groups].sort(
    (a, b) => effectiveRate(a) - effectiveRate(b) || a.name.localeCompare(b.name)
  )
})
</script>

<style scoped>
.plaza-description {
  line-height: 1.7;
  overflow-wrap: anywhere;
}

.plaza-description :deep(h1),
.plaza-description :deep(h2),
.plaza-description :deep(h3) {
  margin: 0.75rem 0 0.5rem;
  color: var(--color-text-primary);
  font-weight: 600;
}

.plaza-description :deep(h1:first-child),
.plaza-description :deep(h2:first-child),
.plaza-description :deep(h3:first-child) {
  margin-top: 0;
}

.plaza-description :deep(p) {
  margin-bottom: 0.5rem;
  color: var(--color-text-secondary);
}

.plaza-description :deep(p:last-child) {
  margin-bottom: 0;
}

.plaza-description :deep(a) {
  color: var(--color-primary);
  text-decoration: underline;
  text-underline-offset: 4px;
}

.plaza-description :deep(a:hover) {
  color: var(--color-primary-hover);
}

.plaza-description :deep(ul) {
  margin-bottom: 0.5rem;
  padding-left: 1.25rem;
  list-style-type: disc;
}

.plaza-description :deep(ol) {
  margin-bottom: 0.5rem;
  padding-left: 1.25rem;
  list-style-type: decimal;
}

.plaza-description :deep(li) {
  margin-bottom: 0.125rem;
  color: var(--color-text-secondary);
}

.plaza-description :deep(code) {
  padding: 0.125rem 0.375rem;
  border-radius: var(--radius-xs);
  background: var(--glass-bg-subtle);
  border: 1px solid var(--glass-border);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: var(--font-size-xs);
}

.plaza-description :deep(blockquote) {
  margin: 0.5rem 0;
  padding-left: 0.75rem;
  border-left: 4px solid var(--color-border-strong);
  color: var(--color-text-tertiary);
}
</style>

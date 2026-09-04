<template>
  <AppLayout>
    <div class="dashboard-page">
      <!-- 整页两栏:左 = 统计卡 + 图表 / 右 = 站内通知 + 消费等级 + 快捷操作 -->
      <div class="dashboard-page__grid">
        <div class="dashboard-page__main">
          <LoadingState v-if="loading" variant="section" class="dashboard-page__loading" />
          <div v-else-if="stats" class="dashboard-page__stats">
            <UserDashboardStats
              :stats="stats"
              :balance="user?.balance || 0"
              :is-simple="authStore.isSimpleMode"
            />
          </div>

          <UserDashboardApiKeyUsage
            v-if="stats"
            :trend="apiKeyUsageTrend"
            :key-count="stats.total_api_keys"
            :api-keys="apiKeys"
            :selected-api-key-id="selectedApiKeyId"
            :loading="loadingApiKeys"
            @select-key="handleApiKeySelection"
          />
        </div>
        <div class="dashboard-page__side">
          <UserDashboardAnnouncements />
          <UserDashboardLevel :profile="levelProfile" :loading="loadingLevel" />
          <UserDashboardQuickActions />
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { keysAPI, usageAPI, userLevelAPI } from '@/api'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'
import type { UserLevelDashboard } from '@/api/userLevel'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingState from '@/components/common/LoadingState.vue'
import UserDashboardAnnouncements from '@/components/user/dashboard/UserDashboardAnnouncements.vue'
import UserDashboardLevel from '@/components/user/dashboard/UserDashboardLevel.vue'
import UserDashboardStats from '@/components/user/dashboard/UserDashboardStats.vue'
import UserDashboardQuickActions from '@/components/user/dashboard/UserDashboardQuickActions.vue'
import UserDashboardApiKeyUsage from '@/components/user/dashboard/UserDashboardApiKeyUsage.vue'
import type { ApiKey, TrendDataPoint } from '@/types'
import { formatDateLocalInput } from '@/utils/format'

const authStore = useAuthStore()
const user = computed(() => authStore.user)
const stats = ref<UserStatsType | null>(null)
const apiKeyUsageTrend = ref<TrendDataPoint[]>([])
const apiKeys = ref<ApiKey[]>([])
const selectedApiKeyId = ref<string | null>(null)
const loading = ref(false)
const loadingApiKeys = ref(false)
const levelProfile = ref<UserLevelDashboard | null>(null)
const loadingLevel = ref(false)

const startDate = formatDateLocalInput(new Date(Date.now() - 6 * 86400000))
const endDate = formatDateLocalInput(new Date())
let apiKeyTrendRequestVersion = 0

async function loadStats() {
  loading.value = true
  try {
    await authStore.refreshUser()
    stats.value = await usageAPI.getDashboardStats()
  } catch (error) {
    console.error('Failed to load dashboard stats:', error)
  } finally {
    loading.value = false
  }
}

async function loadApiKeys() {
  try {
    const response = await keysAPI.list(1, 1000, {
      sort_by: 'name',
      sort_order: 'asc',
    })
    apiKeys.value = response.items ?? []
    if (selectedApiKeyId.value && !apiKeys.value.some((apiKey) => apiKey.id === selectedApiKeyId.value)) {
      selectedApiKeyId.value = null
    }
  } catch (error) {
    console.error('Failed to load API keys for dashboard filter:', error)
  }
}

async function loadApiKeyUsage(apiKeyId = selectedApiKeyId.value) {
  const requestVersion = ++apiKeyTrendRequestVersion
  loadingApiKeys.value = true
  try {
    const response = await usageAPI.getDashboardTrend({
      start_date: startDate,
      end_date: endDate,
      granularity: 'day',
      api_key_id: apiKeyId || undefined,
    })
    if (requestVersion === apiKeyTrendRequestVersion) {
      apiKeyUsageTrend.value = response.trend ?? []
    }
  } catch (error) {
    if (requestVersion === apiKeyTrendRequestVersion) {
      console.error('Failed to load API key usage:', error)
      apiKeyUsageTrend.value = []
    }
  } finally {
    if (requestVersion === apiKeyTrendRequestVersion) {
      loadingApiKeys.value = false
    }
  }
}

function handleApiKeySelection(apiKeyId: string | null) {
  selectedApiKeyId.value = apiKeyId
  void loadApiKeyUsage(apiKeyId)
}

async function loadLevel() {
  loadingLevel.value = true
  try {
    levelProfile.value = await userLevelAPI.getCurrent()
  } catch (error) {
    console.error('Failed to load user level:', error)
    levelProfile.value = null
  } finally {
    loadingLevel.value = false
  }
}

function refreshAll() {
  void loadStats()
  void loadApiKeys()
  void loadApiKeyUsage()
  void loadLevel()
}

onMounted(refreshAll)
</script>

<style scoped>
/* 紧凑仪表盘:目标是常规桌面分辨率下一屏展示 */
.dashboard-page {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.dashboard-page__loading {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2.5rem 0;
}

/* ---- 统计卡两行(每行 3 块):行距与卡内边距压缩 ---- */
.dashboard-page__stats {
  display: grid;
  gap: 0.75rem;
}

.dashboard-page__stats :deep(.components-user-dashboard-user-dashboard-stats__panel) {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.75rem;
}

@media (min-width: 640px) {
  .dashboard-page__stats :deep(.components-user-dashboard-user-dashboard-stats__panel) {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 1024px) {
  .dashboard-page__stats :deep(.components-user-dashboard-user-dashboard-stats__panel) {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

.dashboard-page__stats :deep(.components-user-dashboard-user-dashboard-stats__panel-2) {
  padding: 0.75rem 0.875rem;
}

.dashboard-page__stats :deep(.components-user-dashboard-user-dashboard-stats__panel-3) {
  gap: 0.625rem;
}

/* 图标容器 3rem → 2.25rem */
.dashboard-page__stats :deep(.components-user-dashboard-user-dashboard-stats__panel-3 > div:first-child) {
  width: 2.25rem;
  height: 2.25rem;
}

.dashboard-page__stats :deep(.components-user-dashboard-user-dashboard-stats__description-2),
.dashboard-page__stats :deep(.components-user-dashboard-user-dashboard-stats__description-4) {
  font-size: var(--font-size-lg);
  line-height: 1.5rem;
}

/* ---- 整页两栏:左统计 + 图表(2/3),右侧栏(1/3) ---- */
.dashboard-page__grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 1rem;
  align-items: start;
}

@media (min-width: 1280px) {
  .dashboard-page__grid {
    grid-template-columns: minmax(0, 2fr) minmax(0, 1fr);
  }
}

.dashboard-page__main {
  display: flex;
  flex-direction: column;
  min-width: 0;
  gap: 1rem;
}

.dashboard-page__side {
  display: flex;
  flex-direction: column;
  min-width: 0;
  gap: 1rem;
}

/* ---- 图表卡:压缩头部/统计条/图表高度 ---- */
.dashboard-page__main :deep(.dashboard-key-usage__header) {
  padding: 0.75rem 1.25rem;
}

.dashboard-page__main :deep(.dashboard-key-usage__total) {
  padding: 0.625rem 1.25rem;
}

.dashboard-page__main :deep(.dashboard-key-usage__key-filter) {
  padding: 0.75rem 1.25rem 0;
}

.dashboard-page__main :deep(.dashboard-key-usage__chart-toolbar) {
  padding: 0.75rem 1.25rem 0;
}

.dashboard-page__main :deep(.dashboard-key-usage__chart) {
  height: 15rem;
  padding: 0.75rem 1.25rem 1rem;
}

/* ---- 右栏:等级卡紧凑(窄栏内两列摘要) ---- */
.dashboard-page__side :deep(.dashboard-level__header) {
  padding: 0.75rem 1.25rem;
}

.dashboard-page__side :deep(.dashboard-level__body) {
  gap: 0.5rem;
  padding: 0.875rem 1.25rem 1rem;
}

/* ---- 右栏:公告卡改单列紧凑列表 ---- */
.dashboard-page__side :deep(.dashboard-announcements__header) {
  padding: 0.75rem 1.25rem;
}

.dashboard-page__side :deep(.dashboard-announcements__list) {
  grid-template-columns: minmax(0, 1fr);
}

.dashboard-page__side :deep(.dashboard-announcements__item) {
  min-height: 3rem;
  padding: 0.5rem 1.25rem;
  border-right: 0;
}

.dashboard-page__side :deep(.dashboard-announcements__item + .dashboard-announcements__item) {
  border-top: 1px solid var(--color-border-subtle);
}

.dashboard-page__side :deep(.dashboard-announcements__state) {
  min-height: 4rem;
  padding: 0.75rem;
}
</style>

<template>
  <div class="views-key-usage-view__panel">
    <!-- Header (same pattern as HomeView) -->
    <header class="views-key-usage-view__header">
      <nav class="views-key-usage-view__navigation">
        <router-link to="/home" class="views-key-usage-view__router-link">
          <div class="views-key-usage-view__panel-2">
            <img :src="siteLogo || '/logo.svg'" alt="Logo" class="views-key-usage-view__image" />
          </div>
          <span class="views-key-usage-view__text">{{ siteName }}</span>
        </router-link>
        <div class="views-key-usage-view__router-link">
          <LocaleSwitcher />
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="views-key-usage-view__link"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>
          <button
            @click="toggleTheme"
            class="views-key-usage-view__link"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>
        </div>
      </nav>
    </header>

    <!-- Main Content -->
    <main class="views-key-usage-view__main">
      <!-- Hero -->
      <div class="views-key-usage-view__panel-3">
        <h1 class="views-key-usage-view__heading">
          {{ t('keyUsage.title') }}
        </h1>
        <p class="views-key-usage-view__description">
          {{ t('keyUsage.subtitle') }}
        </p>
      </div>

      <!-- Input Section -->
      <div class="views-key-usage-view__panel-4">
        <div class="views-key-usage-view__panel-5">
          <div class="views-key-usage-view__panel-6">
            <div class="views-key-usage-view__panel-7">
              <svg class="views-key-usage-view__icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
              </svg>
            </div>
            <input
              v-model="apiKey"
              :type="keyVisible ? 'text' : 'password'"
              :placeholder="t('keyUsage.placeholder')"
              class="views-key-usage-view__field input-ring"
              @keydown.enter="queryKey"
            />
            <button
              @click="keyVisible = !keyVisible"
              class="views-key-usage-view__action"
            >
              <svg v-if="!keyVisible" class="views-key-usage-view__icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/>
                <line x1="1" y1="1" x2="23" y2="23"/>
              </svg>
              <svg v-else class="views-key-usage-view__icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>
              </svg>
            </button>
          </div>
          <button
            @click="queryKey"
            :disabled="isQuerying"
            class="views-key-usage-view__action-2"
          >
            <svg v-if="isQuerying" class="views-key-usage-view__icon-2" viewBox="0 0 24 24" fill="none">
              <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3" opacity="0.25"/>
              <path d="M12 2a10 10 0 0 1 10 10" stroke="currentColor" stroke-width="3" stroke-linecap="round"/>
            </svg>
            <svg v-else class="views-key-usage-view__icon-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
            </svg>
            {{ isQuerying ? t('keyUsage.querying') : t('keyUsage.query') }}
          </button>
        </div>
        <p class="views-key-usage-view__description-2">
          {{ t('keyUsage.privacyNote') }}
        </p>

        <!-- Date Range Picker -->
        <div v-if="showDatePicker" class="views-key-usage-view__panel-8">
          <div class="views-key-usage-view__panel-9">
            <span class="views-key-usage-view__text-2">{{ t('keyUsage.dateRange') }}</span>
            <button
              v-for="range in dateRanges"
              :key="range.key"
              @click="setDateRange(range.key)"
              class="views-key-usage-view__action-3"
              :class="currentRange === range.key
                ? 'views-key-usage-view__action-6'
                : 'views-key-usage-view__action-7'"
            >{{ range.label }}</button>
            <div v-if="currentRange === 'custom'" class="views-key-usage-view__panel-10">
              <input
                v-model="customStartDate"
                type="date"
                class="views-key-usage-view__field-2 input-ring"
              />
              <span class="views-key-usage-view__text-3">-</span>
              <input
                v-model="customEndDate"
                type="date"
                class="views-key-usage-view__field-2 input-ring"
              />
              <button
                @click="queryKey"
                class="views-key-usage-view__action-4"
              >{{ t('keyUsage.apply') }}</button>
            </div>
          </div>
        </div>
      </div>

      <!-- Results Container -->
      <div v-if="showResults">
        <!-- Loading Skeleton -->
        <div v-if="showLoading" class="views-key-usage-view__panel-11">
          <div class="views-key-usage-view__panel-12">
            <div class="views-key-usage-view__panel-13">
              <div class="views-key-usage-view__panel-14 skeleton"></div>
              <div class="views-key-usage-view__panel-15"><div class="views-key-usage-view__panel-16 skeleton"></div></div>
            </div>
            <div class="views-key-usage-view__panel-13">
              <div class="views-key-usage-view__panel-14 skeleton"></div>
              <div class="views-key-usage-view__panel-15"><div class="views-key-usage-view__panel-16 skeleton"></div></div>
            </div>
          </div>
          <div class="views-key-usage-view__panel-13">
            <div class="views-key-usage-view__panel-17 skeleton"></div>
            <div class="views-key-usage-view__panel-18">
              <div class="views-key-usage-view__panel-19 skeleton"></div>
              <div class="views-key-usage-view__panel-20 skeleton"></div>
              <div class="views-key-usage-view__panel-21 skeleton"></div>
              <div class="views-key-usage-view__panel-22 skeleton"></div>
            </div>
          </div>
        </div>

        <!-- Result Content -->
        <div v-else-if="resultData" class="views-key-usage-view__panel-11">
          <!-- Status Badge -->
          <div v-if="statusInfo" class="views-key-usage-view__panel-23 fade-up">
            <div class="views-key-usage-view__panel-24">
              <span
                class="views-key-usage-view__text-4 pulse-dot"
                :class="statusInfo.isActive ? 'views-key-usage-view__text-13' : 'views-key-usage-view__text-14'"
              ></span>
              <span class="views-key-usage-view__text-5">{{ statusInfo.label }}</span>
              <span class="views-key-usage-view__text-6">|</span>
              <span class="views-key-usage-view__text-2">{{ statusInfo.statusText }}</span>
            </div>
          </div>

          <!-- Ring Cards Grid -->
          <div v-if="ringItems.length > 0" :class="ringGridClass">
            <div
              v-for="(ring, i) in ringItems"
              :key="i"
              class="views-key-usage-view__panel-25 fade-up"
              :class="`fade-up-delay-${Math.min(i + 1, 4)}`"
            >
              <div class="views-key-usage-view__panel-26">
                <h3 class="views-key-usage-view__heading-2">
                  {{ ring.title }}
                </h3>
                <!-- Clock icon -->
                <svg v-if="ring.iconType === 'clock'" class="views-key-usage-view__icon-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
                </svg>
                <!-- Calendar icon -->
                <svg v-else-if="ring.iconType === 'calendar'" class="views-key-usage-view__icon-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/>
                </svg>
                <!-- Platform points icon -->
                <svg v-else class="views-key-usage-view__icon-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="m12 3 1.8 5.2L19 10l-5.2 1.8L12 17l-1.8-5.2L5 10l5.2-1.8L12 3Z"/><path d="m19 16 .7 2.3L22 19l-2.3.7L19 22l-.7-2.3L16 19l2.3-.7L19 16Z"/>
                </svg>
              </div>
              <div class="views-key-usage-view__panel-15">
                <div class="views-key-usage-view__panel-27">
                  <svg class="views-key-usage-view__icon-5" viewBox="0 0 160 160">
                    <circle cx="80" cy="80" r="68" fill="none" :stroke="ringTrackColor" stroke-width="10"/>
                    <circle
                      class="progress-ring"
                      cx="80" cy="80" r="68" fill="none"
                      :stroke="`url(#ring-grad-${i})`"
                      stroke-width="10" stroke-linecap="round"
                      :stroke-dasharray="CIRCUMFERENCE.toFixed(2)"
                      :stroke-dashoffset="getRingOffset(ring)"
                    />
                    <defs>
                      <linearGradient :id="`ring-grad-${i}`" x1="0%" y1="0%" x2="100%" y2="100%">
                        <stop offset="0%" :stop-color="RING_GRADIENTS[i % 4].from"/>
                        <stop offset="100%" :stop-color="RING_GRADIENTS[i % 4].to"/>
                      </linearGradient>
                    </defs>
                  </svg>
                  <div class="views-key-usage-view__panel-28">
                    <template v-if="ring.isBalance">
                      <span class="views-key-usage-view__text-7" :style="{ color: RING_GRADIENTS[i % 4].from }">
                        {{ ring.amount }}
                      </span>
                    </template>
                    <template v-else>
                      <span class="views-key-usage-view__text-8">
                        {{ displayPcts[i] ?? 0 }}%
                      </span>
                      <span class="views-key-usage-view__text-9">{{ t('keyUsage.used') }}</span>
                      <span
                        class="views-key-usage-view__text-10"
                        :style="{ color: RING_GRADIENTS[i % 4].from }"
                      >{{ ring.amount }}</span>
                      <p v-if="ring.resetAt && formatResetTime(ring.resetAt)" class="views-key-usage-view__description-3">
                        ⟳ {{ formatResetTime(ring.resetAt) }}
                      </p>
                    </template>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Detail Card -->
          <div
            v-if="detailRows.length > 0"
            class="views-key-usage-view__panel-29 fade-up fade-up-delay-3"
          >
            <div class="views-key-usage-view__panel-30">
              <h3 class="views-key-usage-view__heading-2">{{ t('keyUsage.detailInfo') }}</h3>
            </div>
            <div class="views-key-usage-view__panel-31">
              <div
                v-for="(row, i) in detailRows"
                :key="i"
                class="views-key-usage-view__panel-32"
              >
                <div class="views-key-usage-view__router-link">
                  <div class="views-key-usage-view__panel-33" :class="row.iconBg">
                    <svg
                      class="views-key-usage-view__icon-3"
                      :class="row.iconColor"
                      viewBox="0 0 24 24" fill="none" stroke="currentColor"
                      stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                      v-html="row.iconSvg"
                    ></svg>
                  </div>
                  <span class="views-key-usage-view__text-11">{{ row.label }}</span>
                </div>
                <span class="views-key-usage-view__text-12" :class="row.valueClass || 'views-key-usage-view__text-15'">
                  {{ row.value }}
                </span>
              </div>
            </div>
          </div>

          <!-- Usage Stats Card -->
          <div
            v-if="usageStatCells.length > 0"
            class="views-key-usage-view__panel-29 fade-up fade-up-delay-3"
          >
            <div class="views-key-usage-view__panel-30">
              <h3 class="views-key-usage-view__heading-2">{{ t('keyUsage.tokenStats') }}</h3>
            </div>
            <div class="views-key-usage-view__panel-34">
              <div
                v-for="(cell, i) in usageStatCells"
                :key="i"
                class="views-key-usage-view__panel-35"
              >
                <div class="views-key-usage-view__panel-36">{{ cell.label }}</div>
                <div class="views-key-usage-view__panel-37">{{ cell.value }}</div>
              </div>
            </div>
          </div>

          <!-- Daily Usage Table -->
          <div
            v-if="showDailyUsage"
            class="views-key-usage-view__panel-29 fade-up fade-up-delay-4"
          >
            <div class="views-key-usage-view__panel-38">
              <h3 class="views-key-usage-view__heading-2">{{ t('keyUsage.dailyDetail') }}</h3>
              <div class="views-key-usage-view__panel-39">
                <button
                  v-for="option in dailyUsageOptions"
                  :key="option.value"
                  @click="setDailyUsageDays(option.value)"
                  class="views-key-usage-view__action-5"
                  :class="dailyUsageDays === option.value
                    ? 'views-key-usage-view__action-8'
                    : 'views-key-usage-view__action-9'"
                >
                  {{ option.label }}
                </button>
              </div>
            </div>
            <div v-if="dailyUsageRows.length > 0" class="views-key-usage-view__panel-40">
              <table class="views-key-usage-view__table">
                <thead>
                  <tr class="views-key-usage-view__row">
                    <th class="views-key-usage-view__heading-3">{{ t('keyUsage.date') }}</th>
                    <th class="views-key-usage-view__heading-4">{{ t('keyUsage.requests') }}</th>
                    <th class="views-key-usage-view__heading-4">{{ t('keyUsage.inputTokens') }}</th>
                    <th class="views-key-usage-view__heading-4">{{ t('keyUsage.outputTokens') }}</th>
                    <th class="views-key-usage-view__heading-4">{{ t('keyUsage.cacheReadTokens') }}</th>
                    <th class="views-key-usage-view__heading-4">{{ t('keyUsage.cacheWriteTokens') }}</th>
                    <th class="views-key-usage-view__heading-4">{{ t('keyUsage.cost') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="row in dailyUsageRows"
                    :key="row.date"
                    class="views-key-usage-view__row-2"
                  >
                    <td class="views-key-usage-view__cell">{{ row.date }}</td>
                    <td class="views-key-usage-view__cell-2">{{ fmtNum(row.requests) }}</td>
                    <td class="views-key-usage-view__cell-2">{{ fmtNum(row.input_tokens) }}</td>
                    <td class="views-key-usage-view__cell-2">{{ fmtNum(row.output_tokens) }}</td>
                    <td class="views-key-usage-view__cell-2">{{ fmtNum(row.cache_read_tokens) }}</td>
                    <td class="views-key-usage-view__cell-2">{{ fmtNum(row.cache_write_tokens) }}</td>
                    <td class="views-key-usage-view__cell-3">{{ points(row.actual_cost) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-else class="views-key-usage-view__panel-41">
              {{ t('keyUsage.noDailyUsage') }}
            </div>
          </div>

          <!-- Model Stats Table -->
          <div
            v-if="modelStats.length > 0"
            class="views-key-usage-view__panel-29 fade-up fade-up-delay-4"
          >
            <div class="views-key-usage-view__panel-30">
              <h3 class="views-key-usage-view__heading-2">{{ t('keyUsage.modelStats') }}</h3>
            </div>
            <div class="views-key-usage-view__panel-40">
              <table class="views-key-usage-view__table">
                <thead>
                  <tr class="views-key-usage-view__row">
                    <th class="views-key-usage-view__heading-3">{{ t('keyUsage.model') }}</th>
                    <th class="views-key-usage-view__heading-4">{{ t('keyUsage.requests') }}</th>
                    <th class="views-key-usage-view__heading-4">{{ t('keyUsage.inputTokens') }}</th>
                    <th class="views-key-usage-view__heading-4">{{ t('keyUsage.outputTokens') }}</th>
                    <th class="views-key-usage-view__heading-4">{{ t('keyUsage.cacheCreationTokens') }}</th>
                    <th class="views-key-usage-view__heading-4">{{ t('keyUsage.cacheReadTokens') }}</th>
                    <th class="views-key-usage-view__heading-4">{{ t('keyUsage.totalTokens') }}</th>
                    <th class="views-key-usage-view__heading-4">{{ t('keyUsage.cost') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="(m, i) in modelStats"
                    :key="i"
                    class="views-key-usage-view__row-2"
                  >
                    <td class="views-key-usage-view__cell">{{ m.model || '-' }}</td>
                    <td class="views-key-usage-view__cell-2">{{ fmtNum(m.requests) }}</td>
                    <td class="views-key-usage-view__cell-2">{{ fmtNum(m.input_tokens) }}</td>
                    <td class="views-key-usage-view__cell-2">{{ fmtNum(m.output_tokens) }}</td>
                    <td class="views-key-usage-view__cell-2">{{ fmtNum(m.cache_creation_tokens) }}</td>
                    <td class="views-key-usage-view__cell-2">{{ fmtNum(m.cache_read_tokens) }}</td>
                    <td class="views-key-usage-view__cell-2">{{ fmtNum(m.total_tokens) }}</td>
                    <td class="views-key-usage-view__cell-3">{{ points(m.actual_cost) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Footer (same pattern as HomeView) -->
    <footer class="views-key-usage-view__footer">
      <div class="views-key-usage-view__panel-42">
        <p class="views-key-usage-view__description-4">
          &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
        </p>
        <div class="views-key-usage-view__panel-43">
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="views-key-usage-view__link-2"
          >{{ t('home.docs') }}</a>
          <a
            :href="githubUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="views-key-usage-view__link-2"
          >GitHub</a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { buildGatewayUrl } from '@/api/client'
import { formatDateLocalInput, formatPointAmount, formatPoints } from '@/utils/format'
import { sanitizeUrl } from '@/utils/url'

const { t, locale } = useI18n()
const appStore = useAppStore()

// ==================== Site Settings (same as HomeView) ====================

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'EasySub2api')
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const githubUrl = 'https://github.com/AsukaCC/EasySub2api'

// ==================== Theme (same as HomeView) ====================

const isDark = ref(document.documentElement.classList.contains('dark'))

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

const currentYear = computed(() => new Date().getFullYear())

// ==================== Key Query State ====================

const apiKey = ref('')
const keyVisible = ref(false)
const isQuerying = ref(false)
const showResults = ref(false)
const showLoading = ref(false)
const showDatePicker = ref(false)
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const resultData = ref<any>(null)
const now = ref(new Date())
let resetTimer: ReturnType<typeof setInterval> | null = null

// ==================== Date Range State ====================

type DateRangeKey = 'today' | '7d' | '30d' | 'custom'
const currentRange = ref<DateRangeKey>('today')
const customStartDate = ref('')
const customEndDate = ref('')
const dailyUsageDays = ref<7 | 30 | 90>(30)

const dateRanges = computed(() => [
  { key: 'today' as const, label: t('keyUsage.dateRangeToday') },
  { key: '7d' as const, label: t('keyUsage.dateRange7d') },
  { key: '30d' as const, label: t('keyUsage.dateRange30d') },
  { key: 'custom' as const, label: t('keyUsage.dateRangeCustom') },
])

const dailyUsageOptions = computed(() => [
  { value: 7 as const, label: t('keyUsage.dateRange7d') },
  { value: 30 as const, label: t('keyUsage.dateRange30d') },
  { value: 90 as const, label: t('keyUsage.dateRange90d') },
])

function setDateRange(key: DateRangeKey) {
  currentRange.value = key
  if (key !== 'custom') {
    queryKey()
  }
}

function getDateParams(): string {
  const now = new Date()
  const params = new URLSearchParams()

  if (currentRange.value === 'custom') {
    if (customStartDate.value && customEndDate.value) {
      params.set('start_date', customStartDate.value)
      params.set('end_date', customEndDate.value)
    }
  } else {
    const end = formatDateLocalInput(now)
    let start: string
    switch (currentRange.value) {
      case 'today': start = end; break
      case '7d': start = formatDateLocalInput(new Date(now.getTime() - 7 * 86400000)); break
      case '30d': start = formatDateLocalInput(new Date(now.getTime() - 30 * 86400000)); break
      default: start = formatDateLocalInput(new Date(now.getTime() - 30 * 86400000))
    }
    params.set('start_date', start)
    params.set('end_date', end)
  }
  params.set('days', String(dailyUsageDays.value))
  params.set('timezone', getBrowserTimezone())
  return params.toString()
}

function setDailyUsageDays(days: 7 | 30 | 90) {
  if (dailyUsageDays.value === days) return
  dailyUsageDays.value = days
  if (resultData.value && apiKey.value.trim()) {
    queryKey()
  }
}

// ==================== Ring Animation ====================

const CIRCUMFERENCE = 2 * Math.PI * 68
const RING_GRADIENTS = [
  { from: '#0a84ff', to: '#7cc2ff' },
  { from: '#6366F1', to: '#A5B4FC' },
  { from: '#10B981', to: '#6EE7B7' },
  { from: '#F59E0B', to: '#FCD34D' },
]

const ringAnimated = ref(false)
const displayPcts = ref<number[]>([])

const ringTrackColor = computed(() => isDark.value ? '#222222' : '#F0F0EE')

interface RingItem {
  title: string
  pct: number
  amount: string
  isBalance?: boolean
  iconType: 'clock' | 'calendar' | 'points'
  resetAt?: string | null
}

function getRingOffset(ring: RingItem): number {
  if (!ringAnimated.value) return CIRCUMFERENCE
  if (ring.isBalance) return 0
  return CIRCUMFERENCE - (Math.min(ring.pct, 100) / 100) * CIRCUMFERENCE
}

function triggerRingAnimation(items: RingItem[]) {
  ringAnimated.value = false
  displayPcts.value = items.map(() => 0)

  nextTick(() => {
    requestAnimationFrame(() => {
      setTimeout(() => {
        ringAnimated.value = true

        // Animate percentage numbers
        const duration = 1000
        const startTime = performance.now()
        const targets = items.map(item => item.isBalance ? 0 : item.pct)

        function tick() {
          const elapsed = performance.now() - startTime
          const p = Math.min(elapsed / duration, 1)
          const ease = 1 - Math.pow(1 - p, 3)
          displayPcts.value = targets.map(target => Math.round(ease * target))
          if (p < 1) requestAnimationFrame(tick)
        }
        requestAnimationFrame(tick)
      }, 50)
    })
  })
}

// ==================== Computed Data ====================

const statusInfo = computed(() => {
  const data = resultData.value
  if (!data) return null

  if (data.mode === 'quota_limited') {
    const isValid = data.isValid !== false
    const statusMap: Record<string, string> = {
      active: 'Active',
      quota_exhausted: 'Quota Exhausted',
      expired: 'Expired',
    }
    return {
      label: t('keyUsage.quotaMode'),
      statusText: statusMap[data.status] || data.status || 'Unknown',
      isActive: isValid && data.status === 'active',
    }
  }

  return {
    label: data.planName || t('keyUsage.walletBalance'),
    statusText: 'Active',
    isActive: true,
  }
})

const ringItems = computed<RingItem[]>(() => {
  const data = resultData.value
  if (!data) return []

  const items: RingItem[] = []

  if (data.mode === 'quota_limited') {
    if (data.quota) {
      const pct = data.quota.limit > 0 ? Math.min(Math.round((data.quota.used / data.quota.limit) * 100), 100) : 0
      items.push({ title: t('keyUsage.totalQuota'), pct, amount: pointRange(data.quota.used, data.quota.limit), iconType: 'points' })
    }
    if (data.rate_limits) {
      const windowLabels: Record<string, string> = { '5h': t('keyUsage.limit5h'), '1d': t('keyUsage.limitDaily'), '7d': t('keyUsage.limit7d') }
      const windowIcons: Record<string, 'clock' | 'calendar'> = { '5h': 'clock', '1d': 'calendar', '7d': 'calendar' }
      for (const rl of data.rate_limits) {
        const pct = rl.limit > 0 ? Math.min(Math.round((rl.used / rl.limit) * 100), 100) : 0
        items.push({
          title: windowLabels[rl.window] || rl.window,
          pct,
          amount: pointRange(rl.used, rl.limit),
          iconType: windowIcons[rl.window] || 'clock',
          resetAt: rl.reset_at,
        })
      }
    }
  } else {
    if (data.subscription) {
      const sub = data.subscription
      const limits = [
        { label: t('keyUsage.limitDaily'), usage: sub.daily_usage_points ?? sub.daily_usage_usd, limit: sub.daily_limit_points ?? sub.daily_limit_usd },
        { label: t('keyUsage.limitWeekly'), usage: sub.weekly_usage_points ?? sub.weekly_usage_usd, limit: sub.weekly_limit_points ?? sub.weekly_limit_usd },
        { label: t('keyUsage.limitMonthly'), usage: sub.monthly_usage_points ?? sub.monthly_usage_usd, limit: sub.monthly_limit_points ?? sub.monthly_limit_usd },
      ]
      for (const l of limits) {
        if (l.limit != null && l.limit > 0) {
          const pct = Math.min(Math.round((l.usage / l.limit) * 100), 100)
          items.push({ title: l.label, pct, amount: pointRange(l.usage, l.limit), iconType: 'calendar' })
        }
      }
    }
    if (!data.subscription && data.balance != null) {
      items.push({ title: t('keyUsage.walletBalance'), pct: 0, amount: points(data.balance), isBalance: true, iconType: 'points' })
    }
  }

  return items
})

const ringGridClass = computed(() => {
  const len = ringItems.value.length
  if (len === 1) return 'views-key-usage-view__state'
  if (len === 2) return 'views-key-usage-view__panel-12'
  return 'views-key-usage-view__state-2'
})

interface DetailRow {
  iconBg: string
  iconColor: string
  iconSvg: string
  label: string
  value: string
  valueClass: string
}

function getUsageColor(pct: number): string {
  if (pct > 90) return 'status-text--danger'
  if (pct > 70) return 'status-text--warning'
  return 'status-text--success'
}

const detailRows = computed<DetailRow[]>(() => {
  const data = resultData.value
  if (!data) return []

  const rows: DetailRow[] = []
  const ICON_SHIELD = '<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>'
  const ICON_CALENDAR = '<rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/>'
  const ICON_POINTS = '<path d="m12 3 1.8 5.2L19 10l-5.2 1.8L12 17l-1.8-5.2L5 10l5.2-1.8L12 3Z"/><path d="m19 16 .7 2.3L22 19l-2.3.7L19 22l-.7-2.3L16 19l2.3-.7L19 16Z"/>'
  const ICON_CHECK = '<polyline points="20 6 9 17 4 12"/>'

  if (data.mode === 'quota_limited') {
    if (data.quota) {
      const remainColor = data.quota.remaining <= 0 ? 'status-text--danger'
        : data.quota.remaining < data.quota.limit * 0.1 ? 'status-text--warning'
        : 'status-text--success'
      rows.push({
        iconBg: 'status-fill--success-subtle', iconColor: 'status-text--success', iconSvg: ICON_SHIELD,
        label: t('keyUsage.remainingQuota'), value: points(data.quota.remaining), valueClass: remainColor,
      })
    }
    if (data.expires_at) {
      const daysLeft = data.days_until_expiry
      let expiryStr = formatDate(data.expires_at)
      if (daysLeft != null) {
        expiryStr += daysLeft > 0 ? ` ${t('keyUsage.daysLeft', { days: daysLeft })}` : daysLeft === 0 ? ` ${t('keyUsage.todayExpires')}` : ''
      }
      rows.push({
        iconBg: 'status-fill--warning-subtle', iconColor: 'status-text--warning', iconSvg: ICON_CALENDAR,
        label: t('keyUsage.expiresAt'), value: expiryStr, valueClass: '',
      })
    }
    if (data.rate_limits) {
      const windowMap: Record<string, string> = { '5h': '5H', '1d': locale.value === 'zh' ? '日' : 'D', '7d': '7D' }
      for (const rl of data.rate_limits) {
        const pct = rl.limit > 0 ? (rl.used / rl.limit) * 100 : 0
        let valueStr = pointRange(rl.used, rl.limit)
        const resetStr = formatResetTime(rl.reset_at)
        if (resetStr) {
          valueStr += ` (⟳ ${resetStr})`
        }
        rows.push({
          iconBg: 'status-fill--neutral-subtle', iconColor: 'status-text--neutral', iconSvg: ICON_POINTS,
          label: `${t('keyUsage.usedQuota')} (${windowMap[rl.window] || rl.window})`,
          value: valueStr,
          valueClass: getUsageColor(pct),
        })
      }
    }
  } else {
    rows.push({
      iconBg: 'status-fill--success-subtle', iconColor: 'status-text--success', iconSvg: ICON_CHECK,
      label: t('keyUsage.subscriptionType'), value: data.planName || t('keyUsage.walletBalance'), valueClass: '',
    })

    if (data.subscription) {
      const sub = data.subscription
      const dailyLimit = sub.daily_limit_points ?? sub.daily_limit_usd
      const dailyUsage = sub.daily_usage_points ?? sub.daily_usage_usd
      if (dailyLimit > 0) {
        const pct = (dailyUsage / dailyLimit) * 100
        rows.push({
          iconBg: 'status-fill--neutral-subtle', iconColor: 'status-text--neutral', iconSvg: ICON_POINTS,
          label: `${t('keyUsage.usedQuota')} (${locale.value === 'zh' ? '日' : 'D'})`, value: pointRange(dailyUsage, dailyLimit), valueClass: getUsageColor(pct),
        })
      }
      const weeklyLimit = sub.weekly_limit_points ?? sub.weekly_limit_usd
      const weeklyUsage = sub.weekly_usage_points ?? sub.weekly_usage_usd
      if (weeklyLimit > 0) {
        const pct = (weeklyUsage / weeklyLimit) * 100
        rows.push({
          iconBg: 'status-fill--info-subtle', iconColor: 'status-text--info', iconSvg: ICON_POINTS,
          label: `${t('keyUsage.usedQuota')} (${locale.value === 'zh' ? '周' : 'W'})`, value: pointRange(weeklyUsage, weeklyLimit), valueClass: getUsageColor(pct),
        })
      }
      const monthlyLimit = sub.monthly_limit_points ?? sub.monthly_limit_usd
      const monthlyUsage = sub.monthly_usage_points ?? sub.monthly_usage_usd
      if (monthlyLimit > 0) {
        const pct = (monthlyUsage / monthlyLimit) * 100
        rows.push({
          iconBg: 'status-fill--success-subtle', iconColor: 'status-text--success', iconSvg: ICON_POINTS,
          label: `${t('keyUsage.usedQuota')} (${locale.value === 'zh' ? '月' : 'M'})`, value: pointRange(monthlyUsage, monthlyLimit), valueClass: getUsageColor(pct),
        })
      }
      if (sub.expires_at) {
        rows.push({
          iconBg: 'status-fill--warning-subtle', iconColor: 'status-text--warning', iconSvg: ICON_CALENDAR,
          label: t('keyUsage.subscriptionExpires'), value: formatDate(sub.expires_at), valueClass: '',
        })
      }
    }

    const remainColor = data.remaining != null
      ? (data.remaining <= 0 ? 'status-text--danger' : data.remaining < 10 ? 'status-text--warning' : 'status-text--success')
      : ''
    rows.push({
      iconBg: 'status-fill--success-subtle', iconColor: 'status-text--success', iconSvg: ICON_SHIELD,
      label: t('keyUsage.remainingQuota'), value: data.remaining != null ? points(data.remaining) : '-', valueClass: remainColor,
    })
  }

  return rows
})

interface StatCell {
  label: string
  value: string
}

const usageStatCells = computed<StatCell[]>(() => {
  const usage = resultData.value?.usage
  if (!usage) return []

  const today = usage.today || {}
  const total = usage.total || {}

  return [
    { label: t('keyUsage.todayRequests'), value: fmtNum(today.requests) },
    { label: t('keyUsage.todayInputTokens'), value: fmtNum(today.input_tokens) },
    { label: t('keyUsage.todayOutputTokens'), value: fmtNum(today.output_tokens) },
    { label: t('keyUsage.todayTokens'), value: fmtNum(today.total_tokens) },
    { label: t('keyUsage.todayCacheCreation'), value: fmtNum(today.cache_creation_tokens) },
    { label: t('keyUsage.todayCacheRead'), value: fmtNum(today.cache_read_tokens) },
    { label: t('keyUsage.todayCost'), value: points(today.actual_cost) },
    { label: t('keyUsage.rpmTpm'), value: `${usage.rpm || 0} / ${usage.tpm || 0}` },
    { label: t('keyUsage.totalRequests'), value: fmtNum(total.requests) },
    { label: t('keyUsage.totalInputTokens'), value: fmtNum(total.input_tokens) },
    { label: t('keyUsage.totalOutputTokens'), value: fmtNum(total.output_tokens) },
    { label: t('keyUsage.totalTokensLabel'), value: fmtNum(total.total_tokens) },
    { label: t('keyUsage.totalCacheCreation'), value: fmtNum(total.cache_creation_tokens) },
    { label: t('keyUsage.totalCacheRead'), value: fmtNum(total.cache_read_tokens) },
    { label: t('keyUsage.totalCost'), value: points(total.actual_cost) },
    { label: t('keyUsage.avgDuration'), value: usage.average_duration_ms ? `${Math.round(usage.average_duration_ms)} ms` : '-' },
  ]
})

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const modelStats = computed<any[]>(() => resultData.value?.model_stats || [])

interface DailyUsageRow {
  date: string
  requests: number
  input_tokens: number
  output_tokens: number
  cache_read_tokens: number
  cache_write_tokens: number
  cost: number
  actual_cost?: number
}

const dailyUsageRows = computed<DailyUsageRow[]>(() => {
  const rows = resultData.value?.daily_usage
  return Array.isArray(rows) ? rows : []
})

const showDailyUsage = computed(() => Boolean(resultData.value && Array.isArray(resultData.value.daily_usage)))

// ==================== Utility Functions ====================

function points(value: number | null | undefined): string {
  if (value == null || value < 0) return '-'
  return formatPoints(Number(value))
}

function pointRange(used: number | null | undefined, limit: number | null | undefined): string {
  if (used == null || limit == null || used < 0 || limit < 0) return '-'
  return `${formatPointAmount(Number(used))} / ${formatPoints(Number(limit))}`
}

function fmtNum(val: number | null | undefined): string {
  if (val == null) return '-'
  return val.toLocaleString()
}

function formatDate(iso: string | null | undefined): string {
  if (!iso) return '-'
  const d = new Date(iso)
  const loc = locale.value === 'zh' ? 'zh-CN' : 'en-US'
  return d.toLocaleDateString(loc, { year: 'numeric', month: 'long', day: 'numeric' })
}

function getBrowserTimezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'
  } catch {
    return 'UTC'
  }
}

// ==================== API Query ====================

async function fetchUsage(key: string) {
  const dateParams = getDateParams()
  const url = buildGatewayUrl('/v1/usage') + (dateParams ? '?' + dateParams : '')
  const res = await fetch(url, {
    headers: { 'Authorization': 'Bearer ' + key },
  })
  if (!res.ok) {
    const body = await res.json().catch(() => null)
    const msg = body?.error?.message || body?.message || `${t('keyUsage.queryFailed')} (${res.status})`
    throw new Error(msg)
  }
  return await res.json()
}

async function queryKey() {
  if (isQuerying.value) return
  const key = apiKey.value.trim()
  if (!key) {
    appStore.showInfo(t('keyUsage.enterApiKey'))
    return
  }

  isQuerying.value = true
  showResults.value = true
  showLoading.value = true
  resultData.value = null

  try {
    const data = await fetchUsage(key)
    resultData.value = data
    showLoading.value = false
    showDatePicker.value = true

    // Trigger ring animations after DOM update
    nextTick(() => {
      triggerRingAnimation(ringItems.value)
    })

    appStore.showSuccess(t('keyUsage.querySuccess'))
  } catch (err) {
    showResults.value = false
    showLoading.value = false
    appStore.showError((err as Error).message || t('keyUsage.queryFailedRetry'))
  } finally {
    isQuerying.value = false
  }
}

// ==================== Lifecycle ====================

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme === 'dark' || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

function formatResetTime(resetAt: string | null | undefined): string {
  if (!resetAt) return ''
  const diff = new Date(resetAt).getTime() - now.value.getTime()
  if (diff <= 0) return t('keyUsage.resetNow')
  const days = Math.floor(diff / 86400000)
  const hours = Math.floor((diff % 86400000) / 3600000)
  const mins = Math.floor((diff % 3600000) / 60000)
  if (days > 0) return `${days}d ${hours}h`
  if (hours > 0) return `${hours}h ${mins}m`
  return `${mins}m`
}

onMounted(() => {
  initTheme()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
  resetTimer = setInterval(() => { now.value = new Date() }, 60000)
})

onUnmounted(() => {
  if (resetTimer) clearInterval(resetTimer)
})
</script>

<style scoped>
/* Input focus ring */
.input-ring {
  transition: box-shadow 0.2s ease, border-color 0.2s ease;
}
.input-ring:focus {
  box-shadow: 0 0 0 3px rgba(10, 132, 255, 0.2);
  border-color: #0a84ff;
  outline: none;
}

/* Ring animation */
.progress-ring {
  transition: stroke-dashoffset 1.2s cubic-bezier(0.4, 0, 0.2, 1);
  transform: rotate(-90deg);
  transform-origin: 50% 50%;
}

/* Skeleton loading */
@keyframes shimmer-kv {
  0%   { background-position: -200% 0; }
  100% { background-position: 200% 0; }
}
.skeleton {
  background: linear-gradient(90deg, #e5e7eb 25%, #f3f4f6 50%, #e5e7eb 75%);
  background-size: 200% 100%;
  animation: shimmer-kv 1.8s ease-in-out infinite;
  border-radius: 8px;
}
:global(.dark) .skeleton {
  background: linear-gradient(90deg, #26262b 25%, #17171a 50%, #26262b 75%);
  background-size: 200% 100%;
}

/* Fade up animation */
@keyframes fade-up-kv {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
}
.fade-up {
  animation: fade-up-kv 0.5s cubic-bezier(0.4, 0, 0.2, 1) forwards;
}
.fade-up-delay-1 { animation-delay: 0.1s; opacity: 0; }
.fade-up-delay-2 { animation-delay: 0.2s; opacity: 0; }
.fade-up-delay-3 { animation-delay: 0.3s; opacity: 0; }
.fade-up-delay-4 { animation-delay: 0.4s; opacity: 0; }

/* Pulse dot */
@keyframes pulse-dot-kv {
  0%, 100% { opacity: 1; box-shadow: 0 0 0 0 currentColor; }
  50% { opacity: 0.6; box-shadow: 0 0 8px 2px currentColor; }
}
.pulse-dot {
  animation: pulse-dot-kv 2s ease-in-out infinite;
}

/* Tabular nums */
.tabular-nums {
  font-variant-numeric: tabular-nums;
  letter-spacing: 0;
}
</style>

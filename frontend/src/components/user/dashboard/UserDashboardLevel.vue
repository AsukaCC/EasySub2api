<template>
  <section class="dashboard-level card" :aria-label="t('dashboard.level.title')">
    <header class="dashboard-level__header">
      <div class="dashboard-level__title-group">
        <span class="dashboard-level__icon"><Icon name="chart" size="md" /></span>
        <div>
          <h2 class="dashboard-level__title">{{ t('dashboard.level.title') }}</h2>
          <p class="dashboard-level__description">{{ t('dashboard.level.description') }}</p>
        </div>
      </div>
      <div class="dashboard-level__header-actions">
        <span v-if="profile" class="dashboard-level__badge">L{{ profile.level }}</span>
        <button
          v-if="profile"
          type="button"
          class="dashboard-level__view-full"
          @click="showFullDialog = true"
        >
          {{ t('dashboard.level.viewFull') }}
        </button>
      </div>
    </header>

    <LoadingState v-if="loading" variant="section" size="md" class="dashboard-level__state" />
    <div v-else-if="profile" class="dashboard-level__body">
      <!-- 仅展示距下一级的进度,完整阶梯见弹窗 -->
      <div class="dashboard-level__progress" :aria-hidden="false">
        <div class="dashboard-level__progress-track">
          <span class="dashboard-level__progress-fill" :style="{ width: `${progress}%` }" />
        </div>
        <div class="dashboard-level__progress-labels">
          <span>L{{ profile.level }} · {{ formatPoints(profile.usage_7d) }}</span>
          <span v-if="nextLevel">L{{ nextLevel.level }} · ≥ {{ formatPoints(nextLevel.threshold) }}</span>
          <span v-else>L3</span>
        </div>
      </div>
      <p class="dashboard-level__caption">
        <template v-if="nextLevel">
          {{ t('dashboard.level.toNext', { level: `L${nextLevel.level}`, amount: formatPoints(nextLevel.remaining) }) }}
        </template>
        <template v-else>{{ t('dashboard.level.maxLevel') }}</template>
      </p>
      <div
        v-if="nextLevel && profile.next_level_multiplier != null"
        class="dashboard-level__next-rate"
      >
        <div class="dashboard-level__next-rate-copy">
          <span class="dashboard-level__label">
            {{ t('dashboard.level.nextMultiplier', { level: `L${nextLevel.level}` }) }}
          </span>
          <small v-if="profile.next_multiplier_group" class="dashboard-level__group">
            {{ profile.next_multiplier_group }}
          </small>
        </div>
        <strong class="dashboard-level__next-rate-value">
          ×{{ formatMultiplier(profile.next_level_multiplier) }}
        </strong>
      </div>
    </div>
    <div v-else class="dashboard-level__state dashboard-level__muted">
      {{ t('dashboard.level.unavailable') }}
    </div>

    <!-- 完整进度弹窗 -->
    <BaseDialog
      :show="showFullDialog"
      :title="t('dashboard.level.fullTitle')"
      width="normal"
      @close="showFullDialog = false"
    >
      <div v-if="profile" class="dashboard-level__full">
        <div class="dashboard-level__summary">
          <div>
            <span class="dashboard-level__label">{{ t('dashboard.level.current') }}</span>
            <strong class="dashboard-level__level">L{{ profile.level }}</strong>
          </div>
          <div>
            <span class="dashboard-level__label">{{ t('dashboard.level.spend7d') }}</span>
            <strong class="dashboard-level__value">{{ formatPoints(profile.usage_7d) }}</strong>
          </div>
          <div>
            <span class="dashboard-level__label">{{ t('dashboard.level.multiplier') }}</span>
            <strong v-if="profile.level_multiplier != null" class="dashboard-level__value">×{{ formatMultiplier(profile.level_multiplier) }}</strong>
            <strong v-else class="dashboard-level__value dashboard-level__muted">—</strong>
          </div>
          <div v-if="profile.effective_multiplier != null && !sameMultiplier">
            <span class="dashboard-level__label">{{ t('dashboard.level.effectiveMultiplier') }}</span>
            <strong class="dashboard-level__value">×{{ formatMultiplier(profile.effective_multiplier) }}</strong>
            <small v-if="profile.multiplier_group" class="dashboard-level__group">{{ profile.multiplier_group }}</small>
          </div>
          <div v-if="nextLevel && profile.next_level_multiplier != null">
            <span class="dashboard-level__label">
              {{ t('dashboard.level.nextMultiplier', { level: `L${nextLevel.level}` }) }}
            </span>
            <strong class="dashboard-level__value">×{{ formatMultiplier(profile.next_level_multiplier) }}</strong>
            <small v-if="profile.next_multiplier_group" class="dashboard-level__group">{{ profile.next_multiplier_group }}</small>
          </div>
        </div>

        <div class="dashboard-level__ladder">
          <h4 class="dashboard-level__ladder-title">{{ t('dashboard.level.ladder') }}</h4>
          <div v-for="tier in ladder" :key="tier.level" class="dashboard-level__tier">
            <div class="dashboard-level__tier-head">
              <span class="dashboard-level__tier-name">L{{ tier.level }}</span>
              <span class="dashboard-level__tier-threshold">≥ {{ formatPoints(tier.threshold) }}</span>
              <span class="dashboard-level__tier-status" :class="`dashboard-level__tier-status--${tier.status}`">
                {{ t(`dashboard.level.${tier.status}`) }}
              </span>
            </div>
            <div class="dashboard-level__progress-track">
              <span class="dashboard-level__progress-fill" :style="{ width: `${tier.progress}%` }" />
            </div>
          </div>
        </div>
      </div>
    </BaseDialog>
  </section>
</template>

<script setup lang="ts">
import LoadingState from '@/components/common/LoadingState.vue'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import type { UserLevelDashboard } from '@/api/userLevel'
import { formatPoints } from '@/utils/format'

const props = defineProps<{
  profile: UserLevelDashboard | null
  loading: boolean
}>()

const { t } = useI18n()

const showFullDialog = ref(false)

const formatMultiplier = (value: number) => Number(value || 0).toFixed(2)

// 下一级信息:门槛与差额;满级时为 null
const nextLevel = computed(() => {
  const p = props.profile
  if (!p || p.level >= 3) return null
  const threshold = p.level === 1 ? p.l2_min_spend : p.l3_min_spend
  return {
    level: p.level + 1,
    threshold,
    remaining: Math.max(0, threshold - Math.max(0, p.usage_7d)),
  }
})

// 当前级到下一级的进度(满级 100%)
const progress = computed(() => {
  if (!props.profile) return 0
  const spend = Math.max(0, props.profile.usage_7d)
  if (props.profile.level >= 3) return 100
  if (props.profile.level === 2) {
    const span = props.profile.l3_min_spend - props.profile.l2_min_spend
    return span > 0 ? Math.min(100, Math.max(0, ((spend - props.profile.l2_min_spend) / span) * 100)) : 50
  }
  return props.profile.l2_min_spend > 0 ? Math.min(100, (spend / props.profile.l2_min_spend) * 100) : 0
})

type TierStatus = 'achieved' | 'inProgress' | 'locked'

// 弹窗内的完整阶梯:每级门槛、状态与到达该级的进度
// 每级的进度区间 = [上一级门槛, 本级门槛],已达成恒为 100%。
const ladder = computed(() => {
  const p = props.profile
  if (!p) return []
  const spend = Math.max(0, p.usage_7d)
  const tiers = [
    { level: 1, from: 0, threshold: 0 },
    { level: 2, from: 0, threshold: p.l2_min_spend },
    { level: 3, from: p.l2_min_spend, threshold: p.l3_min_spend },
  ]
  return tiers.map((tier) => {
    let status: TierStatus
    if (p.level >= tier.level) {
      status = 'achieved'
    } else if (tier.level === p.level + 1) {
      status = 'inProgress'
    } else {
      status = 'locked'
    }
    let tierProgress = 0
    if (status === 'achieved') {
      tierProgress = 100
    } else if (status === 'inProgress') {
      const span = tier.threshold - tier.from
      tierProgress = span > 0 ? Math.min(100, Math.max(0, ((spend - tier.from) / span) * 100)) : 0
    }
    return { level: tier.level, threshold: tier.threshold, status, progress: tierProgress }
  })
})

const sameMultiplier = computed(() => {
  if (!props.profile || props.profile.level_multiplier == null || props.profile.effective_multiplier == null) return true
  return Math.abs(props.profile.level_multiplier - props.profile.effective_multiplier) < 0.0001
})
</script>

<style scoped>
.dashboard-level { overflow: visible; }
.dashboard-level__header,
.dashboard-level__title-group { display: flex; align-items: center; }
.dashboard-level__header { justify-content: space-between; gap: 1rem; padding: 1rem 1.5rem; border-bottom: 1px solid var(--color-border-subtle); }
.dashboard-level__title-group { min-width: 0; gap: .75rem; }
.dashboard-level__header-actions { display: flex; flex: 0 0 auto; align-items: center; gap: .5rem; }
.dashboard-level__icon { display: inline-flex; flex: 0 0 auto; align-items: center; justify-content: center; width: 2.25rem; height: 2.25rem; border-radius: var(--radius-md); color: var(--color-text-brand); background: var(--color-primary-subtle); }
.dashboard-level__title { margin: 0; color: var(--color-text-primary); font-size: var(--font-size-base); font-weight: 650; }
.dashboard-level__description { margin: .125rem 0 0; color: var(--color-text-tertiary); font-size: var(--font-size-xs); }
.dashboard-level__badge { flex: 0 0 auto; padding: .3rem .7rem; border-radius: 999px; color: var(--color-text-brand); background: var(--color-primary-subtle); font-size: var(--font-size-xs); font-weight: 700; }
.dashboard-level__view-full {
  flex: 0 0 auto;
  padding: .3rem .7rem;
  border: 1px solid var(--glass-border);
  border-radius: 999px;
  color: var(--color-text-secondary);
  background: transparent;
  font-size: var(--font-size-xs);
  font-weight: 600;
  cursor: pointer;
  transition: color 160ms ease, border-color 160ms ease, background-color 160ms ease;
}
.dashboard-level__view-full:hover {
  border-color: color-mix(in srgb, var(--theme-accent) 35%, transparent);
  color: var(--color-text-brand);
  background: var(--color-primary-subtle);
}
.dashboard-level__state { min-height: 5rem; display: flex; align-items: center; justify-content: center; padding: 1rem; }
.dashboard-level__body { display: grid; gap: .5rem; padding: 1rem 1.5rem 1.15rem; }
.dashboard-level__label { color: var(--color-text-tertiary); font-size: var(--font-size-xs); }
.dashboard-level__level { color: var(--color-text-primary); font-size: var(--font-size-xl); }
.dashboard-level__value { color: var(--color-text-primary); font-size: var(--font-size-lg); }
.dashboard-level__muted { color: var(--color-text-tertiary); }
.dashboard-level__group { overflow: hidden; color: var(--color-text-tertiary); font-size: var(--font-size-xs); text-overflow: ellipsis; white-space: nowrap; }
.dashboard-level__caption { margin: 0; color: var(--color-text-secondary); font-size: var(--font-size-sm); font-weight: 500; }
.dashboard-level__next-rate { display: flex; align-items: center; justify-content: space-between; gap: .75rem; padding: .625rem .75rem; border: 1px solid var(--glass-border); border-radius: var(--radius-lg); background: var(--glass-layer-inset-bg); box-shadow: 0 1px 0 var(--glass-highlight) inset; -webkit-backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate)); backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate)); }
.dashboard-level__next-rate-copy { display: grid; min-width: 0; gap: .125rem; }
.dashboard-level__next-rate-value { flex: 0 0 auto; color: var(--color-text-brand); font-size: var(--font-size-lg); font-weight: 700; }
.dashboard-level__progress { display: grid; gap: .4rem; }
.dashboard-level__progress-track { height: .4rem; overflow: hidden; border-radius: 999px; background: var(--color-surface-muted); }
.dashboard-level__progress-fill { display: block; height: 100%; border-radius: inherit; background: linear-gradient(90deg, var(--theme-accent), color-mix(in srgb, var(--theme-accent) 55%, white)); transition: width .25s ease; }
.dashboard-level__progress-labels { display: flex; justify-content: space-between; gap: .75rem; color: var(--color-text-tertiary); font-size: var(--font-size-xs); }

/* ---- 弹窗内容 ---- */
.dashboard-level__full { display: grid; gap: 1.25rem; }
.dashboard-level__summary { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 1rem; }
.dashboard-level__summary > div { min-width: 0; display: grid; gap: .25rem; }
.dashboard-level__ladder { display: grid; gap: .75rem; }
.dashboard-level__ladder-title { margin: 0; color: var(--color-text-primary); font-size: var(--font-size-sm); font-weight: 650; }
.dashboard-level__tier { display: grid; gap: .375rem; padding: .75rem .875rem; border: 1px solid var(--glass-border); border-radius: var(--radius-lg); background: var(--color-surface-muted); }
.dashboard-level__tier-head { display: flex; align-items: center; gap: .625rem; }
.dashboard-level__tier-name { color: var(--color-text-primary); font-size: var(--font-size-sm); font-weight: 700; }
.dashboard-level__tier-threshold { color: var(--color-text-tertiary); font-size: var(--font-size-xs); }
.dashboard-level__tier-status { margin-left: auto; padding: .125rem .5rem; border-radius: 999px; font-size: var(--font-size-xs); font-weight: 600; }
.dashboard-level__tier-status--achieved { color: #059669; background: rgba(16, 185, 129, 0.14); }
.dashboard-level__tier-status--inProgress { color: var(--color-text-brand); background: var(--color-primary-subtle); }
.dashboard-level__tier-status--locked { color: var(--color-text-tertiary); background: rgba(120, 120, 128, 0.14); }
.dark .dashboard-level__tier-status--achieved { color: #34d399; background: rgba(6, 95, 70, 0.4); }
</style>

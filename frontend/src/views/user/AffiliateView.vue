<template>
  <AppLayout>
    <div class="views-user-affiliate-view__panel">
      <div v-if="loading" class="views-user-affiliate-view__panel-2">
        <div
          class="views-user-affiliate-view__panel-3"
        ></div>
      </div>

      <template v-else-if="detail">
        <div class="views-user-affiliate-view__panel-4">
          <div class="views-user-affiliate-view__panel-5 card">
            <p class="views-user-affiliate-view__description">
              <Icon name="sparkles" size="sm" class="views-user-affiliate-view__icon" />
              {{ t('affiliate.stats.rebateRate') }}
            </p>
            <p class="views-user-affiliate-view__description-2">
              {{ formattedRebateRate }}<span class="views-user-affiliate-view__text">%</span>
            </p>
            <p class="views-user-affiliate-view__description-3">
              {{ t('affiliate.stats.rebateRateHint') }}
            </p>
          </div>
          <div class="views-user-affiliate-view__panel-5 card">
            <p class="views-user-affiliate-view__description-4">{{ t('affiliate.stats.invitedUsers') }}</p>
            <p class="views-user-affiliate-view__description-5">
              {{ formatCount(detail.aff_count) }}
            </p>
          </div>
          <div class="views-user-affiliate-view__panel-5 card">
            <p class="views-user-affiliate-view__description-4">{{ t('affiliate.stats.availableQuota') }}</p>
            <p class="views-user-affiliate-view__description-6">
              {{ formatPoints(detail.aff_quota) }}
            </p>
          </div>
          <div class="views-user-affiliate-view__panel-5 card">
            <p class="views-user-affiliate-view__description-4">{{ t('affiliate.stats.totalQuota') }}</p>
            <p class="views-user-affiliate-view__description-5">
              {{ formatPoints(detail.aff_history_quota) }}
            </p>
            <p v-if="detail.aff_frozen_quota > 0" class="views-user-affiliate-view__description-7">
              {{ t('affiliate.stats.frozenQuota') }}: {{ formatPoints(detail.aff_frozen_quota) }}
            </p>
          </div>
        </div>

        <div class="views-user-affiliate-view__panel-6 card">
          <h3 class="views-user-affiliate-view__heading">{{ t('affiliate.title') }}</h3>
          <p class="views-user-affiliate-view__description-8">{{ t('affiliate.description') }}</p>

          <div class="views-user-affiliate-view__panel-7">
            <div class="views-user-affiliate-view__panel-8">
              <p class="views-user-affiliate-view__description-9">{{ t('affiliate.yourCode') }}</p>
              <div class="views-user-affiliate-view__panel-9">
                <code class="views-user-affiliate-view__code">{{ detail.aff_code }}</code>
                <button class="views-user-affiliate-view__action btn btn-secondary btn-sm" @click="copyCode">
                  <Icon name="copy" size="sm" />
                  <span>{{ t('affiliate.copyCode') }}</span>
                </button>
              </div>
            </div>

            <div class="views-user-affiliate-view__panel-8">
              <p class="views-user-affiliate-view__description-9">{{ t('affiliate.inviteLink') }}</p>
              <div class="views-user-affiliate-view__panel-9">
                <code class="views-user-affiliate-view__code-2">{{ inviteLink }}</code>
                <button class="views-user-affiliate-view__action btn btn-secondary btn-sm" @click="copyInviteLink">
                  <Icon name="copy" size="sm" />
                  <span>{{ t('affiliate.copyLink') }}</span>
                </button>
              </div>
            </div>
          </div>

          <div class="views-user-affiliate-view__panel-10">
            <p class="views-user-affiliate-view__description-10">{{ t('affiliate.tips.title') }}</p>
            <ul class="views-user-affiliate-view__list">
              <li>1. {{ t('affiliate.tips.line1') }}</li>
              <li>2. {{ t('affiliate.tips.line2', { rate: `${formattedRebateRate}%` }) }}</li>
              <li>3. {{ t('affiliate.tips.line3') }}</li>
              <li v-if="detail.aff_frozen_quota > 0">4. {{ t('affiliate.tips.line4') }}</li>
            </ul>
          </div>
        </div>

        <div class="views-user-affiliate-view__panel-6 card">
          <div class="views-user-affiliate-view__panel-11">
            <div>
              <h3 class="views-user-affiliate-view__heading">{{ t('affiliate.transfer.title') }}</h3>
              <p class="views-user-affiliate-view__description-8">{{ t('affiliate.transfer.description') }}</p>
            </div>
            <button
              class="btn btn-primary"
              :disabled="transferring || detail.aff_quota <= 0"
              @click="transferQuota"
            >
              <Icon v-if="transferring" name="refresh" size="sm" class="views-user-affiliate-view__icon-2" />
              <Icon v-else name="sparkles" size="sm" />
              <span>{{ transferring ? t('affiliate.transfer.transferring') : t('affiliate.transfer.button') }}</span>
            </button>
          </div>
          <p v-if="detail.aff_quota <= 0" class="views-user-affiliate-view__element">
            {{ t('affiliate.transfer.empty') }}
          </p>
        </div>

        <div class="views-user-affiliate-view__panel-6 card">
          <h3 class="views-user-affiliate-view__heading">{{ t('affiliate.invitees.title') }}</h3>
          <div v-if="detail.invitees.length === 0" class="views-user-affiliate-view__panel-12">
            {{ t('affiliate.invitees.empty') }}
          </div>
          <div v-else class="views-user-affiliate-view__panel-13">
            <table class="views-user-affiliate-view__table">
              <thead>
                <tr class="views-user-affiliate-view__row">
                  <th class="views-user-affiliate-view__heading-2">{{ t('affiliate.invitees.columns.email') }}</th>
                  <th class="views-user-affiliate-view__heading-2">{{ t('affiliate.invitees.columns.username') }}</th>
                  <th class="views-user-affiliate-view__heading-3">{{ t('affiliate.invitees.columns.rebate') }}</th>
                  <th class="views-user-affiliate-view__heading-2">{{ t('affiliate.invitees.columns.joinedAt') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="item in detail.invitees"
                  :key="item.user_id"
                  class="views-user-affiliate-view__row-2"
                >
                  <td class="views-user-affiliate-view__cell">{{ item.email || '-' }}</td>
                  <td class="views-user-affiliate-view__cell-2">{{ item.username || '-' }}</td>
                  <td class="views-user-affiliate-view__cell-3">{{ formatPoints(item.total_rebate) }}</td>
                  <td class="views-user-affiliate-view__cell-2">{{ formatDateTime(item.created_at) || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import userAPI from '@/api/user'
import type { UserAffiliateDetail } from '@/types'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { useClipboard } from '@/composables/useClipboard'
import { formatDateTime, formatPoints } from '@/utils/format'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const { copyToClipboard } = useClipboard()

const loading = ref(true)
const transferring = ref(false)
const detail = ref<UserAffiliateDetail | null>(null)

const inviteLink = computed(() => {
  if (!detail.value) return ''
  if (typeof window === 'undefined') return `/register?aff=${encodeURIComponent(detail.value.aff_code)}`
  return `${window.location.origin}/register?aff=${encodeURIComponent(detail.value.aff_code)}`
})

// Rebate rate is a percentage in the range [0, 100]; backend already clamps it.
// We trim trailing zeros (e.g. 20.00 → "20", 12.50 → "12.5") for a cleaner UI.
const formattedRebateRate = computed(() => {
  const v = detail.value?.effective_rebate_rate_percent ?? 0
  const rounded = Math.round(v * 100) / 100
  return Number.isInteger(rounded) ? String(rounded) : rounded.toString()
})

function formatCount(value: number): string {
  return value.toLocaleString()
}

async function loadAffiliateDetail(silent = false): Promise<void> {
  if (!silent) {
    loading.value = true
  }
  try {
    detail.value = await userAPI.getAffiliateDetail()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.loadFailed')))
  } finally {
    if (!silent) {
      loading.value = false
    }
  }
}

async function copyCode(): Promise<void> {
  if (!detail.value?.aff_code) return
  await copyToClipboard(detail.value.aff_code, t('affiliate.codeCopied'))
}

async function copyInviteLink(): Promise<void> {
  if (!inviteLink.value) return
  await copyToClipboard(inviteLink.value, t('affiliate.linkCopied'))
}

async function transferQuota(): Promise<void> {
  if (!detail.value || detail.value.aff_quota <= 0 || transferring.value) return
  transferring.value = true
  try {
    const resp = await userAPI.transferAffiliateQuota()
    appStore.showSuccess(t('affiliate.transfer.success', { amount: formatPoints(resp.transferred_quota) }))
    await Promise.all([
      loadAffiliateDetail(true),
      authStore.refreshUser().catch(() => undefined),
    ])
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.transferFailed')))
  } finally {
    transferring.value = false
  }
}

onMounted(() => {
  void loadAffiliateDetail()
})
</script>

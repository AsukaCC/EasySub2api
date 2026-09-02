<template>
  <AppLayout>
    <div class="page-stack">
      <div v-if="loading" class="views-user-affiliate-view__panel-2">
        <div
          class="views-user-affiliate-view__panel-3"
        ></div>
      </div>

      <template v-else-if="detail">
        <section v-if="detail.can_bind_inviter" class="affiliate-bind card card-body">
          <div class="affiliate-bind__intro">
            <div>
              <h3 class="affiliate-bind__title">{{ t('affiliate.binding.title') }}</h3>
              <p class="affiliate-bind__description">{{ t('affiliate.binding.description') }}</p>
            </div>
            <span class="affiliate-bind__status">{{ t('affiliate.binding.unbound') }}</span>
          </div>

          <div class="affiliate-bind__form">
            <label class="affiliate-bind__field">
              <span class="input-label">{{ t('affiliate.binding.label') }}</span>
              <input
                v-model="inviteCode"
                class="input affiliate-bind__input"
                type="text"
                maxlength="32"
                autocomplete="off"
                :disabled="binding"
                :placeholder="t('affiliate.binding.placeholder')"
                @input="normalizeInviteCode"
                @keyup.enter="openBindConfirm"
              />
            </label>
            <button
              type="button"
              class="btn btn-primary affiliate-bind__action"
              :disabled="binding || !canSubmitInviteCode"
              @click="openBindConfirm"
            >
              <Icon :name="binding ? 'refresh' : 'link'" size="sm" :class="{ 'affiliate-bind__spinner': binding }" />
              <span>{{ binding ? t('affiliate.binding.binding') : t('affiliate.binding.action') }}</span>
            </button>
          </div>

          <div class="affiliate-bind__notice">
            <Icon name="infoCircle" size="sm" />
            <p>
              {{ detail.invitee_binding_reward_points > 0
                ? t('affiliate.binding.reward', {
                    points: formatPoints(detail.invitee_binding_reward_points),
                    days: detail.invitee_binding_reward_validity_days
                  })
                : t('affiliate.binding.noReward') }}
              {{ t('affiliate.binding.locked') }}
            </p>
          </div>
        </section>

        <div v-else-if="detail.inviter_id" class="affiliate-bound card-body card">
          <Icon name="checkCircle" size="sm" />
          <span>{{ t('affiliate.binding.bound') }}</span>
        </div>

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
              {{ t(detail.rebate_recipient === 'invitee'
                ? 'affiliate.stats.rebateRateHintInvitee'
                : 'affiliate.stats.rebateRateHintInviter') }}
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

        <div class="views-user-affiliate-view__panel-6 card-body card">
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
                <button
                  class="views-user-affiliate-view__action btn btn-secondary btn-sm"
                  :disabled="regenerating || detail.code_regeneration_remaining <= 0"
                  @click="showRegenerateConfirm = true"
                >
                  <Icon name="refresh" size="sm" />
                  <span>{{ t('affiliate.regenerate.button') }}</span>
                </button>
              </div>
              <p class="views-user-affiliate-view__description-3">
                {{ t('affiliate.regenerate.remaining', {
                  remaining: detail.code_regeneration_remaining,
                  limit: detail.code_regeneration_limit,
                  resetAt: formatDateTime(detail.code_regeneration_reset_at)
                }) }}
              </p>
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
              <li>2. {{ t(detail.rebate_recipient === 'invitee'
                ? 'affiliate.tips.line2Invitee'
                : 'affiliate.tips.line2Inviter', { rate: `${formattedRebateRate}%` }) }}</li>
              <li>3. {{ t('affiliate.tips.line3') }}</li>
              <li v-if="detail.aff_frozen_quota > 0">4. {{ t('affiliate.tips.line4') }}</li>
            </ul>
          </div>
        </div>

        <div class="views-user-affiliate-view__panel-6 card-body card">
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

        <div class="views-user-affiliate-view__panel-6 card-body card">
          <h3 class="views-user-affiliate-view__heading">{{ t('affiliate.invitees.title') }}</h3>
          <div v-if="detail.invitees.length === 0" class="views-user-affiliate-view__panel-12 card-body">
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

    <ConfirmDialog
      :show="showBindConfirm"
      :title="t('affiliate.binding.confirmTitle')"
      :message="t('affiliate.binding.confirmMessage', { code: inviteCode })"
      :confirm-text="t('affiliate.binding.confirmButton')"
      @confirm="bindInviteCode"
      @cancel="showBindConfirm = false"
    />
    <ConfirmDialog
      :show="showRegenerateConfirm"
      :title="t('affiliate.regenerate.confirmTitle')"
      :message="t('affiliate.regenerate.confirmMessage', {
        remaining: detail?.code_regeneration_remaining ?? 0
      })"
      :confirm-text="t('affiliate.regenerate.confirmButton')"
      danger
      @confirm="regenerateCode"
      @cancel="showRegenerateConfirm = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
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
const regenerating = ref(false)
const binding = ref(false)
const inviteCode = ref('')
const showBindConfirm = ref(false)
const showRegenerateConfirm = ref(false)
const detail = ref<UserAffiliateDetail | null>(null)

const canSubmitInviteCode = computed(() => {
  const length = inviteCode.value.trim().length
  return length >= 4 && length <= 32
})

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

function normalizeInviteCode(event: Event): void {
  inviteCode.value = (event.target as HTMLInputElement).value.toUpperCase()
}

function openBindConfirm(): void {
  if (!canSubmitInviteCode.value || binding.value) return
  showBindConfirm.value = true
}

async function bindInviteCode(): Promise<void> {
  if (!canSubmitInviteCode.value || binding.value) return
  showBindConfirm.value = false
  binding.value = true
  try {
    const response = await userAPI.bindAffiliateCode(inviteCode.value)
    detail.value = response.detail
    inviteCode.value = ''
    const reward = response.binding.invitee_reward
    appStore.showSuccess(reward.applied
      ? t('affiliate.binding.successWithReward', {
          points: formatPoints(reward.points),
          expiresAt: reward.expires_at ? formatDateTime(reward.expires_at) : '-'
        })
      : t('affiliate.binding.success'))
    await authStore.refreshUser().catch(() => undefined)
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.binding.failed'), {
      AFFILIATE_CODE_INVALID: t('affiliate.binding.errors.invalid'),
      AFFILIATE_ALREADY_BOUND: t('affiliate.binding.errors.alreadyBound'),
      AFFILIATE_INVITE_CYCLE: t('affiliate.binding.errors.cycle'),
      AFFILIATE_INVITER_UNAVAILABLE: t('affiliate.binding.errors.unavailable'),
      FEATURE_DISABLED: t('affiliate.binding.errors.disabled')
    }))
  } finally {
    binding.value = false
  }
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

async function regenerateCode(): Promise<void> {
  if (!detail.value || regenerating.value || detail.value.code_regeneration_remaining <= 0) return
  showRegenerateConfirm.value = false
  regenerating.value = true
  try {
    const result = await userAPI.regenerateAffiliateCode()
    detail.value.aff_code = result.aff_code
    detail.value.code_regeneration_limit = result.limit
    detail.value.code_regeneration_used = result.used
    detail.value.code_regeneration_remaining = result.remaining
    detail.value.code_regeneration_reset_at = result.reset_at
    appStore.showSuccess(t('affiliate.regenerate.success'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.regenerate.failed')))
    await loadAffiliateDetail(true)
  } finally {
    regenerating.value = false
  }
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

<style scoped>
.affiliate-bind {
  display: grid;
  gap: 1rem;
}

.affiliate-bind__intro,
.affiliate-bind__form,
.affiliate-bind__notice,
.affiliate-bound {
  display: flex;
  align-items: center;
}

.affiliate-bind__intro {
  justify-content: space-between;
  gap: 1rem;
}

.affiliate-bind__title {
  color: var(--color-text-primary);
  font-size: var(--type-section-title-size);
  line-height: var(--line-height-tight);
  font-weight: var(--font-weight-semibold);
}

.affiliate-bind__description {
  margin-top: 0.25rem;
  color: var(--color-text-secondary);
  font-size: var(--type-caption-size);
  line-height: var(--type-caption-line-height);
}

.affiliate-bind__status {
  flex: 0 0 auto;
  color: var(--color-text-brand);
  font-size: var(--type-caption-size);
}

.affiliate-bind__form {
  align-items: flex-end;
  gap: 0.75rem;
}

.affiliate-bind__field {
  flex: 1 1 auto;
}

.affiliate-bind__input {
  text-transform: uppercase;
}

.affiliate-bind__action {
  flex: 0 0 auto;
}

.affiliate-bind__notice,
.affiliate-bound {
  gap: 0.5rem;
  color: var(--color-text-secondary);
  font-size: var(--type-caption-size);
  line-height: var(--type-caption-line-height);
}

.affiliate-bind__notice {
  padding: 0.75rem;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-md);
  background: var(--glass-layer-inset-bg);
  backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate));
}

.affiliate-bound {
  color: var(--color-text-success);
}

.affiliate-bind__spinner {
  animation: affiliate-bind-spin 0.8s linear infinite;
}

@keyframes affiliate-bind-spin {
  to { transform: rotate(360deg); }
}

@media (max-width: 640px) {
  .affiliate-bind__intro,
  .affiliate-bind__form {
    align-items: stretch;
    flex-direction: column;
  }
}

@media (prefers-reduced-motion: reduce) {
  .affiliate-bind__spinner {
    animation: none;
  }
}
</style>

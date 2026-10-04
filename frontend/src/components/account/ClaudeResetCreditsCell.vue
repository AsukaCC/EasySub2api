<template>
  <div v-if="visible" class="claude-reset">
    <button type="button" class="btn btn-secondary" :disabled="loading || redeeming" @click="load">
      {{ loading ? t('admin.accounts.claudeReset.loading') : t('admin.accounts.claudeReset.query') }}
    </button>
    <span v-if="data">{{ t('admin.accounts.claudeReset.count', { count: data.available_count }) }}</span>
    <button v-if="data?.credits.some(credit => credit.redeemable)" type="button" class="btn btn-secondary" :disabled="redeeming || loading" @click="confirming = true">
      {{ redeeming ? t('admin.accounts.claudeReset.redeeming') : t('admin.accounts.claudeReset.reset') }}
    </button>
    <span v-if="error" role="alert">{{ t('admin.accounts.claudeReset.unavailable') }}</span>
    <span v-if="outcome" role="status">{{ t(`admin.accounts.claudeReset.outcomes.${outcome}`) }}</span>
    <ConfirmDialog :show="confirming" :title="t('admin.accounts.claudeReset.reset')" :message="t('admin.accounts.claudeReset.confirm')" @confirm="redeem" @cancel="confirming = false" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Account } from '@/types'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { queryClaudeResetCredits, redeemClaudeResetCredit } from '@/api/admin/accounts'

const props = defineProps<{ account: Account }>()
const { t } = useI18n()
const data = ref<Awaited<ReturnType<typeof queryClaudeResetCredits>> | null>(null)
const loading = ref(false)
const redeeming = ref(false)
const error = ref(false)
const confirming = ref(false)
const outcome = ref('')
let confirmationKey = ''
let generation = 0
const visible = computed(() => props.account.platform === 'anthropic' && props.account.type === 'oauth')

watch(() => props.account.id, () => {
  generation++
  data.value = null
  error.value = false
  outcome.value = ''
  confirming.value = false
  loading.value = false
  redeeming.value = false
  confirmationKey = ''
})

async function load() {
  if (!visible.value || loading.value) return
  const current = generation
  loading.value = true
  error.value = false
  try {
    const result = await queryClaudeResetCredits(props.account.id)
    if (current === generation) data.value = result
  } catch {
    if (current === generation) error.value = true
  } finally {
    if (current === generation) loading.value = false
  }
}

async function redeem() {
  confirming.value = false
  if (!data.value || redeeming.value) return
  const current = generation
  redeeming.value = true
  error.value = false
  // Reuse the confirmation after transport failures to avoid consuming another grant.
  confirmationKey ||= crypto.randomUUID()
  try {
    const result = await redeemClaudeResetCredit(props.account.id, confirmationKey)
    if (current !== generation) return
    const outcomes = ['reset', 'already_used', 'not_limited', 'cooldown', 'ineligible', 'unknown']
    outcome.value = outcomes.includes(result.outcome) ? result.outcome : 'unknown'
    if (result.outcome !== 'unknown') confirmationKey = ''
    if (result.credits) data.value = result.credits
    await load()
  } catch {
    if (current === generation) error.value = true
  } finally {
    if (current === generation) redeeming.value = false
  }
}
</script>

<style scoped>
.claude-reset { display: flex; align-items: center; flex-wrap: wrap; gap: 0.5rem; font-size: var(--type-control-size); color: var(--color-text-secondary); }
</style>

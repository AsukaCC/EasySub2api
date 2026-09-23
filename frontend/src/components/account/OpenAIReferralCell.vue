<template>
  <div class="codex-capabilities">
    <div class="codex-capabilities__row">
      <span>{{ t('upstreamUpdate.credits') }}: {{ credits?.unlimited ? t('upstreamUpdate.unlimited') : credits?.balance ?? t('upstreamUpdate.unknown') }}</span>
      <button type="button" class="btn btn-secondary btn-xs" :disabled="busy" :title="t('upstreamUpdate.refresh')" :aria-label="t('upstreamUpdate.refresh')" @click="queryCredits"><Icon name="refresh" size="sm" /></button>
      <button v-if="!account.parent_account_id" type="button" class="btn btn-secondary btn-xs" @click="open = true">{{ t('upstreamUpdate.invitations') }}</button>
    </div>
    <p v-if="error && !open" role="alert">{{ error }}</p>
    <BaseDialog :show="open" :title="t('upstreamUpdate.invitations')" width="normal" @close="open = false">
      <div class="codex-capabilities__dialog">
        <button type="button" class="btn btn-secondary" :disabled="busy" @click="queryReferrals"><Icon name="refresh" size="sm" />{{ t('upstreamUpdate.refresh') }}</button>
        <p>{{ t('upstreamUpdate.available') }}: {{ eligibility?.available_invites ?? t('upstreamUpdate.unknown') }}</p>
        <p v-if="eligibility?.title">{{ eligibility.title }}</p>
        <ul v-if="eligibility?.rules"><li v-for="rule in eligibility.rules" :key="rule">{{ rule }}</li></ul>
        <label>{{ t('upstreamUpdate.recipient') }}<input v-model="email" class="input" type="email" autocomplete="off" :disabled="busy" /></label>
        <label><input v-model="consent" type="checkbox" :disabled="busy" />{{ t('upstreamUpdate.confirmConsent') }}</label>
        <p v-if="error" role="alert">{{ error }}</p>
        <p v-if="sent" role="status">{{ t('upstreamUpdate.sent') }}</p>
        <button type="button" class="btn btn-primary" :disabled="busy || uncertain || !consent || !email.trim() || !eligibility?.should_show || !eligibility.available_invites" @click="send">{{ t('upstreamUpdate.send') }}</button>
      </div>
    </BaseDialog>
  </div>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import type { Account } from '@/types'
import type { OpenAIReferralEligibility } from '@/types/openaiReferrals'
import { refreshCodexCredits, refreshCodexReferrals, sendCodexReferral, type CodexCredits } from '@/api/admin/upstreamCapabilities'
const props = defineProps<{ account: Account }>()
const { t } = useI18n()
const credits = ref<CodexCredits>(), eligibility = ref<OpenAIReferralEligibility>()
const busy = ref(false), open = ref(false), error = ref(''), email = ref(''), consent = ref(false), sent = ref(false), uncertain = ref(false)
let epoch = 0
watch(() => props.account.id, () => { epoch++; credits.value = undefined; eligibility.value = undefined; busy.value = false; open.value = false; error.value = ''; email.value = ''; consent.value = false; sent.value = false; uncertain.value = false })
async function queryCredits() {
  const current = ++epoch; busy.value = true; error.value = ''
  try { const value = await refreshCodexCredits(props.account.id); if (epoch === current) credits.value = value.credits }
  catch { if (epoch === current) error.value = t('upstreamUpdate.queryFailed') }
  finally { if (epoch === current) busy.value = false }
}
async function queryReferrals() {
  const current = ++epoch; busy.value = true; error.value = ''
  try { const value = await refreshCodexReferrals(props.account.id); if (epoch === current) eligibility.value = value.eligibility }
  catch { if (epoch === current) error.value = t('upstreamUpdate.queryFailed') }
  finally { if (epoch === current) busy.value = false }
}
async function send() {
  if (busy.value || uncertain.value || !consent.value || !eligibility.value) return
  const current = ++epoch; busy.value = true; error.value = ''; sent.value = false
  try { const result = await sendCodexReferral(props.account.id, email.value.trim(), eligibility.value.program_id); if (epoch === current) { sent.value = result.sent; consent.value = false; eligibility.value = result.eligibility } }
  catch { if (epoch === current) { uncertain.value = true; error.value = t('upstreamUpdate.unknownSend') } }
  finally { if (epoch === current) busy.value = false }
}
</script>
<style scoped>
.codex-capabilities { font-size: 11px; }
.codex-capabilities__row { display: flex; align-items: center; flex-wrap: wrap; gap: 6px; }
.codex-capabilities__dialog { display: grid; gap: 12px; font-size: 13px; }
.codex-capabilities__dialog label { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }
.codex-capabilities__dialog input[type=email] { width: 100%; min-width: 0; }
</style>

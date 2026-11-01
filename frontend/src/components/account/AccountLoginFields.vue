<template>
  <section class="account-login-fields">
    <h3>{{ t('admin.accounts.loginInfo.title') }}</h3>
    <p v-if="loading" role="status">{{ t('common.loading') }}</p>
    <div v-if="loadFailed" class="account-login-error" role="alert">
      {{ t('admin.accounts.loginInfo.loadFailed') }}
      <button type="button" class="btn btn-secondary btn-sm" :title="t('common.retry')" :aria-label="t('common.retry')" @click="retry++"><Icon name="refresh" size="sm" /></button>
    </div>
    <div v-for="field in fields" :key="field.key" class="account-login-field">
      <label>
        <span class="input-label">{{ field.label }}</span>
        <span class="account-login-input">
          <input :name="field.key" :value="valueFor(field.key)" :type="field.secret && !revealed[field.key] ? 'password' : 'text'"
            :readonly="readonly" :disabled="loading" class="input" maxlength="4096" :autocomplete="field.secret ? 'new-password' : 'off'" spellcheck="false"
            :placeholder="field.key === 'login_totp_secret' ? t('admin.accounts.loginInfo.secretPlaceholder') : ''"
            @input="setValue(field.key, ($event.target as HTMLInputElement).value)" />
          <button v-if="field.secret" type="button" class="btn btn-secondary" :title="revealed[field.key] ? t('admin.accounts.loginInfo.hide') : t('admin.accounts.loginInfo.show')"
            :aria-label="revealed[field.key] ? t('admin.accounts.loginInfo.hide') : t('admin.accounts.loginInfo.show')" :aria-pressed="!!revealed[field.key]"
            @click="revealed[field.key] = !revealed[field.key]"><Icon :name="revealed[field.key] ? 'eyeOff' : 'eye'" size="sm" /></button>
          <button type="button" class="btn btn-secondary" :disabled="!valueFor(field.key)" :title="t('common.copy')" :aria-label="t('common.copy')" @click="copyToClipboard(valueFor(field.key))"><Icon name="copy" size="sm" /></button>
        </span>
      </label>
    </div>
    <p v-if="invalid" class="account-login-error" role="alert">{{ t('admin.accounts.loginInfo.invalidSecret') }}</p>
    <div v-if="code" class="account-login-code">
      <span>{{ t('admin.accounts.loginInfo.code') }}</span>
      <output>{{ code }}</output>
      <span>{{ t('admin.accounts.loginInfo.remaining', { seconds: remaining }) }}</span>
      <button type="button" class="btn btn-secondary" :title="t('common.copy')" :aria-label="t('common.copy')" @click="copyCode"><Icon name="copy" size="sm" /></button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useClipboard } from '@/composables/useClipboard'
import { useAccountTotp } from '@/composables/useAccountTotp'
import { getAccountLoginCredentials, type AccountLoginCredentials } from '@/api/admin/accountLogin'
import type { Account } from '@/types'

const props = withDefaults(defineProps<{ modelValue?: AccountLoginCredentials; account?: Account | null; active?: boolean; readonly?: boolean }>(), {
  modelValue: () => ({}), account: null, active: true, readonly: false
})
const emit = defineEmits<{ 'update:modelValue': [value: AccountLoginCredentials] }>()
const { t } = useI18n()
const { copyToClipboard } = useClipboard()
type FieldKey = keyof AccountLoginCredentials
const fields = computed<{ key: FieldKey; label: string; secret: boolean }[]>(() => [
  { key: 'login_username', label: t('admin.accounts.loginInfo.username'), secret: false },
  { key: 'login_password', label: t('admin.accounts.loginInfo.password'), secret: true },
  { key: 'login_totp_secret', label: t('admin.accounts.loginInfo.secret'), secret: true }
])
const loaded = ref<AccountLoginCredentials>({})
const revealed = ref<Partial<Record<FieldKey, boolean>>>({})
const loading = ref(false)
const loadFailed = ref(false)
const retry = ref(0)
let controller: AbortController | undefined
function valueFor(key: FieldKey) { return props.modelValue[key] ?? loaded.value[key] ?? '' }
function setValue(key: FieldKey, value: string) { emit('update:modelValue', { ...props.modelValue, [key]: value }) }
const { code, remaining, invalid, refresh } = useAccountTotp(computed(() => valueFor('login_totp_secret')), computed(() => props.active))
function copyCode() {
  refresh()
  if (code.value) void copyToClipboard(code.value)
}
watch([() => props.active, () => props.account?.id, retry], async ([active, id], _, onCleanup) => {
  controller?.abort()
  loaded.value = {}
  revealed.value = {}
  loadFailed.value = false
  loading.value = false
  if (!active || !id || !fields.value.some(field => props.account?.credentials_status?.[`has_${field.key}`])) return
  const request = new AbortController()
  controller = request
  onCleanup(() => request.abort())
  loading.value = true
  try {
    const result = await getAccountLoginCredentials(id, request.signal)
    if (!request.signal.aborted) loaded.value = result
  } catch {
    if (!request.signal.aborted) loadFailed.value = true
  } finally {
    if (!request.signal.aborted) loading.value = false
  }
}, { immediate: true })
onUnmounted(() => { controller?.abort(); loaded.value = {} })
</script>

<style scoped>
.account-login-fields { display: grid; gap: 12px; min-width: 0; border-top: 1px solid var(--color-border); padding-top: 16px; }
.account-login-fields h3 { margin: 0; font-size: var(--font-size-sm); font-weight: 600; }
.account-login-field, .account-login-field label { min-width: 0; }
.account-login-input { display: flex; gap: 8px; align-items: center; }
.account-login-input input { flex: 1; min-width: 0; }
.account-login-input button, .account-login-code button { flex-shrink: 0; width: 36px; height: 36px; padding: 8px; }
.account-login-error { color: var(--color-text-danger); font-size: var(--font-size-sm); }
.account-login-code { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; font-size: var(--font-size-sm); }
.account-login-code output { font: 600 22px monospace; font-variant-numeric: tabular-nums; min-width: 8ch; }
</style>

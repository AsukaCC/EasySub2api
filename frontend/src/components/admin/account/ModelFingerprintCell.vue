<template>
  <button type="button" class="fingerprint-cell" :title="t('admin.accounts.fingerprint.title')" @click="emit('open', account)">
    <ModelFingerprintResult v-if="displaySnapshot" :snapshot="displaySnapshot" compact />
    <span v-else class="fingerprint-empty"><Icon name="chart" size="sm" />{{ t('admin.accounts.fingerprint.notTested') }}</span>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@/components/icons'
import type { Account } from '@/types'
import type { ModelFingerprintSnapshot } from '@/api/admin/modelFingerprint'
import { useModelFingerprint, useModelFingerprintExpiry } from '@/composables/useModelFingerprint'
import ModelFingerprintResult from './ModelFingerprintResult.vue'

const props = defineProps<{ account: Account }>()
const emit = defineEmits<{ open: [account: Account]; update: [accountId: string, snapshot: ModelFingerprintSnapshot | null] }>()
const { t } = useI18n()
const saved = computed(() => props.account.extra?.model_fingerprint as ModelFingerprintSnapshot | undefined)
const activeId = computed(() => saved.value?.status === 'running' ? props.account.id : null)
const { snapshot } = useModelFingerprint(activeId, value => emit('update', props.account.id, value))
const displaySnapshot = useModelFingerprintExpiry(
  computed(() => snapshot.value || saved.value || null),
  () => emit('update', props.account.id, null)
)
</script>

<style scoped>
.fingerprint-cell { display: block; width: 156px; max-width: 100%; min-height: 28px; padding: 4px; color: inherit; background: transparent; border: 0; cursor: pointer; border-radius: 4px; }
.fingerprint-cell:hover { background: var(--color-surface-hover); }
.fingerprint-cell:focus-visible { outline: 2px solid var(--color-primary); outline-offset: 2px; }
.fingerprint-empty { display: flex; align-items: center; gap: 6px; color: var(--color-text-tertiary); font-size: var(--font-size-xs); }
</style>

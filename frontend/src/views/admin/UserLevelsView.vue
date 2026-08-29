<template>
  <AppLayout>
    <div class="user-levels-view">
      <header class="user-levels-view__header">
        <div>
          <h1>{{ t('admin.users.levels.title') }}</h1>
          <p>{{ t('admin.users.levels.description') }}</p>
        </div>
      </header>
      <form class="user-levels-view__form" @submit.prevent="save">
        <div class="user-levels-view__grid">
          <label><span>{{ t('admin.users.levels.l2') }}</span><input v-model.number="form.l2_min_spend" class="input" type="number" min="0" step="0.01" /></label>
          <label><span>{{ t('admin.users.levels.l3') }}</span><input v-model.number="form.l3_min_spend" class="input" type="number" min="0" step="0.01" /></label>
        </div>
        <div class="user-levels-view__range">
          <div><b>L1</b><span>{{ t('admin.users.levels.rangeBelow', { amount: money(form.l2_min_spend) }) }}</span></div>
          <div><b>L2</b><span>{{ t('admin.users.levels.rangeBetween', { from: money(form.l2_min_spend), to: money(form.l3_min_spend) }) }}</span></div>
          <div><b>L3</b><span>{{ t('admin.users.levels.rangeAbove', { amount: money(form.l3_min_spend) }) }}</span></div>
        </div>
        <div class="user-levels-view__actions"><span v-if="message" class="user-levels-view__message">{{ message }}</span><button class="btn btn-primary" type="submit" :disabled="loading"><Icon v-if="loading" name="refresh" size="sm" />{{ loading ? t('common.saving') : t('common.save') }}</button></div>
      </form>
      <p class="user-levels-view__note">{{ t('admin.users.levels.windowNote') }}</p>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api'
import { useAppStore } from '@/stores/app'
import { formatPoints } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(false)
const message = ref('')
const form = reactive({ l2_min_spend: 50, l3_min_spend: 200 })
const money = (value: number) => formatPoints(Number(value || 0))

async function load() {
  try { const settings = await adminAPI.users.getLevelSettings(); form.l2_min_spend = settings.l2_min_spend; form.l3_min_spend = settings.l3_min_spend } catch (error) { appStore.showError(t('admin.users.levels.loadFailed')) }
}
async function save() {
  if (form.l3_min_spend <= form.l2_min_spend) { appStore.showError(t('admin.users.levels.invalidRange')); return }
  loading.value = true; message.value = ''
  try { await adminAPI.users.updateLevelSettings(form); message.value = t('admin.users.levels.saved') } catch (error) { appStore.showError(t('admin.users.levels.saveFailed')) } finally { loading.value = false }
}
onMounted(load)
</script>

<style scoped>
.user-levels-view { max-width: 980px; margin: 0 auto; padding: 2rem; }
.user-levels-view__header { margin-bottom: 1.5rem; }
.user-levels-view__header h1 { margin: 0; font-size: var(--font-size-2xl); }
.user-levels-view__header p, .user-levels-view__note { color: var(--color-text-secondary); }
.user-levels-view__form { border: 1px solid var(--glass-border); background: var(--glass-bg); border-radius: var(--radius-lg); padding: 1.5rem; }
.user-levels-view__grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 1rem; }
.user-levels-view__grid label { display: grid; gap: .4rem; color: var(--color-text-secondary); }
.user-levels-view__range { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: .75rem; margin-top: 1.5rem; }
.user-levels-view__range div { display: grid; gap: .3rem; padding: 1rem; border-left: 3px solid var(--color-primary); background: var(--color-surface-muted); }
.user-levels-view__range span { color: var(--color-text-secondary); font-size: var(--font-size-sm); }
.user-levels-view__actions { display: flex; justify-content: flex-end; align-items: center; gap: 1rem; margin-top: 1.5rem; }
.user-levels-view__message { color: var(--color-success); font-size: var(--font-size-sm); }
@media (max-width: 700px) { .user-levels-view { padding: 1rem; } .user-levels-view__grid, .user-levels-view__range { grid-template-columns: 1fr; } }
</style>

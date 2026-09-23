<template>
  <section class="opencode-settings">
    <h3>{{ t('upstreamUpdate.openCode') }}</h3>
    <form v-if="settings" class="opencode-settings__form" @submit.prevent="save">
      <label><input v-model="settings.enabled" type="checkbox" />{{ t('upstreamUpdate.autoRefresh') }}</label>
      <label>{{ t('upstreamUpdate.interval') }}<input v-model.number="settings.interval_minutes" class="input" type="number" min="5" max="1440" required /></label>
      <label>{{ t('upstreamUpdate.debounce') }}<input v-model.number="settings.debounce_minutes" class="input" type="number" min="1" max="60" required /></label>
      <button class="btn btn-secondary" type="submit" :disabled="busy">{{ t('common.save') }}</button>
    </form>
    <button v-else class="btn btn-secondary" :disabled="busy" @click="load"><Icon name="refresh" size="sm" />{{ t('upstreamUpdate.refresh') }}</button>
    <p v-if="error" role="alert">{{ error }}</p>
  </section>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { getOpenCodeSettings, saveOpenCodeSettings, type OpenCodeGoUsageSettings } from '@/api/admin/upstreamCapabilities'
const { t } = useI18n()
const settings = ref<OpenCodeGoUsageSettings>(), busy = ref(false), error = ref('')
async function load() { busy.value = true; error.value = ''; try { settings.value = await getOpenCodeSettings() } catch { error.value = t('upstreamUpdate.loadFailed') } finally { busy.value = false } }
async function save() { if (!settings.value) return; busy.value = true; error.value = ''; try { settings.value = await saveOpenCodeSettings(settings.value) } catch { error.value = t('upstreamUpdate.saveFailed') } finally { busy.value = false } }
onMounted(load)
</script>
<style scoped>
.opencode-settings { padding: 16px 0; border-top: 1px solid var(--border-color); }
.opencode-settings h3 { font-size: 14px; margin-bottom: 12px; }
.opencode-settings__form { display: flex; align-items: end; gap: 12px; flex-wrap: wrap; }
.opencode-settings label { display: grid; gap: 6px; font-size: 12px; }
.opencode-settings label:first-child { display: flex; align-items: center; align-self: center; }
.opencode-settings input[type=number] { width: 150px; max-width: 100%; }
</style>

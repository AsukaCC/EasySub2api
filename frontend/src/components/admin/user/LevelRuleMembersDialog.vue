<template>
  <BaseDialog :show="!!rule" :title="`${rule?.name ?? ''} · ${t('admin.users.levels.members')}`" width="extra-wide" @close="!busy && emit('close')">
    <div class="level-members">
      <div class="level-members__toolbar" role="tablist">
        <button v-for="tab in modes" :key="tab.value" type="button" role="tab" :aria-selected="mode === tab.value" :disabled="busy" class="btn" :class="mode === tab.value ? 'btn-primary' : 'btn-ghost'" @click="mode = tab.value; page = 1; void load()">{{ tab.label }}</button>
      </div>
      <form class="level-members__toolbar" @submit.prevent="page = 1; load()">
        <input v-model="search" class="input" :placeholder="t('admin.users.levels.memberSearch')" :aria-label="t('admin.users.levels.memberSearch')" />
        <button class="btn btn-secondary" type="submit" :disabled="loading || busy" :title="t('common.search')" :aria-label="t('common.search')"><Icon name="search" size="sm" /></button>
      </form>
      <p v-if="error" role="alert">{{ error }}</p>
      <div class="level-members__table">
        <table :aria-busy="loading">
          <thead><tr><th><input type="checkbox" :checked="rows.length > 0 && selected.length === rows.length" :disabled="loading || busy || rows.length === 0" :aria-label="t('common.selectAll')" @change="selected = ($event.target as HTMLInputElement).checked ? rows.map(u => u.id) : []" /></th><th>{{ t('admin.users.email') }}</th><th>{{ t('admin.users.username') }}</th><th>{{ t('common.actions') }}</th></tr></thead>
          <tbody>
            <tr v-if="loading"><td colspan="4">{{ t('common.loading') }}</td></tr>
            <tr v-else-if="!rows.length"><td colspan="4">{{ t('admin.users.levels.noMembers') }}</td></tr>
            <tr v-for="user in loading ? [] : rows" :key="user.id">
              <td><input v-model="selected" type="checkbox" :value="user.id" :disabled="busy" :aria-label="user.email" /></td><td>{{ user.email }}</td><td>{{ user.username || '-' }}</td>
              <td><button type="button" class="btn btn-secondary btn-sm" :disabled="busy || (mode === 'members' && !target)" @click="assign([user.id], mode === 'add' ? rule!.id : target)">{{ t(mode === 'add' ? 'admin.users.levels.addMember' : 'admin.users.levels.transfer') }}</button></td>
            </tr>
          </tbody>
        </table>
      </div>
      <Pagination :page="page" :page-size="20" :total="total" :show-page-size-selector="false" @update:page="changePage" />
      <div class="level-members__toolbar">
        <Select
          v-if="mode === 'members'"
          v-model="target"
          class="level-members__target"
          :options="targetOptions"
          :placeholder="t('admin.users.levels.targetRule')"
          :disabled="busy"
          :searchable="false"
          :aria-label="t('admin.users.levels.targetRule')"
        />
        <button type="button" class="btn btn-primary" :disabled="busy || loading || !selected.length || (mode === 'members' && !target)" @click="assign(selected, mode === 'add' ? rule!.id : target)"><Icon name="users" size="sm" />{{ t('admin.users.levels.applyMembers', { count: selected.length }) }}</button>
        <button v-if="mode === 'members' && !rule?.is_default" type="button" class="btn btn-secondary" :disabled="busy || loading || !selected.length" @click="assign(selected, '')">{{ t('admin.users.levels.resetDefault') }}</button>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { UserLevelRule } from '@/api/admin/users'
import type { AdminUser } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { extractApiErrorMessage } from '@/utils/apiError'

const props = defineProps<{ rule: UserLevelRule | null; rules: UserLevelRule[] }>()
const emit = defineEmits<{ close: []; changed: [] }>()
const { t } = useI18n()
const mode = ref<'members' | 'add'>('members')
const modes = computed(() => [{ value: 'members' as const, label: t('admin.users.levels.members') }, { value: 'add' as const, label: t('admin.users.levels.addMember') }])
const targetOptions = computed(() => props.rules.filter(r => r.enabled && r.id !== props.rule?.id).map(item => ({ value: item.id, label: item.name })))
const search = ref(''), target = ref(''), error = ref('')
const page = ref(1), total = ref(0)
const rows = ref<AdminUser[]>([]), selected = ref<string[]>([])
const loading = ref(false), busy = ref(false)
let loadVersion = 0

async function load() {
  const id = props.rule?.id
  const version = ++loadVersion
  if (!id) return
  loading.value = true
  error.value = ''
  selected.value = []
  try {
    const data = mode.value === 'members'
      ? await adminAPI.users.getLevelRuleMembers(id, page.value, 20, search.value)
      : await adminAPI.users.list(page.value, 20, { search: search.value })
    if (version !== loadVersion) return
    rows.value = data.items
    total.value = data.total
  } catch (e) {
    if (version === loadVersion) { rows.value = []; total.value = 0; error.value = extractApiErrorMessage(e, t('admin.users.levels.loadFailed')) }
  } finally { if (version === loadVersion) loading.value = false }
}
function changePage(value: number) { if (!busy.value) { page.value = value; void load() } }
async function assign(ids: string[], ruleID: string) {
  const name = props.rules.find(r => r.id === ruleID)?.name ?? t('admin.users.levels.defaultRule')
  if (!window.confirm(t('admin.users.levels.confirmTransfer', { count: ids.length, name }))) return
  busy.value = true
  try {
    await adminAPI.users.batchAssignLevelRules({ user_ids: [...ids], rule_ids: ruleID ? [ruleID] : [], operation: 'replace' })
    emit('changed')
    await load()
  } catch (e) { error.value = extractApiErrorMessage(e, t('admin.users.levels.saveFailed')) }
  finally { busy.value = false }
}
watch(() => props.rule?.id, () => { mode.value = 'members'; page.value = 1; search.value = ''; target.value = ''; rows.value = []; void load() }, { immediate: true })
</script>

<style scoped>
.level-members { display: flex; flex-direction: column; gap: 1rem; min-width: 0; }
.level-members__toolbar { display: flex; align-items: center; flex-wrap: wrap; gap: .5rem; }
.level-members__toolbar .input,
.level-members__target { flex: 1 1 180px; width: auto; min-width: 0; }
.level-members__table { overflow-x: auto; min-height: 160px; }
table { border-collapse: collapse; width: 100%; min-width: 520px; }
th, td { text-align: left; padding: .6rem; border-bottom: 1px solid var(--glass-border); overflow-wrap: anywhere; }
td:last-child { min-width: 90px; }
</style>

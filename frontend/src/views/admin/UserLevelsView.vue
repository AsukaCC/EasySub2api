<template>
  <AppLayout>
    <section class="user-level-rules-view">
      <header class="user-level-rules-view__header">
        <div>
          <h1>{{ t('admin.users.levels.title') }}</h1>
          <span class="user-level-rules-view__count">{{ rules.length }}</span>
        </div>
        <div class="user-level-rules-view__actions">
          <button class="btn btn-ghost" type="button" :disabled="loading" :aria-label="t('common.refresh')" @click="load">
            <Icon name="refresh" size="sm" :class="{ 'is-spinning': loading }" />
          </button>
          <button class="btn btn-primary" type="button" @click="beginCreate">
            <Icon name="plus" size="sm" />
            {{ t('admin.users.levels.create') }}
          </button>
        </div>
      </header>

      <div v-if="errorMessage" class="user-level-rules-view__error">{{ errorMessage }}</div>
      <div v-if="loading && rules.length === 0" class="user-level-rules-view__state">{{ t('common.loading') }}</div>
      <div v-else-if="rules.length === 0" class="user-level-rules-view__state">{{ t('admin.users.levels.empty') }}</div>
      <div v-else class="user-level-rules-view__table-wrap">
        <table class="user-level-rules-view__table">
          <thead>
            <tr>
              <th>{{ t('admin.users.levels.name') }}</th>
              <th>{{ t('admin.users.levels.window') }}</th>
              <th>{{ t('admin.users.levels.status') }}</th>
              <th>{{ t('admin.users.levels.tiers') }}</th>
              <th>{{ t('admin.users.levels.members') }}</th>
              <th>{{ t('admin.users.levels.references') }}</th>
              <th class="user-level-rules-view__actions-head">{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="rule in rules" :key="rule.id">
              <td>
                <strong>{{ rule.name }}</strong>
                <small>{{ rule.id }}</small>
              </td>
              <td><span class="user-level-rules-view__window">{{ rule.window_days }}d</span></td>
              <td>
                <span class="user-level-rules-view__status" :class="rule.enabled ? 'is-enabled' : 'is-disabled'">
                  {{ rule.enabled ? t('admin.users.levels.enabled') : t('admin.users.levels.disabled') }}
                </span>
              </td>
              <td>{{ rule.tiers.length }}</td>
              <td>{{ rule.assigned_user_count }}</td>
              <td>{{ rule.reference_count }}</td>
              <td class="user-level-rules-view__row-actions">
                <button class="btn btn-ghost btn-sm" type="button" :aria-label="t('common.edit')" @click="beginEdit(rule)">
                  <Icon name="edit" size="sm" />
                </button>
                <button class="btn btn-ghost btn-sm user-level-rules-view__delete" type="button" :aria-label="t('common.delete')" @click="removeRule(rule)">
                  <Icon name="trash" size="sm" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <BaseDialog :show="editorOpen" :title="editingID ? t('admin.users.levels.edit') : t('admin.users.levels.create')" width="extra-wide" @close="closeEditor">
      <form class="user-level-rule-editor" @submit.prevent="saveRule">
        <div class="user-level-rule-editor__meta">
          <label>
            <span>{{ t('admin.users.levels.name') }}</span>
            <input v-model="draft.name" class="input" type="text" maxlength="100" required />
          </label>
          <label>
            <span>{{ t('admin.users.levels.window') }}</span>
            <select v-model.number="draft.window_days" class="input">
              <option :value="7">7d</option>
              <option :value="14">14d</option>
              <option :value="30">30d</option>
            </select>
          </label>
          <label class="user-level-rule-editor__toggle">
            <span>{{ t('admin.users.levels.status') }}</span>
            <input v-model="draft.enabled" type="checkbox" />
            {{ draft.enabled ? t('admin.users.levels.enabled') : t('admin.users.levels.disabled') }}
          </label>
        </div>

        <div class="user-level-rule-editor__tiers-header">
          <div>
            <h4>{{ t('admin.users.levels.tiers') }}</h4>
            <span>{{ t('admin.users.levels.tierHint') }}</span>
          </div>
          <button class="btn btn-secondary btn-sm" type="button" @click="addTier">
            <Icon name="plus" size="sm" />
            {{ t('admin.users.levels.addTier') }}
          </button>
        </div>

        <div class="user-level-rule-editor__tiers">
          <article v-for="(tier, index) in draft.tiers" :key="tier.id || `new-${index}`" class="user-level-rule-editor__tier">
            <div class="user-level-rule-editor__tier-order">{{ index + 1 }}</div>
            <label>
              <span>{{ t('admin.users.levels.tierName') }}</span>
              <input v-model="tier.name" class="input" type="text" maxlength="100" required />
            </label>
            <label>
              <span>{{ t('admin.users.levels.minSpend') }}</span>
              <input v-model.number="tier.min_spend" class="input" type="number" min="0" step="0.01" :disabled="index === 0" />
            </label>
            <label>
              <span>{{ t('admin.users.levels.multiplier') }}</span>
              <input v-model="tier.default_multiplier" class="input" type="number" min="0.01" max="100" step="0.0001" placeholder="-" />
            </label>
            <div class="user-level-rule-editor__tier-actions">
              <button class="btn btn-ghost btn-sm" type="button" :disabled="index === 0" :aria-label="t('common.moveUp')" @click="moveTier(index, -1)"><Icon name="chevronUp" size="sm" /></button>
              <button class="btn btn-ghost btn-sm" type="button" :disabled="index === draft.tiers.length - 1" :aria-label="t('common.moveDown')" @click="moveTier(index, 1)"><Icon name="chevronDown" size="sm" /></button>
              <button class="btn btn-ghost btn-sm user-level-rules-view__delete" type="button" :disabled="index === 0" :aria-label="t('common.delete')" @click="removeTier(index)"><Icon name="trash" size="sm" /></button>
            </div>
          </article>
        </div>

        <div class="user-level-rule-editor__footer">
          <span v-if="editorError" class="user-level-rule-editor__error">{{ editorError }}</span>
          <div>
            <button class="btn btn-ghost" type="button" @click="closeEditor">{{ t('common.cancel') }}</button>
            <button class="btn btn-primary" type="submit" :disabled="saving">
              <Icon v-if="saving" name="refresh" size="sm" class="is-spinning" />
              {{ saving ? t('common.saving') : t('common.save') }}
            </button>
          </div>
        </div>
      </form>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api'
import type { UserLevelRule, UserLevelRuleTierInput } from '@/api/admin/users'
import { useAppStore } from '@/stores/app'

interface TierDraft {
  id?: string
  name: string
  sort_order: number
  min_spend: number
  default_multiplier?: number | string | null
}

const { t } = useI18n()
const appStore = useAppStore()
const rules = ref<UserLevelRule[]>([])
const loading = ref(false)
const saving = ref(false)
const errorMessage = ref('')
const editorError = ref('')
const editorOpen = ref(false)
const editingID = ref('')
const draft = reactive<{ name: string; window_days: 7 | 14 | 30; enabled: boolean; tiers: TierDraft[] }>({
  name: '', window_days: 7, enabled: true, tiers: []
})

async function load() {
  loading.value = true
  errorMessage.value = ''
  try {
    rules.value = await adminAPI.users.listLevelRules()
  } catch {
    errorMessage.value = t('admin.users.levels.loadFailed')
  } finally {
    loading.value = false
  }
}

function resetDraft() {
  draft.name = ''
  draft.window_days = 7
  draft.enabled = true
  draft.tiers = [{ id: '', name: t('admin.users.levels.baseTier'), sort_order: 0, min_spend: 0, default_multiplier: null }]
  editorError.value = ''
}

function beginCreate() {
  editingID.value = ''
  resetDraft()
  editorOpen.value = true
}

function beginEdit(rule: UserLevelRule) {
  editingID.value = rule.id
  draft.name = rule.name
  draft.window_days = rule.window_days
  draft.enabled = rule.enabled
  draft.tiers = rule.tiers.slice().sort((a, b) => a.sort_order - b.sort_order).map((tier) => ({
    id: tier.id,
    name: tier.name,
    sort_order: tier.sort_order,
    min_spend: tier.min_spend,
    default_multiplier: tier.default_multiplier ?? null
  }))
  editorError.value = ''
  editorOpen.value = true
}

function closeEditor() {
  if (saving.value) return
  editorOpen.value = false
}

function addTier() {
  const previous = draft.tiers[draft.tiers.length - 1]
  draft.tiers.push({
    id: '', name: `${t('admin.users.levels.tier')} ${draft.tiers.length + 1}`,
    sort_order: draft.tiers.length, min_spend: Number(previous?.min_spend || 0) + 1,
    default_multiplier: null
  })
}

function removeTier(index: number) {
  if (index === 0) return
  draft.tiers.splice(index, 1)
  syncTierOrder()
}

function moveTier(index: number, direction: -1 | 1) {
  const target = index + direction
  if (index <= 0 || target <= 0 || target >= draft.tiers.length) return
  const [tier] = draft.tiers.splice(index, 1)
  draft.tiers.splice(target, 0, tier)
  syncTierOrder()
}

function syncTierOrder() {
  draft.tiers.forEach((tier, index) => {
    tier.sort_order = index
    if (index === 0) tier.min_spend = 0
  })
}

function normalizedTiers(): UserLevelRuleTierInput[] {
  syncTierOrder()
  return draft.tiers.map((tier) => ({
    id: tier.id || undefined,
    name: tier.name.trim(),
    sort_order: tier.sort_order,
    min_spend: Number(tier.min_spend || 0),
    default_multiplier: tier.default_multiplier === '' || tier.default_multiplier == null
      ? null
      : Number(tier.default_multiplier)
  }))
}

async function saveRule() {
  editorError.value = ''
  const tiers = normalizedTiers()
  for (let index = 1; index < tiers.length; index += 1) {
    if (tiers[index].min_spend <= tiers[index - 1].min_spend) {
      editorError.value = t('admin.users.levels.invalidThresholds')
      return
    }
  }
  saving.value = true
  try {
    if (editingID.value) {
      await adminAPI.users.updateLevelRule(editingID.value, {
        name: draft.name,
        window_days: draft.window_days,
        enabled: draft.enabled,
        tiers
      })
    } else {
      const created = await adminAPI.users.createLevelRule({ name: draft.name, window_days: draft.window_days })
      const createdBaseID = created.tiers?.find((tier) => tier.sort_order === 0)?.id || created.tiers?.[0]?.id
      if (createdBaseID && tiers[0] && !tiers[0].id) {
        tiers[0].id = createdBaseID
      }
      const needsFollowUpUpdate =
        created.enabled !== draft.enabled ||
        tiers.length !== 1 ||
        tiers[0]?.name !== t('admin.users.levels.baseTier') ||
        tiers[0]?.default_multiplier != null
      if (needsFollowUpUpdate) {
        await adminAPI.users.updateLevelRule(created.id, {
          name: draft.name,
          window_days: draft.window_days,
          enabled: draft.enabled,
          tiers
        })
      }
    }
    editorOpen.value = false
    await load()
  } catch {
    editorError.value = t('admin.users.levels.saveFailed')
    appStore.showError(t('admin.users.levels.saveFailed'))
  } finally {
    saving.value = false
  }
}

async function removeRule(rule: UserLevelRule) {
  if (rule.reference_count > 0 || rule.assigned_user_count > 0) {
    appStore.showError(t('admin.users.levels.deleteReferenced'))
    return
  }
  if (!window.confirm(t('admin.users.levels.confirmDelete', { name: rule.name }))) return
  try {
    await adminAPI.users.deleteLevelRule(rule.id)
    await load()
  } catch {
    appStore.showError(t('admin.users.levels.deleteFailed'))
  }
}

onMounted(load)
</script>

<style scoped>
.user-level-rules-view { width: min(1180px, 100%); margin: 0 auto; padding: 1.5rem; }
.user-level-rules-view__header { display: flex; align-items: center; justify-content: space-between; gap: 1rem; margin-bottom: 1rem; }
.user-level-rules-view__header > div:first-child { display: flex; align-items: center; gap: .65rem; }
.user-level-rules-view__header h1 { margin: 0; font-size: 1.35rem; }
.user-level-rules-view__count { min-width: 1.6rem; padding: .15rem .45rem; border-radius: 999px; background: var(--color-surface-muted); color: var(--color-text-secondary); text-align: center; font-size: var(--font-size-sm); }
.user-level-rules-view__actions, .user-level-rules-view__row-actions { display: flex; align-items: center; gap: .4rem; }
.user-level-rules-view__table-wrap { overflow-x: auto; border: 1px solid var(--glass-border); background: var(--glass-bg); border-radius: 8px; }
.user-level-rules-view__table { width: 100%; min-width: 760px; border-collapse: collapse; }
.user-level-rules-view__table th, .user-level-rules-view__table td { padding: .8rem .9rem; border-bottom: 1px solid var(--glass-border); text-align: left; white-space: nowrap; }
.user-level-rules-view__table th { color: var(--color-text-secondary); font-size: var(--font-size-sm); font-weight: 600; }
.user-level-rules-view__table tbody tr:last-child td { border-bottom: 0; }
.user-level-rules-view__table td:first-child { min-width: 220px; }
.user-level-rules-view__table td:first-child small { display: block; max-width: 220px; overflow: hidden; color: var(--color-text-tertiary); font-size: .7rem; text-overflow: ellipsis; }
.user-level-rules-view__window { color: var(--color-primary); font-variant-numeric: tabular-nums; }
.user-level-rules-view__status { display: inline-flex; align-items: center; padding: .2rem .5rem; border-radius: 999px; font-size: var(--font-size-sm); }
.user-level-rules-view__status.is-enabled { background: color-mix(in srgb, var(--color-success) 14%, transparent); color: var(--color-success); }
.user-level-rules-view__status.is-disabled { background: var(--color-surface-muted); color: var(--color-text-secondary); }
.user-level-rules-view__actions-head { text-align: right !important; }
.user-level-rules-view__row-actions { justify-content: flex-end; }
.user-level-rules-view__delete { color: var(--color-danger); }
.user-level-rules-view__error, .user-level-rule-editor__error { color: var(--color-danger); font-size: var(--font-size-sm); }
.user-level-rules-view__state { padding: 3rem 1rem; border: 1px dashed var(--glass-border); color: var(--color-text-secondary); text-align: center; }
.user-level-rule-editor { display: grid; gap: 1.25rem; }
.user-level-rule-editor__meta { display: grid; grid-template-columns: minmax(0, 1.5fr) 180px 180px; gap: 1rem; }
.user-level-rule-editor label { display: grid; gap: .4rem; color: var(--color-text-secondary); font-size: var(--font-size-sm); }
.user-level-rule-editor__toggle { align-content: start; }
.user-level-rule-editor__toggle input { width: 1rem; height: 1rem; }
.user-level-rule-editor__tiers-header { display: flex; align-items: center; justify-content: space-between; gap: 1rem; padding-top: .25rem; border-top: 1px solid var(--glass-border); }
.user-level-rule-editor__tiers-header h4 { margin: 1rem 0 .25rem; color: var(--color-text-primary); }
.user-level-rule-editor__tiers-header span { color: var(--color-text-secondary); font-size: var(--font-size-sm); }
.user-level-rule-editor__tiers { display: grid; gap: .65rem; }
.user-level-rule-editor__tier { display: grid; grid-template-columns: 2rem minmax(150px, 1.2fr) minmax(130px, 1fr) minmax(130px, 1fr) auto; align-items: end; gap: .7rem; padding: .75rem; border: 1px solid var(--glass-border); border-radius: 6px; background: var(--color-surface-muted); }
.user-level-rule-editor__tier-order { align-self: center; color: var(--color-text-secondary); font-variant-numeric: tabular-nums; text-align: center; }
.user-level-rule-editor__tier-actions { display: flex; gap: .2rem; }
.user-level-rule-editor__footer { display: flex; align-items: center; justify-content: space-between; gap: 1rem; padding-top: .5rem; border-top: 1px solid var(--glass-border); }
.user-level-rule-editor__footer > div { display: flex; gap: .5rem; }
.is-spinning { animation: user-level-rules-spin .8s linear infinite; }
@keyframes user-level-rules-spin { to { transform: rotate(360deg); } }
@media (max-width: 760px) {
  .user-level-rules-view { padding: 1rem; }
  .user-level-rules-view__header { align-items: flex-start; }
  .user-level-rule-editor__meta, .user-level-rule-editor__tier { grid-template-columns: 1fr; }
  .user-level-rule-editor__tier-order { display: none; }
  .user-level-rule-editor__footer { align-items: flex-start; flex-direction: column; }
}
</style>

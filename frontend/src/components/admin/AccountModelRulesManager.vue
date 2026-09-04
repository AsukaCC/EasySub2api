<template>
  <div class="account-model-rules-page">
    <div class="account-model-rules-page__toolbar">
      <p class="input-hint account-model-rules-page__description">
        {{ t('admin.accounts.modelRules.description') }}
      </p>
      <div class="account-model-rules-page__toolbar-actions">
        <button type="button" class="icon-btn" :disabled="loading" :title="t('common.refresh')" @click="loadRules">
          <Icon name="refresh" size="sm" :class="loading ? 'account-model-rules-page__spin' : ''" />
        </button>
        <Select v-model="platformFilter" :options="platformFilterOptions" class="account-model-rules-page__filter" />
        <Select v-model="tierFilter" :options="tierFilterOptions" class="account-model-rules-page__filter" :disabled="loadingFilterTiers" />
        <button type="button" class="btn btn-primary" @click="openCreate">
          <Icon name="plus" size="sm" />
          <span>{{ t('admin.accounts.modelRules.createRule') }}</span>
        </button>
      </div>
    </div>

    <div class="account-model-rules-page__surface">
      <LoadingState v-if="loading" variant="section" :label="t('common.loading')" />
      <div v-else-if="rules.length === 0" class="account-model-rules-page__empty">
        <Icon name="swap" size="lg" />
        <strong>{{ t('admin.accounts.modelRules.noRules') }}</strong>
        <span>{{ t('admin.accounts.modelRules.createFirst') }}</span>
      </div>
      <div v-else class="account-model-rules-page__table-wrap">
        <table class="account-model-rules-page__table">
          <thead>
            <tr>
              <th>{{ t('admin.accounts.modelRules.name') }}</th>
              <th>{{ t('admin.accounts.modelRules.scope') }}</th>
              <th>{{ t('admin.accounts.modelRules.routeConfig') }}</th>
              <th>{{ t('admin.accounts.modelRules.boundAccounts') }}</th>
              <th>{{ t('admin.accounts.modelRules.descriptionLabel') }}</th>
              <th>{{ t('admin.accounts.modelRules.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="rule in rules" :key="rule.id">
              <td class="account-model-rules-page__name">{{ rule.name }}</td>
              <td>
                <span class="account-model-rules-page__scope">
                  <PlatformIcon :platform="rule.platform" size="xs" />
                  {{ platformLabel(rule.platform) }} · {{ scopeTierLabel(rule.subscription_tier) }}
                </span>
              </td>
              <td>
                <span>{{ t('admin.accounts.modelRules.routeCount', { count: (rule.model_routes || []).length }) }}</span>
                <span class="account-model-rules-page__preview">{{ routePreview(rule) }}</span>
              </td>
              <td>{{ rule.bound_account_count || 0 }}</td>
              <td>{{ rule.description || '—' }}</td>
              <td>
                <div class="account-model-rules-page__row-actions">
                  <button type="button" class="icon-btn" :title="t('common.edit')" @click="openEdit(rule)">
                    <Icon name="edit" size="sm" />
                  </button>
                  <button
                    type="button"
                    class="icon-btn icon-btn--danger"
                    :disabled="(rule.bound_account_count || 0) > 0"
                    :title="(rule.bound_account_count || 0) > 0 ? t('admin.accounts.modelRules.deleteBoundHint') : t('common.delete')"
                    @click="openDelete(rule)"
                  >
                    <Icon name="trash" size="sm" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>

  <BaseDialog
    :show="showForm"
    :title="editingRule ? t('admin.accounts.modelRules.editRule') : t('admin.accounts.modelRules.createRule')"
    width="wide"
    @close="closeForm"
  >
    <form class="account-model-rules-form" @submit.prevent="handleSubmit">
      <div class="account-model-rules-form__grid">
        <div>
          <label class="input-label">{{ t('admin.accounts.modelRules.name') }}</label>
          <input v-model="form.name" class="input" maxlength="100" required />
        </div>
        <div>
          <label class="input-label">{{ t('admin.accounts.modelRules.platform') }}</label>
          <Select v-model="form.platform" data-testid="rule-platform" :options="platformOptions" :disabled="scopeLocked" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.accounts.modelRules.subscriptionTier') }}</label>
          <Select v-model="form.subscriptionTier" data-testid="rule-tier" :options="formTierOptions" :disabled="scopeLocked || loadingFormTiers" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.accounts.modelRules.syncSourceAccount') }}</label>
          <Select
            v-model="form.sourceAccountId"
            data-testid="rule-source-account"
            :options="sourceAccountOptions"
            :disabled="loadingSourceAccounts || sourceAccounts.length === 0"
            searchable="auto"
            clearable
          />
        </div>
      </div>

      <div>
        <label class="input-label">{{ t('admin.accounts.modelRules.descriptionLabel') }}</label>
        <textarea v-model="form.description" class="input account-model-rules-form__textarea" rows="2" />
      </div>

      <div class="account-model-rules-form__routes-header">
        <div>
          <label class="input-label">{{ t('admin.accounts.modelRules.routeConfig') }}</label>
          <p class="input-hint">{{ t('admin.accounts.modelRules.emptyRoutesHint') }}</p>
        </div>
        <div class="account-model-rules-form__actions">
          <button type="button" class="btn btn-secondary btn-sm" :disabled="!form.sourceAccountId || syncingModels" @click="syncModels">
            <LoadingButtonContent :loading="syncingModels" :loading-text="t('admin.accounts.modelRules.loadingModels')">
              <Icon name="sync" size="sm" />
              <span>{{ t('admin.accounts.modelRules.syncUpstreamModels') }}</span>
            </LoadingButtonContent>
          </button>
          <button type="button" class="btn btn-secondary btn-sm" :disabled="syncedModels.length === 0" @click="importSyncedModels">
            <Icon name="download" size="sm" />
            <span>{{ t('admin.accounts.modelRules.importSyncedModels') }}</span>
          </button>
          <button type="button" class="btn btn-secondary btn-sm" @click="addRoute">
            <Icon name="plus" size="sm" />
            <span>{{ t('admin.accounts.modelRules.addRoute') }}</span>
          </button>
        </div>
      </div>

      <datalist id="account-model-rule-synced-models">
        <option v-for="model in syncedModels" :key="model" :value="model" />
      </datalist>

      <div v-if="form.routes.length === 0" class="account-model-rules-form__empty-routes">
        {{ t('admin.accounts.modelRules.emptyRoutes') }}
      </div>
      <div v-else class="account-model-rules-form__route-list">
        <div v-for="(route, index) in form.routes" :key="route.id" class="account-model-rules-form__route-row">
          <input v-model="route.request_model" class="input" list="account-model-rule-synced-models" :placeholder="t('admin.accounts.modelRules.fromModel')" />
          <span aria-hidden="true">→</span>
          <input v-model="route.upstream_model" class="input" list="account-model-rule-synced-models" :placeholder="t('admin.accounts.modelRules.toModel')" />
          <Select
            v-if="form.platform === 'openai'"
            v-model="route.reasoning_effort"
            data-testid="rule-reasoning-effort"
            :options="reasoningEffortOptions"
            :placeholder="t('admin.accounts.modelRules.reasoningEffortFollowRequest')"
            clearable
          />
          <button type="button" class="icon-btn icon-btn--danger" :title="t('admin.accounts.modelRules.removeRoute')" @click="removeRoute(index)">
            <Icon name="trash" size="sm" />
          </button>
        </div>
      </div>
      <p v-if="syncWarning" class="input-hint account-model-rules-form__warning">{{ syncWarning }}</p>
      <p v-if="formError" class="account-model-rules-form__error">{{ formError }}</p>
    </form>

    <template #footer>
      <button type="button" class="btn btn-secondary" @click="closeForm">{{ t('common.cancel') }}</button>
      <button type="button" class="btn btn-primary" :disabled="submitting" :aria-busy="submitting" @click="handleSubmit">
        <LoadingButtonContent :loading="submitting" :loading-text="t('common.saving')">
          {{ t('admin.accounts.modelRules.save') }}
        </LoadingButtonContent>
      </button>
    </template>
  </BaseDialog>

  <ConfirmDialog
    :show="showDeleteConfirm"
    :title="t('admin.accounts.modelRules.deleteRule')"
    :message="t('admin.accounts.modelRules.deleteConfirm', { name: deletingRule?.name || '' })"
    :confirm-text="t('common.delete')"
    :cancel-text="t('common.cancel')"
    danger
    @confirm="confirmDelete"
    @cancel="cancelDelete"
  />
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import LoadingButtonContent from '@/components/common/LoadingButtonContent.vue'
import LoadingState from '@/components/common/LoadingState.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import { adminAPI } from '@/api/admin'
import type { AccountModelRoute, AccountModelRule } from '@/api/admin/accountModelRules'
import type { Account, AccountPlatform } from '@/types'
import { accountPlatformOptions } from '@/utils/accountPlatforms'
import { isValidWildcardPattern } from '@/composables/useModelWhitelist'
import { useAppStore } from '@/stores/app'

interface RouteRow extends AccountModelRoute { id: number }

const { t } = useI18n()
const appStore = useAppStore()
const rules = ref<AccountModelRule[]>([])
const loading = ref(false)
const submitting = ref(false)
const platformFilter = ref<AccountPlatform | ''>('')
const tierFilter = ref('')
const filterTiers = ref<Array<{ value: string; label: string; account_count: number }>>([])
const loadingFilterTiers = ref(false)
const formTiers = ref<Array<{ value: string; label: string; account_count: number }>>([])
const loadingFormTiers = ref(false)
const sourceAccounts = ref<Account[]>([])
const loadingSourceAccounts = ref(false)
const syncingModels = ref(false)
const syncedModels = ref<string[]>([])
const syncWarning = ref('')
const showForm = ref(false)
const editingRule = ref<AccountModelRule | null>(null)
const deletingRule = ref<AccountModelRule | null>(null)
const showDeleteConfirm = ref(false)
const formError = ref('')
let nextRouteID = 1
let formScopeVersion = 0

const form = reactive({
  name: '',
  description: '',
  platform: 'anthropic' as AccountPlatform,
  subscriptionTier: '',
  sourceAccountId: '',
  routes: [] as RouteRow[]
})

const platformOptions = computed(() => accountPlatformOptions(t))
const platformFilterOptions = computed(() => [{ value: '', label: t('admin.accounts.allPlatforms') }, ...platformOptions.value])
const tierFilterOptions = computed(() => [
  { value: '', label: t('admin.accounts.allSubscriptionTiers') },
  { value: '__all__', label: t('admin.accounts.modelRules.allTiersOnly') },
  ...filterTiers.value.filter(tier => tier.value !== '__unrecognized__').map(tier => ({ value: tier.value, label: `${tier.label} (${tier.account_count})` }))
])
const formTierOptions = computed(() => [
  { value: '', label: t('admin.accounts.modelRules.allTiers') },
  ...formTiers.value.filter(tier => tier.value !== '__unrecognized__').map(tier => ({ value: tier.value, label: `${tier.label} (${tier.account_count})` }))
])
const sourceAccountOptions = computed(() => sourceAccounts.value.map(account => ({
  value: account.id,
  label: `${account.name}${account.subscription_tier ? ` · ${account.subscription_tier}` : ''}`
})))
const reasoningEffortOptions = computed(() => ['minimal', 'low', 'medium', 'high', 'xhigh', 'max'].map(value => ({ value, label: value })))
const scopeLocked = computed(() => Boolean(editingRule.value && (editingRule.value.bound_account_count || 0) > 0))

const platformLabel = (platform: string) => t(`admin.accounts.platforms.${platform}`, platform)
const scopeTierLabel = (tier: string | null) => tier || t('admin.accounts.modelRules.allTiers')
const routePreview = (rule: AccountModelRule) => (rule.model_routes || []).slice(0, 2).map(route => `${route.request_model} → ${route.upstream_model}`).join(', ') || t('admin.accounts.modelRules.passthroughAll')

async function loadTierOptions(platform: string, target: 'filter' | 'form') {
  const loadingRef = target === 'filter' ? loadingFilterTiers : loadingFormTiers
  loadingRef.value = true
  try {
    const tiers = await adminAPI.accounts.listSubscriptionTiers(platform)
    if (target === 'filter') filterTiers.value = tiers
    else formTiers.value = tiers
  } catch (error) {
    if (target === 'filter') filterTiers.value = []
    else formTiers.value = []
    console.error('Failed to load subscription tiers:', error)
  } finally {
    loadingRef.value = false
  }
}

async function loadRules() {
  loading.value = true
  try {
    rules.value = await adminAPI.accountModelRules.list(platformFilter.value, tierFilter.value)
  } catch (error) {
    appStore.showError(t('admin.accounts.modelRules.loadFailed'))
    console.error('Failed to load account model rules:', error)
  } finally {
    loading.value = false
  }
}

async function loadSourceAccounts() {
  const version = ++formScopeVersion
  loadingSourceAccounts.value = true
  form.sourceAccountId = ''
  sourceAccounts.value = []
  syncedModels.value = []
  try {
    const accounts: Account[] = []
    let page = 1
    let total = 0
    do {
      const result = await adminAPI.accounts.list(page, 100, {
        platform: form.platform,
        ...(form.subscriptionTier ? { subscription_tier: form.subscriptionTier } : {}),
        sort_by: 'name',
        sort_order: 'asc'
      })
      if (version !== formScopeVersion) return
      accounts.push(...result.items)
      total = result.total
      page += 1
    } while (accounts.length < total)
    sourceAccounts.value = accounts
  } catch (error) {
    if (version === formScopeVersion) console.error('Failed to load source accounts:', error)
  } finally {
    if (version === formScopeVersion) loadingSourceAccounts.value = false
  }
}

function resetForm() {
  form.name = ''
  form.description = ''
  form.platform = 'anthropic'
  form.subscriptionTier = ''
  form.sourceAccountId = ''
  form.routes = []
  syncedModels.value = []
  syncWarning.value = ''
  formError.value = ''
}

function openCreate() {
  editingRule.value = null
  resetForm()
  showForm.value = true
  void loadTierOptions(form.platform, 'form')
  void loadSourceAccounts()
}

function openEdit(rule: AccountModelRule) {
  editingRule.value = rule
  form.name = rule.name
  form.description = rule.description || ''
  form.platform = rule.platform
  form.subscriptionTier = rule.subscription_tier || ''
  form.sourceAccountId = ''
  form.routes = (rule.model_routes || []).map(route => ({ id: nextRouteID++, ...route, reasoning_effort: route.reasoning_effort || '' }))
  syncedModels.value = []
  syncWarning.value = ''
  formError.value = ''
  showForm.value = true
  void loadTierOptions(form.platform, 'form')
  void loadSourceAccounts()
}

function closeForm() {
  showForm.value = false
  editingRule.value = null
  formScopeVersion += 1
  resetForm()
}

function addRoute() {
  form.routes.push({ id: nextRouteID++, request_model: '', upstream_model: '', reasoning_effort: '' })
}

function removeRoute(index: number) {
  form.routes.splice(index, 1)
}

async function syncModels() {
  if (!form.sourceAccountId) return
  syncingModels.value = true
  syncWarning.value = ''
  try {
    const result = await adminAPI.accounts.syncUpstreamModels(form.sourceAccountId)
    syncedModels.value = Array.from(new Set(result.models.map(model => model.trim()).filter(Boolean))).sort()
    syncWarning.value = result.warnings?.map(warning => warning.message).join('；') || ''
    if (syncedModels.value.length === 0) appStore.showInfo(t('admin.accounts.modelRules.noModels'))
  } catch (error: any) {
    const message = error?.response?.data?.detail || error?.message
    appStore.showError(message || t('admin.accounts.modelRules.loadModelsFailed'))
  } finally {
    syncingModels.value = false
  }
}

function importSyncedModels() {
  const existing = new Set(form.routes.map(route => route.request_model.trim()).filter(Boolean))
  for (const model of syncedModels.value) {
    if (!existing.has(model)) {
      form.routes.push({ id: nextRouteID++, request_model: model, upstream_model: model, reasoning_effort: '' })
      existing.add(model)
    }
  }
}

function buildRoutes(): AccountModelRoute[] | null {
  const routes: AccountModelRoute[] = []
  const seen = new Set<string>()
  for (const row of form.routes) {
    const requestModel = row.request_model.trim()
    const upstreamModel = row.upstream_model.trim()
    if (!requestModel && !upstreamModel) continue
    if (!requestModel || !upstreamModel || !isValidWildcardPattern(requestModel) || upstreamModel.includes('*') || seen.has(requestModel)) {
      formError.value = t('admin.accounts.modelRules.routeInvalid')
      return null
    }
    seen.add(requestModel)
    routes.push({
      request_model: requestModel,
      upstream_model: upstreamModel,
      ...(form.platform === 'openai' && row.reasoning_effort ? { reasoning_effort: row.reasoning_effort } : {})
    })
  }
  return routes
}

async function handleSubmit() {
  formError.value = ''
  if (!form.name.trim()) {
    formError.value = t('admin.accounts.modelRules.nameRequired')
    return
  }
  if (!editingRule.value && !form.sourceAccountId) {
    formError.value = t('admin.accounts.modelRules.sourceAccountRequired')
    return
  }
  const routes = buildRoutes()
  if (!routes) return
  submitting.value = true
  try {
    const payload = {
      name: form.name.trim(),
      description: form.description.trim() || null,
      platform: form.platform,
      subscription_tier: form.subscriptionTier || null,
      model_routes: routes
    }
    if (editingRule.value) {
      await adminAPI.accountModelRules.update(editingRule.value.id, payload)
      appStore.showSuccess(t('admin.accounts.modelRules.updateSuccess'))
    } else {
      await adminAPI.accountModelRules.create(payload)
      appStore.showSuccess(t('admin.accounts.modelRules.createSuccess'))
    }
    closeForm()
    await loadRules()
  } catch (error: any) {
    const message = error?.response?.data?.detail || error?.response?.data?.message || error?.message
    appStore.showError(message || t('admin.accounts.modelRules.saveFailed'))
  } finally {
    submitting.value = false
  }
}

function openDelete(rule: AccountModelRule) {
  if ((rule.bound_account_count || 0) > 0) return
  deletingRule.value = rule
  showDeleteConfirm.value = true
}

function cancelDelete() {
  deletingRule.value = null
  showDeleteConfirm.value = false
}

async function confirmDelete() {
  if (!deletingRule.value) return
  try {
    await adminAPI.accountModelRules.delete(deletingRule.value.id)
    appStore.showSuccess(t('admin.accounts.modelRules.deleteSuccess'))
    cancelDelete()
    await loadRules()
  } catch (error: any) {
    appStore.showError(error?.response?.data?.message || t('admin.accounts.modelRules.deleteFailed'))
  }
}

watch(platformFilter, async platform => {
  tierFilter.value = ''
  await loadTierOptions(platform, 'filter')
  await loadRules()
})
watch(tierFilter, () => void loadRules())
watch(() => form.platform, async () => {
  if (!showForm.value || scopeLocked.value) return
  form.subscriptionTier = ''
  await loadTierOptions(form.platform, 'form')
  await loadSourceAccounts()
}, { flush: 'sync' })
watch(() => form.subscriptionTier, () => {
  if (showForm.value) void loadSourceAccounts()
}, { flush: 'sync' })

onMounted(async () => {
  await loadTierOptions('', 'filter')
  await loadRules()
})
</script>

<style scoped>
.account-model-rules-page { display: flex; height: calc(100vh - var(--app-shell-height) - 3.25rem); min-height: 0; flex-direction: column; gap: 1rem; }
.account-model-rules-page__toolbar, .account-model-rules-page__toolbar-actions, .account-model-rules-page__scope, .account-model-rules-page__row-actions, .account-model-rules-form__routes-header, .account-model-rules-form__actions { display: flex; align-items: center; gap: .5rem; }
.account-model-rules-page__toolbar { justify-content: space-between; gap: 1rem; }
.account-model-rules-page__description { margin: 0; max-width: 54rem; }
.account-model-rules-page__filter { min-width: 12rem; }
.account-model-rules-page__surface { flex: 1; min-height: 0; overflow: hidden; border: 1px solid var(--glass-border); border-radius: 8px; background: var(--glass-bg); }
.account-model-rules-page__empty { display: flex; min-height: 12rem; height: 100%; flex-direction: column; align-items: center; justify-content: center; gap: .5rem; color: var(--color-text-secondary); }
.account-model-rules-page__table-wrap { height: 100%; overflow: auto; }
.account-model-rules-page__table { width: 100%; min-width: 850px; border-collapse: collapse; }
.account-model-rules-page__table th, .account-model-rules-page__table td { padding: .75rem; border-bottom: 1px solid var(--color-border); text-align: left; vertical-align: top; }
.account-model-rules-page__table th { color: var(--color-text-secondary); font-size: var(--font-size-xs); }
.account-model-rules-page__name { font-weight: 600; }
.account-model-rules-page__preview { display: block; max-width: 20rem; overflow: hidden; color: var(--color-text-secondary); font-size: var(--font-size-xs); text-overflow: ellipsis; white-space: nowrap; }
.account-model-rules-form { display: flex; flex-direction: column; gap: 1rem; }
.account-model-rules-form__grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 1rem; }
.account-model-rules-form__textarea { resize: vertical; }
.account-model-rules-form__routes-header { justify-content: space-between; align-items: flex-start; }
.account-model-rules-form__routes-header .input-hint { margin: .25rem 0 0; }
.account-model-rules-form__actions { flex-wrap: wrap; justify-content: flex-end; }
.account-model-rules-form__route-list { display: flex; flex-direction: column; gap: .5rem; }
.account-model-rules-form__route-row { display: grid; grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr) minmax(9rem, .6fr) auto; align-items: center; gap: .5rem; }
.account-model-rules-form__empty-routes { padding: 1rem; border: 1px dashed var(--color-border); border-radius: 6px; color: var(--color-text-secondary); text-align: center; }
.account-model-rules-form__error { margin: 0; color: var(--color-text-danger); }
.account-model-rules-form__warning { margin: 0; color: var(--color-text-warning); }
.icon-btn { display: inline-flex; width: 2.25rem; height: 2.25rem; align-items: center; justify-content: center; border: 1px solid var(--color-border); border-radius: 6px; background: transparent; color: inherit; cursor: pointer; }
.icon-btn--danger { color: var(--color-text-danger); }
.icon-btn:disabled { cursor: not-allowed; opacity: .45; }
.account-model-rules-page__spin { animation: spin 1s linear infinite; }
@media (max-width: 800px) { .account-model-rules-page__toolbar, .account-model-rules-page__toolbar-actions, .account-model-rules-form__routes-header { align-items: stretch; flex-direction: column; } .account-model-rules-page__filter { width: 100%; } .account-model-rules-form__grid { grid-template-columns: 1fr; } .account-model-rules-form__route-row { grid-template-columns: 1fr; } .account-model-rules-form__route-row > span { display: none; } }
</style>

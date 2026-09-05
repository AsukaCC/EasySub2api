<template>
  <div class="components-admin-account-account-table-filters__panel">
    <SearchInput
      :model-value="searchQuery"
      :placeholder="t('admin.accounts.searchAccounts')"
      class="components-admin-account-account-table-filters__search-input"
      @update:model-value="$emit('update:searchQuery', $event)"
      @search="$emit('change')"
    />
    <Select :model-value="filters.platform" class="components-admin-account-account-table-filters__field" :options="pOpts" @update:model-value="updatePlatform" @change="$emit('change')" />
    <Select :model-value="filters.subscription_tier" class="components-admin-account-account-table-filters__field" :options="tierOpts" :disabled="loadingTiers" @update:model-value="updateSubscriptionTier" @change="$emit('change')" />
    <Select :model-value="filters.type" class="components-admin-account-account-table-filters__field" :options="tOpts" @update:model-value="updateType" @change="$emit('change')" />
    <Select :model-value="filters.expiry_status" class="components-admin-account-account-table-filters__field" :options="expiryOpts" @update:model-value="updateExpiryStatus" @change="$emit('change')" />
    <Select :model-value="filters.status" class="components-admin-account-account-table-filters__field" :options="sOpts" @update:model-value="updateStatus" @change="$emit('change')" />
    <Select :model-value="filters.privacy_mode" class="components-admin-account-account-table-filters__field" :options="privacyOpts" @update:model-value="updatePrivacyMode" @change="$emit('change')" />
    <Select :model-value="filters.group" class="components-admin-account-account-table-filters__field" :options="gOpts" @update:model-value="updateGroup" @change="$emit('change')" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'; import { useI18n } from 'vue-i18n'; import Select from '@/components/common/Select.vue'; import SearchInput from '@/components/common/SearchInput.vue'
import type { AdminGroup } from '@/types'
import { accountPlatformOptions } from '@/utils/accountPlatforms'
import { adminAPI } from '@/api/admin'
import type { AccountSubscriptionTierOption } from '@/api/admin/accounts'
const props = defineProps<{ searchQuery: string; filters: Record<string, any>; groups?: AdminGroup[] }>()
const emit = defineEmits(['update:searchQuery', 'update:filters', 'change']); const { t } = useI18n()
const tiers = ref<AccountSubscriptionTierOption[]>([])
const loadingTiers = ref(false)
const updatePlatform = (value: string | number | boolean | null) => { emit('update:filters', { ...props.filters, platform: value, subscription_tier: '' }) }
const updateSubscriptionTier = (value: string | number | boolean | null) => { emit('update:filters', { ...props.filters, subscription_tier: value }) }
const updateType = (value: string | number | boolean | null) => { emit('update:filters', { ...props.filters, type: value }) }
const updateExpiryStatus = (value: string | number | boolean | null) => { emit('update:filters', { ...props.filters, expiry_status: value }) }
const updateStatus = (value: string | number | boolean | null) => { emit('update:filters', { ...props.filters, status: value }) }
const updatePrivacyMode = (value: string | number | boolean | null) => { emit('update:filters', { ...props.filters, privacy_mode: value }) }
const updateGroup = (value: string | number | boolean | null) => { emit('update:filters', { ...props.filters, group: value }) }
const pOpts = computed(() => [
  { value: '', label: t('admin.accounts.allPlatforms') },
  ...accountPlatformOptions(t)
])
const tierOpts = computed(() => [
  { value: '', label: t('admin.accounts.allSubscriptionTiers') },
  ...tiers.value.map(tier => ({
    value: tier.value,
    label: `${tier.value === '__unrecognized__' ? t('admin.accounts.subscriptionTierUnrecognized') : tier.label} (${tier.account_count})`
  }))
])
const tOpts = computed(() => [{ value: '', label: t('admin.accounts.allTypes') }, { value: 'oauth', label: t('admin.accounts.oauthType') }, { value: 'setup-token', label: t('admin.accounts.setupToken') }, { value: 'apikey', label: t('admin.accounts.apiKey') }, { value: 'bedrock', label: 'AWS Bedrock' }])
const expiryOpts = computed(() => [
  { value: '', label: t('admin.accounts.allExpiryStatuses') },
  { value: 'expiring', label: t('admin.accounts.expiringWithin7Days') },
  { value: 'expired', label: t('admin.accounts.expiredAccounts') }
])
const sOpts = computed(() => [{ value: '', label: t('admin.accounts.allStatus') }, { value: 'active', label: t('admin.accounts.status.active') }, { value: 'inactive', label: t('admin.accounts.status.inactive') }, { value: 'error', label: t('admin.accounts.status.error') }, { value: 'rate_limited', label: t('admin.accounts.status.rateLimited') }, { value: 'temp_unschedulable', label: t('admin.accounts.status.tempUnschedulable') }, { value: 'unschedulable', label: t('admin.accounts.status.unschedulable') }])
const privacyOpts = computed(() => [
  { value: '', label: t('admin.accounts.allPrivacyModes') },
  { value: '__unset__', label: t('admin.accounts.privacyUnset') },
  { value: 'training_off', label: 'Privacy' },
  { value: 'training_set_cf_blocked', label: 'CF' },
  { value: 'training_set_failed', label: 'Fail' }
])
const gOpts = computed(() => [
  { value: '', label: t('admin.accounts.allGroups') },
  { value: 'ungrouped', label: t('admin.accounts.ungroupedGroup') },
  ...(props.groups || []).map(g => ({ value: String(g.id), label: g.name }))
])

watch(() => props.filters.platform, async (platform) => {
  loadingTiers.value = true
  try {
    tiers.value = await adminAPI.accounts.listSubscriptionTiers(typeof platform === 'string' ? platform : '')
  } catch (error) {
    tiers.value = []
    console.error('Failed to load account subscription tiers:', error)
  } finally {
    loadingTiers.value = false
  }
}, { immediate: true })
</script>

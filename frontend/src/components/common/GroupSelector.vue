<template>
  <div>
    <label class="input-label">
      {{ t('admin.users.groups') }}
      <span class="components-common-group-selector__text">{{ t('common.selectedCount', { count: modelValue.length }) }}</span>
    </label>
    <div
      v-if="isSearchable"
      class="components-common-group-selector__panel"
    >
      <Icon name="search" size="sm" class="components-common-group-selector__icon" />
      <input
        v-model="searchText"
        type="text"
        :placeholder="t('common.searchPlaceholder')"
        class="components-common-group-selector__field"
      />
    </div>
    <div
      :class="[
        'components-common-group-selector__panel-3',
        isSearchable
          ? 'components-common-group-selector__panel-4'
          : 'components-common-group-selector__panel-5'
      ]"
    >
      <label
        v-for="group in filteredGroups"
        :key="group.id"
        class="components-common-group-selector__label"
        :title="t('admin.groups.rateAndAccounts', { rate: group.rate_multiplier, count: group.account_count || 0 })"
      >
        <input
          type="checkbox"
          :value="group.id"
          :checked="modelValue.includes(group.id)"
          @change="handleChange(group.id, ($event.target as HTMLInputElement).checked)"
          class="components-common-group-selector__field-2"
        />
        <GroupBadge
          :name="group.name"
          :platform="group.platform"
          :subscription-type="group.subscription_type"
          :rate-multiplier="group.rate_multiplier"
          class="components-common-group-selector__group-badge"
        />
        <span class="components-common-group-selector__text-2">{{ group.account_count || 0 }}</span>
      </label>
      <div
        v-if="filteredGroups.length === 0"
        class="components-common-group-selector__panel-2"
      >
        {{ t('common.noGroupsAvailable') }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import GroupBadge from './GroupBadge.vue'
import Icon from '@/components/icons/Icon.vue'
import type { AdminGroup, GroupPlatform } from '@/types'

const { t } = useI18n()

interface Props {
  modelValue: string[]
  groups: AdminGroup[]
  platform?: GroupPlatform // Optional platform filter
  searchable?: boolean | 'auto'
}

const props = withDefaults(defineProps<Props>(), {
  searchable: 'auto'
})
const emit = defineEmits<{
  'update:modelValue': [value: string[]]
}>()

const searchText = ref('')

const isSearchable = computed(() => {
  if (props.searchable === 'auto') return props.groups.length > 5
  return props.searchable
})

// Filter groups by platform if specified
const filteredGroups = computed(() => {
  let result: AdminGroup[] = props.groups
  if (props.platform) {
    // 只能选择同 platform 的分组；composite 分组可接收任意具体平台账号
    result = result.filter((g) => g.platform === props.platform || g.platform === 'composite')
  }
  if (isSearchable.value && searchText.value) {
    const q = searchText.value.toLowerCase()
    result = result.filter(
      (g) => g.name.toLowerCase().includes(q) || g.description?.toLowerCase().includes(q)
    )
  }
  return result
})

const handleChange = (groupId: string, checked: boolean) => {
  const newValue = checked
    ? [...props.modelValue, groupId]
    : props.modelValue.filter((id) => id !== groupId)
  emit('update:modelValue', newValue)
}
</script>

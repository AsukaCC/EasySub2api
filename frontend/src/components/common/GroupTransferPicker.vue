<template>
  <div class="components-common-group-transfer-picker__panel">
    <div class="components-common-group-transfer-picker__panel-2">
      <Icon name="search" size="sm" class="components-common-group-transfer-picker__icon" />
      <input
        v-model="query"
        type="search"
        :placeholder="searchPlaceholder"
        class="components-common-group-transfer-picker__field"
      />
    </div>

    <div class="components-common-group-transfer-picker__panel-3">
      <section class="components-common-group-transfer-picker__section" role="table" :aria-label="availableLabel">
        <div class="components-common-group-transfer-picker__panel-4">
          <span class="components-common-group-transfer-picker__text">{{ availableLabel }}</span>
          <span class="components-common-group-transfer-picker__text-2">{{ filteredAvailableGroups.length }}</span>
        </div>
        <div class="components-common-group-transfer-picker__panel-5">
          <button
            v-for="group in filteredAvailableGroups"
            :key="group.id"
            type="button"
            class="components-common-group-transfer-picker__action"
            @click="moveToSelected(group.id)"
          >
            <GroupBadge
              :name="group.name"
              :platform="group.platform"
              :subscription-type="group.subscription_type"
              :rate-multiplier="group.rate_multiplier"
              :peak-rate-enabled="group.peak_rate_enabled"
              :peak-start="group.peak_start"
              :peak-end="group.peak_end"
              :peak-rate-multiplier="group.peak_rate_multiplier"
              class="components-common-group-transfer-picker__group-badge"
            />
            <Icon name="chevronRight" size="sm" class="components-common-group-transfer-picker__icon-2" />
          </button>
          <p v-if="filteredAvailableGroups.length === 0" class="components-common-group-transfer-picker__description">
            {{ emptyLabel }}
          </p>
        </div>
      </section>

      <section class="components-common-group-transfer-picker__section-2" role="table" :aria-label="selectedLabel">
        <div class="components-common-group-transfer-picker__panel-6">
          <span class="components-common-group-transfer-picker__text-3">{{ selectedLabel }}</span>
          <span class="components-common-group-transfer-picker__text-4">{{ selectedGroups.length }}</span>
        </div>
        <div class="components-common-group-transfer-picker__panel-5">
          <button
            v-for="group in selectedGroups"
            :key="group.id"
            type="button"
            class="components-common-group-transfer-picker__action-2"
            @click="moveToAvailable(group.id)"
          >
            <Icon name="chevronLeft" size="sm" class="components-common-group-transfer-picker__icon-3" />
            <GroupBadge
              :name="group.name"
              :platform="group.platform"
              :subscription-type="group.subscription_type"
              :rate-multiplier="group.rate_multiplier"
              :peak-rate-enabled="group.peak_rate_enabled"
              :peak-start="group.peak_start"
              :peak-end="group.peak_end"
              :peak-rate-multiplier="group.peak_rate_multiplier"
              class="components-common-group-transfer-picker__group-badge"
            />
          </button>
          <p v-if="selectedGroups.length === 0" class="components-common-group-transfer-picker__description">
            {{ emptyLabel }}
          </p>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { GroupPlatform, SubscriptionType } from '@/types'
import GroupBadge from './GroupBadge.vue'
import Icon from '@/components/icons/Icon.vue'

interface GroupTransferGroup {
  id: string
  name: string
  description: string | null
  platform: GroupPlatform
  subscription_type: SubscriptionType
  rate_multiplier: number
  peak_rate_enabled: boolean
  peak_start: string
  peak_end: string
  peak_rate_multiplier: number
}

const props = defineProps<{
  groups: GroupTransferGroup[]
  modelValue: string[]
  availableLabel: string
  selectedLabel: string
  searchPlaceholder: string
  emptyLabel: string
}>()

const emit = defineEmits<{
  (event: 'update:modelValue', value: string[]): void
}>()

const query = ref('')

const selectedSet = computed(() => new Set(props.modelValue))
const selectedGroups = computed(() => props.groups.filter((group) => selectedSet.value.has(group.id)))
const filteredAvailableGroups = computed(() => {
  const normalized = query.value.trim().toLowerCase()
  return props.groups.filter((group) => {
    if (selectedSet.value.has(group.id)) return false
    if (!normalized) return true
    return group.name.toLowerCase().includes(normalized) || (group.description || '').toLowerCase().includes(normalized)
  })
})

const moveToSelected = (groupID: string) => {
  if (selectedSet.value.has(groupID)) return
  emit('update:modelValue', [...props.modelValue, groupID])
}

const moveToAvailable = (groupID: string) => {
  emit('update:modelValue', props.modelValue.filter((id) => id !== groupID))
}
</script>

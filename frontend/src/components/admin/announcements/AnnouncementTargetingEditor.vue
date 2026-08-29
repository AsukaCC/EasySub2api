<template>
  <div class="components-admin-announcements-announcement-targeting-editor__panel">
    <div class="components-admin-announcements-announcement-targeting-editor__panel-2">
      <div>
        <div class="components-admin-announcements-announcement-targeting-editor__panel-3">
          {{ t('admin.announcements.form.targetingMode') }}
        </div>
        <div class="components-admin-announcements-announcement-targeting-editor__panel-4">
          {{ mode === 'all' ? t('admin.announcements.form.targetingAll') : t('admin.announcements.form.targetingCustom') }}
        </div>
      </div>

      <div class="components-admin-announcements-announcement-targeting-editor__panel-5">
        <label class="components-admin-announcements-announcement-targeting-editor__label">
          <input
            type="radio"
            name="announcement-targeting-mode"
            value="all"
            :checked="mode === 'all'"
            @change="setMode('all')"
            class="components-admin-announcements-announcement-targeting-editor__field"
          />
          {{ t('admin.announcements.form.targetingAll') }}
        </label>
        <label class="components-admin-announcements-announcement-targeting-editor__label">
          <input
            type="radio"
            name="announcement-targeting-mode"
            value="custom"
            :checked="mode === 'custom'"
            @change="setMode('custom')"
            class="components-admin-announcements-announcement-targeting-editor__field"
          />
          {{ t('admin.announcements.form.targetingCustom') }}
        </label>
      </div>
    </div>

    <div v-if="mode === 'custom'" class="components-admin-announcements-announcement-targeting-editor__panel-6">
      <div class="components-admin-announcements-announcement-targeting-editor__panel-7">
        <div class="components-admin-announcements-announcement-targeting-editor__panel-3">
          OR
          <span class="components-admin-announcements-announcement-targeting-editor__text">
            ({{ anyOf.length }}/50)
          </span>
        </div>
        <button
          type="button"
          class="btn btn-secondary"
          :disabled="anyOf.length >= 50"
          @click="addOrGroup"
        >
          <Icon name="plus" size="sm" class="components-admin-announcements-announcement-targeting-editor__icon" />
          {{ t('admin.announcements.form.addOrGroup') }}
        </button>
      </div>

      <div v-if="anyOf.length === 0" class="components-admin-announcements-announcement-targeting-editor__panel-8">
        {{ t('admin.announcements.form.targetingCustom') }}: {{ t('admin.announcements.form.addOrGroup') }}
      </div>

      <div
        v-for="(group, groupIndex) in anyOf"
        :key="groupIndex"
        class="components-admin-announcements-announcement-targeting-editor__panel-9"
      >
        <div class="components-admin-announcements-announcement-targeting-editor__panel-10">
          <div class="components-admin-announcements-announcement-targeting-editor__panel-11">
            <div class="components-admin-announcements-announcement-targeting-editor__panel-3">
              {{ t('admin.announcements.form.targetingCustom') }} #{{ groupIndex + 1 }}
              <span class="components-admin-announcements-announcement-targeting-editor__text-2">AND ({{ (group.all_of?.length || 0) }}/50)</span>
            </div>
            <div class="components-admin-announcements-announcement-targeting-editor__panel-4">
              {{ t('admin.announcements.form.addAndCondition') }}
            </div>
          </div>

          <button
            type="button"
            class="btn btn-secondary"
            @click="removeOrGroup(groupIndex)"
          >
            <Icon name="trash" size="sm" class="components-admin-announcements-announcement-targeting-editor__icon" />
            {{ t('common.delete') }}
          </button>
        </div>

        <div class="components-admin-announcements-announcement-targeting-editor__panel-12">
          <div
            v-for="(cond, condIndex) in (group.all_of || [])"
            :key="condIndex"
            class="components-admin-announcements-announcement-targeting-editor__panel-13"
          >
            <div class="components-admin-announcements-announcement-targeting-editor__panel-14">
              <div class="components-admin-announcements-announcement-targeting-editor__panel-15">
                <label class="input-label">{{ t('admin.announcements.form.conditionType') }}</label>
                <Select
                  :model-value="cond.type"
                  :options="conditionTypeOptions"
                  @update:model-value="(v) => setConditionType(groupIndex, condIndex, v as any)"
                />
              </div>

              <div v-if="cond.type === 'subscription'" class="components-admin-announcements-announcement-targeting-editor__panel-16">
                <label class="input-label">{{ t('admin.announcements.form.selectPackages') }}</label>
                <GroupSelector
                  v-model="subscriptionSelections[groupIndex][condIndex]"
                  :groups="groups"
                />
              </div>

              <div v-else-if="cond.type === 'user'" class="components-admin-announcements-announcement-targeting-editor__panel-16">
                <label class="input-label">{{ t('admin.announcements.form.selectUsers') }}</label>
                <OpenAIFastPolicyUserSelector
                  :model-value="cond.user_ids ?? []"
                  @update:model-value="(ids) => setUserIDs(groupIndex, condIndex, ids)"
                />
              </div>

              <div v-else-if="cond.type === 'level'" class="components-admin-announcements-announcement-targeting-editor__panel-16">
                <label class="input-label">{{ t('admin.announcements.form.selectLevels') }}</label>
                <div class="components-admin-announcements-announcement-targeting-editor__levels">
                  <label
                    v-for="level in userLevelOptions"
                    :key="level.value"
                    class="components-admin-announcements-announcement-targeting-editor__level"
                  >
                    <input
                      type="checkbox"
                      :checked="cond.levels?.includes(level.value)"
                      @change="setLevels(groupIndex, condIndex, level.value, ($event.target as HTMLInputElement).checked)"
                    />
                    {{ level.label }}
                  </label>
                </div>
              </div>

              <div v-else class="components-admin-announcements-announcement-targeting-editor__panel-17">
                <div class="components-admin-announcements-announcement-targeting-editor__panel-18">
                  <label class="input-label">{{ t('admin.announcements.form.operator') }}</label>
                  <Select
                    :model-value="cond.operator"
                    :options="balanceOperatorOptions"
                    @update:model-value="(v) => setOperator(groupIndex, condIndex, v as any)"
                  />
                </div>
                <div class="components-admin-announcements-announcement-targeting-editor__panel-19">
                  <label class="input-label">{{ t('admin.announcements.form.balanceValue') }}</label>
                  <input
                    :value="String(cond.value ?? '')"
                    type="number"
                    step="any"
                    class="input"
                    @input="(e) => setBalanceValue(groupIndex, condIndex, (e.target as HTMLInputElement).value)"
                  />
                </div>
              </div>

              <div class="components-admin-announcements-announcement-targeting-editor__panel-20">
                <button
                  type="button"
                  class="btn btn-secondary"
                  @click="removeAndCondition(groupIndex, condIndex)"
                >
                  <Icon name="trash" size="sm" class="components-admin-announcements-announcement-targeting-editor__icon" />
                  {{ t('common.delete') }}
                </button>
              </div>
            </div>
          </div>

          <div class="components-admin-announcements-announcement-targeting-editor__panel-20">
            <button
              type="button"
              class="btn btn-secondary"
              :disabled="(group.all_of?.length || 0) >= 50"
              @click="addAndCondition(groupIndex)"
            >
              <Icon name="plus" size="sm" class="components-admin-announcements-announcement-targeting-editor__icon" />
              {{ t('admin.announcements.form.addAndCondition') }}
            </button>
          </div>
        </div>
      </div>

      <div v-if="validationError" class="components-admin-announcements-announcement-targeting-editor__panel-21">
        {{ validationError }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type {
  AdminGroup,
  AnnouncementTargeting,
  AnnouncementCondition,
  AnnouncementConditionGroup,
  AnnouncementConditionType,
  AnnouncementOperator
} from '@/types'

import Select from '@/components/common/Select.vue'
import GroupSelector from '@/components/common/GroupSelector.vue'
import Icon from '@/components/icons/Icon.vue'
import OpenAIFastPolicyUserSelector from '@/views/admin/settings/OpenAIFastPolicyUserSelector.vue'

const { t } = useI18n()

const props = defineProps<{
  modelValue: AnnouncementTargeting
  groups: AdminGroup[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: AnnouncementTargeting): void
}>()

const anyOf = computed(() => props.modelValue?.any_of ?? [])

type Mode = 'all' | 'custom'
const mode = computed<Mode>(() => (anyOf.value.length === 0 ? 'all' : 'custom'))

const conditionTypeOptions = computed(() => [
  { value: 'subscription', label: t('admin.announcements.form.conditionSubscription') },
  { value: 'balance', label: t('admin.announcements.form.conditionBalance') },
  { value: 'user', label: t('admin.announcements.form.conditionUser') },
  { value: 'level', label: t('admin.announcements.form.conditionLevel') }
])

const userLevelOptions = computed(() => [
  { value: 1, label: t('admin.announcements.form.level1') },
  { value: 2, label: t('admin.announcements.form.level2') },
  { value: 3, label: t('admin.announcements.form.level3') },
])

const balanceOperatorOptions = computed(() => [
  { value: 'gt', label: t('admin.announcements.operators.gt') },
  { value: 'gte', label: t('admin.announcements.operators.gte') },
  { value: 'lt', label: t('admin.announcements.operators.lt') },
  { value: 'lte', label: t('admin.announcements.operators.lte') },
  { value: 'eq', label: t('admin.announcements.operators.eq') }
])

function setMode(next: Mode) {
  if (next === 'all') {
    emit('update:modelValue', { any_of: [] })
    return
  }
  if (anyOf.value.length === 0) {
    emit('update:modelValue', { any_of: [{ all_of: [defaultSubscriptionCondition()] }] })
  }
}

function defaultSubscriptionCondition(): AnnouncementCondition {
  return {
    type: 'subscription' as AnnouncementConditionType,
    operator: 'in' as AnnouncementOperator,
    group_ids: []
  }
}

function defaultBalanceCondition(): AnnouncementCondition {
  return {
    type: 'balance' as AnnouncementConditionType,
    operator: 'gte' as AnnouncementOperator,
    value: 0
  }
}

function defaultUserCondition(): AnnouncementCondition {
  return {
    type: 'user' as AnnouncementConditionType,
    operator: 'in' as AnnouncementOperator,
    user_ids: []
  }
}

function defaultLevelCondition(): AnnouncementCondition {
  return {
    type: 'level' as AnnouncementConditionType,
    operator: 'in' as AnnouncementOperator,
    levels: []
  }
}

type TargetingDraft = {
  any_of: AnnouncementConditionGroup[]
}

function updateTargeting(mutator: (draft: TargetingDraft) => void) {
  const draft: TargetingDraft = JSON.parse(JSON.stringify(props.modelValue ?? { any_of: [] }))
  if (!draft.any_of) draft.any_of = []
  mutator(draft)
  emit('update:modelValue', draft)
}

function addOrGroup() {
  updateTargeting((draft) => {
    if (draft.any_of.length >= 50) return
    draft.any_of.push({ all_of: [defaultSubscriptionCondition()] })
  })
}

function removeOrGroup(groupIndex: number) {
  updateTargeting((draft) => {
    draft.any_of.splice(groupIndex, 1)
  })
}

function addAndCondition(groupIndex: number) {
  updateTargeting((draft) => {
    const group = draft.any_of[groupIndex]
    if (!group.all_of) group.all_of = []
    if (group.all_of.length >= 50) return
    group.all_of.push(defaultSubscriptionCondition())
  })
}

function removeAndCondition(groupIndex: number, condIndex: number) {
  updateTargeting((draft) => {
    const group = draft.any_of[groupIndex]
    if (!group?.all_of) return
    group.all_of.splice(condIndex, 1)
  })
}

function setConditionType(groupIndex: number, condIndex: number, nextType: AnnouncementConditionType) {
  updateTargeting((draft) => {
    const group = draft.any_of[groupIndex]
    if (!group?.all_of) return

    if (nextType === 'subscription') {
      group.all_of[condIndex] = defaultSubscriptionCondition()
    } else if (nextType === 'balance') {
      group.all_of[condIndex] = defaultBalanceCondition()
    } else if (nextType === 'user') {
      group.all_of[condIndex] = defaultUserCondition()
    } else {
      group.all_of[condIndex] = defaultLevelCondition()
    }
  })
}

function setUserIDs(groupIndex: number, condIndex: number, userIDs: string[]) {
  updateTargeting((draft) => {
    const condition = draft.any_of[groupIndex]?.all_of?.[condIndex]
    if (!condition) return
    condition.user_ids = [...new Set(userIDs.filter(Boolean))]
    condition.operator = 'in'
  })
}

function setLevels(groupIndex: number, condIndex: number, level: number, checked: boolean) {
  updateTargeting((draft) => {
    const condition = draft.any_of[groupIndex]?.all_of?.[condIndex]
    if (!condition) return
    const levels = new Set(condition.levels ?? [])
    if (checked) levels.add(level)
    else levels.delete(level)
    condition.levels = Array.from(levels).sort((a, b) => a - b)
    condition.operator = 'in'
  })
}

function setOperator(groupIndex: number, condIndex: number, op: AnnouncementOperator) {
  updateTargeting((draft) => {
    const group = draft.any_of[groupIndex]
    if (!group?.all_of) return

    const cond = group.all_of[condIndex]
    if (!cond) return

    cond.operator = op
  })
}

function setBalanceValue(groupIndex: number, condIndex: number, raw: string) {
  const n = raw === '' ? 0 : Number(raw)
  updateTargeting((draft) => {
    const group = draft.any_of[groupIndex]
    if (!group?.all_of) return

    const cond = group.all_of[condIndex]
    if (!cond) return

    cond.value = Number.isFinite(n) ? n : 0
  })
}

// Keep group_ids selection in a parallel reactive map for each condition editor.
// Then we mirror it back to targeting.group_ids via a watcher.
const subscriptionSelections = reactive<Record<number, Record<number, string[]>>>({})

function ensureSelectionPath(groupIndex: number, condIndex: number) {
  if (!subscriptionSelections[groupIndex]) subscriptionSelections[groupIndex] = {}
  if (!subscriptionSelections[groupIndex][condIndex]) subscriptionSelections[groupIndex][condIndex] = []
}

// Sync from modelValue to subscriptionSelections (one-way: model -> local state)
watch(
  () => props.modelValue,
  (v) => {
    const groups = v?.any_of ?? []
    for (let gi = 0; gi < groups.length; gi++) {
      const allOf = groups[gi]?.all_of ?? []
      for (let ci = 0; ci < allOf.length; ci++) {
        const c = allOf[ci]
        if (c?.type === 'subscription') {
          ensureSelectionPath(gi, ci)
          // Only update if different to avoid triggering unnecessary updates
          const newIds = (c.group_ids ?? []).slice()
          const currentIds = subscriptionSelections[gi]?.[ci] ?? []
          if (JSON.stringify(newIds.sort()) !== JSON.stringify(currentIds.sort())) {
            subscriptionSelections[gi][ci] = newIds
          }
        }
      }
    }
  },
  { immediate: true }
)

// Sync from subscriptionSelections to modelValue (one-way: local state -> model)
// Use a debounced approach to avoid infinite loops
let syncTimeout: ReturnType<typeof setTimeout> | null = null
watch(
  () => subscriptionSelections,
  () => {
    // Debounce the sync to avoid rapid fire updates
    if (syncTimeout) clearTimeout(syncTimeout)

    syncTimeout = setTimeout(() => {
      // Build the new targeting state
      const newTargeting: TargetingDraft = JSON.parse(JSON.stringify(props.modelValue ?? { any_of: [] }))
      if (!newTargeting.any_of) newTargeting.any_of = []

      const groups = newTargeting.any_of ?? []
      for (let gi = 0; gi < groups.length; gi++) {
        const allOf = groups[gi]?.all_of ?? []
        for (let ci = 0; ci < allOf.length; ci++) {
          const c = allOf[ci]
          if (c?.type === 'subscription') {
            ensureSelectionPath(gi, ci)
            c.operator = 'in' as AnnouncementOperator
            c.group_ids = (subscriptionSelections[gi]?.[ci] ?? []).slice()
          }
        }
      }

      // Only emit if there's an actual change (deep comparison)
      if (JSON.stringify(props.modelValue) !== JSON.stringify(newTargeting)) {
        emit('update:modelValue', newTargeting)
      }
    }, 0)
  },
  { deep: true }
)

const validationError = computed(() => {
  if (mode.value !== 'custom') return ''

  const groups = anyOf.value
  if (groups.length === 0) return t('admin.announcements.form.addOrGroup')

  if (groups.length > 50) return 'any_of > 50'

  for (const g of groups) {
    const allOf = g?.all_of ?? []
    if (allOf.length === 0) return t('admin.announcements.form.addAndCondition')
    if (allOf.length > 50) return 'all_of > 50'

    for (const c of allOf) {
      if (c.type === 'subscription') {
        if (!c.group_ids || c.group_ids.length === 0) return t('admin.announcements.form.selectPackages')
      }
      if (c.type === 'user') {
        if (!c.user_ids || c.user_ids.length === 0) return t('admin.announcements.form.selectUsers')
      }
      if (c.type === 'level') {
        if (!c.levels || c.levels.length === 0) return t('admin.announcements.form.selectLevels')
      }
    }
  }

  return ''
})
</script>

<style scoped>
.components-admin-announcements-announcement-targeting-editor__levels {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.components-admin-announcements-announcement-targeting-editor__level {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.45rem 0.7rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
  cursor: pointer;
}
</style>

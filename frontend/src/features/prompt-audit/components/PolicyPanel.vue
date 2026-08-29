<template>
  <section aria-labelledby="prompt-policy-title" class="features-prompt-audit-components-policy-panel__section">
    <div>
      <h2 id="prompt-policy-title" class="features-prompt-audit-components-policy-panel__heading">{{ t('admin.promptAudit.policy.title') }}</h2>
      <p class="features-prompt-audit-components-policy-panel__description">{{ t('admin.promptAudit.policy.description') }}</p>
    </div>

    <div class="features-prompt-audit-components-policy-panel__panel">
      <div class="features-prompt-audit-components-policy-panel__panel-2">
        <fieldset>
          <legend class="features-prompt-audit-components-policy-panel__legend">{{ t('admin.promptAudit.policy.scope') }}</legend>
          <div class="features-prompt-audit-components-policy-panel__panel-3">
            <label class="features-prompt-audit-components-policy-panel__label">
              <input type="radio" name="prompt-audit-scope" :checked="draft.all_groups" @change="patch({ all_groups: true, group_ids: [] })" />
              {{ t('admin.promptAudit.policy.allGroups') }}
            </label>
            <label class="features-prompt-audit-components-policy-panel__label">
              <input type="radio" name="prompt-audit-scope" :checked="!draft.all_groups" @change="patch({ all_groups: false })" />
              {{ t('admin.promptAudit.policy.selectedGroups') }}
            </label>
          </div>
        </fieldset>

        <div v-if="!draft.all_groups" class="features-prompt-audit-components-policy-panel__panel-4">
          <label class="features-prompt-audit-components-policy-panel__label-2">
            <span>{{ t('admin.promptAudit.policy.searchGroups') }}</span>
            <input v-model="groupSearch" type="search" class="features-prompt-audit-components-policy-panel__field input" :aria-label="t('admin.promptAudit.policy.searchGroups')" />
          </label>
          <div class="features-prompt-audit-components-policy-panel__panel-5">
            <label v-for="group in filteredGroups" :key="group.id" class="features-prompt-audit-components-policy-panel__label-3">
              <span class="features-prompt-audit-components-policy-panel__text">
                <input type="checkbox" :checked="draft.group_ids.includes(group.id)" @change="toggleGroup(group.id)" />
                {{ group.name }}
              </span>
              <span class="features-prompt-audit-components-policy-panel__text-2">{{ group.platform }} · {{ group.status }}</span>
            </label>
            <p v-if="filteredGroups.length === 0" class="features-prompt-audit-components-policy-panel__description-2">{{ t('admin.promptAudit.policy.noGroups') }}</p>
          </div>
          <div v-if="missingGroupIds.length" class="features-prompt-audit-components-policy-panel__panel-6">
            {{ t('admin.promptAudit.policy.missingGroups') }}: {{ missingGroupIds.join(', ') }}
          </div>
          <p class="features-prompt-audit-components-policy-panel__description-3">{{ t('admin.promptAudit.policy.selectedCount', { count: draft.group_ids.length }) }}</p>
        </div>

        <fieldset class="features-prompt-audit-components-policy-panel__fieldset">
          <legend class="features-prompt-audit-components-policy-panel__legend">{{ t('admin.promptAudit.policy.scanners') }}</legend>
          <div class="features-prompt-audit-components-policy-panel__panel-7">
            <label v-for="scanner in SCANNER_CATALOG" :key="scanner.id" class="features-prompt-audit-components-policy-panel__label-4">
              <input type="checkbox" :checked="draft.scanners.includes(scanner.id)" :aria-label="scannerLabel(scanner.id)" @change="toggleScanner(scanner.id)" />
              <span>{{ scannerLabel(scanner.id) }}</span>
            </label>
          </div>
        </fieldset>
      </div>

      <div class="features-prompt-audit-components-policy-panel__panel-8">
        <label class="features-prompt-audit-components-policy-panel__label-2">
          <span>{{ t('admin.promptAudit.policy.workerCount') }}</span>
          <input :value="draft.worker_count" type="number" min="1" max="32" class="features-prompt-audit-components-policy-panel__field input" :aria-label="t('admin.promptAudit.policy.workerCount')" @input="patch({ worker_count: Number(($event.target as HTMLInputElement).value) })" />
        </label>
        <label class="features-prompt-audit-components-policy-panel__label-2">
          <span>{{ t('admin.promptAudit.policy.queueCapacity') }}</span>
          <input :value="draft.queue_capacity" type="number" min="1" max="100000" class="features-prompt-audit-components-policy-panel__field input" :aria-label="t('admin.promptAudit.policy.queueCapacity')" @input="patch({ queue_capacity: Number(($event.target as HTMLInputElement).value) })" />
        </label>
        <div class="features-prompt-audit-components-policy-panel__panel-9">
          <p class="features-prompt-audit-components-policy-panel__description-4">{{ t('admin.promptAudit.policy.strategy') }}</p>
          <p class="features-prompt-audit-components-policy-panel__description-5">priority · {{ t('admin.promptAudit.policy.strategyHint') }}</p>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { PromptAuditDraft, PromptAuditGroup } from '../types'
import { cloneData, SCANNER_CATALOG } from '../viewModel'

const props = defineProps<{ draft: PromptAuditDraft; groups: PromptAuditGroup[] }>()
const emit = defineEmits<{ (event: 'update:draft', value: PromptAuditDraft): void }>()
const { t } = useI18n()
const groupSearch = ref('')

const filteredGroups = computed(() => {
  const query = groupSearch.value.trim().toLowerCase()
  if (!query) return props.groups
  return props.groups.filter((group) => `${group.name} ${group.id} ${group.platform}`.toLowerCase().includes(query))
})
const knownGroupIds = computed(() => new Set(props.groups.map((group) => group.id)))
const missingGroupIds = computed(() => props.draft.group_ids.filter((id) => !knownGroupIds.value.has(id)))

function patch(value: Partial<PromptAuditDraft>) {
  emit('update:draft', { ...cloneData(props.draft), ...value })
}
function toggleGroup(id: string) {
  const selected = new Set(props.draft.group_ids)
  if (selected.has(id)) selected.delete(id)
  else selected.add(id)
  patch({ group_ids: [...selected].sort() })
}
function toggleScanner(id: string) {
  const selected = new Set(props.draft.scanners)
  if (selected.has(id)) selected.delete(id)
  else selected.add(id)
  patch({ scanners: SCANNER_CATALOG.map((item) => item.id).filter((item) => selected.has(item)) })
}
function scannerLabel(id: string): string {
  return t(`admin.promptAudit.scanners.${id}`)
}
</script>

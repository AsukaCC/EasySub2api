<template>
  <section class="reasoning-policy" :aria-label="t('admin.groups.form.reasoningEffortMappings')">
    <div class="reasoning-policy__field">
      <label :for="`${idPrefix}-max-effort`" class="input-label">
        {{ t("admin.groups.form.maxReasoningEffort") }}
      </label>
      <Select
        :id="`${idPrefix}-max-effort`"
        :model-value="maxEffort"
        :options="reasoningEffortOptions"
        :placeholder="t('admin.groups.form.maxReasoningEffortUnlimited')"
        :aria-label="t('admin.groups.form.maxReasoningEffort')"
        :searchable="false"
        clearable
        @update:model-value="updateMaxEffort"
      />
      <p class="input-hint">{{ t("admin.groups.form.maxReasoningEffortHint") }}</p>
    </div>

    <div class="reasoning-policy__field">
      <label :for="`${idPrefix}-over-limit`" class="input-label">
        {{ t("admin.groups.form.maxReasoningEffortOverLimit") }}
      </label>
      <Select
        :id="`${idPrefix}-over-limit`"
        :model-value="overLimit"
        :options="overLimitOptions"
        :aria-label="t('admin.groups.form.maxReasoningEffortOverLimit')"
        :searchable="false"
        :disabled="!maxEffort"
        @update:model-value="updateOverLimit"
      />
      <p class="input-hint">{{ t("admin.groups.form.maxReasoningEffortOverLimitHint") }}</p>
    </div>

    <div class="reasoning-policy__mapping-section">
      <div class="reasoning-policy__header">
        <div>
          <label class="input-label reasoning-policy__label">
            {{ t("admin.groups.form.reasoningEffortMappings") }}
          </label>
          <p class="input-hint reasoning-policy__hint">
            {{ t("admin.groups.form.reasoningEffortMappingsHint") }}
          </p>
        </div>
        <button
          type="button"
          class="btn btn-secondary btn-sm reasoning-policy__add"
          :disabled="mappingCount >= maxMappings"
          @click="addMapping"
        >
          <Icon name="plus" size="sm" />
          {{ t("admin.groups.form.addReasoningEffortMapping") }}
        </button>
      </div>

      <div class="reasoning-policy__counter" :class="{ 'reasoning-policy__counter--limit': mappingCount >= maxMappings }">
        {{ t("admin.groups.form.reasoningEffortMappingCount", { count: mappingCount, max: maxMappings }) }}
      </div>

      <p
        v-if="showValidation && Object.values(validationErrors).some((error) => error.limit)"
        class="reasoning-policy__error"
        role="alert"
      >
        {{ t("admin.groups.form.mappingLimit", { max: maxMappings }) }}
      </p>

      <div v-if="mappings.length > 0" class="reasoning-policy__list">
        <div
          v-for="row in mappings"
          :key="row.id"
          class="reasoning-policy__row"
        >
          <div class="reasoning-policy__scope">
            <div class="reasoning-policy__field">
              <label :for="`${idPrefix}-${row.id}-match-type`" class="input-label">
                {{ t("admin.groups.form.reasoningEffortMatchType") }}
              </label>
              <Select
                :id="`${idPrefix}-${row.id}-match-type`"
                :model-value="row.match_type || ''"
                :options="matchTypeOptions"
                :placeholder="t('admin.groups.form.reasoningEffortMatchTypePlaceholder')"
                :error="showValidation && !!validationErrors[row.id]?.match_type"
                :aria-label="t('admin.groups.form.reasoningEffortMatchType')"
                :searchable="false"
                clearable
                @update:model-value="updateMapping(row.id, 'match_type', $event)"
              />
              <p
                v-if="showValidation && validationErrors[row.id]?.match_type"
                class="reasoning-policy__error"
                role="alert"
              >
                {{ mappingErrorText(validationErrors[row.id]?.match_type) }}
              </p>
            </div>

            <div class="reasoning-policy__field">
              <label :for="`${idPrefix}-${row.id}-model`" class="input-label">
                {{ t("admin.groups.form.reasoningEffortModel") }}
              </label>
              <input
                :id="`${idPrefix}-${row.id}-model`"
                :value="row.model || ''"
                type="text"
                :maxlength="maxModelLength"
                autocomplete="off"
                class="input"
                :class="{ 'reasoning-policy__input--error': showValidation && !!validationErrors[row.id]?.model }"
                :placeholder="t('admin.groups.form.reasoningEffortModelPlaceholder')"
                :aria-label="t('admin.groups.form.reasoningEffortModel')"
                :aria-invalid="showValidation && !!validationErrors[row.id]?.model"
                @input="onModelInput(row.id, $event)"
              />
              <p
                v-if="showValidation && validationErrors[row.id]?.model"
                class="reasoning-policy__error"
                role="alert"
              >
                {{ mappingErrorText(validationErrors[row.id]?.model) }}
              </p>
            </div>
          </div>

          <div class="reasoning-policy__rule">
            <div class="reasoning-policy__field">
              <label :for="`${idPrefix}-${row.id}-from`" class="input-label">
                {{ t("admin.groups.form.reasoningEffortFrom") }}
              </label>
              <Select
                :id="`${idPrefix}-${row.id}-from`"
                :model-value="row.from"
                :options="reasoningEffortOptions"
                :placeholder="t('admin.groups.form.reasoningEffortFromPlaceholder')"
                :error="showValidation && !!validationErrors[row.id]?.from"
                :aria-label="t('admin.groups.form.reasoningEffortFrom')"
                :searchable="false"
                clearable
                @update:model-value="updateMapping(row.id, 'from', $event)"
              />
              <p
                v-if="showValidation && validationErrors[row.id]?.from"
                class="reasoning-policy__error"
                role="alert"
              >
                {{ mappingErrorText(validationErrors[row.id]?.from) }}
              </p>
            </div>

            <div class="reasoning-policy__arrow" aria-hidden="true">
              <Icon name="arrowRight" size="sm" />
            </div>

            <div class="reasoning-policy__field">
              <label :for="`${idPrefix}-${row.id}-to`" class="input-label">
                {{ t("admin.groups.form.reasoningEffortTo") }}
              </label>
              <Select
                :id="`${idPrefix}-${row.id}-to`"
                :model-value="row.to"
                :options="reasoningEffortOptions"
                :placeholder="t('admin.groups.form.reasoningEffortToPlaceholder')"
                :error="showValidation && !!validationErrors[row.id]?.to"
                :aria-label="t('admin.groups.form.reasoningEffortTo')"
                :searchable="false"
                clearable
                @update:model-value="updateMapping(row.id, 'to', $event)"
              />
              <p
                v-if="showValidation && validationErrors[row.id]?.to"
                class="reasoning-policy__error"
                role="alert"
              >
                {{ mappingErrorText(validationErrors[row.id]?.to) }}
              </p>
            </div>

            <button
              type="button"
              class="btn btn-ghost btn-icon reasoning-policy__remove"
              :title="t('admin.groups.form.removeReasoningEffortMapping')"
              :aria-label="t('admin.groups.form.removeReasoningEffortMapping')"
              @click="removeMapping(row.id)"
            >
              <Icon name="trash" size="sm" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useI18n } from "vue-i18n";
import type { GroupPlatform } from "@/types";
import Icon from "@/components/icons/Icon.vue";
import Select from "@/components/common/Select.vue";
import {
  createReasoningEffortMappingRow,
  MAX_REASONING_EFFORT_MAPPINGS,
  MAX_REASONING_EFFORT_MODEL_LENGTH,
  normalizeReasoningEffortMatchType,
  reasoningEffortOptionsForPlatform,
  reasoningEffortOverLimitDeny,
  reasoningEffortOverLimitDowngrade,
  validateReasoningEffortMappings,
  type ReasoningEffortMappingErrorCode,
  type ReasoningEffortMappingRow,
} from "@/views/admin/groupsReasoningEffort";

const props = defineProps<{
  idPrefix: string;
  platform: GroupPlatform;
  maxEffort: string;
  overLimit: string;
  mappings: ReasoningEffortMappingRow[];
}>();

const emit = defineEmits<{
  (event: "update:maxEffort", value: string): void;
  (event: "update:overLimit", value: string): void;
  (event: "update:mappings", value: ReasoningEffortMappingRow[]): void;
}>();

const { t } = useI18n();
const showValidation = ref(false);
const maxMappings = MAX_REASONING_EFFORT_MAPPINGS;
const maxModelLength = MAX_REASONING_EFFORT_MODEL_LENGTH;
const reasoningEffortOptions = computed(() =>
  reasoningEffortOptionsForPlatform(props.platform),
);
const matchTypeOptions = computed(() => [
  { value: "exact", label: t("admin.groups.form.reasoningEffortMatchExact") },
  { value: "prefix", label: t("admin.groups.form.reasoningEffortMatchPrefix") },
  { value: "suffix", label: t("admin.groups.form.reasoningEffortMatchSuffix") },
]);
const overLimitOptions = computed(() => [
  {
    value: reasoningEffortOverLimitDowngrade,
    label: t("admin.groups.form.maxReasoningEffortOverLimitDowngrade"),
  },
  {
    value: reasoningEffortOverLimitDeny,
    label: t("admin.groups.form.maxReasoningEffortOverLimitDeny"),
  },
]);
const validationErrors = computed(() =>
  validateReasoningEffortMappings(props.mappings, props.platform),
);
const mappingCount = computed(() => props.mappings.length);

const asString = (value: string | number | boolean | null): string =>
  value == null ? "" : String(value);

const updateMaxEffort = (value: string | number | boolean | null) => {
  emit("update:maxEffort", asString(value));
};

const updateOverLimit = (value: string | number | boolean | null) => {
  emit(
    "update:overLimit",
    asString(value) || reasoningEffortOverLimitDowngrade,
  );
};

const updateMapping = (
  id: string,
  field: "from" | "to" | "match_type",
  value: string | number | boolean | null,
) => {
  const nextValue =
    field === "match_type"
      ? normalizeReasoningEffortMatchType(asString(value)) || undefined
      : asString(value);
  emit(
    "update:mappings",
    props.mappings.map((row) =>
      row.id === id ? { ...row, [field]: nextValue } : row,
    ),
  );
};

const onModelInput = (id: string, event: Event) => {
  const target = event.target as HTMLInputElement | null;
  const value = (target?.value ?? "").slice(0, maxModelLength);
  emit(
    "update:mappings",
    props.mappings.map((row) =>
      row.id === id ? { ...row, model: value } : row,
    ),
  );
};

const addMapping = () => {
  if (mappingCount.value >= maxMappings) return;
  emit("update:mappings", [
    ...props.mappings,
    createReasoningEffortMappingRow(),
  ]);
};

const removeMapping = (id: string) => {
  emit(
    "update:mappings",
    props.mappings.filter((row) => row.id !== id),
  );
};

const mappingErrorText = (
  code: ReasoningEffortMappingErrorCode | undefined,
): string => (code ? t(`admin.groups.form.${code}`) : "");

const validate = (): boolean => {
  showValidation.value = true;
  return Object.keys(validationErrors.value).length === 0;
};

const resetValidation = () => {
  showValidation.value = false;
};

defineExpose({ validate, resetValidation });
</script>

<style scoped lang="scss">
.reasoning-policy {
  display: grid;
  gap: 1rem;
}

.reasoning-policy__mapping-section {
  display: grid;
  gap: 0.75rem;
  padding-top: 1rem;
  border-top: 1px solid var(--glass-border);
}

.reasoning-policy__header,
.reasoning-policy__scope,
.reasoning-policy__rule {
  display: grid;
  gap: 0.75rem;
}

.reasoning-policy__header {
  align-items: start;
  grid-template-columns: minmax(0, 1fr) auto;
}

.reasoning-policy__scope {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.reasoning-policy__rule {
  align-items: end;
  grid-template-columns: minmax(0, 1fr) 1.5rem minmax(0, 1fr) 2.75rem;
}

.reasoning-policy__label,
.reasoning-policy__hint {
  margin-bottom: 0;
}

.reasoning-policy__counter {
  justify-self: end;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
}

.reasoning-policy__counter--limit,
.reasoning-policy__error {
  color: var(--color-text-danger);
}

.reasoning-policy__list {
  display: grid;
  gap: 0.75rem;
}

.reasoning-policy__row {
  display: grid;
  gap: 0.75rem;
  padding: 0.875rem;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-md);
  background: var(--glass-layer-inset-bg);
}

.reasoning-policy__arrow {
  display: flex;
  min-height: 2.75rem;
  align-items: center;
  justify-content: center;
  color: var(--color-text-tertiary);
}

.reasoning-policy__remove {
  align-self: end;
}

.reasoning-policy__error {
  margin: 0.25rem 0 0;
  font-size: var(--font-size-xs);
}

.reasoning-policy__input--error {
  border-color: var(--color-danger-border);
}

@media (max-width: 48rem) {
  .reasoning-policy__header,
  .reasoning-policy__scope,
  .reasoning-policy__rule {
    grid-template-columns: 1fr;
  }

  .reasoning-policy__arrow {
    display: none;
  }

  .reasoning-policy__remove {
    justify-self: end;
  }
}
</style>

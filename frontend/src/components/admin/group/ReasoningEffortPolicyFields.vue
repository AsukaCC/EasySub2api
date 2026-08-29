<template>
  <div class="components-admin-group-reasoning-effort-policy-fields__panel">
    <div>
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

    <div class="components-admin-group-reasoning-effort-policy-fields__panel-2">
      <div class="components-admin-group-reasoning-effort-policy-fields__panel-3">
        <label class="components-admin-group-reasoning-effort-policy-fields__label input-label">
          {{ t("admin.groups.form.reasoningEffortMappings") }}
        </label>
        <button
          type="button"
          class="components-admin-group-reasoning-effort-policy-fields__action"
          @click="addMapping"
        >
          <Icon name="plus" size="sm" />
          {{ t("admin.groups.form.addReasoningEffortMapping") }}
        </button>
      </div>

      <div v-if="mappings.length > 0" class="components-admin-group-reasoning-effort-policy-fields__panel-4">
        <div
          v-for="row in mappings"
          :key="row.id"
          class="components-admin-group-reasoning-effort-policy-fields__panel-5"
        >
          <div class="components-admin-group-reasoning-effort-policy-fields__panel-6">
            <div>
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
                :aria-describedby="showValidation && validationErrors[row.id]?.from ? `${idPrefix}-${row.id}-from-error` : undefined"
                :searchable="false"
                clearable
                @update:model-value="updateMapping(row.id, 'from', $event)"
              />
              <p
                v-if="showValidation && validationErrors[row.id]?.from"
                :id="`${idPrefix}-${row.id}-from-error`"
                class="components-admin-group-reasoning-effort-policy-fields__description"
                role="alert"
              >
                {{ mappingErrorText(validationErrors[row.id]?.from) }}
              </p>
            </div>

            <div class="components-admin-group-reasoning-effort-policy-fields__panel-7">
              <Icon name="arrowRight" size="sm" />
            </div>

            <div>
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
                :aria-describedby="showValidation && validationErrors[row.id]?.to ? `${idPrefix}-${row.id}-to-error` : undefined"
                :searchable="false"
                clearable
                @update:model-value="updateMapping(row.id, 'to', $event)"
              />
              <p
                v-if="showValidation && validationErrors[row.id]?.to"
                :id="`${idPrefix}-${row.id}-to-error`"
                class="components-admin-group-reasoning-effort-policy-fields__description"
                role="alert"
              >
                {{ mappingErrorText(validationErrors[row.id]?.to) }}
              </p>
            </div>

            <button
              type="button"
              class="components-admin-group-reasoning-effort-policy-fields__action-2"
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
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useI18n } from "vue-i18n";
import type { GroupPlatform } from "@/types";
import Icon from "@/components/icons/Icon.vue";
import Select from "@/components/common/Select.vue";
import {
  createReasoningEffortMappingRow,
  reasoningEffortOptionsForPlatform,
  validateReasoningEffortMappings,
  type ReasoningEffortMappingErrorCode,
  type ReasoningEffortMappingRow,
} from "@/views/admin/groupsReasoningEffort";

const props = defineProps<{
  idPrefix: string;
  platform: GroupPlatform;
  maxEffort: string;
  mappings: ReasoningEffortMappingRow[];
}>();

const emit = defineEmits<{
  (event: "update:maxEffort", value: string): void;
  (event: "update:mappings", value: ReasoningEffortMappingRow[]): void;
}>();

const { t } = useI18n();
const showValidation = ref(false);
const reasoningEffortOptions = computed(() =>
  reasoningEffortOptionsForPlatform(props.platform),
);
const validationErrors = computed(() =>
  validateReasoningEffortMappings(props.mappings, props.platform),
);

const asString = (value: string | number | boolean | null): string =>
  value == null ? "" : String(value);

const updateMaxEffort = (value: string | number | boolean | null) => {
  emit("update:maxEffort", asString(value));
};

const updateMapping = (
  id: string,
  field: "from" | "to",
  value: string | number | boolean | null,
) => {
  emit(
    "update:mappings",
    props.mappings.map((row) =>
      row.id === id ? { ...row, [field]: asString(value) } : row,
    ),
  );
};

const addMapping = () => {
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

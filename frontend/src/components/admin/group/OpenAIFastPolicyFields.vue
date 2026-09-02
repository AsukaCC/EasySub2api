<template>
  <section class="openai-fast-policy" :aria-label="t('admin.groups.openaiFast.title')">
    <h4 class="openai-fast-policy__title">{{ t("admin.groups.openaiFast.title") }}</h4>

    <div class="openai-fast-policy__option">
      <div>
        <label class="openai-fast-policy__label" :for="`${idPrefix}-force`">
          {{ t("admin.groups.openaiFast.force") }}
        </label>
        <p class="input-hint openai-fast-policy__hint">
          {{ t("admin.groups.openaiFast.hint") }}
        </p>
      </div>
      <Toggle
        :id="`${idPrefix}-force`"
        :model-value="force"
        :aria-label="t('admin.groups.openaiFast.force')"
        @update:model-value="emit('update:force', $event)"
      />
    </div>

    <div class="openai-fast-policy__option">
      <div>
        <label class="openai-fast-policy__label" :for="`${idPrefix}-free`">
          {{ t("admin.groups.openaiFast.free") }}
        </label>
        <p class="input-hint openai-fast-policy__hint">
          {{ t("admin.groups.openaiFast.freeHint") }}
        </p>
      </div>
      <Toggle
        :id="`${idPrefix}-free`"
        :model-value="free"
        :aria-label="t('admin.groups.openaiFast.free')"
        @update:model-value="emit('update:free', $event)"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import Toggle from "@/components/common/Toggle.vue";

defineProps<{
  idPrefix: string;
  force: boolean;
  free: boolean;
}>();

const emit = defineEmits<{
  (event: "update:force", value: boolean): void;
  (event: "update:free", value: boolean): void;
}>();

const { t } = useI18n();
</script>

<style scoped lang="scss">
.openai-fast-policy {
  display: grid;
  gap: 0.75rem;
  padding-top: 1rem;
  border-top: 1px solid var(--glass-border);
}

.openai-fast-policy__title {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
}

.openai-fast-policy__option {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.openai-fast-policy__label {
  display: block;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  cursor: pointer;
}

.openai-fast-policy__hint {
  max-width: 48rem;
  margin: 0.25rem 0 0;
}
</style>

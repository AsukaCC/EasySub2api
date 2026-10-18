<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import { availableLocales } from '@/i18n'
import type { GroupI18n } from '@/types'

const props = defineProps<{ modelValue: GroupI18n }>()
const emit = defineEmits<{ 'update:modelValue': [value: GroupI18n] }>()
const { t } = useI18n()
const language = ref('')
const options = computed(() => [
  { value: '', label: t('admin.groups.translations') },
  ...availableLocales.map(locale => ({ value: locale.code, label: locale.name }))
])
function update(field: 'name' | 'description', event: Event) {
  if (!language.value) return
  emit('update:modelValue', {
    ...props.modelValue,
    [language.value]: { ...props.modelValue[language.value], [field]: (event.target as HTMLInputElement).value }
  })
}
</script>

<template>
  <div>
    <label class="input-label">{{ t('admin.groups.translations') }}</label>
    <Select v-model="language" :options="options" />
    <template v-if="language">
      <label class="input-label">{{ t('admin.groups.form.name') }}</label>
      <input class="input" :value="modelValue[language]?.name || ''" maxlength="100" @input="update('name', $event)" />
      <label class="input-label">{{ t('admin.groups.form.description') }}</label>
      <textarea class="input" rows="3" :value="modelValue[language]?.description || ''" @input="update('description', $event)" />
    </template>
  </div>
</template>

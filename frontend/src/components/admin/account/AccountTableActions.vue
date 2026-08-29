<template>
  <div class="account-table-actions">
    <slot name="before"></slot>
    <button @click="$emit('refresh')" :disabled="loading" class="btn btn-secondary">
      <Icon name="refresh" size="md" :class="{ 'account-table-actions__spinner': loading }" />
    </button>
    <slot name="after"></slot>
    <slot name="beforeCreate"></slot>
    <button @click="$emit('create')" class="btn btn-primary">{{ t('admin.accounts.createAccount') }}</button>
    <slot name="afterCreate"></slot>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

defineProps(['loading'])
defineEmits(['refresh', 'create'])

const { t } = useI18n()
</script>

<style scoped>
.account-table-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem;
}

.account-table-actions__spinner {
  animation: account-table-actions-spin 1s linear infinite;
}

@keyframes account-table-actions-spin {
  to { transform: rotate(360deg); }
}
</style>

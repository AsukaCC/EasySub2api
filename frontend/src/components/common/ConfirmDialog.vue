<template>
  <BaseDialog :show="show" :title="title" width="narrow" @close="handleCancel">
    <div class="confirm-dialog__content">
      <p class="confirm-dialog__message">{{ message }}</p>
      <slot></slot>
    </div>

    <template #footer>
      <div class="confirm-dialog__actions">
        <button
          @click="handleCancel"
          type="button"
          class="confirm-dialog__button confirm-dialog__button--cancel"
        >
          {{ cancelText }}
        </button>
        <button
          @click="handleConfirm"
          type="button"
          :class="[
            'confirm-dialog__button',
            danger
              ? 'confirm-dialog__button--danger'
              : 'confirm-dialog__button--confirm'
          ]"
        >
          {{ confirmText }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from './BaseDialog.vue'

const { t } = useI18n()

interface Props {
  show: boolean
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  danger?: boolean
}

interface Emits {
  (e: 'confirm'): void
  (e: 'cancel'): void
}

const props = withDefaults(defineProps<Props>(), {
  danger: false
})

const confirmText = computed(() => props.confirmText || t('common.confirm'))
const cancelText = computed(() => props.cancelText || t('common.cancel'))

const emit = defineEmits<Emits>()

const handleConfirm = () => {
  emit('confirm')
}

const handleCancel = () => {
  emit('cancel')
}
</script>

<style scoped>
.confirm-dialog__content > * + * {
  margin-top: 1rem;
}

.confirm-dialog__message {
  color: var(--color-text-secondary);
  font-size: var(--type-control-size);
  line-height: var(--type-control-line-height);
}

.confirm-dialog__actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

.confirm-dialog__button {
  padding: 0.5rem 1rem;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-md);
  background: var(--glass-layer-inset-bg);
  color: var(--color-text-primary);
  font-size: var(--type-control-size);
  font-weight: var(--font-weight-medium);
  backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate));
  transition: border-color 150ms ease, box-shadow 150ms ease, backdrop-filter 150ms ease;
}

.confirm-dialog__button--confirm {
  border-color: var(--color-primary-border);
  background: var(--glass-tint-brand);
  color: var(--color-text-brand);
}

.confirm-dialog__button--danger {
  border-color: color-mix(in srgb, var(--color-text-danger) 35%, transparent);
  background: var(--glass-tint-danger);
  color: var(--color-text-danger);
}

.confirm-dialog__button:hover {
  border-color: var(--glass-border-hover);
  backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
  box-shadow: var(--glass-shadow-hover), 0 1px 0 var(--glass-highlight) inset;
}

.confirm-dialog__button:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

.confirm-dialog__button--danger:focus-visible {
  outline-color: var(--color-text-danger);
}
</style>

<template>
  <div class="card">
    <!-- Header -->
    <div class="components-payment-payment-provider-list__panel">
      <div class="components-payment-payment-provider-list__panel-2">
        <div>
          <h2 class="components-payment-payment-provider-list__heading">
            {{ t('admin.settings.payment.providerManagement') }}
          </h2>
          <p class="components-payment-payment-provider-list__description">
            {{ t('admin.settings.payment.providerManagementDesc') }}
          </p>
        </div>
        <div class="components-payment-payment-provider-list__panel-3">
          <button
            type="button"
            @click="emit('refresh')"
            :disabled="loading"
            class="btn btn-secondary btn-sm"
            :title="t('common.refresh')"
          >
            <Icon name="refresh" size="sm" :class="loading ? 'components-payment-payment-provider-list__icon-2' : ''" />
          </button>
          <button
            type="button"
            @click="emit('create')"
            :disabled="!canCreate"
            :class="canCreate
              ? 'btn btn-primary btn-sm'
              : 'components-payment-payment-provider-list__action-2 btn btn-secondary btn-sm'"
          >
            {{ t('admin.settings.payment.createProvider') }}
          </button>
        </div>
      </div>
    </div>

    <!-- List -->
    <div class="components-payment-payment-provider-list__panel-4">
      <!-- Loading -->
      <div v-if="loading && !providers.length" class="components-payment-payment-provider-list__panel-5">
        <div class="components-payment-payment-provider-list__panel-6" />
      </div>

      <!-- Provider cards (draggable) -->
      <VueDraggable
        v-if="providers.length"
        v-model="localProviders"
        :animation="200"
        handle=".drag-handle"
        class="components-payment-payment-provider-list__vue-draggable"
        @end="onDragEnd"
      >
        <div v-for="p in localProviders" :key="p.id" class="components-payment-payment-provider-list__panel-7">
          <div class="components-payment-payment-provider-list__panel-8 drag-handle">
            <svg class="components-payment-payment-provider-list__icon" viewBox="0 0 20 20" fill="currentColor">
              <path d="M7 2a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM13 2a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM7 8a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM13 8a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM7 14a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM13 14a2 2 0 1 0 0 4 2 2 0 0 0 0-4z"/>
            </svg>
          </div>
          <div class="components-payment-payment-provider-list__panel-9">
            <ProviderCard
              :provider="p"
              :enabled="isEnabled(p.provider_key)"
              :available-types="getTypes(p.provider_key)"
              @toggle-field="(field) => emit('toggleField', p, field)"
              @toggle-type="(type) => emit('toggleType', p, type)"
              @edit="emit('edit', p)"
              @delete="emit('delete', p)"
            />
          </div>
        </div>
      </VueDraggable>

      <!-- Empty -->
      <div v-else-if="!loading" class="components-payment-payment-provider-list__panel-10">
        <p class="components-payment-payment-provider-list__description-2">
          {{ canCreate
            ? t('admin.settings.payment.noProviders')
            : t('admin.settings.payment.enableTypesFirst') }}
        </p>
        <button
          type="button"
          v-if="canCreate"
          @click="emit('create')"
          class="components-payment-payment-provider-list__action btn btn-primary btn-sm"
        >
          {{ t('admin.settings.payment.createProvider') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { VueDraggable } from 'vue-draggable-plus'
import Icon from '@/components/icons/Icon.vue'
import ProviderCard from './ProviderCard.vue'
import type { ProviderInstance } from '@/types/payment'
import type { TypeOption } from './providerConfig'
import { getAvailableTypes } from './providerConfig'

const props = defineProps<{
  providers: ProviderInstance[]
  loading: boolean
  canCreate: boolean
  enabledPaymentTypes: string[]
  allPaymentTypes: TypeOption[]
  redirectLabel: string
}>()

const emit = defineEmits<{
  refresh: []
  create: []
  edit: [provider: ProviderInstance]
  delete: [provider: ProviderInstance]
  toggleField: [provider: ProviderInstance, field: 'enabled' | 'refund_enabled' | 'allow_user_refund']
  toggleType: [provider: ProviderInstance, type: string]
  reorder: [providers: { id: string; sort_order: number }[]]
}>()

const { t } = useI18n()

const localProviders = ref<ProviderInstance[]>([])

watch(() => props.providers, (val) => {
  localProviders.value = [...val]
}, { immediate: true })

function onDragEnd() {
  const updates = localProviders.value.map((p, idx) => ({
    id: p.id,
    sort_order: idx,
  }))
  emit('reorder', updates)
}

function isEnabled(providerKey: string): boolean {
  return props.enabledPaymentTypes.includes(providerKey)
}

function getTypes(providerKey: string): TypeOption[] {
  return getAvailableTypes(providerKey, props.allPaymentTypes, props.redirectLabel)
    .map(opt => opt.label === opt.value
      ? { ...opt, label: t(`payment.methods.${opt.value}`, opt.value) }
      : opt,
    )
}
</script>

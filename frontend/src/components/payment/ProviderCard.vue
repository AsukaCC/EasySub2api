<template>
  <div
    :class="[
      'components-payment-provider-card__panel-5',
      enabled ? 'components-payment-provider-card__panel-6' : 'components-payment-provider-card__panel-7',
    ]"
    :title="!enabled ? t('admin.settings.payment.typeDisabled') + ' — ' + t('admin.settings.payment.enableTypesFirst') : undefined"
  >
    <div :class="[
      'components-payment-provider-card__panel-8',
      !enabled && 'components-payment-provider-card__panel-9',
    ]">
      <!-- Left: icon + name + key badge + type badges -->
      <div class="components-payment-provider-card__panel">
        <div :class="[
          'components-payment-provider-card__panel-10',
          provider.enabled && enabled ? 'components-payment-provider-card__panel-11' : 'components-payment-provider-card__panel-12',
        ]">
          <Icon
            name="server"
            size="sm"
            :class="provider.enabled && enabled ? 'components-payment-provider-card__icon' : 'components-payment-provider-card__icon-2'"
          />
        </div>
        <span class="components-payment-provider-card__text">{{ provider.name }}</span>
        <span class="components-payment-provider-card__text-2">{{ keyLabel }}</span>
        <span v-if="provider.payment_mode" class="components-payment-provider-card__text-2">· {{ modeLabel }}</span>
        <span v-if="enabled && availableTypes.length" class="components-payment-provider-card__text-3">|</span>
        <div v-if="enabled" class="components-payment-provider-card__panel-2">
          <button
            v-for="pt in availableTypes"
            :key="pt.value"
            type="button"
            @click="emit('toggleType', pt.value)"
            :class="[
              'components-payment-provider-card__action-3',
              isSelected(pt.value)
                ? 'components-payment-provider-card__action-4'
                : 'components-payment-provider-card__action-5',
            ]"
          >{{ pt.label }}</button>
        </div>
      </div>

      <!-- Right: toggles + actions -->
      <div class="components-payment-provider-card__panel-3">
        <ToggleSwitch :label="t('common.enabled')" :checked="provider.enabled" @toggle="emit('toggleField', 'enabled')" />
        <ToggleSwitch :label="t('admin.settings.payment.refundEnabled')" :checked="provider.refund_enabled" @toggle="emit('toggleField', 'refund_enabled')" />
        <ToggleSwitch v-if="provider.refund_enabled" :label="t('admin.settings.payment.allowUserRefund')" :checked="provider.allow_user_refund" @toggle="emit('toggleField', 'allow_user_refund')" />
        <div class="components-payment-provider-card__panel-4">
          <button type="button" @click="emit('edit')" class="components-payment-provider-card__action">
            <Icon name="edit" size="sm" />
            <span class="components-payment-provider-card__text-4">{{ t('common.edit') }}</span>
          </button>
          <button type="button" @click="emit('delete')" class="components-payment-provider-card__action-2">
            <Icon name="trash" size="sm" />
            <span class="components-payment-provider-card__text-4">{{ t('common.delete') }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import ToggleSwitch from './ToggleSwitch.vue'
import type { ProviderInstance } from '@/types/payment'
import type { TypeOption } from './providerConfig'
import { PAYMENT_MODE_QRCODE, PAYMENT_MODE_POPUP, PAYMENT_MODE_REDIRECT } from './providerConfig'

const PROVIDER_KEY_LABELS: Record<string, string> = {
  easypay: 'admin.settings.payment.providerEasypay',
  alipay: 'admin.settings.payment.providerAlipay',
  wxpay: 'admin.settings.payment.providerWxpay',
  stripe: 'admin.settings.payment.providerStripe',
  airwallex: 'admin.settings.payment.providerAirwallex',
}

const props = defineProps<{
  provider: ProviderInstance
  enabled: boolean
  availableTypes: TypeOption[]
}>()

const emit = defineEmits<{
  toggleField: [field: 'enabled' | 'refund_enabled' | 'allow_user_refund']
  toggleType: [type: string]
  edit: []
  delete: []
}>()

const { t } = useI18n()

const keyLabel = computed(() => t(PROVIDER_KEY_LABELS[props.provider.provider_key] || props.provider.provider_key))

const modeLabel = computed(() => {
  if (props.provider.payment_mode === PAYMENT_MODE_QRCODE) return t('admin.settings.payment.modeQRCode')
  if (props.provider.payment_mode === PAYMENT_MODE_POPUP) return t('admin.settings.payment.modePopup')
  if (props.provider.payment_mode === PAYMENT_MODE_REDIRECT) return t('admin.settings.payment.modeRedirect')
  return ''
})

function isSelected(type: string): boolean {
  return Array.isArray(props.provider.supported_types) && props.provider.supported_types.includes(type)
}
</script>
